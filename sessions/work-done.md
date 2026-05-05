# Aufgebaut: worldweathernews.com — Materialsammlung für die Präsentation

Diese Datei fasst zusammen, was in den Setup-Sessions 1–8 entstanden ist.
Sie ist als **Materialsammlung für eine Präsentation** gedacht: jedes Kapitel
liefert die Mission, das Ergebnis, die Schlüssel-Entscheidung mit Begründung,
einen prägnanten Stolperstein und einen Folien-tauglichen Take-away.

Der laufende Status pro Session steht in `sessions/STATUS.md` (knappe
Tagesnotizen, Single-Source-of-Truth für „was ist done"). Diese Datei hier ist
die ausführliche Variante mit Begründungen.

---

## Projekt auf einer Folie

**Was wir bauen:** Eine globale Wetter- und Klimaplattform mit Community-Features.
Aggregiert Daten nationaler Wetterdienste (DWD, NOAA, Met Office, JMA …),
visualisiert Klimaanomalien, baut eine Citizen-Science-Community auf.

**Wer es baut:** Diplom-Meteorologe und IT-Architekt in Personalunion. Alleinige
Entscheidungsgewalt, eigene Hosting-Infrastruktur — keine Abstimmungsschleifen
mit Cloud-Anbietern oder externen Dienstleistern.

**Architektur-Identität in fünf Worten:** Self-hosted. Type-safe. Single-Source.
Observability-ready. Kein Lock-in.

**Tech-Stack im Überblick:**

| Schicht       | Technologie                           | Warum                                           |
| ------------- | ------------------------------------- | ----------------------------------------------- |
| Backend       | Go 1.25 + Chi + sqlc + pgx            | Performance, kein ORM-Overhead                  |
| Workers       | Python 3.12 + uv                      | xarray/cfgrib für GRIB/NetCDF reif              |
| Frontend      | SvelteKit 2 + Svelte 5 Runes          | Schlanker als Next.js, exzellente DX            |
| Datenbank     | PostgreSQL 16 + TimescaleDB + PostGIS | Geo + Zeitreihen aus einer Hand                 |
| Reverse Proxy | Caddy                                 | Auto-SSL, einfache Config                       |
| Container     | Docker Compose                        | Wachstumspfad zu K3s offen, jetzt nicht         |
| CI            | GitHub Actions                        | Schnell, kostenlos, verbreitet                  |
| Schema        | OpenAPI 3.1 als SoT                   | Server-Stubs (Go) + Client-Types (TS) generiert |

**Was bewusst NICHT drin ist:** Kubernetes, ORMs, npm/yarn (pnpm),
pip/poetry (uv), Cloud-Lock-in-Services im kritischen Pfad.

---

## Phase 1 — Fundament (Sessions 1–2)

### Session 1 — Repo-Skelett und Tooling

**Mission:** Reproduzierbares Tool-Setup, das auf jedem Rechner eines
zukünftigen Mitwirkenden in einem Befehl läuft.

**Was am Ende stand:**

- Monorepo-Struktur: `apps/`, `packages/`, `infra/`, `docs/`, `sessions/`
- `.mise.toml` mit gepinnten Versionen für Go, Node, Python, pnpm, uv,
  golangci-lint, sqlc, goose, pre-commit, yamllint
- Top-Level-`Makefile` als einheitliches Interface (`make bootstrap`,
  `make dev`, `make lint`, `make test`)
- ADR-Verzeichnis vorbereitet (`docs/adr/`)
- Abschluss-Commit: `065703f`

**Schlüssel-Entscheidung:** Versionsmanagement über **mise** statt asdf, nvm
oder Homebrew.

- _Begründung:_ mise unterstützt Go, Node, Python und Universal-Binaries
  (`ubi:`-Backend für sqlc/goose) in einer einzigen Tool-Definition. Keine
  drei verschiedenen Versionsmanager parallel pflegen, keine
  „funktioniert-bei-mir"-Diskussionen.

**Stolperstein:** `ubi:`-Backend war zur Setup-Zeit noch experimental und
musste in `.mise.toml` mit `[settings] experimental = true` aktiviert werden.
Fallback per `go install` ist dokumentiert für Systeme, auf denen ubi nicht
greift.

**Take-away für Folie:** _Ein Befehl (`mise install`) bringt jede neue
Maschine auf den exakt selben Toolchain-Stand — über drei Sprachen hinweg._

---

### Session 2 — Pre-commit, Makefile, lokale Workflows

**Mission:** Code, der ins Repo geht, ist immer bereits formatiert, gelinted
und frei von häufigen Fehlern. Vor jedem Commit, lokal, ohne CI-Wartezeit.

**Was am Ende stand:**

- `.pre-commit-config.yaml` mit Hooks für: trailing-whitespace, EOF-fixer,
  YAML-Lint, JSON-Check, Secret-Detection (gitleaks), prettier (als
  `repo: local`-Hook), ruff/ruff-format, golangci-lint, hadolint
- `make fmt` / `make lint` / `make test` als sprach-übergreifende Eintrittspunkte
- `pre-session-checklist.md` für saubere Session-Starts
- Abschluss-Commit: `0207fd4`

**Schlüssel-Entscheidung:** Sprach-spezifische Formatter (prettier, ruff,
golangci-lint) als **`repo: local`-Hooks**, die das Workspace-Tool aufrufen
— nicht über `additional_dependencies` im pre-commit-Config.

- _Begründung:_ Sonst laufen zwei Versionen parallel (Hook-Version vs.
  App-Version), und Diff-Verhalten zwischen lokalem `pnpm exec prettier`
  und CI driftet auseinander. Eine Wahrheit pro Sprache.

**Stolperstein:** `check-json`-Hook scheiterte an JSONC-Files
(`tsconfig.json`, `.vscode/*.json`); musste explizit ausgeschlossen werden.

**Take-away für Folie:** _Wir haben CI bereits lokal, bevor wir CI in der
Cloud haben._

---

## Phase 2 — Services (Sessions 3–6)

### Session 3 — Compose-Stack mit Postgres, Redis, Caddy

**Mission:** Lokale Entwicklungsumgebung, die produktionsnah ist:
gleiche Datenbank, gleicher Cache, gleiches TLS-Verhalten — nur lokal.

**Was am Ende stand:**

- `infra/compose/compose.dev.yml` mit fünf Services: postgres
  (TimescaleDB-HA-Image inkl. PostGIS + TimescaleDB), redis, caddy, mailhog,
  postgres-init
- `infra/caddy/Caddyfile` mit Subdomain-Routing
  (`api.localhost`, `app.localhost`)
- `make dev` startet den Stack, `make dev-down` stoppt, `make dev-reset`
  löscht Volumes
- Smoke-Tests: alle fünf Services kommen sauber hoch
- Abschluss-Commit: `9a02e28`

**Schlüssel-Entscheidung:** **TimescaleDB-HA-Image** als Postgres-Basis
statt vanilla postgres + nachinstallierte Extensions.

- _Begründung:_ Liefert PostGIS und TimescaleDB out-of-the-box. Keine
  Init-Scripts mit `CREATE EXTENSION`, keine Versions-Drift zwischen
  Image-Update und Extension-Update. Geo + Zeitreihen sind beide
  Kernfunktionen — das ist kein Add-on.

**Stolperstein:** Caddyfile musste auf `http://`-Schema umgestellt werden,
sonst bindet Caddy mit `auto_https off` trotzdem auf Port 443 und kollidiert
mit ggf. laufenden Services. Eine Stunde gesucht, ein Zeichen geändert.

**Take-away für Folie:** _Production-Stack-light als Compose-File: fünf
Services, ein Befehl, keine Cloud-Abhängigkeiten._

---

### Session 4 — Go-Backend-Skelett

**Mission:** Ein Go-API-Server, der Health, Readiness, Metrics, strukturiertes
Logging und Graceful Shutdown korrekt macht — bevor der erste Endpunkt
fachliche Logik hat.

**Was am Ende stand:**

- Modul `github.com/relations4u/worldweathernews/apps/backend`
- Chi v5 als HTTP-Router, pgx/v5 für DB-Access, Viper für Config
- Endpunkte: `/healthz`, `/readyz`, `/metrics` (Prometheus),
  `/api/v1/ping`
- Strukturiertes JSON-Logging via `slog` mit Trace-IDs in jedem Eintrag
- Multi-Stage-Dockerfile mit **distroless als final stage**: 26.2 MB
- `golangci-lint` v2 mit migriertem `.golangci.yml`
- Abschluss-Commit: `de176bc`

**Schlüssel-Entscheidung:** **sqlc + pgx direkt, kein ORM** (kein GORM, kein
ent, kein xorm).

- _Begründung:_ sqlc generiert typsicheren Go-Code aus reinen SQL-Queries.
  Kein DSL-Lernaufwand, kein N+1-Problem durch Lazy-Loading-Magie, keine
  Performance-Überraschungen durch ORM-Abstraktionen. Wer SQL liest,
  versteht den Code. Migration-Pfad zu beliebiger anderer DB-Technologie
  bleibt offen.

**Stolperstein:** `go.mod` wurde durch eine transitive Viper-Dependency
automatisch von 1.23 auf 1.25 hochgezogen. Dockerfile-Builder-Stage musste
nachgezogen werden, sonst CI-Bruch. CLAUDE.md wurde anschließend an Go 1.25
angepasst.

**Take-away für Folie:** _Production-Backend in 26 MB Container-Image.
Distroless statt Alpine: kein Shell, keine Angriffsfläche, kein Bloat._

---

### Session 5 — SvelteKit-Frontend-Skelett

**Mission:** Ein Frontend-Skelett, das Backend-Connectivity prüft, mobile-first
denkt und auf Komponenten-Bibliotheken setzt, die mit dem Stack altern können.

**Was am Ende stand:**

- SvelteKit 2 + Svelte 5 (Runes-Syntax) + TypeScript strict
- Tailwind v3 + shadcn-svelte (Badge-Komponente manuell installiert,
  shadcn-CLI hätte interaktiv gehängt)
- `@sveltejs/adapter-node` für Self-Hosting im Container
- Backend-Connectivity-Check auf der Landing-Page (ruft `/api/v1/ping`)
- Multi-Stage-Dockerfile mit non-root in Production
- ESLint 10, prettier, svelte-check, vitest — alle grün
- Abschluss-Commit: `6a5708f`

**Schlüssel-Entscheidung:** **SvelteKit statt Next.js, React oder Nuxt.**

- _Begründung:_ Kleinere Bundle-Größen, weniger Indirektion, Compiler statt
  Runtime-Diffing. Svelte 5 Runes geben uns moderne Reaktivität ohne
  Boilerplate. Self-Hosting via adapter-node ist ein erstklassiger Pfad,
  nicht ein Workaround wie Next-on-Node.

**Stolperstein:** Frontend-Container muss im Dev-Mode als **root** laufen.
`node:22-alpine` läuft default als uid=1000, der hat aber kein Schreibrecht
auf `/usr/local/lib`, wo pnpm globale Dev-Server-Tools ablegen will.
Production-Image bleibt non-root — der Unterschied ist dokumentiert.

**Take-away für Folie:** _Frontend wie ein Backend behandeln: Container,
Health-Check, Self-Hosted, kein Vercel-Lock-in._

---

### Session 6 — Python-Workers-Skelett

**Mission:** Worker-Service-Skelett, der periodisch Daten von externen
Wetterdienst-APIs zieht, normalisiert und in die DB schreibt — mit derselben
Observability wie das Backend.

**Was am Ende stand:**

- `apps/pyworkers/` mit pydantic-settings (typsichere Config),
  structlog (strukturiertes JSON-Logging), httpx (async HTTP),
  asyncpg, redis-py, prometheus-client
- APScheduler 3.x (AsyncIOScheduler) für periodische Jobs
- `click`-CLI als Eintrittspunkt
- Heartbeat alle 30 s, Metriken auf `:9100`
- mypy strict mit gezielten Overrides für Libraries ohne `py.typed`
  (asyncpg, apscheduler)
- pytest mit asyncio_mode = "auto"
- Abschluss-Commit: `6cd78cd`

**Schlüssel-Entscheidung:** **uv als Package-Manager und Resolver, nicht pip
oder poetry.**

- _Begründung:_ uv ist um Größenordnungen schneller bei Resolve und
  Install. Deterministisch via `uv.lock`. Macht den Pip-Cache-Tanz und die
  poetry-Versionskonflikte überflüssig. Container-Build-Zeit halbiert sich.

**Stolperstein:** APScheduler 4.x ist noch Alpha — **bewusst auf 3.x
gepinnt** (`apscheduler>=3.10,<4`). Das Production-Image ist 250 MB; das
Soft-Target von 200 MB wurde verfehlt. Optimierung (uv-Cache-Mount,
multi-stage mit `--no-install-project`) ist vertagt, weil die Größe
funktional unkritisch ist.

**Take-away für Folie:** _Drei Sprachen, ein Logging-Format
(strukturiertes JSON mit Trace-ID), eine Metrics-Schnittstelle (Prometheus)._

---

## Phase 3 — CI/CD (Sessions 7–8)

### Session 7 — OpenAPI-Schema und Type-Generation

**Mission:** Frontend und Backend dürfen sich nie über die API uneinig sein.
Das Schema ist kein Doku-Artefakt, sondern Quelle für generierten Code.

**Was am Ende stand:**

- `packages/api-schema/openapi.yaml` als Single Source of Truth (OpenAPI 3.1)
- Server-Stubs für Go via `oapi-codegen` v2.4.1
- Client-Types für TypeScript via `openapi-typescript` 7.10
- Generierter Code committed (Konvention: `// Code generated …; DO NOT EDIT.`)
- `scripts/check-generated.sh` als CI-Gate (`git diff --exit-code` nach
  Re-Generation)
- ADR-0001: „OpenAPI as Source of Truth" begründet die Entscheidung
- `/api/v1/locations` validiert `required: q` automatisch (HTTP 400 ohne
  Parameter)
- Abschluss-Commit: `1e37154`

**Schlüssel-Entscheidung:** **OpenAPI als Single Source of Truth, nicht
Code-First.**

- _Begründung:_ Bei Code-First muss eine Sprache (typischerweise Go) die
  Spec emittieren, andere Sprachen importieren sie. Das verzerrt den Stack:
  TypeScript-Types wirken wie ein Side-Effekt. Schema-First trägt die Wahrheit
  in einer Sprache (YAML), die allen Tools vertraut ist. Code wird daraus
  generiert — auf beiden Seiten.

**Stolperstein:** `oapi-codegen` warnt bei OpenAPI 3.1 (es unterstützt
offiziell 3.0), funktioniert aber. pnpm 11 brauchte `allowBuilds: true`
für `core-js` und `protobufjs` plus `verify-deps-before-run=false` in
`.npmrc`. Der `pnpm-workspace` musste um `packages/api-schema` ergänzt
werden, damit das Schema Workspace-Member ist.

**Take-away für Folie:** _Ein YAML-File. Ein `make gen`. Server und Client
im Lockstep — Drift mathematisch unmöglich._

---

### Session 8 — GitHub Actions CI-Workflows

**Mission:** Jeder Push und jeder PR muss durch denselben Qualitätsfilter wie
ein Production-Deploy. Ohne grünes CI keine Merge-Möglichkeit auf `main`.

**Was am Ende stand:**

- Fünf Workflow-Files in `.github/workflows/`:
  - `ci-backend.yml` — Go: lint, test (`-race`), build
  - `ci-frontend.yml` — TS: lint, check (svelte-check), test, build
  - `ci-pyworkers.yml` — Python: ruff, mypy, pytest, pip-audit
  - `ci-shared.yml` — yamllint, OpenAPI lint (redocly), commit-lint,
    Markdown-Link-Check, Generated-Files-Drift-Check
  - `security-scan.yml` — Dependabot + Trivy
- `.commitlintrc.yaml` für Conventional Commits
- README-Badges für Workflow-Status
- Branch Protection auf `main`: nur per PR mit grünem CI mergebar,
  signierte Commits erforderlich
- Toolchain-Alignment: Go 1.25, golangci-lint 2.12.1, Node 22, Python 3.12
  in CI gepinnt — identisch zu lokalem `.mise.toml`
- Abschluss-Commits: `fac75c9`, `90c4078`, `cb24876` (PR #15),
  `1ba7009` (PR #16, finaler Fix-PR)

**Schlüssel-Entscheidung:** **CI-Tools für jede Sprache nutzen die
Workspace-Tools, nicht eigene Versionen.**

- _Begründung:_ Was lokal grün ist, ist auch in CI grün und umgekehrt.
  golangci-lint-Version aus `.mise.toml`, ruff-Version aus
  `pyproject.toml`-Dev-Deps, prettier aus dem Frontend-Workspace.
  Keine Versions-Drift zwischen Entwickler-Maschine und CI-Runner.

**Stolperstein — und gleichzeitig die Story dieser Session:** Der erste PR
(#15) wurde gemerged, der Folge-PR (#16) für Lückenschluss hatte fünf rote
Checks:

1. **Frontend Type-Check:** `PUBLIC_API_BASE_URL` nicht aus
   `$env/static/public` exportiert — gefixt durch committetes
   `apps/frontend/.env` mit Default + `.gitignore`-Ausnahme
2. **Frontend Build:** gleiche Wurzel
3. **Commit-Lint:** GitHub Action hatte keine `pull-requests: read`
   Permission — `Resource not accessible by integration`
4. **Markdown-Link-Check:** lychee fand neun tote Links
   (Platzhalter-URLs, Production-URLs, die noch nicht existieren) — gefixt
   durch `.lycheeignore` + `--accept`-Codes + `fail: false` in
   Skeleton-Phase
5. **OpenAPI-Lint:** `redocly` mit `recommended`-Ruleset hatte zwei
   `security-defined`-Errors — gefixt durch Top-Level `security: []` (alle
   Endpoints bewusst öffentlich für jetzt) plus `redocly.yaml` mit
   selektiven Severities

Alle fünf Fixes wurden in einem PR gebündelt, der heute (2026-05-05)
mergebar war (mergeStateStatus: CLEAN, mergeable: MERGEABLE) und gemerged
wurde.

**Take-away für Folie:** _Zwei CI-Iterationen, ein durchgängig grüner
Pipeline-Stand. Branch-Protection erzwingt jetzt: kein Code auf main ohne
grünes CI plus signierten Commit._

---

## Aktueller Stand nach Session 8

### Was läuft

- **Lokale Entwicklung:** `make dev` startet einen produktionsnahen Stack in
  unter einer Minute
- **Code-Qualität:** Pre-commit lokal + GitHub Actions in der Cloud — beide
  prüfen dieselben Regeln mit denselben Versionen
- **API-Vertrag:** OpenAPI-Schema mit generierten Server-Stubs und
  Client-Types, CI sichert Drift-Freiheit
- **Container:** Backend 26 MB (distroless), PyWorkers 250 MB (slim),
  Frontend non-root in Production
- **Branch-Protection:** `main` ist gegen direkte Pushes geschützt, alle
  Merges laufen über PRs mit grünem CI

### Was noch aussteht (Sessions 9–12)

| Session | Phase | Inhalt                                                           |
| ------- | ----- | ---------------------------------------------------------------- |
| 9       | CI/CD | Release-Workflow + Container-Registry (ghcr.io)                  |
| 10      | Ops   | Observability-Stack lokal (Prometheus, Grafana, Loki, Tempo)     |
| 11      | Ops   | Ansible + SOPS + Terraform-Skelett                               |
| 12      | Ops   | Dokumentation finalisieren, Lizenz-Entscheidung, ADRs nachziehen |

### Bewusste offene Punkte

- **Hosting-Provider** — voraussichtlich Hetzner Cloud (Deutschland),
  endgültige Entscheidung in Session 11
- **Git-Hosting** — aktuell GitHub, Migration zu self-hosted Forgejo
  wird offengehalten
- **Lizenz** — vermutlich proprietär, eventuell AGPL für Backend / MIT für
  Schemas (Session 12)
- **Email-Provider** — Postmark oder Brevo, Entscheidung mit erstem
  produktiven User-Flow
- **Backup-Ziel** — S3-kompatibel oder eigenes NAS, Entscheidung in Session 11

---

## Anhang: Toolchain-Pinning

Alle Versionen sind reproduzierbar gepinnt. Wer das Repo klont, bekommt
deterministisch denselben Stack.

### `.mise.toml`

| Tool           | Version |
| -------------- | ------- |
| Go             | 1.25    |
| Node           | 22      |
| Python         | 3.12    |
| pnpm           | 9       |
| uv             | latest  |
| golangci-lint  | 2.12.1  |
| sqlc           | 1.27.0  |
| goose          | 3.22.1  |
| pre-commit     | 4.0.1   |
| yamllint       | 1.35.1  |
| task (go-task) | latest  |

### Backend (Go) — wesentliche Dependencies

- `github.com/go-chi/chi/v5` v5.2.5 — HTTP-Router
- `github.com/jackc/pgx/v5` v5.9.2 — PostgreSQL-Driver
- `github.com/getkin/kin-openapi` v0.137.0 — OpenAPI-Validation
- `github.com/oapi-codegen/runtime` v1.4.0 — Generated-Code-Runtime
- Viper, slog, prometheus/client_golang

### Frontend (TS) — wesentliche Dependencies

- `@sveltejs/kit` 2.57+
- `@sveltejs/adapter-node` 5.5+
- Svelte 5, TypeScript-ESLint 8.59+
- Tailwind v3, shadcn-svelte
- vitest, @testing-library/svelte

### PyWorkers (Python) — wesentliche Dependencies

- pydantic 2.7+, pydantic-settings 2.3+
- structlog 24.1+ (JSON-Logging)
- httpx 0.27+, asyncpg 0.29+, redis 5.0+
- apscheduler 3.10–<4 (4.x ist Alpha, bewusst auf 3.x)
- prometheus-client 0.20+
- click 8.1+
- pytest 8.0+, pytest-asyncio 0.23+, pytest-cov 5.0+

---

## Querverweise

- **Knappe Tagesnotizen:** `sessions/STATUS.md`
- **Vision und Regeln:** `CLAUDE.md` (Repo-Root)
- **Architektur-Entscheidungen:** `docs/adr/0001-openapi-as-source-of-truth.md`
- **Wie deploye ich?** `docs/deployment.md` (Session 9 erweitert)
- **Was tun wenn X kaputt?** `docs/runbook.md` (wächst mit den Sessions)
