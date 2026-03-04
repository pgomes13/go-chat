#!/usr/bin/env bash
# Grants required IAM roles to the Compute Engine default service account.

grant_iam_roles() {
  local project_number
  project_number=$(gcloud projects describe "$PROJECT_ID" --format="value(projectNumber)")
  local compute_sa="${project_number}-compute@developer.gserviceaccount.com"

  info "Granting IAM roles to $compute_sa..."
  for role in \
    roles/cloudbuild.builds.builder \
    roles/storage.admin \
    roles/artifactregistry.writer \
    roles/secretmanager.secretAccessor; do
    gcloud projects add-iam-policy-binding "$PROJECT_ID" \
      --member="serviceAccount:$compute_sa" \
      --role="$role" \
      --quiet 2>/dev/null | grep -q "etag" && true
  done
}
