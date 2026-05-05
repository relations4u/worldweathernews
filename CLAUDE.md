# worldweathernews.com

Wetter- und Klima-Plattform mit Community-Features. Self-hosted Monorepo.

**Maintainer**: Diplom-Meteorologe, IT-Architekt, betreibt eigene Hosting-Infrastruktur.

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

## Tech-Stack (verbindlich)

| Schicht            | Technologie                         | Begründung                                                            |
| ------------------ | ----------------------------------- | --------------------------------------------------------------------- |
| Backend-API        | Go 1.23                             | Performance, Concurrency für API-Aggregation, Self-Hosting-freundlich |
| HTTP-Framework     | Chi                                 | Idiomatisch, nah an net/http, kein Magic                              |
| DB-Access          | sqlc + pgx/v5                       | Typsicher ohne ORM-Overhead                                           |
| Config (Go)        | Viper                               | ENV + Datei + Defaults                                                |
| Workers            | Python 3.12                         | Reife GRIB/NetCDF-Bibliotheken (xarray, cfgrib)                       |
| Python-Pkg-Manager | uv                                  | Schneller als pip/poetry, deterministisch                             |
| Config (Python)    | pydantic-settings                   | Typsicher, Validation beim Start                                      |
| Frontend           | SvelteKit + TypeScript              | Schlanker als Next.js, exzellente DX                                  |
| Styling            | Tailwind CSS + shadcn-svelte        | Schnelles, konsistentes UI                                            |
| Frontend-Adapter   | @sveltejs/adapter-node              | Self-Hosting im Container                                             |
| Datenbank          | PostgreSQL 16                       | Mit PostGIS (Geo) und TimescaleDB (Zeitreihen)                        |
| Cache/Queue        | Redis 7                             | Standard, vielseitig                                                  |
| Reverse Proxy      | Caddy                               | Auto-SSL, einfache Config                                             |
| Container-Runtime  | Docker + Docker Compose             | Wachstumspfad zu K3s offen, aber nicht jetzt                          |
| CI                 | GitHub Actions                      | Schnell, kostenlos, sehr verbreitet                                   |
| Container-Registry | ghcr.io                             | Integriert mit GitHub                                                 |
| Deployment         | Ansible                             | Klassisch, durchschaubar, idempotent                                  |
| Infrastruktur      | Terraform                           | Provider-Abstraktion, deklarativ                                      |
| Secrets            | SOPS + age                          | In Git verschlüsselt, kein externer Vault nötig                       |
| Monitoring         | Prometheus + Grafana + Loki + Tempo | Self-hostbar, ein UI                                                  |
| Errors             | Sentry                              | Quasi-Standard                                                        |
| Uptime             | Uptime Kuma                         | Simpel, schön, self-hosted                                            |
| Migrations         | goose                               | Sprach-agnostisch (Go-Tool, aber File-basiert)                        |

---

## Repo-Struktur (Monorepo)

```
worldweathernews/
├── apps/
│   ├── backend/        # Go API
│   ├── frontend/       # SvelteKit
│   └── pyworkers/      # Python Worker-Service
├── packages/
│   ├── api-schema/     # OpenAPI-Spec (Single Source of Truth)
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
├── .github/workflows/  # CI/CD
├── .mise.toml          # Tool-Versionen
├── Makefile            # Top-Level-Tasks
└── compose.yml         # → infra/compose/compose.dev.yml
```

---

## Conventions

### Git

- **Trunk-based Development**: kurze Feature-Branches, Merge in `main`
- **Conventional Commits**: `feat:`, `fix:`, `chore:`, `docs:`, `refactor:`, `test:`, `perf:`, `ci:`
- Scope optional: `feat(backend):`, `fix(frontend):`
- **Keine Force-Pushes auf `main`**
- **`main` ist immer deploybar**
- **Semver** für Releases, Tags `v*`

### Code-Style

- Go: `gofmt`, `goimports`, `golangci-lint` müssen grün sein
- Python: `ruff` (Lint + Format), `mypy` strict
- TypeScript: `eslint`, `prettier`, `svelte-check`
- Strukturiertes Logging in JSON: `slog` (Go), `structlog` (Python), `pino` (Node)
- Trace-IDs in jedem Log-Eintrag

### Naming-Konventionen

- ENV-Variablen: `WWN_*` (Backend Go), `WWN_PY_*` (Python), `PUBLIC_*` (Frontend client-side)
- Container-Images: `ghcr.io/relations4u/wwn-backend`, `wwn-frontend`, `wwn-pyworkers`
- Verzeichnisse: kebab-case
- Go-Module: `github.com/relations4u/worldweathernews/apps/backend`

### API-Design

- **OpenAPI ist Single Source of Truth** in `packages/api-schema/openapi.yaml`
- Server-Stubs (Go) und Client-Types (TS) werden generiert
- API-Versionierung im Pfad: `/api/v1/...`
- Errors als Problem Details (RFC 7807)
- Pagination: cursor-basiert, nicht offset

### Datenbank

- Migrations in `infra/migrations/` (sprachunabhängig)
- Jede Migration muss ein `down`-Script haben (Reversibilität)
- Keine direkten Schema-Änderungen ohne Migration
- Foreign Keys explizit, ON DELETE CASCADE bewusst

### Tests

- Go: Tests neben Code (`foo.go` + `foo_test.go`), `_race`-Flag im CI
- Python: `tests/`-Verzeichnis, pytest mit `asyncio_mode = "auto"`
- Frontend: Unit-Tests via Vitest, E2E später mit Playwright
- Mindest-Coverage: kein hartes Gate, aber sichtbar in CI

---

## Workflow-Regeln für Claude Code

Diese Regeln gelten **immer** und überschreiben alle Default-Verhaltensweisen.

### Plan vor Ausführung

- Bei jeder Änderung, die mehr als 3 Dateien betrifft: **erst Plan zeigen**
- Plan beinhaltet: Welche Dateien, welche Änderungen, welche Risiken
- Auf explizite Freigabe warten
- Plan-Mode nutzen wenn verfügbar (Shift+Tab in Claude Code)

### Commits

- **Nie eigenständig committen.** Der Maintainer committet selbst.
- Am Ende jeder Session: `git status` + Commit-Vorschlag formulieren
- Commit-Messages folgen Conventional Commits

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

### Generated Code

- Generierter Code (sqlc, oapi-codegen, openapi-typescript) wird committed
- CI prüft via `git diff --exit-code` ob aktuell
- Generated Files mit `// Code generated by ...; DO NOT EDIT.` markieren

### Sicherheit

- Keine Secrets in Code, Logs, Tests, Konfigurationen
- Keine `.env`-Files committen, nur `.env.example`
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

---

## Don'ts (harte Regeln)

- **Kein Kubernetes/K3s/Helm** in dieser Phase. Wir bleiben bei Compose. Wachstumspfad ist offen, aber jetzt nicht.
- **Kein ORM in Go.** sqlc + pgx direkt.
- **Kein npm oder yarn.** pnpm.
- **Kein pip oder poetry.** uv.
- **Keine Cloud-locked-in Services im kritischen Pfad.** Alles, was wir nutzen, muss self-hostbar sein oder leicht ersetzbar.
- **Keine Mock-Daten in Production-Code.** Tests dürfen mocken, App-Code nicht.
- **Keine kommentierten Code-Blöcke** im Repo. Wenn weg, dann weg. Git erinnert sich.
- **Keine `any`-Types in TypeScript** ohne Begründung im Kommentar.
- **Keine `interface{}`/`any` in Go-Public-APIs**.
- **Keine globalen Variablen** ohne sehr guten Grund.
- **Keine TODO-Kommentare ohne Issue-Referenz** (nach Phase 1).

---

## Wo finde ich was

| Anliegen                       | Ort                                    |
| ------------------------------ | -------------------------------------- |
| Architektur-Diagramme          | `docs/architecture.md`                 |
| Was tun wenn X kaputt?         | `docs/runbook.md`                      |
| Wie deploye ich?               | `docs/deployment.md`                   |
| Wie entwickle ich Feature X?   | `docs/development.md`                  |
| Warum diese Tech-Entscheidung? | `docs/adr/`                            |
| Config-Reference               | `docs/config-reference.md` (generiert) |
| Secrets-Workflow               | `docs/secrets.md`                      |
| Service-spezifische Doku       | `apps/<service>/README.md`             |

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

## Offene Fragen (werden im Verlauf beantwortet)

- [ ] Hosting-Provider: TBD (Hetzner Cloud Deutschland wahrscheinlich)
- [ ] Git-Hosting: TBD (GitHub vs. self-hosted Forgejo)
- [ ] DNS-Routing: TBD (Subdomains vs. Pfad-basiert)
- [ ] Email-Provider: TBD (Postmark, Brevo, eigener SMTP?)
- [ ] i18n-Library: TBD (svelte-i18n vs. Paraglide vs. Inlang)
- [ ] Backup-Ziel: TBD (S3-kompatibel, eigenes NAS?)
- [ ] Org/User-Name auf GitHub für Container-Registry: TBD
- [ ] Lizenz: TBD (vermutlich proprietär, evtl. AGPL für Backend, MIT für Schemas)

Wenn eine dieser Fragen für deine Aufgabe relevant wird: **fragen**, nicht annehmen.

---

## Session-Tracking

Die Sessions zur Initial-Einrichtung sind dokumentiert in `sessions/step01.md` bis
`sessions/step12.md`. Wenn du in einer Session arbeitest, halte dich an die dort
definierten Aufgaben. Bei Abweichung: zurück zur Datei, fragen.

Stand der Sessions wird in `sessions/STATUS.md` gepflegt — am Ende jeder Session
ein kurzer Eintrag.

---

## Dockerfile-Konventionen

- Alle `go install`-Befehle MÜSSEN eine `@vX.Y.Z`-Version haben (DL3062)
- Aktuelle Air-Adresse: github.com/air-verse/air (nicht mehr cosmtrek/air)
- Base-Images mit konkretem Tag, nicht `:latest` (DL3007)
- `apt-get install` immer mit `--no-install-recommends` und `rm -rf /var/lib/apt/lists/*`

---

## Pre-commit-Konventionen

- Sprach-spezifische Formatter (prettier, ruff) als `repo: local` Hooks
  einbinden, die das jeweilige Workspace-Tool aufrufen
  (`pnpm --filter ... exec prettier`, `uv run ruff`)
- Keine `additional_dependencies` für Sprach-Tools — nutzt sonst zwei
  Versionen parallel (Hook vs. App)
- `check-json`-Hook schließt JSONC-Files aus (tsconfig.json, .vscode/\*.json,
  .code-workspace)
- `pre-commit clean && pre-commit install --install-hooks` nach jeder
  Änderung an `.pre-commit-config.yaml`

---

## Maintainer-Identität

Maintainer commits MUST be made with:

- user.name: Hans-Werner Roitzsch
- user.email: <echte-adresse>@example.de (eine bei GitHub registrierte Adresse)
- SSH-Signing aktiv (gpg.format=ssh, user.signingkey=~/.ssh/id_ed25519.pub)
- SSH-Public-Key bei GitHub als BEIDES registriert: Authentication Key + Signing Key

Falls neuer Maintainer hinzukommt: gleicher Setup-Pfad in pre-session-checklist.md.

History pre-2026-05-04 enthält Commits unter Platzhalter-Mail "deine@email.tld" —
das ist Setup-Artefakt, kein Sicherheitsproblem. Nicht rewriten, nur ab Punkt X
neue Commits sauber führen.

---

## Changelog dieser Datei

Diese Datei wächst mit dem Projekt. Wenn du etwas Strukturelles lernst, das hier
fehlt: vorschlagen, mit Begründung. Ich entscheide, ob es rein kommt.

- 2026-05-03: Initiale Version
