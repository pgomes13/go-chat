#!/usr/bin/env bash
# Builds the Go binary locally to catch compile errors before deploying.

build_app() {
  info "Building application..."
  go build ./... || error "Build failed — fix errors before deploying"
  info "Build succeeded"
}
