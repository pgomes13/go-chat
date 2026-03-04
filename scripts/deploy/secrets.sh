#!/usr/bin/env bash
# Syncs .env secrets to GCP Secret Manager.

sync_secrets() {
  info "Syncing secrets to Secret Manager..."
  upsert_secret "go-chat-google-secret" "$GOOGLE_CLIENT_SECRET"
  upsert_secret "go-chat-mongo-uri"     "$MONGO_URI"

  SESSION_SECRET="${SESSION_SECRET:-$(openssl rand -base64 32)}"
  upsert_secret "go-chat-session-secret" "$SESSION_SECRET"
}
