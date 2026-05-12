#!/usr/bin/env bash
# Prüft, ob die generierten Files committed und aktuell sind:
#   - aus packages/api-schema/openapi.yaml: api.gen.go + types.gen.ts
#     via oapi-codegen + openapi-typescript
#   - aus infra/migrations/: kuratierte schema.sql + sqlc-Output unter
#     apps/backend/internal/storage/db/ via build-sqlc-schema.py + sqlc
# Wird in CI nach `make gen` aufgerufen.

set -euo pipefail

ROOT="$(git rev-parse --show-toplevel)"
cd "$ROOT"

echo "==> Regenerating OpenAPI + sqlc artifacts..."
make gen

echo "==> Checking for diff..."
if ! git diff --exit-code -- \
    apps/backend/internal/api/api.gen.go \
    apps/frontend/src/lib/api/types.gen.ts \
    apps/backend/internal/storage/schema.sql \
    apps/backend/internal/storage/db/; then
    echo
    echo "✗ Generated files are out of date."
    echo "  Run 'make gen' locally and commit the changes."
    exit 1
fi

echo "✓ Generated files are up to date."
