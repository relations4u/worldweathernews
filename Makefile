.PHONY: help bootstrap env dev dev-full dev-monitoring dev-down dev-reset dev-logs dev-psql dev-redis test lint fmt build gen gen-check migrate clean release backend-dev backend-test backend-lint frontend-dev frontend-test frontend-lint frontend-check pyworkers-dev pyworkers-test pyworkers-lint pyworkers-typecheck

.DEFAULT_GOAL := help

help: ## Zeige diese Hilfe
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

bootstrap: env ## Erst-Setup nach Repo-Clone
	@echo "==> Installing tools via mise..."
	mise install
	@echo "==> Installing pre-commit hooks..."
	pre-commit install --install-hooks
	@echo "==> Bootstrap done. Run 'make dev' to start the local stack."

env: ## .env aus .env.example anlegen, falls fehlt
	@if [ ! -f .env ]; then \
		cp .env.example .env; \
		echo "  .env aus .env.example erstellt — bei Bedarf anpassen."; \
	fi

dev: env ## Lokalen Stack starten (Compose)
	docker compose up -d
	docker compose logs -f --tail=20

dev-full: env ## Lokaler Stack inkl. Monitoring (Prometheus, Grafana, Loki, Promtail, Tempo)
	docker compose --profile monitoring up -d
	docker compose --profile monitoring logs -f --tail=20

dev-monitoring: env ## Nur Monitoring-Services starten (Apps müssen separat laufen)
	docker compose --profile monitoring up -d prometheus grafana loki promtail tempo

dev-down: ## Stack stoppen
	docker compose --profile monitoring down

dev-reset: ## Stack stoppen + Volumes löschen
	docker compose --profile monitoring down -v

dev-logs: ## Logs eines Services tailen (SERVICE=name)
	docker compose logs -f $(SERVICE)

dev-psql: ## psql-Shell auf der DB
	docker compose exec postgres psql -U wwn -d wwn

dev-redis: ## redis-cli auf dem Cache
	docker compose exec redis redis-cli

test: ## Alle Tests ausführen
	$(MAKE) -C apps/backend test
	$(MAKE) -C apps/frontend test
	$(MAKE) -C apps/pyworkers test

backend-dev: ## Backend lokal mit Hot-Reload (air)
	$(MAKE) -C apps/backend dev

backend-test: ## Backend-Tests
	$(MAKE) -C apps/backend test

backend-lint: ## Backend-Lint (golangci-lint)
	$(MAKE) -C apps/backend lint

frontend-dev: ## Frontend lokal mit Hot-Reload (vite)
	$(MAKE) -C apps/frontend dev

frontend-test: ## Frontend-Tests (vitest)
	$(MAKE) -C apps/frontend test

frontend-lint: ## Frontend-Lint (eslint + prettier --check)
	$(MAKE) -C apps/frontend lint

frontend-check: ## Frontend Type-Check (svelte-check)
	$(MAKE) -C apps/frontend check

pyworkers-dev: ## Python-Workers lokal mit Hot-Reload (watchfiles)
	$(MAKE) -C apps/pyworkers dev

pyworkers-test: ## Python-Workers Tests (pytest)
	$(MAKE) -C apps/pyworkers test

pyworkers-lint: ## Python-Workers Lint (ruff check + format --check)
	$(MAKE) -C apps/pyworkers lint

pyworkers-typecheck: ## Python-Workers Type-Check (mypy strict)
	$(MAKE) -C apps/pyworkers typecheck

lint: ## Alle Linter (via pre-commit)
	pre-commit run --all-files

fmt: ## Auto-Format (gofmt, ruff, prettier wo verfügbar)
	@command -v gofmt >/dev/null && find apps/backend -name '*.go' 2>/dev/null | xargs -r gofmt -w || true
	@command -v ruff >/dev/null && ruff format apps/pyworkers 2>/dev/null || true
	@command -v prettier >/dev/null && prettier --write 'apps/frontend/**/*.{js,ts,svelte,html,css,json}' 2>/dev/null || true

build: ## Container bauen
	@echo "Wird ab Session 4 sinnvoll. Aktuell nichts zu bauen."

gen: ## Generierten Code aktualisieren (OpenAPI + sqlc)
	$(MAKE) -C packages/api-schema gen
	$(MAKE) sqlc-schema
	$(MAKE) sqlc-generate

sqlc-schema: ## schema.sql für sqlc aus den goose-Migrations generieren
	python3 scripts/build-sqlc-schema.py

sqlc-generate: ## sqlc-Code generieren (apps/backend/internal/storage/db/)
	cd apps/backend && sqlc generate

gen-check: ## Prüft, ob generierter Code aktuell ist (für CI)
	$(MAKE) gen
	@git diff --exit-code -- \
		apps/backend/internal/api/api.gen.go \
		apps/frontend/src/lib/api/types.gen.ts \
		apps/backend/internal/storage/schema.sql \
		apps/backend/internal/storage/db/ \
		|| (echo "Generated code is out of date. Run 'make gen' and commit." && exit 1)

migrate: ## DB-Migrations anwenden (alle pending)
	bash scripts/migrate.sh up

migrate-status: ## DB-Migrations-Status anzeigen
	bash scripts/migrate.sh status

migrate-down: ## Letzte DB-Migration zurückrollen
	bash scripts/migrate.sh down

migrate-reset: ## Alle DB-Migrations zurückrollen
	bash scripts/migrate.sh reset

release: ## Neuen Release-Tag erstellen (interaktiv) und pushen
	bash scripts/release.sh

clean: ## Aufräumen
	rm -rf bin tmp dist build .turbo
	@find . -name '__pycache__' -type d -exec rm -rf {} + 2>/dev/null || true
	@find . -name '*.pyc' -delete 2>/dev/null || true
