#!/usr/bin/env bash
# Deploys the service to Cloud Run and resolves the OAuth redirect URL.

deploy_service() {
  # Use OAUTH_REDIRECT_URL from .env if already set (all deploys after the first).
  # On first deploy, deploy once to discover the URL, save it to .env, then
  # redeploy with it — the URL never changes after that.
  if [[ -n "${OAUTH_REDIRECT_URL:-}" ]]; then
    SERVICE_URL="${OAUTH_REDIRECT_URL%/auth/google/callback}"
    REDIRECT_URL="$OAUTH_REDIRECT_URL"
    info "Using saved service URL: $SERVICE_URL"
  else
    info "First deploy — discovering service URL..."
    _gcloud_deploy  # deploy without redirect URL to get the URL assigned

    SERVICE_URL=$(gcloud run services describe "$SERVICE" \
      --region "$REGION" --project "$PROJECT_ID" \
      --format "value(status.url)")
    REDIRECT_URL="${SERVICE_URL}/auth/google/callback"

    # Persist so future deploys are single-step.
    printf '\n# Cloud Run\nOAUTH_REDIRECT_URL=%s\n' "$REDIRECT_URL" >> .env
    info "Saved OAUTH_REDIRECT_URL to .env"
  fi

  info "Deploying to ${SERVICE_URL}..."
  _gcloud_deploy --set-env-vars "OAUTH_REDIRECT_URL=${REDIRECT_URL}"

  echo ""
  echo "✔ Deployed: ${SERVICE_URL}"
  echo ""
  echo "  Ensure this is in Google OAuth authorised redirect URIs:"
  echo "  ${REDIRECT_URL}"
}

# Internal: runs gcloud run deploy with common flags; accepts extra args.
_gcloud_deploy() {
  gcloud run deploy "$SERVICE" \
    --source . \
    --region "$REGION" \
    --project "$PROJECT_ID" \
    --allow-unauthenticated \
    --set-env-vars "GOOGLE_CLIENT_ID=${GOOGLE_CLIENT_ID}" \
    --set-secrets "GOOGLE_CLIENT_SECRET=go-chat-google-secret:latest" \
    --set-secrets "MONGO_URI=go-chat-mongo-uri:latest" \
    --set-secrets "SESSION_SECRET=go-chat-session-secret:latest" \
    --quiet \
    "$@"
}
