#!/usr/bin/env bash
# Enables required GCP APIs for the project.

enable_apis() {
  info "Enabling required APIs..."
  gcloud services enable \
    run.googleapis.com \
    cloudbuild.googleapis.com \
    artifactregistry.googleapis.com \
    secretmanager.googleapis.com \
    --project "$PROJECT_ID" --quiet
}
