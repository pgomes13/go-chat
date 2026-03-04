#!/usr/bin/env bash
set -euo pipefail

# ── Config ────────────────────────────────────────────────────────────────────
SERVICE="go-chat"
REGION="${REGION:-australia-southeast1}"
PROJECT_ID=$(gcloud config get-value project 2>/dev/null)

# ── Helpers ───────────────────────────────────────────────────────────────────
info()  { echo "▶ $*"; }
error() { echo "✖ $*" >&2; exit 1; }

secret_exists() { gcloud secrets describe "$1" --project "$PROJECT_ID" &>/dev/null; }

upsert_secret() {
  local name="$1" value="$2"
  if secret_exists "$name"; then
    info "Updating secret: $name"
    echo -n "$value" | gcloud secrets versions add "$name" --data-file=- --project "$PROJECT_ID" --quiet
  else
    info "Creating secret: $name"
    echo -n "$value" | gcloud secrets create "$name" --data-file=- --project "$PROJECT_ID" --quiet
  fi
}

# ── Preflight ─────────────────────────────────────────────────────────────────
[[ -z "$PROJECT_ID" ]] && error "No GCP project set. Run: gcloud config set project PROJECT_ID"
[[ ! -f ".env" ]]      && error ".env file not found"

info "Project: $PROJECT_ID  |  Region: $REGION  |  Service: $SERVICE"

# Load .env — bash source handles # comments natively
set -o allexport
# shellcheck disable=SC1091
source .env
set +o allexport

[[ -z "${GOOGLE_CLIENT_ID:-}"     ]] && error "GOOGLE_CLIENT_ID not set in .env"
[[ -z "${GOOGLE_CLIENT_SECRET:-}" ]] && error "GOOGLE_CLIENT_SECRET not set in .env"
[[ -z "${MONGO_URI:-}"            ]] && error "MONGO_URI not set in .env"

# ── APIs ──────────────────────────────────────────────────────────────────────
info "Enabling required APIs..."
gcloud services enable \
  run.googleapis.com \
  cloudbuild.googleapis.com \
  artifactregistry.googleapis.com \
  secretmanager.googleapis.com \
  --project "$PROJECT_ID" --quiet

# ── IAM ───────────────────────────────────────────────────────────────────────
PROJECT_NUMBER=$(gcloud projects describe "$PROJECT_ID" --format="value(projectNumber)")
COMPUTE_SA="${PROJECT_NUMBER}-compute@developer.gserviceaccount.com"

info "Granting IAM roles to $COMPUTE_SA..."
for role in \
  roles/cloudbuild.builds.builder \
  roles/storage.admin \
  roles/artifactregistry.writer \
  roles/secretmanager.secretAccessor; do
  gcloud projects add-iam-policy-binding "$PROJECT_ID" \
    --member="serviceAccount:$COMPUTE_SA" \
    --role="$role" \
    --quiet 2>/dev/null | grep -q "etag" && true
done

# ── Secrets ───────────────────────────────────────────────────────────────────
info "Syncing secrets to Secret Manager..."
upsert_secret "go-chat-google-secret"  "$GOOGLE_CLIENT_SECRET"
upsert_secret "go-chat-mongo-uri"      "$MONGO_URI"

SESSION_SECRET="${SESSION_SECRET:-$(openssl rand -base64 32)}"
upsert_secret "go-chat-session-secret" "$SESSION_SECRET"

# ── Resolve service URL ───────────────────────────────────────────────────────
# Use OAUTH_REDIRECT_URL from .env if already set (all deploys after the first).
# On first deploy we don't know the URL yet, so deploy once to get it, save it
# to .env, then redeploy with it — the URL never changes after that.

if [[ -n "${OAUTH_REDIRECT_URL:-}" ]]; then
  SERVICE_URL="${OAUTH_REDIRECT_URL%/auth/google/callback}"
  REDIRECT_URL="$OAUTH_REDIRECT_URL"
  info "Using saved service URL: $SERVICE_URL"
else
  info "First deploy — discovering service URL..."
  gcloud run deploy "$SERVICE" \
    --source . \
    --region "$REGION" \
    --project "$PROJECT_ID" \
    --allow-unauthenticated \
    --set-env-vars "GOOGLE_CLIENT_ID=${GOOGLE_CLIENT_ID}" \
    --set-secrets "GOOGLE_CLIENT_SECRET=go-chat-google-secret:latest" \
    --set-secrets "MONGO_URI=go-chat-mongo-uri:latest" \
    --set-secrets "SESSION_SECRET=go-chat-session-secret:latest" \
    --quiet

  SERVICE_URL=$(gcloud run services describe "$SERVICE" \
    --region "$REGION" --project "$PROJECT_ID" \
    --format "value(status.url)")
  REDIRECT_URL="${SERVICE_URL}/auth/google/callback"

  # Persist so future deploys are single-step.
  echo "" >> .env
  echo "# Cloud Run" >> .env
  echo "OAUTH_REDIRECT_URL=${REDIRECT_URL}" >> .env
  info "Saved OAUTH_REDIRECT_URL to .env"
fi

# ── Deploy ────────────────────────────────────────────────────────────────────
info "Deploying to ${SERVICE_URL}..."
gcloud run deploy "$SERVICE" \
  --source . \
  --region "$REGION" \
  --project "$PROJECT_ID" \
  --allow-unauthenticated \
  --set-env-vars "GOOGLE_CLIENT_ID=${GOOGLE_CLIENT_ID}" \
  --set-env-vars "OAUTH_REDIRECT_URL=${REDIRECT_URL}" \
  --set-secrets "GOOGLE_CLIENT_SECRET=go-chat-google-secret:latest" \
  --set-secrets "MONGO_URI=go-chat-mongo-uri:latest" \
  --set-secrets "SESSION_SECRET=go-chat-session-secret:latest" \
  --quiet

echo ""
echo "✔ Deployed: ${SERVICE_URL}"
echo ""
echo "  Ensure this is in Google OAuth authorised redirect URIs:"
echo "  ${REDIRECT_URL}"
