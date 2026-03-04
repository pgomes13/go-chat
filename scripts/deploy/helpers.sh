#!/usr/bin/env bash
# Shared utility functions sourced by other deploy scripts.

info()  { echo "▶ $*"; }
error() { echo "✖ $*" >&2; exit 1; }

secret_exists() {
  gcloud secrets describe "$1" --project "$PROJECT_ID" &>/dev/null
}

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
