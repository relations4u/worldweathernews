# Session 2 — Pre-commit, Makefile, lokale Workflows

**Phase**: A (Fundament)
**Geschätzte Dauer**: 1 Stunde
**Vorbedingung**: Session 1 abgeschlossen, `mise install` durchgelaufen.

## Ziel

Am Ende dieser Session funktionieren `make bootstrap`, `make lint`, und Pre-commit-Hooks
greifen bei jedem Commit. Auto-Format mit `make fmt` läuft.

`make test` und `make build` existieren als Targets, schlagen aber graceful fehl
oder zeigen "no services yet" — Service-Dateien kommen erst ab Session 4.

## Aufgaben

### 1. `.pre-commit-config.yaml`

Hooks für:

**Generic** (jeder Commit):
- `check-merge-conflict`
- `check-added-large-files` (Limit z.B. 1MB)
- `end-of-file-fixer`
- `trailing-whitespace`
- `check-yaml`
- `check-json`
- `mixed-line-ending`

**Secrets-Scanning**:
- `gitleaks` (über pre-commit-Hook von gitleaks)

**Go**:
- `gofmt`
- `goimports`
- `golangci-lint` (system-Hook, nutzt lokales `golangci-lint` aus mise)

**Python**:
- `ruff check`
- `ruff format`

**JavaScript/TypeScript/Svelte**:
- `prettier` mit Plugins für Svelte und Tailwind
- (eslint kommt erst, wenn das Frontend existiert — Session 5)

**YAML**:
- `yamllint`

**Dockerfiles**:
- `hadolint` (greift später, wenn Dockerfiles existieren)

**Konfiguration**:
- Hooks die für leere Verzeichnisse fehlschlagen würden: konditional aktivieren
  oder mit `exclude`-Pattern entschärfen
- `--fail-fast` ausschalten, wir wollen alle Findings sehen

### 2. `.yamllint`-Config

Pragmatisch:
- `extends: default`
- Line-length disabled (oder auf 200 hochgesetzt)
- Indentation: 2
- Document-start (`---`): not required
- Truthy: nur `true`/`false`, nicht `yes`/`no`

### 3. `.golangci.yml`

Pragmatisches Setup:

```yaml
run:
  timeout: 5m
linters:
  enable:
    - errcheck
    - govet
    - staticcheck
    - revive
    - gosec
    - misspell
    - gofmt
    - goimports
    - ineffassign
    - unused
    - bodyclose
    - sqlclosecheck
    - errorlint
issues:
  exclude-rules:
    - path: _test\.go
      linters:
        - gosec
```

Begründung als Kommentar im File: Tests dürfen pragmatisch sein (z.B. nichtkonst.
Strings für SQL in Test-Setup).

### 4. `apps/pyworkers/pyproject.toml` (Stub)

Minimal:

```toml
[project]
name = "wwn-pyworkers"
version = "0.0.1"
description = "Worker services for worldweathernews.com"
requires-python = ">=3.12"
dependencies = []

[tool.ruff]
line-length = 100
target-version = "py312"

[tool.ruff.lint]
select = ["E", "F", "I", "N", "UP", "B", "SIM", "RUF"]

[tool.ruff.format]
quote-style = "double"

[tool.mypy]
python_version = "3.12"
strict = true

[tool.pytest.ini_options]
asyncio_mode = "auto"
```

Echte Dependencies kommen in Session 6.

### 5. Top-Level `Makefile`

Targets mit `##`-Doc-Kommentaren für `make help`:

```makefile
.PHONY: help bootstrap dev dev-full dev-down dev-reset test lint fmt build gen migrate clean

.DEFAULT_GOAL := help

help: ## Zeige diese Hilfe
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

bootstrap: ## Erst-Setup nach Repo-Clone
	@echo "==> Installing tools via mise..."
	mise install
	@echo "==> Installing pre-commit hooks..."
	pre-commit install --install-hooks
	@echo "==> Bootstrap done. Run 'make dev' to start."

dev: ## Lokalen Stack starten (Compose)
	@if [ ! -f compose.yml ] && [ ! -f infra/compose/compose.dev.yml ]; then \
		echo "Compose-File fehlt. Wird in Session 3 erstellt."; exit 1; \
	fi
	docker compose up -d
	docker compose logs -f --tail=20

dev-down: ## Stack stoppen
	docker compose down

dev-reset: ## Stack stoppen + Volumes löschen
	docker compose down -v

test: ## Alle Tests ausführen
	@$(MAKE) -C apps/backend test 2>/dev/null || echo "  backend: noch nicht da"
	@$(MAKE) -C apps/frontend test 2>/dev/null || echo "  frontend: noch nicht da"
	@$(MAKE) -C apps/pyworkers test 2>/dev/null || echo "  pyworkers: noch nicht da"

lint: ## Alle Linter
	pre-commit run --all-files

fmt: ## Auto-Format
	@command -v gofmt >/dev/null && find apps/backend -name '*.go' 2>/dev/null | xargs -r gofmt -w || true
	@command -v ruff >/dev/null && ruff format apps/pyworkers 2>/dev/null || true
	@command -v prettier >/dev/null && prettier --write 'apps/frontend/**/*.{js,ts,svelte,html,css,json}' 2>/dev/null || true

build: ## Container bauen
	@echo "Wird ab Session 4 sinnvoll. Aktuell nichts zu bauen."

gen: ## Generierten Code aktualisieren
	@echo "Wird in Session 7 implementiert."

migrate: ## DB-Migrations anwenden
	@echo "Wird in Session 4/9 implementiert."

clean: ## Aufräumen
	rm -rf bin tmp dist build .turbo
	find . -name '__pycache__' -type d -exec rm -rf {} + 2>/dev/null || true
	find . -name '*.pyc' -delete 2>/dev/null || true
```

Wichtig:
- `.PHONY` korrekt für alle Targets
- Tabs für Indentation (Make-Anforderung)
- Graceful: Targets schlagen nicht hart fehl wenn Sub-Verzeichnisse noch leer sind

### 6. Service-Stub-Makefiles (vorbereitend)

`apps/backend/Makefile`, `apps/frontend/Makefile`, `apps/pyworkers/Makefile`:

Jeweils Minimum mit `test`-, `lint`-, `dev`-Target, aber Inhalt ist nur:

```makefile
.PHONY: test lint dev
test:
	@echo "backend tests: not yet implemented (Session 4)"
lint:
	@echo "backend lint: not yet implemented (Session 4)"
dev:
	@echo "backend dev: not yet implemented (Session 4)"
```

(analog für die anderen, mit jeweiliger Session-Referenz)

## Vorgehen (verbindlich)

1. Plan zeigen
2. Freigabe abwarten
3. Implementieren
4. **Pre-commit lokal einmalig laufen lassen**: `pre-commit run --all-files`
   - Erwartet: einige minor Findings auf bestehenden Files, kommit-able mit `--no-verify` falls nötig
5. **Make-Targets prüfen**:
   - `make help` zeigt Liste
   - `make lint` läuft
   - `make fmt` läuft
   - `make test` läuft graceful durch
6. Status-Output zeigen
7. **Nicht committen.** Nach meinem Review committe ich.

## Erfolgs-Kriterien

- [ ] `pre-commit run --all-files` läuft ohne Crash (Findings ok, Crash nicht)
- [ ] `make help` listet alle Targets mit Beschreibung
- [ ] `make lint` startet Pre-commit
- [ ] `make fmt` formatiert idempotent
- [ ] `make test` schlägt nicht hart fehl
- [ ] Pre-commit-Hook ist installiert (`.git/hooks/pre-commit` existiert)

## Mögliche Stolpersteine

- **gitleaks Installation**: kann via pre-commit-CI-runs schwierig sein. Alternativen: `detect-secrets` (Yelp).
- **prettier ohne package.json**: Top-Level erstmal ohne. Wenn prettier-Hook ohne node_modules nicht funktioniert: prettier-Hook erst nach Frontend-Setup aktivieren, klar dokumentieren.
- **golangci-lint ohne Go-Code**: Hook kann fehlschlagen wenn keine Go-Files da sind. Mit `files:` Pattern auf `.go`-Dateien beschränken.
- **hadolint ohne Dockerfiles**: dito, mit `files:` Pattern auf `Dockerfile*` beschränken.

## Suggested Commit-Message

```
chore: add pre-commit hooks and top-level Makefile
```
