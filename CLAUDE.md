# worldweathernews.com

Wetter- und Klima-Plattform mit Community-Features. Self-hosted Monorepo.

**Maintainer**: Heinz W. Richter <hwr@relations4u.de> — Diplom-Meteorologe,
IT-Architekt, betreibt eigene Hosting-Infrastruktur (Proxmox).

Diese Datei ist die zentrale Spielregel-Datei. Du (Claude Code) liest sie zu Beginn
jeder Session. Wenn du Lücken findest oder etwas widersprüchlich erscheint: fragen,
nicht raten.

---

## Projekt-Vision

Eine globale Plattform, die:

- Wetterdaten und Vorhersagen aus nationalen Wetterdiensten weltweit aggregiert
- Klimadaten visualisiert (Anomalien, Trends, historische Vergleiche)
- Eine Community von Beobachtern, Citizen Scientists und Wetter-Interessierten aufbaut
- Auf mobilen Geräten und Desktop gleichermaßen gut funktioniert
- Sich dynamisch an Nutzer-Region und -Präferenzen anpasst
- Zur Laufzeit konfigurierbar ist (kein Deployment für Inhalts-Änderungen)

---

## Entwicklungs-Setup (verbindlich)

Die Entwicklung erfolgt **nicht direkt auf macOS**, sondern in einer dedizierten
Linux-VM. Begründung: Plattform-Parität mit Production (Hetzner-Server),
saubere Versions-Reproduzierbarkeit, Snapshot-basierte Sicherheit.

| Komponente        | Wert                                   |
| ----------------- | -------------------------------------- |
| Dev-Host          | Mac (Editor via VS Code Remote-SSH)    |
| Dev-Compute       | Proxmox-VM `wwn-dev`                   |
| Dev-OS            | Ubuntu 24.04 LTS Server                |
| Dev-Architektur   | AMD64 (Ryzen 7 host)                   |
| Production-Twin   | Hetzner Cloud, Ubuntu 24.04 LTS, AMD64 |
| Toolchain-Manager | mise (siehe `.mise.toml`)              |

Setup-Anleitung: siehe `vm-setup.md` im Repo-Root.

---

## Tech-Stack (verbindlich)

| Schicht            | Technologie                         | Version                 | Begründung                                      |
| ------------------ | ----------------------------------- | ----------------------- | ----------------------------------------------- |
| Backend-API        | Go                                  | 1.25                    | pgx v5.9.2 erfordert Go ≥ 1.25                  |
| HTTP-Framework     | Chi                                 | v5                      | Idiomatisch, nah an net/http, kein Magic        |
| DB-Access          | sqlc + pgx/v5                       | sqlc 1.27.0, pgx 5.9.2+ | Typsicher ohne ORM-Overhead                     |
| Config (Go)        | Viper                               | 1.21+                   | ENV + Datei + Defaults                          |
| Lint (Go)          | golangci-lint                       | 2.12.1                  | v2-Schema, gebaut mit Go 1.26                   |
| Workers            | Python                              | 3.12                    | Reife GRIB/NetCDF-Bibliotheken (xarray, cfgrib) |
| Python-Pkg-Manager | uv                                  | 0.11+                   | Schneller als pip/poetry, deterministisch       |
| Config (Python)    | pydantic-settings                   | aktuell                 | Typsicher, Validation beim Start                |
| Lint (Python)      | ruff                                | aktuell                 | Lint + Format in einem                          |
| Frontend           | SvelteKit + TypeScript              | Node 22 / pnpm 9        | Schlanker als Next.js, exzellente DX            |
| Styling            | Tailwind CSS + shadcn-svelte        | aktuell                 | Schnelles, konsistentes UI                      |
| Frontend-Adapter   | @sveltejs/adapter-node              | aktuell                 | Self-Hosting im Container                       |
| Datenbank          | PostgreSQL                          | 16                      | Mit PostGIS (Geo) und TimescaleDB (Zeitreihen)  |
| Cache/Queue        | Redis                               | 7                       | Standard, vielseitig                            |
| Reverse Proxy      | Caddy                               | 2                       | Auto-SSL, einfache Config                       |
| Container-Runtime  | Docker + Docker Compose             | aktuell                 | Wachstumspfad zu K3s offen, aber nicht jetzt    |
| CI                 | GitHub Actions                      | —                       | Schnell, kostenlos, sehr verbreitet             |
| Container-Registry | ghcr.io                             | —                       | Integriert mit GitHub                           |
| Deployment         | Ansible                             | aktuell                 | Klassisch, durchschaubar, idempotent            |
| Infrastruktur      | Terraform                           | aktuell                 | Provider-Abstraktion, deklarativ                |
| Secrets            | SOPS + age                          | aktuell                 | In Git verschlüsselt, kein externer Vault nötig |
| Monitoring         | Prometheus + Grafana + Loki + Tempo | aktuell                 | Self-hostbar, ein UI                            |
| Errors             | Sentry / GlitchTip (TBD)            | aktuell                 | Quasi-Standard                                  |
| Uptime             | Uptime Kuma                         | aktuell                 | Simpel, schön, self-hosted                      |
| Migrations         | goose                               | 3.22.1                  | Sprach-agnostisch (Go-Tool, aber File-basiert)  |
| API-Schema         | OpenAPI                             | 3.1                     | Single Source of Truth, redocly als Linter      |

---

## Versions-Pinning — verbindlich

**Goldene Regel:** Vier Quellen müssen dieselbe Major.Minor zeigen — Drift
zwischen ihnen ist die häufigste CI-Failure-Ursache und hat während des
Setups massiv Zeit gekostet.

1. `apps/backend/go.mod` → `go`-Direktive
2. `.mise.toml` → `go = "X.Y"` und alle anderen Tool-Pins
3. `.github/workflows/ci-*.yml` → `go-version: 'X.Y'`, `version: 'X.Y.Z'` etc.
4. `apps/backend/Dockerfile` → `FROM golang:X.Y-alpine` (analog für andere)

### Aktuelle Pflicht-Pins (Stand 6. Mai 2026)

| Tool                 | Pin    | Quelle                                |
| -------------------- | ------ | ------------------------------------- |
| Go                   | 1.25   | go.mod, .mise.toml, alle CI-Workflows |
| golangci-lint        | 2.12.1 | .mise.toml, ci-backend.yml            |
| golangci-lint-action | @v8    | ci-backend.yml                        |
| Node                 | 22     | .mise.toml, ci-frontend.yml           |
| pnpm                 | 9      | .mise.toml, ci-frontend.yml           |
| Python               | 3.12   | .mise.toml, ci-pyworkers.yml          |
| sqlc                 | 1.27.0 | .mise.toml                            |
| goose                | 3.22.1 | .mise.toml                            |
| yamllint             | 1.35.1 | .mise.toml                            |
| pre-commit           | 4.0.1  | .mise.toml                            |
| Terraform            | 1.15   | .mise.toml, infra/terraform/          |
| sops                 | 3.12   | .mise.toml, GitHub-Release auf VMs    |
| ansible-core         | 2.20   | .mise.toml, infra/ansible/            |
| ansible-lint         | 26.4   | .mise.toml, infra/ansible/            |

### Verbote

- **NIEMALS `latest`** als Pin — weder in `.mise.toml`, go.mod, GitHub Actions
  noch in Dockerfiles
- **NIEMALS `stable`** in setup-go (zieht jeweils neueste Major)
- **KEINE Patch-Versions** wenn Major.Minor reicht (z. B. `"1.25"` statt `"1.25.9"`)
- **KEINE manuelle `toolchain`-Direktive** in go.mod (nur was `go mod tidy` setzt)
- **KEINE separaten Tool-Versionen in pre-commit** wo der Workspace eines hat
  (siehe Pre-commit-Strategie unten)

### Wie Versions-Updates ablaufen

1. Maintainer entscheidet bewusst, eine Version anzuheben
2. `.mise.toml` anpassen
3. `mise install`
4. Lokal testen: `make lint && make test`
5. Falls grün: gleiche Version in allen vier Quellen oben aktualisieren
6. Commit mit Begründung warum die Version angehoben wurde
7. CI muss grün durchlaufen, sonst Rollback

### Häufige Versions-Fallen

- Claude Code wählt im Initial-Setup oft "neueste verfügbare" Versionen.
  Bei jedem Skelett-Schritt: Versions-Konsistenz prüfen vor Commit.
- `go mod tidy` kann `toolchain`-Zeilen automatisch ergänzen, wenn Dependencies
  höhere Versionen verlangen. Akzeptieren, aber dann ALLE anderen Quellen mitziehen.
- pre-commit cached Hook-Versionen aggressiv. Nach `.pre-commit-config.yaml`-
  Änderungen: `pre-commit clean && pre-commit install --install-hooks`.
- Beispiel-Falle aus diesem Projekt: pgx v5.9.2 verlangt Go 1.25 — das zwang
  die ganze Toolchain auf Go 1.25 + golangci-lint v2.12.1.

---

## App-Release-Pinning — abgegrenzt vom Toolchain-Pinning

Der vorige Abschnitt regelt **Toolchain-Pins** (Go, Node, Python, golangci-lint, …) —
also Versionen, die in den Build-Container eingebrannt werden. Dieser Abschnitt
regelt **App-Release-Pins** — welche Service-Image-Tag-Version aktuell auf
wwn-prod produziert.

### Authoritative Quellen

- **Letztes signiertes Tag im Repo**: `git tag --sort=-v:refname | head -1`
- **Live auf wwn-prod**:
  `ssh hwr@10.100.100.21 'sudo -u deploy docker ps --format "{{.Image}}" | grep ghcr.io'`

Beide müssen synchron sein. Stand 6. Mai 2026: `v0.0.2` live (alle drei
Services).

### Pinning-Regeln

- **Keine `latest` als App-Release-Tag** — gilt analog zum Toolchain-Pin-Verbot.
- **Kein implizites Default in Production**: `infra/ansible/inventories/production/group_vars/all.yml`
  setzt `default_versions: 0.0.0` als bewussten Fail-fast-Marker. Ohne explizites
  `-e target_version=X.Y.Z` deployt man auf `0.0.0` → Image existiert nicht in
  ghcr.io → Container starten nicht. Das ist gewollt.
- **Eine Version für alle drei Services** — Release-Pipeline taggt alle drei
  zugleich, Deploy zieht alle drei zugleich. Mismatched Versions zwischen
  backend/frontend/pyworkers sind ein Bug.
- **Frontend-Build-Args sind Teil des Release-Pins** — `PUBLIC_API_BASE_URL`
  wird zur Build-Zeit ins JS-Bundle eingebrannt (siehe `apps/frontend/Dockerfile`
  und `.github/workflows/release.yml`). Wer das Frontend für eine andere
  Umgebung baut, MUSS den build-arg setzen, sonst landet der Dockerfile-Default
  `http://api.localhost` im Bundle.

### Release-Workflow

1. `make release` (oder direkt `git tag -s vX.Y.Z -m "Release vX.Y.Z" && git push origin vX.Y.Z`).
2. Release-Pipeline (`.github/workflows/release.yml`) baut, signiert
   (cosign keyless via Sigstore), erstellt SBOMs (Syft), scannt (Trivy)
   und published 3 Images zu ghcr.io.
3. CI grün abwarten — Erfahrungswert: ~3 Min, gelegentlich Syft-CDN-502
   beim SBOM-Step → `gh run rerun <id> --failed` reicht.
4. `bash scripts/deploy.sh production X.Y.Z` (interaktive Bestätigung).
5. Smoke-Tests aus `docs/runbook.md` abklappern.

Direkter `docker pull` und `docker run` außerhalb dieses Workflows ist
nicht der zugelassene Weg in Produktion. cosign-Verify zur Pull-Time auf
wwn-prod ist offen (siehe `docs/backlog.md` → Sicherheit).

### Wo finde ich was

- Pro-Release-Notes: `gh release list` und `gh release view vX.Y.Z`
- Tag-History: `git log --tags --simplify-by-decoration --oneline | head`
- Release-Pipeline-Status: GitHub Actions → "Release"-Workflow

---

## Maintainer-Identität

Alle Commits MÜSSEN signiert sein mit:

- **Author Name**: `Heinz W. Richter`
- **Author Email**: `hwr@relations4u.de`
- **Signing**: SSH-Signing mit `~/.ssh/id_ed25519.pub`
- **Konfiguration**: `gpg.format=ssh`, `commit.gpgsign=true`, `tag.gpgsign=true`

Verifikation vor Commit:

```bash
git config --get user.email   # MUSS hwr@relations4u.de zeigen
git config --get gpg.format   # MUSS ssh zeigen
git log --show-signature -1   # MUSS Good signature zeigen
```

Bei abweichender Identität: SOFORT stoppen, Maintainer fragen, niemals committen.

**Historische Commits** vor dem 4. Mai 2026 sind teilweise unter Platzhalter-Mail
`deine@email.tld` signiert (Setup-Artefakt). Diese werden NICHT gerewritten.

SSH-Public-Key bei GitHub MUSS sowohl als **Authentication Key** als auch als
**Signing Key** registriert sein. Verifizieren mit `gh ssh-key list`.

---

## Repo-Struktur (Monorepo)

```
worldweathernews/
├── apps/
│   ├── backend/        # Go API
│   ├── frontend/       # SvelteKit
│   └── pyworkers/      # Python Worker-Service
├── packages/
│   ├── api-schema/     # OpenAPI-Spec + redocly.yaml (Single Source of Truth)
│   └── shared-types/   # Generierte TS-Types
├── infra/
│   ├── compose/        # Docker-Compose-Files
│   ├── caddy/          # Reverse-Proxy-Config
│   ├── monitoring/     # Prometheus, Grafana, Loki
│   ├── ansible/        # Server-Konfiguration
│   ├── terraform/      # Server-Provisionierung
│   ├── secrets/        # SOPS-verschlüsselte ENV-Files
│   └── migrations/     # goose-DB-Migrations
├── scripts/            # Helper-Scripts
├── docs/               # Architektur, Runbook, ADRs
├── sessions/           # Setup-Session-Prompts (step01.md - step12.md)
├── .github/
│   ├── workflows/      # CI/CD
│   └── dependabot.yml  # Auto-Updates
├── .mise.toml          # Tool-Versionen
├── .pre-commit-config.yaml
├── .lycheeignore       # Markdown-Link-Ignore-Patterns
├── Makefile            # Top-Level-Tasks
├── compose.yml         # → infra/compose/compose.dev.yml
├── CLAUDE.md           # Diese Datei
├── pre-session-checklist.md
└── vm-setup.md
```

---

## Conventions

### Git und PR-Workflow

- **Trunk-based Development**: kurze Feature-Branches, Merge in `main` via PR
- **Branch-Protection auf `main`** (aktiv ab Session 8):
  - Direct Push verboten
  - Pull Request mit grünem CI Pflicht
  - Signed commits Pflicht
  - Linear history (Squash-Merges)
- **Feature-Branch-Schema**:
  - `feat/<thema>` — Neue Features
  - `fix/<thema>` — Bugfixes
  - `chore/<thema>` — Build, Tooling, Refactoring
  - `ci/<thema>` — CI/CD-Änderungen
  - `docs/<thema>` — Reine Doku
- **Conventional Commits**: `feat:`, `fix:`, `chore:`, `docs:`, `refactor:`, `test:`, `perf:`, `ci:`
  - Scope optional: `feat(backend):`, `fix(frontend):`
- **Keine Force-Pushes auf `main`**
- **`main` ist immer deploybar**
- **Semver** für Releases, Tags `v*`

### Code-Style

- Go: `gofmt`, `goimports`, `golangci-lint v2.12.1` müssen grün sein
- Python: `ruff` (Lint + Format), `mypy` strict
- TypeScript: `eslint`, `prettier`, `svelte-check`
- Strukturiertes Logging in JSON: `slog` (Go), `structlog` (Python), `pino` (Node)
- Trace-IDs in jedem Log-Eintrag

### Naming-Konventionen

- ENV-Variablen: `WWN_*` (Backend Go), `WWN_PY_*` (Python), `PUBLIC_*` (Frontend client-side)
- Container-Images: `ghcr.io/relations4u/wwn-backend`, `wwn-frontend`, `wwn-pyworkers`
- Verzeichnisse: kebab-case
- Go-Module: `github.com/relations4u/worldweathernews/apps/backend`
- GitHub-Org: `relations4u`

### API-Design

- **OpenAPI 3.1 ist Single Source of Truth** in `packages/api-schema/openapi.yaml`
- Server-Stubs (Go) und Client-Types (TS) werden generiert
- API-Versionierung im Pfad: `/api/v1/...`
- Errors als Problem Details (RFC 7807)
- Pagination: cursor-basiert, nicht offset
- Default: alle Endpoints öffentlich (`security: []` auf Root-Level)
- Schema-Linting: `redocly lint --config packages/api-schema/redocly.yaml`
- Stilistische Schema-Findings (operation-summary, info-license, unused-components)
  bleiben Warnings bis Session 12 (Schema-Pflege-Pass)

### Datenbank

- Migrations in `infra/migrations/` (sprachunabhängig)
- Jede Migration muss ein `down`-Script haben (Reversibilität)
- Keine direkten Schema-Änderungen ohne Migration
- Foreign Keys explizit, ON DELETE CASCADE bewusst

### Tests

- Go: Tests neben Code (`foo.go` + `foo_test.go`), `-race`-Flag im CI
- Python: `tests/`-Verzeichnis, pytest mit `asyncio_mode = "auto"`
- Frontend: Unit-Tests via Vitest, E2E später mit Playwright
- Mindest-Coverage: kein hartes Gate, aber sichtbar in CI

### Dockerfiles

- Alle `go install`/`pip install`/`npm install -g`-Befehle MÜSSEN eine konkrete
  Version pinnen (DL3062). `@latest` verboten.
- Aktuelle Air-Adresse: `github.com/air-verse/air` (NICHT mehr `cosmtrek/air`)
- Aktuelle Air-Version: v1.65.0 (regelmäßig prüfen)
- Base-Images mit konkretem Tag, nicht `:latest` (DL3007)
- `apt-get install` immer mit `--no-install-recommends` und
  `rm -rf /var/lib/apt/lists/*` am Ende derselben RUN-Layer
- Multi-stage Builds bevorzugen, finales Runtime-Image so klein wie möglich
- HEALTHCHECK in jedem Service-Image
- OCI-Labels (org.opencontainers.image.source etc.)
- Non-root user in Runtime-Stage

### Pre-commit-Strategie

- **Sprach-spezifische Formatter als `repo: local` Hooks**, die Workspace-Tools
  nutzen (`pnpm --filter ... exec prettier`, `uv run ruff`)
- **KEINE `additional_dependencies`** für Sprach-Tools — würde sonst zwei
  Versionen parallel pflegen (Hook vs. App)
- `check-json`-Hook schließt JSONC-Files aus (tsconfig.json, .vscode/\*.json,
  .code-workspace)
- Nach jeder Änderung an `.pre-commit-config.yaml`:
  `pre-commit clean && pre-commit install --install-hooks`

### YAML-Datei-Erstellung

- **NIEMALS heredoc in zsh** für YAML-Files — hat in der Vergangenheit
  binäre Korruption erzeugt (BOM, CRLF)
- Stattdessen: `printf '%s\n' ...` oder vi öffnen und manuell schreiben lassen
- Maintainer nutzt **vi**, nicht nano. In Anweisungen entsprechend formulieren.
- Verifikation neu erstellter YAML-Files:
  ```bash
  xxd file.yaml | head -3   # KEIN "efbb bf" am Anfang
  file file.yaml             # NICHT "with BOM" oder "CRLF line terminators"
  yamllint file.yaml         # syntaktisch korrekt
  ```

---

## Workflow-Regeln für Claude Code

Diese Regeln gelten **immer** und überschreiben alle Default-Verhaltensweisen.

### Plan vor Ausführung

- Bei jeder Änderung, die mehr als 3 Dateien betrifft: **erst Plan zeigen**
- Plan beinhaltet: Welche Dateien, welche Änderungen, welche Risiken
- Auf explizite Freigabe warten
- Plan-Mode nutzen wenn verfügbar (Shift+Tab in Claude Code)

### Commits und Push

- **Nie eigenständig committen.** Der Maintainer committet selbst.
- **Nie auf `main` direkt pushen.** Branch-Protection würde es blockieren —
  immer Feature-Branch + PR.
- Am Ende jeder Session: `git status` + Commit-Vorschlag formulieren
- Commit-Messages folgen Conventional Commits
- Vor Commit-Vorschlag: Identität verifizieren
  (`git config --get user.email` muss `hwr@relations4u.de` zeigen)

### Bei Unklarheit

- **Fragen statt annehmen.** Lieber eine Frage zu viel als eine falsche Annahme.
- Wenn mehrere sinnvolle Optionen existieren: alle nennen, Empfehlung aussprechen, fragen.

### Qualität

- Vor Fertig-Meldung: Linter und Tests laufen lassen
- Bei roten Tests: erst grün machen, dann melden
- Bei Lint-Findings: fixen, nicht ignorieren
- Wenn etwas nicht funktioniert: ehrlich sagen, nicht überspielen

### Dependency-Auswahl

- **Bevor** du eine neue Dependency hinzufügst: prüfen ob sie in der Tech-Stack-Tabelle steht
- Wenn nicht: fragen, mit Begründung
- Bevorzugung: bekannte, gut gepflegte Libraries; Stdlib wo möglich
- Bei jeder Dependency-Wahl: prüfen, ob sie mit Go 1.25 / Node 22 / Python 3.12 kompatibel ist

### Generated Code

- Generierter Code (sqlc, oapi-codegen, openapi-typescript) wird committed
- CI prüft via `git diff --exit-code` ob aktuell
- Generated Files mit `// Code generated by ...; DO NOT EDIT.` markieren
- Aus Lint ausschließen via `.golangci.yml` exclusions

### Sicherheit

- Keine Secrets in Code, Logs, Tests, Konfigurationen
- `.env`-Files: nur `.env.example` committen, außer es sind reine Public-Defaults
  (z. B. `apps/frontend/.env` mit `PUBLIC_API_BASE_URL=http://localhost:8080`
  ist OK, weil ohnehin Build-Zeit-konstant)
- SQL via parametrisierte Queries, niemals String-Konkatenation
- User-Input validieren

---

## Wichtige Kommandos

Diese werden im Verlauf der Setup-Sessions implementiert. Ist ein Kommando noch
nicht da, ist das ein Hinweis, dass es noch fehlt.

### Top-Level

- `make bootstrap` — Erst-Setup nach Repo-Clone
- `make dev` — Lokale Entwicklung starten
- `make dev-full` — Inkl. Monitoring-Stack
- `make dev-down` — Stack stoppen
- `make dev-reset` — Stack stoppen + Volumes löschen
- `make test` — Alle Tests
- `make lint` — Alle Linter
- `make fmt` — Auto-Format
- `make build` — Alle Container bauen
- `make gen` — Generierten Code aktualisieren (OpenAPI → Go/TS)
- `make migrate` — DB-Migrations anwenden
- `make clean` — Aufräumen

### Per Service

- `make backend-dev`, `make backend-test`, `make backend-lint`
- `make frontend-dev`, `make frontend-test`, `make frontend-lint`
- `make pyworkers-dev`, `make pyworkers-test`, `make pyworkers-lint`

### Git und PR-Workflow

- `git checkout -b <branch-prefix>/<thema>` — Branch für die Session
- `gh pr create --base main` — PR eröffnen nach erstem Push
- `gh pr checks --watch` — CI-Status live verfolgen
- `gh pr merge --squash --delete-branch` — Mergen wenn alle Checks grün

---

## Don'ts (harte Regeln)

- **Kein Kubernetes/K3s/Helm** in dieser Phase. Wir bleiben bei Compose.
  Wachstumspfad ist offen, aber jetzt nicht.
- **Kein ORM in Go.** sqlc + pgx direkt.
- **Kein npm oder yarn.** pnpm.
- **Kein pip oder poetry.** uv.
- **Keine Cloud-locked-in Services im kritischen Pfad.** Alles, was wir nutzen,
  muss self-hostbar sein oder leicht ersetzbar.
- **Keine Mock-Daten in Production-Code.** Tests dürfen mocken, App-Code nicht.
- **Keine kommentierten Code-Blöcke** im Repo. Wenn weg, dann weg. Git erinnert sich.
- **Keine `any`-Types in TypeScript** ohne Begründung im Kommentar.
- **Keine `interface{}`/`any` in Go-Public-APIs**.
- **Keine globalen Variablen** ohne sehr guten Grund.
- **Keine TODO-Kommentare ohne Issue-Referenz** (nach Phase 1).
- **Keine direkten Pushes auf `main`** — Branch-Protection blockt es ohnehin.
- **Keine `latest`-Pins** in irgendeiner Toolchain-Datei.
- **Keine YAML-Files via heredoc** in zsh — printf oder vi nutzen.

---

## Wo finde ich was

| Anliegen                       | Ort                                                                                            |
| ------------------------------ | ---------------------------------------------------------------------------------------------- |
| Architektur-Diagramme          | `docs/architecture.md`                                                                         |
| Datenquellen-Übersicht         | `docs/data-sources.md`                                                                         |
| Was tun wenn X kaputt?         | `docs/runbook.md`                                                                              |
| Wie deploye ich?               | `docs/deployment.md`                                                                           |
| Wie entwickle ich Feature X?   | `docs/development.md`                                                                          |
| Warum diese Tech-Entscheidung? | `docs/adr/0001-…` bis `docs/adr/0005-…`                                                        |
| Folge-Tracker (Backlog)        | `docs/backlog.md`                                                                              |
| Lizenz                         | `LICENSE` (AGPL-3.0)                                                                           |
| Config-Reference               | `docs/config-reference.md` (generiert)                                                         |
| Secrets-Workflow               | `docs/secrets.md`                                                                              |
| Service-spezifische Doku       | `apps/<service>/README.md`                                                                     |
| Setup-Sessions (12 Stück)      | `sessions/step01.md` – `step12.md`                                                             |
| Status der Sessions            | `sessions/STATUS.md`                                                                           |
| VM-Setup-Anleitung Dev         | `vm-setup.md` (Entwicklungs-VM)                                                                |
| VM-Setup-Anleitung Prod        | `vm-prod-setup.md` (Forschungs-Production-VM mit Caddy)                                        |
| VM-Setup-Anleitung Mon         | `vm-mon-setup.md` (Monitoring-VM, intern only)                                                 |
| Pre-Session-Checkliste         | `pre-session-checklist.md`                                                                     |
| HTTP-Handler                   | `apps/backend/internal/http/handler/`                                                          |
| Locations-/Wetter-Handler      | `apps/backend/internal/http/handler/api.go` (`ListLocations`, `GetLocationDetail`)             |
| sqlc-Queries (Hand)            | `apps/backend/internal/storage/queries/`                                                       |
| sqlc-Output (generiert)        | `apps/backend/internal/storage/db/`                                                            |
| sqlc-Schema (generiert)        | `apps/backend/internal/storage/schema.sql` (aus den goose-Up-Sections, via `make sqlc-schema`) |
| DB-Migrationen                 | `infra/migrations/`                                                                            |
| Migrations-Wrapper             | `scripts/migrate.sh` (`make migrate*`)                                                         |
| Worker-Jobs                    | `apps/pyworkers/pyworkers/jobs/`                                                               |
| Open-Meteo-Worker              | `apps/pyworkers/pyworkers/jobs/open_meteo.py`                                                  |
| DWD-POI-Worker                 | `apps/pyworkers/pyworkers/jobs/dwd.py`                                                         |
| Frontend-Routes                | `apps/frontend/src/routes/`                                                                    |
| Wetter-Route                   | `apps/frontend/src/routes/wetter/` (CSR-only, siehe ssr=false-Note)                            |
| WeatherCard-Component          | `apps/frontend/src/lib/components/WeatherCard.svelte`                                          |
| StationsMap-Component          | `apps/frontend/src/lib/components/StationsMap.svelte` (MapLibre, lazy in onMount)              |
| Karten-Config (Tile-URL)       | `apps/frontend/src/lib/config/map.ts` (zentraler Wechselpunkt der Tile-Quelle)                 |
| Wind-Helfer (compass/Pfeil)    | `apps/frontend/src/lib/wind.ts` (geteilt: WeatherCard + StationsMap)                           |
| Frontend-API-Client            | `apps/frontend/src/lib/api/client.ts`                                                          |
| Frontend-Default-Env           | `apps/frontend/.env` (committed!)                                                              |
| OpenAPI-Schema                 | `packages/api-schema/openapi.yaml`                                                             |
| OpenAPI-Lint-Config            | `packages/api-schema/redocly.yaml`                                                             |
| Generierte Files               | markiert mit Header `// Code generated by ...`                                                 |
| Secrets                        | `infra/secrets/<env>/*.sops.*`                                                                 |
| Media-Bucket-Doku              | `docs/media-storage.md`                                                                        |
| Media-Bucket-Credentials       | `infra/secrets/production/media-storage.env`                                                   |
| Media-Bucket-Policy + CORS     | `infra/object-storage/{bucket-policy,cors}.json`                                               |
| CMS-Authoring-Guide            | `docs/cms.md`                                                                                  |
| Sveltia-Loader                 | `apps/frontend/static/admin/index.html`                                                        |
| Sveltia-Konfiguration          | `apps/frontend/static/admin/config.yml`                                                        |
| CMS-OAuth-Proxy                | `apps/cms-auth/` (self-hosted Go-Service auf wwn-prod hinter Caddy)                            |
| Markdown-Content-Pages         | `apps/frontend/src/content/pages/{de,en}/<slug>.md`                                            |
| Content-Components             | `apps/frontend/src/lib/content-components/`                                                    |
| i18n-Messages                  | `apps/frontend/messages/{de-de,en}.json` + `apps/frontend/project.inlang/`                     |
| Compose-Configs                | `infra/compose/`                                                                               |
| Lychee-Ignores                 | `.lycheeignore`                                                                                |
| Hosting-Hardware               | `docs/architecture.md` (Hardware-Übersicht)                                                    |
| DNS-Records                    | Cloudflare-Dashboard für `worldweathernews.com`, Joker.com für `hw7.eu`                        |
| ProtonMail-DNS-Konfiguration   | Proton-Webclient → Settings → All settings → Domain names → worldweathernews.com → Configure   |
| DynDNS-Konfiguration           | Firewall-Hardware vor dem Proxmox-Host                                                         |
| Production-Caddyfile           | `infra/caddy/prod/Caddyfile` (eigener Compose-Stack auf wwn-prod)                              |
| Public Status-Page             | https://status.worldweathernews.com (ab Session 10)                                            |

---

## Externe Datenquellen

**Aktiv (Stand Iteration 2.2):**

- **DWD POI** (opendata.dwd.de) — GeoNutzV, kein API-Key. Sechs
  Stationen (Potsdam, Berlin, Hamburg, Brocken, Zugspitze, Helgoland),
  sechs Variablen (T, Niederschlag, Wind-Speed/Direction, Druck,
  Feuchte), halbstündliches Polling. Default-Source für alle drei
  Stadt-Slugs. Details in `docs/data-sources.md`.
- **Open-Meteo** (open-meteo.com) — CC-BY-4.0, EU-basiert, kein API-Key.
  Drei Stadt-Locations (Potsdam, Berlin, Hamburg), sechs Variablen,
  current 10-min + hourly 60-min. Per `?source=open-meteo`-Param
  erreichbar; in 2.2 Backfill für pressure_msl + relative_humidity_2m.

**Geplant für spätere Iterationen:**

- **DWD MOSMIX / ICON** — Forecast-Pfade ergänzend zu POI
  (Iteration 2.2b und Klima-Folge)
- **NOAA** (USA) — National Weather Service API
- **Met Office** (UK), **JMA** (Japan), **Météo-France** etc. — phasenweise
- **EUMETSAT** — Satellitenbilder
- **USGS** — Erdbebendaten
- **NOAA Space Weather** — Aurora-Vorhersagen

Worker-Pattern: Pull alle X Minuten je Quelle, normalisieren, in DB. Cache in Redis.

---

## Beantwortete Entscheidungen

Diese Fragen sind im Verlauf der Setup-Phase entschieden worden:

- ✅ **GitHub-Org**: `relations4u`
- ✅ **Repo**: `github.com/relations4u/worldweathernews` (privat)
- ✅ **Dev-Setup**: Proxmox-VM mit Ubuntu 24.04 LTS, nicht macOS direkt
- ✅ **Hosting-Provider (Forschungs-Phase)**: Self-Hosting auf eigenem Proxmox
  (Ryzen 7, 32 GB RAM, 500 GB HD, Hardware-Firewall davor, 1000/500 Mbit
  Anschluss). Hetzner-Migration später als Option dokumentiert.
- ✅ **Git-Hosting**: GitHub (nicht Forgejo)
- ✅ **Toolchain-Manager**: mise mit projekt-lokaler `.mise.toml`
- ✅ **Go-Version**: 1.25 (durch pgx v5.9.2 Dependency erzwungen)
- ✅ **Commit-Signing**: SSH-Signing (nicht GPG)
- ✅ **Editor-Konvention**: vi (Maintainer-Präferenz)
- ✅ **Domain-Registrar**: Joker.com (Domain `worldweathernews.com`)
- ✅ **DNS-Provider**: Cloudflare (Free-Plan, DNS-only-Modus, kein Proxy)
- ✅ **DynDNS-Anker**: `gate.hw7.eu` (separate Domain bei Joker, einziger
  Update-Punkt für Public-IP-Wechsel)
- ✅ **Subdomain-Schema**: `research.` für Forschungs-Phase (nicht `forschung.`,
  nicht `resa.`)
- ✅ **Apex-Strategie**: Cloudflare CNAME-Flattening
- ✅ **Mail-Provider**: ProtonMail (kostenpflichtig, Domain `worldweathernews.com`
  verified, MX/DKIM/SPF/DMARC aktiv konfiguriert)
- ✅ **Lizenz**: AGPL-3.0 (in Session 12 entschieden, siehe `LICENSE`)

### Entscheidungen ab 2026-05-06 (Caddy-Online + Session 11)

- ✅ **Caddy-Stack-Topologie**: eigenständiger Compose-Stack unter
  `/srv/wwn/caddy` auf wwn-prod mit `network_mode: host` (echte Client-IPs
  in den Logs, kein Docker-iptables-NAT). **Nicht** Teil des
  App-Compose-Stacks. Deployed via `infra/deploy/deploy-caddy.sh`
  (rsync + ssh docker compose).
- ✅ **HSTS-Strategie für die Stub-Phase**: `max-age=31536000` **ohne**
  `includeSubDomains` und ohne `preload`. Bewusste Abweichung vom
  ursprünglichen Plan (5. Mai) — Begründung: zukünftige Subdomains
  (z. B. Monitoring intern) bekommen ggf. lange kein TLS, und
  `includeSubDomains` würde sie dann lockout'en. Verschärfung auf
  `includeSubDomains` und perspektivisch `preload`, sobald alle
  geplanten Subdomains zuverlässig TLS haben.
- ✅ **ACME-Email für Caddy**: `ops@worldweathernews.com`
  (ProtonMail-Postfach).
- ✅ **SOPS-Setup**: age-Keys, Public-Key im Repo unter `.sops.yaml`,
  Privater Key beim Maintainer unter `~/.config/sops/age/keys.txt`
  (Mode 0600). Secret-Files unter `infra/secrets/<env>/<service>.env`.
  Pre-commit-Hook `forbid-unencrypted-secrets` blockt Plaintext-Commits.
- ✅ **Ansible-User**: `deploy` (Default im Inventory). Erster Bootstrap
  auf dem manuell angelegten wwn-prod via Override
  `-e ansible_user=hwr`, danach übernimmt der Default.
- ✅ **Ansible vs. Caddy**: app-Rolle deployt Postgres/Redis/Backend/
  Frontend/Pyworkers — Caddy bewusst nicht angefasst, läuft weiter
  als eigener Stack unter `/srv/wwn/caddy`. Caddy-Block in
  `infra/compose/compose.prod.yml` wird in eigener Folge-PR entfernt.
- ✅ **Terraform-Provider-Strategie**: `bpg/proxmox ~> 0.66` aktiv für die
  Forschungs-Phase. `hetznercloud/hcloud ~> 1.48` als Migrations-Stub
  bleibt im Repo, aber unbenutzt.
- ✅ **DNS-Management**: bleibt manuell in der Cloudflare-UI. Kein
  Terraform-Modul für DNS-Records — sonst rauscht der Zustand mit
  Hand-Edits auseinander.
- ✅ **Inventory-Scope**: nur `production`, kein `staging` in der
  Forschungs-Phase. SOPS-`creation_rules` akzeptieren `staging` schon
  als Pfad-Pattern, das Verzeichnis ist aber leer.
- ✅ **Commit-Workflow für größere Änderungen**: Feature-Branch + PR.
  Direct-Commits auf main nur für eng abgegrenzte Hotfixes nach
  expliziter Maintainer-Freigabe (Beispiel: zwei Caddy-Deploy-Commits
  am 6. Mai 2026).
- ✅ **`terraform import` vor erstem `apply`** für die bestehenden
  manuell erstellten VMs (wwn-prod, wwn-mon) — Workflow im
  `infra/terraform/README.md` dokumentiert, ist Maintainer-Hausaufgabe
  (NICHT in der Skelett-Session ausgeführt).

### Entscheidungen ab 2026-05-11

- ✅ **CMS-OAuth-Proxy self-hosted statt Cloudflare-Worker**:
  Sveltia-Auth läuft als kleiner Go-Service `apps/cms-auth/` (Chi-Router,
  ~170 LOC, Distroless-Image) im App-Compose-Stack auf wwn-prod hinter
  Caddy unter `cms-auth.worldweathernews.com`. Begründung: Maintainer-
  Prinzip „Cloudflare-Abhängigkeit minimieren". Der vormalige Worker
  unter `infra/cloudflare-worker-cms-auth/` wurde nach erfolgreichem
  Cutover am 11. Mai 2026 aus dem Repo entfernt (Worker selbst via
  `wrangler delete` im Account `hwr-06e` durch Maintainer abgebaut).
  Migration-Checkliste in `docs/cms.md` → „Maintainer-Aufgaben für
  Erst-Aktivierung". Setzt A.4 aus
  `sessions/feature1/feature-decisions.md` ausser Kraft.

- ✅ **Self-hosting-Prinzip als generelle Architektur-Leitlinie**:
  Aus dem cms-auth-Refactor abgeleitet. Neue Services landen als
  Container im App-Compose-Stack auf wwn-prod, nicht als Cloudflare-
  Worker oder vergleichbare Drittanbieter-Compute-Schicht. Begründung:
  DNS ist bei Cloudflare bereits ein kritischer Pfad; jeder weitere
  kritische Pfad beim selben Anbieter hebt das Migrations-Risiko
  unnötig. Reine **Edge-/Cache-Schichten** vor self-hosted Origins
  sind die einzige Ausnahme (kein kritischer Pfad). Konkrete Folgen:
  Iteration 1.3b (Image-Pipeline) wird `apps/cms-media-upload/`
  Go-Service. Track 3 LLM-Calls sind davon **nicht** betroffen —
  Cloud-LLM-API-Aufrufe sind keine Compute-Hosting-Frage. Volle
  Regel als A.19 in `sessions/feature1/feature-decisions.md`.
- Worker-Scheduling-Pattern (W1 als Default)
- Frontend-Position: eigene Route pro Feature

---

## Hosting-Architektur

Die Plattform wird in der Forschungs-Phase **selbst gehostet** auf eigener
Hardware, nicht in einer Public Cloud. Begründung: Kostenkontrolle, volle
Hardware-Kontrolle, DSGVO-Klarheit, Lerneffekt, ausreichende Bandbreite.

### Hardware

| Komponente | Spec                                                       |
| ---------- | ---------------------------------------------------------- |
| Hypervisor | Proxmox VE auf AMD Ryzen 7                                 |
| RAM        | 32 GB                                                      |
| Storage    | 500 GB SSD                                                 |
| Firewall   | Dedizierte Hardware-Firewall vor dem Proxmox-Host          |
| Anschluss  | 1000 Mbit Down / 500 Mbit Up                               |
| DNS-Anker  | `gate.hw7.eu` (separate Domain bei Joker, DynDNS-gepflegt) |

### VM-Aufteilung

```
[Internet] → [Hardware-Firewall (NAT 80/443)] → [Proxmox-Host]
                                                       │
                                                       ├─ wwn-dev   (8 GB) — Entwicklung
                                                       ├─ wwn-prod  (8 GB) — Forschungs-Production
                                                       └─ wwn-mon   (4 GB) — Monitoring (separate VM)

                                                       Reserve: ~12 GB
```

**Aufgabentrennung:**

- **wwn-dev:** Entwicklung mit Toolchain, IDE, Claude Code
- **wwn-prod:** Application-Stack (Backend, Frontend, Workers, Postgres, Redis, Caddy) — public erreichbar via research.worldweathernews.com
- **wwn-mon:** Observability-Stack (Prometheus, Grafana, Loki, Tempo, Uptime Kuma) — intern only, Zugang via VPN/SSH-Tunnel

Die Trennung von prod und mon hat zwei Hauptvorteile: Wenn wwn-prod abstürzt,
bleibt Telemetrie auf wwn-mon verfügbar (forensische Analyse). Plus die
hohe I/O-Last von Prometheus/Loki konkurriert nicht mit Application-Code.

**IP-Schema im manager-Netz (10.100.100.0/24):**

| Host               | IP             | MAC                 | Status                   |
| ------------------ | -------------- | ------------------- | ------------------------ |
| OPNsense (Gateway) | 10.100.100.1   | —                   | aktiv                    |
| wwn-dev            | 10.100.100.113 | `bc:24:11:44:7e:dd` | aktiv (DHCP-Reservation) |
| wwn-prod           | 10.100.100.21  | `bc:24:11:00:21:21` | Basis installiert        |
| wwn-mon            | 10.100.100.22  | `bc:24:11:00:22:22` | Basis installiert        |

MAC-Konvention: letzte zwei Bytes spiegeln IP-Suffix (`:21:21` für `.21`,
`:22:22` für `.22`). Erlaubt DHCP-Reservation in OPNsense **vor** VM-Erstellung.

**VLAN-Konfiguration:** Die Zuordnung der VMs zum manager-VLAN (VLAN 04,
10.100.100.0/24) erfolgt **port-basiert über die Switches**, nicht über
Proxmox-VLAN-Tags. Proxmox-Bridge `vmbr0` ist tagless konfiguriert; der
Switch-Port am Proxmox-Host ist für VLAN 04 als Access-Port eingerichtet.
Proxmox-VMs an `vmbr0` landen automatisch im manager-Netz, ohne dass im
VM-Network-Profil ein VLAN-Tag gesetzt werden darf (würde sonst Doppel-Tagging
verursachen — siehe Häufige Fallen).

### Forschungs-Produktion ist NICHT echte Production

- Keine SLA-Versprechen
- Kein automatisches Failover bei Hypervisor-Ausfall
- Kein dedizierter DDoS-Schutz (nur Hardware-Firewall)
- Kein dynamisches Skalieren

Der Status wird Nutzern explizit kommuniziert via Banner auf
`research.worldweathernews.com`. DSGVO-Pflichten gelten trotzdem (Impressum,
Datenschutzerklärung, Cookie-Hinweis).

### Migration-Pfad zur echten Production

Wenn die Plattform reif für echte Production ist, wird ein zweiter Stack
auf Hetzner Cloud Deutschland aufgebaut (siehe ADR in `docs/adr/`).
`research.` bleibt als Beta-Instanz weiter laufen. Migration-Schritte werden
in Session 11 vorbereitet (Ansible/Terraform funktionieren beide gegen
Proxmox jetzt und gegen Hetzner später).

---

## Domain- und DNS-Architektur

### Domain und Registrar

- **Domain:** `worldweathernews.com`
- **Registrar:** Joker.com (bleibt unverändert)
- **DNS-Provider:** Cloudflare Free-Plan (Nameserver bei Joker auf
  Cloudflare delegiert)
- **DynDNS-Provider:** Joker.com für Domain `hw7.eu`, gepflegter Hostname
  `gate.hw7.eu`

### Subdomain-Schema

| Hostname                            | Zweck                                  | Phase                                              |
| ----------------------------------- | -------------------------------------- | -------------------------------------------------- |
| `worldweathernews.com`              | Apex — Landing-Page                    | jetzt: simple Landing, später: Production-Frontend |
| `www.worldweathernews.com`          | 301-Redirect auf Apex                  | dauerhaft                                          |
| `research.worldweathernews.com`     | Forschungs-Frontend                    | Forschungs-Phase (jetzt)                           |
| `api.research.worldweathernews.com` | Forschungs-Backend-API                 | Forschungs-Phase (jetzt)                           |
| `api.worldweathernews.com`          | Production-Backend-API                 | später (Production-Phase)                          |
| `cms-auth.worldweathernews.com`     | Self-hosted OAuth-Proxy für Sveltia    | seit 2026-05-11                                    |
| `media.worldweathernews.com`        | Hetzner Object-Storage Read-Only-Proxy | seit Iteration 1.1b                                |
| `status.worldweathernews.com`       | Public Uptime-Page (Uptime Kuma)       | ab Session 10                                      |

### Nicht öffentlich erreichbar

Diese Services bekommen **keine** öffentliche Subdomain — Zugriff nur via
SSH-Tunnel oder VPN:

- Grafana, Prometheus, Loki, Tempo (Monitoring-Stack)
- Datenbank-Admin-Tools
- SOPS-verschlüsselte Konfigurationen

### DNS-Auflösungs-Kette

Einziger DynDNS-Update-Punkt ist `gate.hw7.eu`. Alle WWN-Hostnames sind CNAMEs
darauf, mit `home.worldweathernews.com` als interner Anker:

```
gate.hw7.eu                              A      <Public-IP>          (Joker DynDNS)
home.worldweathernews.com                CNAME  gate.hw7.eu          (Cloudflare)
worldweathernews.com (Apex)              CNAME  gate.hw7.eu          (Cloudflare CNAME-Flattening)
www.worldweathernews.com                 CNAME  worldweathernews.com (Cloudflare)
research.worldweathernews.com            CNAME  home.worldweathernews.com (Cloudflare)
api.research.worldweathernews.com        CNAME  home.worldweathernews.com (Cloudflare)
cms-auth.worldweathernews.com            CNAME  home.worldweathernews.com (Cloudflare)
media.worldweathernews.com               CNAME  home.worldweathernews.com (Cloudflare)
```

### Cloudflare-Konfiguration

- **Plan:** Free
- **Proxy-Modus:** AUS (DNS-only, graue Wolke). Aktivierung später bei Bedarf
  pro Subdomain möglich.
- **TTL:** Auto (300 s im DNS-only-Modus)
- **Apex-Strategie:** CNAME-Flattening (Cloudflare-Spezialfeature, RFC-konform
  via Antwort-Time-Resolution)

### Mail-Strategie (entschieden)

Mail für `@worldweathernews.com` läuft über **ProtonMail** (kostenpflichtiger
Plan, Mailbox in der Schweiz, DSGVO-konform). Domain ist verified, Mailflow
aktiv, Mail-Versand und -Empfang funktionieren.

**Mail-Konfiguration ist betriebskritisch.** Ändere KEINE der folgenden
Records ohne Maintainer-Rücksprache. Falsche Records bedeuten verlorene Mails.

#### Aktive DNS-Records bei Cloudflare

```
;; Mail-Empfang
worldweathernews.com.                MX    10 mail.protonmail.ch
worldweathernews.com.                MX    20 mailsec.protonmail.ch

;; Mail-Versand (DKIM-Signing über 3 Selectors mit Auto-Rotation)
protonmail._domainkey.worldweathernews.com.   CNAME  protonmail.domainkey.<id>.domains.proton.ch
protonmail2._domainkey.worldweathernews.com.  CNAME  protonmail2.domainkey.<id>.domains.proton.ch
protonmail3._domainkey.worldweathernews.com.  CNAME  protonmail3.domainkey.<id>.domains.proton.ch

;; SPF — erlaubt Proton als legitimen Sender
worldweathernews.com.                TXT    "v=spf1 include:_spf.protonmail.ch ~all"

;; DMARC — Phishing-Schutz, p=quarantine konservativ
_dmarc.worldweathernews.com.         TXT    "v=DMARC1; p=quarantine; rua=mailto:hwr@relations4u.de; sp=quarantine; aspf=s; adkim=s; pct=100"

;; Domain-Ownership-Beweis für Proton (einmalig, dauerhaft)
worldweathernews.com.                TXT    "protonmail-verification=<hash>"
```

`<id>` ist die Proton-Account-spezifische Zone (bei diesem Account:
`d26nxgbbk2z2xfxpjqypvwn3ioaxn63tvyfat3wph3itsdk7xyaeq`). Die exakten Werte
stehen im Proton-Dashboard unter Settings → All settings → Domain names →
worldweathernews.com → Configure.

#### DMARC-Upgrade-Pfad

Aktuell `p=quarantine`. Wenn nach 1-2 Wochen `rua`-Reports keine legitimen
Mails fälschlich quarantäniert werden, kann auf `p=reject` upgegradet werden
für maximalen Schutz.

#### Transactional Mails (zukünftig)

Wenn die Plattform User-Registrierung, Alerts oder andere automatisierte
Mails versendet, läuft das über einen separaten Provider (Postmark/Brevo/
SES — Entscheidung später). SPF wird dann erweitert:

```
v=spf1 include:_spf.protonmail.ch include:<provider> ~all
```

Sender-Adresse für Transactional würde `noreply@worldweathernews.com` oder
`noreply@research.worldweathernews.com` sein, separat von normaler Korrespondenz.

### CAA-Records (Certificate-Aussteller)

Aktuell permissiv konfiguriert — mehrere CAs erlaubt (Let's Encrypt, Google
Trust Services, Sectigo, DigiCert, ssl.com), plus iodef für Bericht bei
Mis-Issuance:

```
worldweathernews.com.  CAA  0 issue     "letsencrypt.org"
worldweathernews.com.  CAA  0 issue     "pki.goog; cansignhttpexchanges=yes"
worldweathernews.com.  CAA  0 issue     "comodoca.com"
worldweathernews.com.  CAA  0 issue     "digicert.com; cansignhttpexchanges=yes"
worldweathernews.com.  CAA  0 issuewild "letsencrypt.org"
worldweathernews.com.  CAA  0 iodef     "mailto:hwr@relations4u.de"
```

Die permissive Liste ist Cloudflare-Default und bleibt. Dein Caddy holt
trotzdem nur Let's-Encrypt-Zertifikate — die anderen CAs werden nur
referenziert für eventuelle Cloudflare-Universal-SSL-Aktivierung später.

### TLS-Zertifikate

- **Caddy** holt automatisch Let's-Encrypt-Zertifikate via HTTP-01-Challenge
- Voraussetzung: Port 80 muss von außen auf den Caddy-Container weitergeleitet sein
- Wildcard-Zertifikate (DNS-01) werden NICHT genutzt — pro Hostname ein
  Einzel-Zertifikat
- Zertifikat-Storage: persistentes Volume in Compose (`caddy_data`)

### Was passiert bei IP-Wechsel

1. DynDNS-Client auf der Firewall/dem Router meldet neue IP an Joker.com
2. Joker aktualisiert A-Record für `gate.hw7.eu`
3. Cloudflare löst CNAME-Kette frisch auf, sobald TTLs auslaufen
4. Resolver weltweit erfahren neue IP nach 1–10 Min
5. Caddy braucht keinen Neustart — Hostnames bleiben stabil

### CORS und Cross-Origin

- Backend setzt `Access-Control-Allow-Origin` explizit auf
  `https://research.worldweathernews.com` für Forschungs-Phase
- Caddy macht zusätzlich Preflight-Handling für saubere OPTIONS-Antworten
- Frontend ruft API mit absoluter URL `PUBLIC_API_BASE_URL=https://api.research.worldweathernews.com`

---

## Offene Fragen (werden im Verlauf beantwortet)

- [ ] **Transactional**-Mail-Provider (für User-Registrierung, Alerts, etc.):
      Postmark, Brevo, AWS SES, eigener SMTP? — Erst in Feature-Phase relevant.
      Normale Mailbox-Mails laufen über ProtonMail, das ist entschieden.
- [ ] i18n-Library: svelte-i18n vs. Paraglide vs. Inlang — Entscheidung
      in der ersten Feature-Session.
- [ ] Backup-Ziel: S3-kompatibel (Hetzner Storage Box?), eigenes NAS, BorgBase?
      Empfehlung im `docs/backlog.md` als Operations-Punkt: Proxmox Backup
      Server als eigene VM auf demselben Host, plus externer Off-Site-Sync.
- [ ] Sentry vs. GlitchTip — bewusst aus Session 10c rausgehalten. Aktuelle
      Telemetrie: Logs via Loki, Traces via Tempo. Error-Tracking-Tool kommt
      in der Feature-Phase, wenn Application-Code Error-Capture-fähig ist
      (z. B. Sentry-SDK in Backend/Frontend integriert).

Wenn eine dieser Fragen für deine Aufgabe relevant wird: **fragen**, nicht annehmen.

---

## Session-Tracking

Die Sessions zur Initial-Einrichtung sind dokumentiert in `sessions/step01.md` bis
`sessions/step12.md`. Wenn du in einer Session arbeitest, halte dich an die dort
definierten Aufgaben. Bei Abweichung: zurück zur Datei, fragen.

Stand der Sessions wird in `sessions/STATUS.md` gepflegt — am Ende jeder Session
ein kurzer Eintrag.

### Aktueller Status (6. Mai 2026)

| Phase | Session                                            | Status                                                     |
| ----- | -------------------------------------------------- | ---------------------------------------------------------- |
| A     | 1 — Repo-Skelett, mise                             | ✅ Abgeschlossen                                           |
| A     | 2 — Pre-commit, Makefile                           | ✅ Abgeschlossen                                           |
| B     | 3 — Compose-Stack                                  | ✅ Abgeschlossen                                           |
| B     | 4 — Go-Backend                                     | ✅ Abgeschlossen                                           |
| B     | 5 — SvelteKit-Frontend                             | ✅ Abgeschlossen                                           |
| B     | 6 — Python-Workers                                 | ✅ Abgeschlossen                                           |
| C     | 7 — OpenAPI + Codegen                              | ✅ Abgeschlossen                                           |
| C     | 8 — CI-Workflows                                   | ✅ Abgeschlossen                                           |
| C     | 9 — Release + ghcr.io                              | ✅ Abgeschlossen                                           |
| D     | 10a — wwn-prod und wwn-mon Basis-Setup             | ✅ Abgeschlossen (siehe unten)                             |
| D     | 10b — Caddy auf wwn-prod, Public-Erreichbarkeit    | ✅ Abgeschlossen (6. Mai 2026, Snapshot `caddy-online`)    |
| D     | 10c — Observability-Stack auf wwn-mon              | ✅ Abgedeckt durch Session 11a (Rolle `monitoring-stack`)  |
| D     | 11 — Ansible + SOPS + Terraform-Skelett            | ✅ Code-Skelett gemerged (#22, #23)                        |
| D     | 11a — Komplettes Deployment auf wwn-prod + wwn-mon | ✅ Abgeschlossen (6. Mai 2026, v0.0.2 live)                |
| D     | 12 — Dokumentation, ADRs, Runbook                  | ✅ Abgeschlossen (6. Mai 2026, Setup-Phase damit komplett) |

**Session 10a-Ergebnisse** (Basis-Setup beider VMs, Stand 5. Mai 2026):

✅ wwn-prod (10.100.100.21):

- Ubuntu 24.04 LTS Server installiert
- Hardening (SSH key-only, ufw mit minimalen Regeln, fail2ban)
- Docker CE + docker-compose-plugin
- sops 3.9.4 + age (via GitHub-Releases, nicht apt)
- Verzeichnisstruktur unter `/opt/wwn/` angelegt
- `wwnops`-User für Container-Operationen

✅ wwn-mon (10.100.100.22):

- Identisches Basis-Setup wie wwn-prod
- ufw-Regeln intern only (kein Public-Forward)
- Bereit für Observability-Stack

✅ Netzwerk:

- DHCP-Reservations in OPNsense für alle drei VMs
- VLAN-Tagging port-basiert über Switches (kein Proxmox-Tag)
- DNS-Auflösung über Cloudflare → gate.hw7.eu funktioniert
- DynDNS-Update auf FritzBox aktiv

**Session 10b-Ergebnisse** (Caddy live auf wwn-prod, Stand 6. Mai 2026):

✅ Caddy als eigenständiger Compose-Stack unter `/srv/wwn/caddy` mit
`network_mode: host`. Vier Let's-Encrypt-Zertifikate ausgestellt für
worldweathernews.com (Apex), www, research, api.research. HSTS aktiv
mit `max-age=31536000` ohne `includeSubDomains` (Begründung in
„Beantwortete Entscheidungen ab 2026-05-06"). Stub-Antworten via
`respond` 200 — Cutover auf `reverse_proxy` ist Teil von Session 11a.
Snapshot `caddy-online` gesetzt. Deploy-Pfad: `infra/deploy/deploy-caddy.sh`.

**Session 10c-Inhalte** (Observability-Stack auf wwn-mon):

Vollständig in Session 11a umgesetzt — neue Ansible-Rolle
`monitoring-stack` deployt Prometheus/Loki/Tempo/Grafana auf wwn-mon,
`monitoring-agent` (Promtail + node-exporter) läuft auf wwn-prod.

**Session 11a-Ergebnisse** (komplettes Deployment beider VMs, Stand 6. Mai 2026):

✅ wwn-prod (10.100.100.21):

- App-Stack v0.0.2 läuft: backend, frontend, pyworkers (alle
  healthy), postgres (TimescaleDB), redis, plus monitoring-agent
  (promtail, node-exporter). Erstdeploy in 11a war v0.0.1-rc4;
  noch am 6. Mai über v0.0.1-rc3-Diagnose-Patches auf v0.0.2 gehoben
  (siehe Changelog).
- Caddy unverändert (Stand-alone-Stack), aber Caddyfile auf
  `reverse_proxy 127.0.0.1:{3000,8080}` umgestellt; alle vier
  LE-Zertifikate haben unveränderte `notBefore`-Dates seit Session 10b
- App-Backend-CORS auf die drei aktiven Origins (Apex, www, research)
  konfiguriert via SOPS-encrypted backend.env

✅ wwn-mon (10.100.100.22):

- monitoring-stack läuft: Prometheus, Loki, Tempo, Grafana 11.3.0
  mit drei provisionierten Dashboards (Backend, Pyworkers, Infra)
  unter Folder `worldweathernews`
- Grafana an 0.0.0.0:3000 gebunden, UFW erlaubt :3000 nur aus
  10.100.100.0/24 — Zugriff via `http://10.100.100.22:3000`,
  Admin-Passwort in `infra/secrets/production/grafana.env`
- Loki :3100 und Tempo OTLP :4317/:4318 von wwn-prod erreichbar

✅ Public Smokes (alle 200):

- `https://worldweathernews.com` — SvelteKit-Frontend, „Backend
  connected" mit Trace-IDs
- `https://www.worldweathernews.com` — 301 → Apex
- `https://research.worldweathernews.com` — SvelteKit-Frontend
- `https://api.research.worldweathernews.com/api/v1/ping` — Backend-
  JSON mit `traceId`

⚠️ Bekannt-offen (in `prometheus.yml` als Comment dokumentiert):

- Backend-/Pyworkers-`/metrics`-Ports binden 127.0.0.1 only —
  Prometheus auf wwn-mon kann sie nicht scrapen. Entscheidung
  zwischen LAN-Bind+ufw vs. Push-Sidecar steht aus.
- node-exporter für wwn-mon nicht im Stack (nur node-exporter auf
  wwn-prod ist UP) — als Folge-PR, falls Host-Metriken für wwn-mon
  benötigt werden.

**Session 12-Ergebnisse** (Doku-Finalisierung, Stand 6. Mai 2026):

✅ Vier Tranchen, alle gemerged:

- **PR #33** — `README.md` rewrite, `CONTRIBUTING.md`,
  STATUS+CLAUDE-Status-Updates
- **PR #34** — `docs/architecture.md` (Mermaid-Diagramm), `docs/development.md`
- **PR #35** — `docs/deployment.md` (full), `docs/runbook.md`
  (10 konkrete Szenarien)
- **PR #36** — ADRs 0002 (Go-für-Backend), 0003 (Monorepo),
  0004 (Compose-vor-K3s), 0005 (SOPS+age) im MADR-Format,
  `LICENSE` (AGPL-3.0), `docs/backlog.md` als Folge-Tracker,
  In-Context-TODO-Triage

✅ Alle Docs reflektieren den Ist-Zustand nach Session 11a, nicht
den theoretischen Endzustand. `runbook.md` Szenario 2 ist exakt das
4-Optionen-Diagnose-Pattern für „Backend offline / Failed to fetch",
das die 11a-Pipeline durchgemacht hat. `deployment.md` dokumentiert
Bind-Mount-Inode-Falle und deploy-User-NOPASSWD-Scope explizit.

✅ `docs/backlog.md` ist die low-ceremony Vorstufe für GitHub-Issues:
Operations-Punkte (Backend-Metriken-Scrape-Lücke, mon-node-exporter,
Caddy-Admin, Tracing-Sampler, Tempo-S3, automatische Postgres-Backups,
Container-Resource-Exporters), Tooling-Punkte (Mailpit-Migration,
i18n-Library-Wahl, default_versions-Tagging-Strategie),
Sicherheits-Punkte (DMARC `p=reject`, HSTS `includeSubDomains`,
cosign-verify im Deploy).

✅ Damit ist die initiale Setup-Phase formal abgeschlossen.
Nächster Schritt ist Feature-Arbeit (Open-Meteo, Locations-Suche,
Auth, Karten), nicht mehr Infrastruktur. Eigene Session-Struktur
für Feature-Sessions wird aufgebaut, sobald wir dort ankommen.

---

## Häufige Fallen (Lessons Learned aus Setup-Phase)

- **Caddy-WebSocket-Upgrade** muss explizit erlaubt sein für Vite-HMR
- **sqlc** benötigt evtl. explizite Type-Overrides für PostGIS-Typen
- **prettier-Plugin-Resolution** in Monorepo: lokale Hooks via pnpm-Filter, nicht
  pre-commit-mirror mit additional_dependencies
- **golangci-lint v1 vs. v2** — Config-Schemas inkompatibel; v2-Config braucht
  v2-Binary, sonst "can't load config" mit verwirrenden Versions-Hinweisen
- **`go mod tidy` ergänzt `toolchain`-Zeile** wenn Dependencies höhere Go-Version
  verlangen — alle Quellen gleichziehen
- **Fine-grained PATs unterstützen ghcr.io NICHT** — die Permission „Packages"
  existiert in der Fine-grained-Welt nicht (seit Jahren in „Beta", siehe
  GitHub-Community-Discussions). Für ghcr-Pull/Push klassischen PAT mit
  `read:packages` bzw. `write:packages` Scope nutzen. Username im docker-login
  ist der persönliche User (`hwrichter`), nicht der Org-Name (`relations4u`).
- **Branch-Protection erst NACH Session 8** aktivieren, sonst eigene erste
  Commits blockiert
- **mise-Hook in zsh aktivieren**: `eval "$(mise activate zsh)"` in `~/.zshrc`
- **YAML in zsh via heredoc kann BOM/CRLF erzeugen** — printf oder vi nutzen
- **prettier auch im pre-commit auf YAML** — manuell vorab laufen lassen,
  dann staged + unstaged trennen, sonst Stash-Konflikte
- **Maintainer-Mail prüfen** vor erstem Commit — `deine@email.tld` als
  Platzhalter aus Anleitungen wandert sonst in Git-Config
- **GitHub trennt Authentication- und Signing-Key** — beide explizit registrieren
- **Frontend `.env` mit Public-Defaults darf committed werden** — sonst scheitert
  CI Type-check, weil `$env/static/public` keine Member exportiert
- **Apex-CNAME ist per RFC 1034 verboten** — Lösung: Cloudflare CNAME-Flattening
  oder pragmatisch Apex direkt als A-Record + Subdomains als CNAMEs darauf
- **DynDNS auf einen Anker-Hostname** — alle anderen Hostnames als CNAMEs
  darauf zeigen lassen, dann ein einziger Update-Punkt bei IP-Wechsel.
  Cross-Zone-CNAMEs (Cloudflare → Joker) funktionieren standardkonform.
- **Self-Hosting bedeutet keine echte Production** — keine SLA, kein Failover,
  kein DDoS-Schutz. Nutzer transparent informieren via Banner auf
  `research.worldweathernews.com`.
- **Mail-Records auch ohne eigenen Mailserver setzen** (SPF `-all`, DMARC
  `p=reject`, CAA für Let's Encrypt) — schützt Domain vor Phishing-Missbrauch
  und Cert-Mis-Issuance. Bei aktivem Mailserver (z. B. Proton): SPF mit
  `include:` für Provider, DMARC `p=quarantine` als Start, später `p=reject`.
- **Cloudflare Quick-Scan importiert vorhandene Records mit** — bei
  Zone-Migration prüfen, ob ProtonMail-, alte SPF-, DKIM-, MX-Records korrekt
  übernommen wurden. `dig MX domain.com` und `dig protonmail._domainkey.domain.com`
  als Sanity-Check nach Migration.
- **Bei aktiver Mailbox: Mail-Authentication-Test** an
  `check-auth@verifier.port25.com` schicken nach DNS-Änderungen — bekommst
  detaillierten Report ob SPF/DKIM/DMARC korrekt funktionieren
- **VLAN-Tagging port-basiert über Switches**, nicht über Proxmox-VM-Network-Profil:
  wenn Switch-Port als Access-Port für ein VLAN konfiguriert ist, **darf kein**
  VLAN-Tag in der Proxmox-VM-Network-Config gesetzt werden (sonst Doppel-Tagging,
  Frames werden vom Switch verworfen). Symptom: VM bekommt keine IP via DHCP.
- **MAC-Adresse bei VM-Erstellung manuell setzen**, nicht Proxmox-Auto-MAC.
  Konvention: letzte Bytes spiegeln IP wider (z. B. `BC:24:11:00:21:21` für
  IP `.21`). Erlaubt OPNsense-DHCP-Reservation **vor** VM-Erstellung anzulegen,
  greift dann beim ersten Boot direkt.
- **sops nicht in Default-Ubuntu-Repos** — direkt via GitHub-Release als .deb
  installieren (`https://github.com/getsops/sops/releases`), nicht via apt
  oder snap. apt liefert stale-Versionen, snap macht Confinement-Probleme.
- **Block-Logs auf falsche IPs ernst nehmen, dann einsortieren** — wenn die
  Firewall regelmäßig blockiert, immer die Quell-MAC identifizieren (DHCP-Leases
  oder ARP-Table). Oft sind es vergessene Geräte mit veralteten Configs (z. B.
  Webcam mit alter NAS-FTP-Backup-Adresse). Falsch konfigurierte IoT-Geräte
  gehören idealerweise in das IoT-VLAN (`opt3`), nicht ins manager-VLAN.
- **Caddy `network_mode: host`** auf dem Production-Host: Client-IPs landen
  unverfälscht in Logs, Caddy bindet direkt an :80/:443 des Hosts. Trade-off:
  keine Container-interne Service-Discovery — `reverse_proxy` muss auf
  `127.0.0.1:PORT` zeigen, App-Container müssen ihre Ports auf Loopback binden.
- **HSTS `includeSubDomains` mit Bedacht aktivieren** — schaltet ALLE
  Subdomains der Apex auf HTTPS-only, auch zukünftige. Erst aktivieren, wenn
  alle geplanten Subdomains zuverlässig TLS haben. Apex und Leaf-Subdomains
  einzeln HSTS'en ist sicherer für Setups, in denen interne Subdomains noch
  HTTP-only sein dürfen.
- **`sudo` über plain SSH braucht Terminal** — Deploy-Skripte laufen mit
  `BatchMode=yes` und ohne TTY. `ssh host sudo cmd` schlägt mit
  „sudo: a terminal is required" fehl. Lösung: einmalig manuell mit
  `ssh -t host sudo …` vorbereiten und Skript defensiv testen
  (`test -d` / `test -w`), ob die Vorbereitung gelaufen ist.
- **`terraform import` vor `terraform apply`** wenn Ressourcen schon manuell
  existieren — sonst legt Terraform sie als „neu" an und kollidiert mit
  vorhandenen IDs. Workflow: tfvars befüllen → init →
  `terraform import 'module.<name>.proxmox_virtual_environment_vm.this' <node>/qemu/<vmid>`
  → `terraform plan` muss „No changes" zeigen, sonst main.tf an den
  Ist-Zustand anpassen, NICHT umgekehrt.
- **`required_providers` muss pro Modul/Environment** stehen — ein
  zentrales `versions.tf` im Terraform-Wurzelverzeichnis wird nicht von
  Sub-Environments erkannt. Sonst landet man bei
  `hashicorp/<provider>` (existiert nicht) statt z. B. `bpg/proxmox`.
- **ansible-lint `var-naming[no-role-prefix]` ist für first-party Rollen
  überzogen** — Variablen wie `deploy_user`, `app_dir`, `image_namespace`
  werden bewusst Cross-Role aus `group_vars/all.yml` genutzt. In
  `infra/ansible/.ansible-lint` skip-listen, dafür im Repo dokumentieren.
- **`partial-become[task]` aus ansible-lint** — wenn `become_user: X`
  auf einem Task steht, muss `become: true` daneben stehen, auch wenn
  der Play schon `become: true` hat. Lint will es explizit pro Task.
- **Pre-commit prettier auto-fix + unstaged-Diff = Stash-Konflikt** —
  wenn prettier ein File modifiziert, das gleichzeitig unstaged geändert
  ist, rollt der Hook seine Fixes zurück. Lösung: vor dem Commit
  Working-Tree clean machen oder die prettier-formatierte Version
  einfach nachstagen und neu committen.
- **`bpg/proxmox` braucht eine Cloud-init-fähige Template-VM** im
  Cluster, bevor `terraform apply` etwas Sinnvolles tut. Workflow:
  Cloud-Image runterladen, daraus VM erzeugen, `qm set --ide2 cloudinit`,
  `qm template`, VMID in `proxmox_template_id` eintragen.
- **`apps/frontend/pnpm-lock.yaml` ist ein standalone Lockfile** —
  der Docker-Build-Context unter `apps/frontend/` nutzt diesen, **nicht**
  den Workspace-Root-Lockfile. pnpm regeneriert den standalone-Lockfile
  von sich aus nicht, weil der Workspace-Root ihn überschattet. Symptom:
  Frontend-Container-Build schlägt mit Lockfile-Drift fehl, obwohl der
  Workspace-Install grün war. Fix-Workflow für Dependency-Updates:
  `pnpm-workspace.yaml` temporär verstecken, `pnpm install` (regeneriert
  standalone), Workspace zurück, `pnpm install` erneut (regeneriert
  Workspace), beide Lockfiles committen. Pattern taucht bei jedem
  Frontend-Dependency-Update wieder auf — siehe `32f571d` (Folge zu #46)
  und PR #62 (Dependabot-Triage 11. Mai).
- **`gh pr merge` schlägt fehl bei Workflow-File-Änderungen ohne
  `workflow`-Scope** auf dem PAT — GraphQL-Error
  `mergePullRequest: workflow scope missing`. Drei Optionen:
  (1) Sofort-Retry — klappt oft beim zweiten Versuch (Race oder
  lazy-loaded Scope-Check). (2) Sauber:
  `gh auth refresh -s workflow` und neu authentifizieren.
  (3) Notlösung: Web-UI-Merge bypasst das CLI-Check. Beobachtet
  bei PR #63 (Tailwind v4) und bei der CF-Worker-Cleanup-Tranche
  am 11. Mai.
- **Tailwind v4: `@theme`-Block in CSS, keine `tailwind.config.js`** —
  Tailwind v4 hat keine JS-Config-Datei mehr. Konfiguration läuft
  über `@theme`-Block in `src/app.css`. shadcn-svelte-HSL-Variablen
  (`--background`, `--primary` …) werden via `--color-*` in den v4-
  Namespace gebrückt, damit Utility-Klassen wie `bg-primary` weiter
  auflösen. **Vite-Integration:** `@tailwindcss/vite` ersetzt das
  PostCSS-Plugin — PostCSS-Pfad kollidiert mit Vites postcss-import,
  das `@import "tailwindcss"` als relative Datei aufzulösen versucht.
  **autoprefixer entfällt** — Lightning CSS macht Vendor-Prefixing
  intern. `.dark`-Strategie über `@custom-variant dark` beibehalten.
  Referenz: PR #63 (11. Mai).
- **Self-hosting-Prinzip für neue Services**: neue Compute-Services
  landen erst-mal als Container im App-Compose-Stack auf wwn-prod,
  nicht als Cloudflare-Worker. Begründung: DNS läuft bereits über
  Cloudflare — jeder zusätzliche kritische Pfad bei demselben Anbieter
  hebt das Migrations-Risiko unnötig. CF-Worker-zu-Go-Container-Port
  ist machbar (~170 LOC bei cms-auth), fügt sich in bestehendes
  Compose-/Caddy-/Monitoring-Setup natürlich ein. Ausnahme: reine
  **Edge-/Cache-Schichten** vor self-hosted Origins sind okay
  (kein kritischer Pfad). Volle Regel als A.19 in
  `sessions/feature1/feature-decisions.md` dokumentiert.
- docker exec ohne -u 0
- OpenAPI 3.1 ohne nullable
- sqlc-Schema via Pre-Processing
- DB-Migrations als Pflicht-Deploy-Step

---

## Changelog dieser Datei

Diese Datei wächst mit dem Projekt. Wenn du etwas Strukturelles lernst, das hier
fehlt: vorschlagen, mit Begründung. Maintainer entscheidet, ob es rein kommt.

- **2026-05-03** — Initiale Version
- **2026-05-05** — Reflexion der Setup-Phase: Versions-Pinning (Go 1.25,
  golangci-lint v2.12.1), Maintainer-Identität (Heinz W. Richter,
  hwr@relations4u.de), Dev-Setup auf Proxmox-VM, GitHub-Org `relations4u`
  bestätigt, Pre-commit-Strategie (local hooks), Häufige Fallen aus
  Sessions 1–8, beantwortete Entscheidungen, aktueller Session-Status,
  vi statt nano, Branch-Protection-Verhalten dokumentiert
- **2026-05-05 (später)** — Hosting- und DNS-Architektur entschieden:
  Self-Hosting auf eigenem Proxmox-Host (Ryzen 7, 32 GB) für Forschungs-Phase
  statt Hetzner Cloud. Hetzner-Migration als ADR für später. Domain bleibt
  bei Joker.com, DNS migriert zu Cloudflare Free-Plan (DNS-only). DynDNS-Anker
  `gate.hw7.eu` als einziger Update-Punkt, alle WWN-Hostnames per CNAME
  (apex via Cloudflare CNAME-Flattening). Subdomain-Schema: `research.` für
  Forschungs-Phase, `api.research.` für deren Backend. Mail-/CAA-Records
  als Domain-Hygiene auch ohne eigenen Mailserver.
- **2026-05-05 (Mail-Korrektur)** — Mail-Strategie korrigiert: Domain nutzt
  aktiv ProtonMail als Mailbox-Provider (war im ersten Hosting-Entwurf
  übersehen). DNS-Records dokumentiert: MX, 3x DKIM-CNAME, SPF, DMARC,
  Proton-Verification. DMARC verschärft von minimalem `p=quarantine` auf
  vollen Record mit `rua`, `sp`, `aspf=s`, `adkim=s`. Transactional-Mail
  separat für später vorgemerkt.
- **2026-05-05 (Setup-Phase)** — Sessions 8 und 9 abgeschlossen, Session 10
  begonnen mit Aufteilung in 10a (wwn-prod + Caddy) und 10b (wwn-mon +
  Observability-Stack). Entscheidung: wwn-mon als separate VM (4 GB) statt
  Co-Location mit wwn-prod — bessere Telemetrie-Isolation bei Crashes,
  saubere Trennung der I/O-Profile. Setup-Anleitungen `vm-prod-setup.md`
  und `vm-mon-setup.md` erstellt. Caddy-Konfiguration mit HSTS-Header
  (max-age=31536000; includeSubDomains, ohne preload bis Plattform stabil),
  HTTP/3 via UDP/443, Security-Headers (X-Frame-Options, Referrer-Policy,
  Permissions-Policy). HTTP-01-Challenge für Let's Encrypt (Port 80 offen).

  **Anmerkung 6. Mai:** `includeSubDomains` wurde am 6. Mai gestrichen
  (siehe „Beantwortete Entscheidungen ab 2026-05-06"). Begründung:
  zukünftige interne Subdomains bekommen ggf. lange kein TLS, die
  HSTS-Vererbung würde sie aussperren.

- **2026-05-05 (Basis-Setup beider VMs)** — wwn-prod und wwn-mon mit
  Ubuntu 24.04 + Hardening + Docker + sops/age installiert. DHCP-Reservations
  greifen (verifiziert: wwn-prod auf 10.100.100.21 mit MAC bc:24:11:00:21:21).
  Korrektur: VLAN-Tagging muss in Proxmox-VM-Config **leer** bleiben — Switch
  macht VLAN-Zuordnung port-basiert. Korrektur: sops nicht in Default-apt,
  via GitHub-Release als .deb installiert. Anomalie geklärt: 10.100.100.213
  war falsch konfigurierte Webcam mit veralteter NAS-FTP-Backup-Adresse zu
  nicht-existenter 10.1.1.169 — Block bleibt korrekt, Webcam-Konfig fixen
  (Aufgabe für später, idealerweise Cam ins IoT-VLAN verschieben). Phase
  10b verschoben auf Caddy-Online und 10c auf Observability — Sessionplan
  entsprechend angepasst, Phase 11 (Ansible) bleibt ausstehend bis Caddy
  manuell läuft, dann als Referenz für Playbook-Verifikation.
- **2026-05-06 (Caddy live + Session-11-Skelett)** — Caddy auf wwn-prod
  als eigenständiger Stack (`/srv/wwn/caddy`, `network_mode: host`)
  deployed; vier Let's-Encrypt-Zertifikate ausgestellt (Apex, www,
  research, api.research). HSTS bewusst **ohne** `includeSubDomains`
  (Abweichung vom 5.-Mai-Plan, dokumentiert oben in Beantwortete
  Entscheidungen). Snapshot `caddy-online` gesetzt. Session 11 als
  Code-Skelett geliefert (Feature-Branch + PR): `.sops.yaml` mit
  age-Pubkey, sechs verschlüsselte Demo-Secrets unter
  `infra/secrets/production/`, Pre-commit-Hook
  `forbid-unencrypted-secrets`, vier Ansible-Rollen
  (`common`/`docker`/`app`/`monitoring-agent`), drei Playbooks
  (`site`/`deploy`/`rollback`), Terraform-Skelett mit `bpg/proxmox`
  aktiv und `hetznercloud/hcloud` als Migrations-Stub, `scripts/deploy.sh`-
  Wrapper mit production-Confirmation. Tooling-Pins ergänzt: terraform
  1.15, sops 3.12, ansible-core 2.20, ansible-lint 26.4. Validierung
  lokal grün (ansible-lint, --syntax-check, terraform fmt/validate,
  pre-commit). Tatsächliches Server-Apply bleibt Maintainer-Hausaufgabe
  — `terraform import` der bestehenden VMs vor dem ersten `apply`
  dokumentiert.
- **2026-05-06 (Session 11a — Komplettes Deployment live)** — wwn-prod
  und wwn-mon vollständig via Ansible bootstrapped, App-Stack v0.0.1-rc4
  läuft auf wwn-prod (backend/frontend/pyworkers alle healthy), zentraler
  Observability-Stack auf wwn-mon (Prometheus/Loki/Tempo/Grafana mit drei
  provisionierten Dashboards). Caddy-Cutover von Stub-`respond` auf
  `reverse_proxy` 127.0.0.1:{3000,8080} mit unveränderten Cert-`notBefore`-
  Dates. Caddy-Cert-Volume zu Bind-Mount `/srv/wwn/caddy/data` migriert.
  Code-Fixes auf dem Weg dahin (alle gemerged): release.yml gibt jetzt
  `PUBLIC_API_BASE_URL` als build-arg an den Frontend-Build (sonst landet
  der Dockerfile-Default `http://api.localhost` im JS-Bundle), Backend
  bekommt `WWN_HTTP_CORSORIGINS` für Apex/www/research-Origins, Caddy lässt
  OPTIONS-Preflights zur Backend-chi-cors durch, monitoring-stack-Rolle
  erstellt `grafana/`+`grafana/provisioning/` mit Mode 0755 explizit (uid 472
  brauchte traversal), monitoring-agent-Play limitiert auf `app`-Hosts (mons
  Stack-Promtail würde sonst kollidieren), neue UFW-`monitoring_scrape_ports`-
  Liste in der common-Rolle. Single-File-Bind-Mount-Inode-Falle bei Caddy
  und monitoring-stack: `restart` nach Config-Update nötig (`deploy-caddy.sh`
  und Ansible-Handler). Frontend-Healthcheck wechselt von `localhost` auf
  `127.0.0.1` (busybox-wget resolved IPv6, Server bindet IPv4). PR-Reihe
  #25–#32, Tag v0.0.1-rc4.
- **2026-05-06 (Session 12 — Doku-Finalisierung, Setup-Phase abgeschlossen)** —
  Vier Doku-Tranchen gemerged (#33..#36): README rewrite + CONTRIBUTING,
  architecture.md mit Mermaid-Diagramm, development.md mit How-Tos,
  deployment.md (Bootstrap, Folge-Deploys, Rollback) und runbook.md mit
  10 Szenarien (Szenario 2 ist exakt das „Backend offline / Failed to
  fetch"-Diagnose-Pattern aus der 11a-Pipeline), ADRs 0002–0005 im
  MADR-Format, AGPL-3.0 LICENSE, docs/backlog.md als low-ceremony
  Folge-Tracker, In-Context-TODO-Triage. Damit alle step12.md-Erfolgs-
  Kriterien ✅ und die initiale Setup-Phase (Sessions 1–12) formal
  abgeschlossen.
- **2026-05-06 (Pflege-Pass nach Session 12)** — sechs Drift-Korrekturen
  in einer Tranche eingebaut: (1) Pflicht-Pins-Tabelle um Terraform 1.15,
  sops 3.12, ansible-core 2.20, ansible-lint 26.4 ergänzt — Session 11
  hatte sie im Changelog erwähnt, aber nicht in der zentralen Tabelle
  geführt. (2) Neue Sektion „App-Release-Pinning" als Abgrenzung zum
  Toolchain-Pinning, dokumentiert v0.0.2 als aktuellen Live-Stand und
  den `default_versions: 0.0.0`-Fail-fast-Marker. (3) „Wo finde ich
  was" um `docs/backlog.md`, ADR-Enumeration (0001–0005) und `LICENSE`
  erweitert. (4) HSTS-Strategie disambiguiert: Setup-Phase-Changelog-
  Eintrag um „Anmerkung 6. Mai" ergänzt, der den 5.-Mai-Plan
  (`includeSubDomains`) explizit als überholt markiert. (5) Status-
  Header-Datum auf 6. Mai korrigiert (war noch auf 5. Mai). (6) Offene-
  Fragen-Sektion bereinigt: Lizenz als AGPL-3.0 entfernt (entschieden,
  in „Beantwortete Entscheidungen" verschoben), Sentry/GlitchTip-Status
  ehrlich als „in Feature-Phase, Telemetrie aktuell via Loki/Tempo"
  formuliert, Backup-Ziel verweist auf `docs/backlog.md` statt auf
  abgeschlossene Session 12. Plus Korrektur in „Häufige Fallen":
  alter Eintrag „Fine-grained PATs statt Classic für ghcr.io — mehr
  Kontrolle" war faktisch falsch (Fine-grained PATs unterstützen
  Packages-Permission seit Jahren nicht), ersetzt durch korrekte
  Anweisung „Classic PAT mit `read:packages`/`write:packages` Scope".
- **2026-05-11 (CMS-OAuth self-hosted)** — Cloudflare-Worker für die
  Sveltia-OAuth ersetzt durch einen self-hosted Go-Service unter
  `apps/cms-auth/` (Chi-Router, Distroless-Image, ~170 LOC Logik 1:1
  vom CF-Worker portiert). Service läuft im App-Compose-Stack auf
  wwn-prod (Bind 127.0.0.1:8090), Caddy proxied unter neuem Host
  `cms-auth.worldweathernews.com`. SOPS-Secret
  `infra/secrets/production/cms-auth.env` ergänzt, Release-Pipeline
  baut viertes Image `wwn-cms-auth`, eigener CI-Workflow
  `ci-cms-auth.yml`. Subdomain-Tabelle und DNS-Auflösungs-Kette
  erweitert. A.4 in `sessions/feature1/feature-decisions.md` mit
  Verweis auf die neue Entscheidung superseded. Maintainer-Hausaufgaben:
  DNS-CNAME setzen, GitHub-OAuth-App-Callback-URL umstellen, GitHub-
  Client-ID/Secret in die neue env-Datei einsetzen, Tag pushen +
  deployen, Caddy reloaden für Cert.
- **2026-05-11 (CF-Worker-Cleanup)** — Nach erfolgreichem Cutover und
  Live-Smoketest des Sveltia-Logins wurde der CF-Worker im Account
  `hwr-06e` via `wrangler delete` abgebaut. Repo-Verzeichnis
  `infra/cloudflare-worker-cms-auth/` entfernt (Folge-PR). Verweise
  in `CLAUDE.md`, `docs/cms.md`, `apps/frontend/static/admin/config.yml`
  und `apps/cms-auth/README.md` auf den vormaligen Pfad bereinigt.
  Damit ist die Sveltia-OAuth komplett self-hosted, der einzige
  verbleibende CF-Touchpoint ist DNS.
- **2026-05-11 (Pflege-Runde nach v0.0.4)** — vier Lessons aus der
  Feature-Phase als generalisierbare Regeln in „Häufige Fallen"
  nachgezogen, damit sie nicht nur in einzelnen PR-Beschreibungen
  versteckt sind:
  (1) **Standalone `apps/frontend/pnpm-lock.yaml`** — Docker-Build
  nutzt diesen, nicht den Workspace-Root. Re-Sync-Workflow für jedes
  Frontend-Dependency-Update dokumentiert (Folge aus #32f571d und PR #62).
  (2) **`gh pr merge` ohne `workflow`-Scope** — drei Optionen
  (Sofort-Retry, `gh auth refresh -s workflow`, Web-UI-Merge),
  beobachtet bei PR #63 und CF-Worker-Cleanup.
  (3) **Tailwind v4 mit `@theme`-Block** statt `tailwind.config.js`,
  `@tailwindcss/vite` als Vite-Plugin, autoprefixer entfällt,
  shadcn-svelte HSL via `--color-*` gebrückt (Referenz PR #63).
  (4) **Self-hosting-Prinzip für neue Services** als allgemeine
  Leitlinie — nicht nur eine Einzelfall-Entscheidung für cms-auth.
  Konsequenz: Iteration 1.3b (Image-Pipeline) wird als Container
  geplant, nicht als CF-Worker. Volle Regel als A.19 in
  `sessions/feature1/feature-decisions.md`.
  **2026-05-12 (Konsolidierung post-Iteration-2.1)**
