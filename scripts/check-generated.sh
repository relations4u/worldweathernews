#!/usr/bin/env bash
# Prüft ob die aus packages/api-schema/openapi.yaml generierten Files
# (Go-Server-Stubs + TypeScript-Types) committed und aktuell sind.
# Wird in CI nach `make gen` aufgerufen.

set -euo pipefail

ROOT="$(git rev-parse --show-toplevel)"
cd "$ROOT"

echo "==> Regenerating OpenAPI artifacts..."
make gen

echo "==> Checking for diff..."
if ! git diff --exit-code -- \
    apps/backend/internal/api/api.gen.go \
    apps/frontend/src/lib/api/types.gen.ts; then
    echo
    echo "✗ Generated files are out of date."
    echo "  Run 'make gen' locally and commit the changes."
    exit 1
fi

echo "✓ Generated files are up to date."
