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

### Aktuelle Pflicht-Pins (Stand Mai 2026)

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

| Anliegen                       | Ort                                                                     |
| ------------------------------ | ----------------------------------------------------------------------- |
| Architektur-Diagramme          | `docs/architecture.md`                                                  |
| Was tun wenn X kaputt?         | `docs/runbook.md`                                                       |
| Wie deploye ich?               | `docs/deployment.md`                                                    |
| Wie entwickle ich Feature X?   | `docs/development.md`                                                   |
| Warum diese Tech-Entscheidung? | `docs/adr/`                                                             |
| Config-Reference               | `docs/config-reference.md` (generiert)                                  |
| Secrets-Workflow               | `docs/secrets.md`                                                       |
| Service-spezifische Doku       | `apps/<service>/README.md`                                              |
| Setup-Sessions (12 Stück)      | `sessions/step01.md` – `step12.md`                                      |
| Status der Sessions            | `sessions/STATUS.md`                                                    |
| VM-Setup-Anleitung             | `vm-setup.md`                                                           |
| Pre-Session-Checkliste         | `pre-session-checklist.md`                                              |
| HTTP-Handler                   | `apps/backend/internal/http/handler/`                                   |
| DB-Migrationen                 | `infra/migrations/`                                                     |
| Worker-Jobs                    | `apps/pyworkers/pyworkers/jobs/`                                        |
| Frontend-Routes                | `apps/frontend/src/routes/`                                             |
| Frontend-Default-Env           | `apps/frontend/.env` (committed!)                                       |
| OpenAPI-Schema                 | `packages/api-schema/openapi.yaml`                                      |
| OpenAPI-Lint-Config            | `packages/api-schema/redocly.yaml`                                      |
| Generierte Files               | markiert mit Header `// Code generated by ...`                          |
| Secrets                        | `infra/secrets/<env>/*.sops.*`                                          |
| Compose-Configs                | `infra/compose/`                                                        |
| Lychee-Ignores                 | `.lycheeignore`                                                         |
| Hosting-Hardware               | `docs/architecture.md` (Hardware-Übersicht)                             |
| DNS-Records                    | Cloudflare-Dashboard für `worldweathernews.com`, Joker.com für `hw7.eu` |
| DynDNS-Konfiguration           | Firewall-Hardware vor dem Proxmox-Host                                  |
| Production-Caddyfile           | `infra/caddy/Caddyfile.prod`                                            |
| Public Status-Page             | https://status.worldweathernews.com (ab Session 10)                     |

---

## Externe Datenquellen (geplant, für spätere Sessions)

Diese werden im Backend integriert. Aktuell nur als Referenz, damit du den Kontext kennst.

- **Open-Meteo** (open-meteo.com) — Erste primäre Quelle, EU-basiert, ohne API-Key
- **DWD** (Deutscher Wetterdienst) — OpenData, MOSMIX, ICON-Modelle
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
                                                       ├─ wwn-dev    (8 GB) — Entwicklung
                                                       ├─ wwn-prod   (8 GB) — Forschungs-Produktion
                                                       └─ wwn-mon    (4 GB) — Monitoring (optional)

                                                       Reserve: ~12 GB
```

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

| Hostname                            | Zweck                            | Phase                                              |
| ----------------------------------- | -------------------------------- | -------------------------------------------------- |
| `worldweathernews.com`              | Apex — Landing-Page              | jetzt: simple Landing, später: Production-Frontend |
| `www.worldweathernews.com`          | 301-Redirect auf Apex            | dauerhaft                                          |
| `research.worldweathernews.com`     | Forschungs-Frontend              | Forschungs-Phase (jetzt)                           |
| `api.research.worldweathernews.com` | Forschungs-Backend-API           | Forschungs-Phase (jetzt)                           |
| `api.worldweathernews.com`          | Production-Backend-API           | später (Production-Phase)                          |
| `status.worldweathernews.com`       | Public Uptime-Page (Uptime Kuma) | ab Session 10                                      |

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
```

### Cloudflare-Konfiguration

- **Plan:** Free
- **Proxy-Modus:** AUS (DNS-only, graue Wolke). Aktivierung später bei Bedarf
  pro Subdomain möglich.
- **TTL:** Auto (300 s im DNS-only-Modus)
- **Apex-Strategie:** CNAME-Flattening (Cloudflare-Spezialfeature, RFC-konform
  via Antwort-Time-Resolution)

### Mail-Strategie

In der Forschungs-Phase **kein eigener Mailserver**. Domain-Mails laufen über
externen Provider, sobald nötig (Feature-Phase).

DNS-Records, die SOFORT Spam-Schutz für die Domain bieten — auch wenn keine
Mails versendet werden:

```
worldweathernews.com.       TXT    "v=spf1 -all"
_dmarc.worldweathernews.com. TXT   "v=DMARC1; p=reject; rua=mailto:hwr@relations4u.de"
```

CAA-Records einschränken die zulässigen Zertifikat-Aussteller auf Let's Encrypt:

```
worldweathernews.com.  CAA  0 issue     "letsencrypt.org"
worldweathernews.com.  CAA  0 issuewild "letsencrypt.org"
worldweathernews.com.  CAA  0 iodef     "mailto:hwr@relations4u.de"
```

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

- [ ] Email-Provider: Postmark, Brevo, eigener SMTP? — Erst in Feature-Phase
      relevant (User-Registrierung, Alerts).
- [ ] i18n-Library: svelte-i18n vs. Paraglide vs. Inlang — Entscheidung
      in der ersten Feature-Session.
- [ ] Backup-Ziel: S3-kompatibel (Hetzner Storage Box?), eigenes NAS, BorgBase?
      Entscheidung in Session 12 (Runbook). Empfehlung: Proxmox Backup Server
      als eigene VM auf demselben Host, plus externer Off-Site-Sync.
- [ ] Lizenz: vermutlich proprietär, evtl. AGPL für Backend, MIT für Schemas.
      Entscheidung am Ende von Session 12.
- [ ] Sentry vs. GlitchTip — Entscheidung in Session 10 (Observability).

Wenn eine dieser Fragen für deine Aufgabe relevant wird: **fragen**, nicht annehmen.

---

## Session-Tracking

Die Sessions zur Initial-Einrichtung sind dokumentiert in `sessions/step01.md` bis
`sessions/step12.md`. Wenn du in einer Session arbeitest, halte dich an die dort
definierten Aufgaben. Bei Abweichung: zurück zur Datei, fragen.

Stand der Sessions wird in `sessions/STATUS.md` gepflegt — am Ende jeder Session
ein kurzer Eintrag.

### Aktueller Status (5. Mai 2026)

| Phase | Session                         | Status                                               |
| ----- | ------------------------------- | ---------------------------------------------------- |
| A     | 1 — Repo-Skelett, mise          | ✅ Abgeschlossen                                     |
| A     | 2 — Pre-commit, Makefile        | ✅ Abgeschlossen                                     |
| B     | 3 — Compose-Stack               | ✅ Abgeschlossen                                     |
| B     | 4 — Go-Backend                  | ✅ Abgeschlossen                                     |
| B     | 5 — SvelteKit-Frontend          | ✅ Abgeschlossen                                     |
| B     | 6 — Python-Workers              | ✅ Abgeschlossen                                     |
| C     | 7 — OpenAPI + Codegen           | ✅ Abgeschlossen                                     |
| C     | 8 — CI-Workflows                | 🔄 In Arbeit (Backend-CI grün, PR offen mit 5 Fixes) |
| C     | 9 — Release + ghcr.io           | ⏳ Ausstehend                                        |
| D     | 10 — Observability              | ⏳ Ausstehend                                        |
| D     | 11 — Ansible + SOPS + Terraform | ⏳ Ausstehend                                        |
| D     | 12 — Dokumentation              | ⏳ Ausstehend                                        |

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
- **Fine-grained PATs** statt Classic für ghcr.io — mehr Kontrolle
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
  und Cert-Mis-Issuance

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
