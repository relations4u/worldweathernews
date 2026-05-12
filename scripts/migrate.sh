#!/usr/bin/env bash
# goose-Wrapper für lokale Entwicklung gegen den dev-Compose-Stack.
#
# Liest POSTGRES_PASSWORD aus .env, baut die lokale DSN (127.0.0.1:5432,
# nicht der Docker-interne 'postgres'-Hostname) und reicht alle Argumente
# an `goose` weiter.
#
# Usage:
#   bash scripts/migrate.sh up                # alle pending Migrations
#   bash scripts/migrate.sh up-by-one         # eine Migration anwenden
#   bash scripts/migrate.sh down              # letzte Migration zurück
#   bash scripts/migrate.sh status            # Status anzeigen
#   bash scripts/migrate.sh reset             # alle Migrations zurück
#   bash scripts/migrate.sh create <name> sql # neue Migration anlegen
#
# Voraussetzung: dev-Stack läuft (`make dev`), .env existiert.

set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

if [ ! -f .env ]; then
	echo "Error: .env not found. Run 'make bootstrap' or copy .env.example to .env." >&2
	exit 1
fi

# POSTGRES_PASSWORD aus .env extrahieren, ohne die ganze .env zu sourcen
# (vermeidet, dass andere Variablen in die Shell-Umgebung leaken).
POSTGRES_PASSWORD="$(grep -E '^POSTGRES_PASSWORD=' .env | head -1 | cut -d= -f2-)"

if [ -z "$POSTGRES_PASSWORD" ]; then
	echo "Error: POSTGRES_PASSWORD not set in .env." >&2
	exit 1
fi

DB_URL="postgres://wwn:${POSTGRES_PASSWORD}@127.0.0.1:5432/wwn?sslmode=disable"

exec goose -dir infra/migrations postgres "$DB_URL" "$@"
