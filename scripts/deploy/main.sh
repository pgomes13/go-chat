#!/usr/bin/env bash
set -euo pipefail

# Resolve the repo root regardless of where the script is called from.
REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Source sub-scripts
# shellcheck source=scripts/deploy/helpers.sh
source "$SCRIPT_DIR/helpers.sh"
# shellcheck source=scripts/deploy/apis.sh
source "$SCRIPT_DIR/apis.sh"
# shellcheck source=scripts/deploy/iam.sh
source "$SCRIPT_DIR/iam.sh"
# shellcheck source=scripts/deploy/secrets.sh
source "$SCRIPT_DIR/secrets.sh"
# shellcheck source=scripts/deploy/build.sh
source "$SCRIPT_DIR/build.sh"
# shellcheck source=scripts/deploy/run.sh
source "$SCRIPT_DIR/run.sh"

# ── Config ────────────────────────────────────────────────────────────────────
SERVICE="go-chat"
REGION="${REGION:-australia-southeast1}"
PROJECT_ID=$(gcloud config get-value project 2>/dev/null)

# ── Preflight ─────────────────────────────────────────────────────────────────
[[ -z "$PROJECT_ID" ]] && error "No GCP project set. Run: gcloud config set project PROJECT_ID"
[[ ! -f "$REPO_ROOT/.env" ]] && error ".env file not found at $REPO_ROOT/.env"

info "Project: $PROJECT_ID  |  Region: $REGION  |  Service: $SERVICE"

# Load .env — bash source handles # comments natively
set -o allexport
# shellcheck source=.env
source "$REPO_ROOT/.env"
set +o allexport

[[ -z "${GOOGLE_CLIENT_ID:-}"     ]] && error "GOOGLE_CLIENT_ID not set in .env"
[[ -z "${GOOGLE_CLIENT_SECRET:-}" ]] && error "GOOGLE_CLIENT_SECRET not set in .env"
[[ -z "${MONGO_URI:-}"            ]] && error "MONGO_URI not set in .env"

# Change to repo root so --source . works correctly
cd "$REPO_ROOT"

# ── Run steps ─────────────────────────────────────────────────────────────────
build_app
enable_apis
grant_iam_roles
sync_secrets
deploy_service
