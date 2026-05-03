# Session 12 — Dokumentation finalisieren

**Phase**: D (Ops)
**Geschätzte Dauer**: 1 Stunde
**Vorbedingung**: Sessions 1-11 abgeschlossen.

## Ziel

Alle Dokumentation ist konsistent, vollständig und aktuell. Ein neuer Entwickler
oder Mitarbeitender kann das Repo klonen und in maximal 30 Minuten lokal
loslaufen lassen. Architektur-Entscheidungen sind als ADRs festgehalten.
TODO-Marker sind bereinigt oder in Issues konvertiert. Lizenz-Frage ist
beantwortet.

Diese Session ist nicht spektakulär, aber wichtig: die ganze Arbeit der vorherigen
Sessions wird hier "fertig gemacht".

## Vor-Klärung

- **Lizenz**: Welche soll's werden? Empfehlungen je nach Strategie:
  - **AGPL-3.0** — wenn das Backend-System Open Source sein soll, Forks aber
    Re-Hosts auch Source veröffentlichen müssen
  - **MIT/Apache-2.0** — maximal permissiv, wenn du die Verbreitung priorisiert
  - **Proprietär (alle Rechte vorbehalten)** — wenn kommerzielle Verwertung
    geplant ist, Code aber im Repo nicht öffentlich
  - **Mixed** — z.B. MIT für `packages/api-schema`, AGPL für `apps/backend`,
    proprietär für Rest. Üblich, aber pflegeintensiv.
- **GitHub-Org-Name** — falls noch der Platzhalter `<org>` im Code steht: jetzt
  global ersetzen.

## Aufgaben

### 1. Top-Level `README.md` polieren

Komplett neuschreiben oder strukturieren:

````markdown
# worldweathernews.com

Globale Wetter- und Klima-Plattform mit Community-Features.

[![CI Backend](https://github.com/<org>/worldweathernews/actions/workflows/ci-backend.yml/badge.svg)](https://github.com/<org>/worldweathernews/actions/workflows/ci-backend.yml)
[![CI Frontend](https://github.com/<org>/worldweathernews/actions/workflows/ci-frontend.yml/badge.svg)](https://github.com/<org>/worldweathernews/actions/workflows/ci-frontend.yml)
[![CI PyWorkers](https://github.com/<org>/worldweathernews/actions/workflows/ci-pyworkers.yml/badge.svg)](https://github.com/<org>/worldweathernews/actions/workflows/ci-pyworkers.yml)
![License](https://img.shields.io/badge/license-<LICENSE>-blue.svg)

## Was ist das?

Eine self-hosted Plattform für Wetter- und Klima-Beobachtungen, die:

- Daten aus nationalen Wetterdiensten weltweit aggregiert
- Klimadaten visualisiert (Anomalien, Trends, historische Vergleiche)
- Eine Community von Beobachtern, Citizen Scientists und Wetter-Interessierten verbindet
- Auf Mobile und Desktop funktioniert

**Status**: in early development. Currently building infrastructure.

## Quickstart

Voraussetzungen: Docker, Docker Compose, [mise](https://mise.jdx.dev/) oder
manuell installiertes Go 1.23 + Node 22 + Python 3.12 + pnpm + uv.

```bash
git clone https://github.com/<org>/worldweathernews.git
cd worldweathernews
mise install
make bootstrap
cp .env.example .env
make dev
```
````

Dann öffnen:

- App: http://app.localhost
- API: http://api.localhost
- Mailhog: http://mail.localhost

## Tech-Stack

| Schicht       | Technologie                              |
| ------------- | ---------------------------------------- |
| Backend       | Go 1.23, Chi, sqlc, pgx                  |
| Workers       | Python 3.12, asyncio, asyncpg, structlog |
| Frontend      | SvelteKit, TypeScript, Tailwind, shadcn  |
| Datenbank     | PostgreSQL 16 + PostGIS + TimescaleDB    |
| Cache         | Redis 7                                  |
| Reverse Proxy | Caddy                                    |
| Container     | Docker + Docker Compose (lokal & prod)   |
| Monitoring    | Prometheus, Grafana, Loki, Tempo         |
| CI/CD         | GitHub Actions, ghcr.io                  |
| Deployment    | Ansible, SOPS+age, Terraform             |

## Repo-Struktur

```
apps/
  backend/        Go API
  frontend/       SvelteKit
  pyworkers/      Python Workers
packages/
  api-schema/     OpenAPI 3.1 (Single Source of Truth)
  shared-types/   Generierte TS-Types
infra/
  compose/        Docker-Compose-Files
  caddy/          Reverse-Proxy-Config
  monitoring/     Prometheus + Grafana + Loki + Tempo
  ansible/        Server-Konfiguration
  terraform/      Server-Provisionierung
  secrets/        SOPS-verschlüsselte ENV-Files
  migrations/     goose-DB-Migrations
docs/             Architektur, Runbook, ADRs
.github/workflows/  CI/CD
```

## Dokumentation

- [Architektur](docs/architecture.md) — System-Diagramm, Service-Verantwortlichkeiten
- [Development](docs/development.md) — Wie ich Feature X hinzufüge
- [Deployment](docs/deployment.md) — Wie ich auf einen Server deploye
- [Runbook](docs/runbook.md) — Was tun wenn X kaputt?
- [Secrets](docs/secrets.md) — SOPS-Workflow
- [ADRs](docs/adr/) — Architecture Decision Records
- [Contributing](CONTRIBUTING.md) — Wie ich beitrage
- [CLAUDE.md](CLAUDE.md) — Spielregeln für KI-gestützte Entwicklung

## Lizenz

<LICENSE_TEXT_OR_LINK>

`````

### 2. `docs/architecture.md`

Vollständige Architektur-Doku mit Mermaid-Diagrammen:

````markdown
# Architektur

## System-Überblick

```mermaid
graph TB
    subgraph "External"
        User[User Browser]
        WeatherAPIs[National Weather Services]
    end

    subgraph "Edge"
        Caddy[Caddy Reverse Proxy<br/>+ ACME]
    end

    subgraph "Application"
        Frontend[SvelteKit<br/>Node Adapter]
        Backend[Go API<br/>Chi + sqlc]
        Workers[Python Workers<br/>asyncio]
    end

    subgraph "Data"
        Postgres[(PostgreSQL 16<br/>+PostGIS<br/>+TimescaleDB)]
        Redis[(Redis 7)]
    end

    subgraph "Observability"
        Prometheus[Prometheus]
        Loki[Loki]
        Tempo[Tempo]
        Grafana[Grafana]
    end

    User --> Caddy
    Caddy --> Frontend
    Caddy --> Backend
    Frontend -.->|API calls| Backend
    Backend --> Postgres
    Backend --> Redis
    Workers --> Postgres
    Workers --> Redis
    Workers --> WeatherAPIs

    Backend -.->|metrics| Prometheus
    Workers -.->|metrics| Prometheus
    Backend -.->|traces| Tempo
    Workers -.->|traces| Tempo
    Backend -.->|logs| Loki
    Workers -.->|logs| Loki
    Grafana --> Prometheus
    Grafana --> Loki
    Grafana --> Tempo
`````

## Service-Verantwortlichkeiten

### Backend (Go)

- HTTP-API für Frontend und ggf. mobile Apps / Drittsysteme
- Authentifizierung und Autorisierung
- Geschäftslogik (Locations, User-Profile, Alerts, ...)
- Caching von hot Read-Pfaden (Redis)
- Wird **nicht** für ETL/Batch-Jobs genutzt

### Pyworkers (Python)

- Pull von externen Wetterdaten (DWD, NOAA, Open-Meteo, ...)
- GRIB- und NetCDF-Parsing
- Normalisierung und Persistierung in TimescaleDB
- Periodische Aggregationen (z.B. tägliche/monatliche Klima-Stats)
- Sentinel-Jobs (Health, Heartbeat)

### Frontend (SvelteKit)

- Server-Side-Rendering für SEO und schnellen First-Paint
- Hydration für Interaktivität
- Karten-Komponente (MapLibre, später)
- Personalisierung (Locations, Einheiten, Sprache)

## Datenfluss

### Read-Pfad (User schaut Wetter an)

```
User → Caddy → Frontend (SSR) → Backend → Redis (Cache) → Postgres (Miss) → Backend → Frontend → User
```

### Write-Pfad (User postet Beobachtung)

```
User → Caddy → Frontend → Backend (Auth) → Postgres
```

### Ingest-Pfad (Wetterdaten holen)

```
Cron → Workers → External API (HTTP/GRIB) → Normalize → Postgres (TimescaleDB Hypertable)
```

## Datenmodell (high-level)

(Wird mit ersten Features konkretisiert. Skizze:)

- `locations` — geographische Orte (UUID, Name, Country, Geo)
- `weather_observations` — Hypertable, time-series der Beobachtungen pro Location
- `users` — Mitglieder
- `posts` — User-generated Content
- `weather_stations` — citizen-science Stations

## Skalierungs-Annahmen

- Single-Host-Deployment ist initial ausreichend (ein Hetzner CCX31 oder vergleichbar)
- Horizontale Skalierung ist über mehrere Backend-Replicas + externer LB möglich,
  aber für Phase 1 nicht nötig
- TimescaleDB skaliert vertikal weit; Sharding kommt erst bei vielen TB

## Externe Abhängigkeiten (Plan)

| Service              | Zweck                | Kritikalität     |
| -------------------- | -------------------- | ---------------- |
| Hetzner Cloud        | Hosting              | Kritisch         |
| ghcr.io              | Container-Images     | Kritisch         |
| GitHub               | Source-Hosting, CI   | Kritisch         |
| Open-Meteo           | Wetterdaten Phase 1  | Hoch (ersetzbar) |
| DWD OpenData         | Deutschland-Daten    | Hoch (ersetzbar) |
| Sentry               | Error-Tracking       | Mittel           |
| (Email-Provider TBD) | Transaktionale Mails | Mittel           |

````

### 3. `docs/development.md`

```markdown
# Development Guide

## Wie ich einen neuen API-Endpoint hinzufüge

1. **OpenAPI-Schema erweitern**
   ```yaml
   # packages/api-schema/openapi.yaml
   /api/v1/weather/{locationId}:
     get:
       operationId: getWeather
       ...
   ```
2. `make gen` (regeneriert Go-Stubs und TS-Types)
3. **Backend-Handler**
   ```go
   // apps/backend/internal/http/handler/weather.go
   func (h *APIHandler) GetWeather(ctx context.Context, req api.GetWeatherRequestObject) (api.GetWeatherResponseObject, error) {
       // ... DB-Query, Cache-Check, Return
   }
   ```
4. **Test schreiben** (Handler + Integration)
5. **Frontend-Aufruf**
   ```ts
   import { getWeather } from '$lib/api/client';
   const result = await getWeather(locationId);
   ```
6. PR aufmachen, CI grün, mergen

## Wie ich eine neue DB-Migration mache

```bash
goose -dir infra/migrations create add_weather_observations sql
```

Migration schreiben (UP + DOWN), dann lokal anwenden:
```bash
make migrate
```

## Wie ich einen neuen Worker-Job hinzufüge

1. Datei `apps/pyworkers/pyworkers/jobs/my_job.py` mit `async def run()` anlegen
2. In `__main__.py` als scheduler-job registrieren
3. Test in `tests/test_my_job.py`
4. Metrics ergänzen falls relevant

## Wie ich einen Bug debugge

1. **Reproduzieren**: lokal mit `make dev-full`
2. **Logs anschauen**: Loki via Grafana, filtern nach `service` und `level=error`
3. **Trace folgen**: Trace-ID aus Log → Tempo-Search → Span-Tree
4. **DB-Stand prüfen**: `make dev-psql`
5. **Cache prüfen**: `make dev-redis`
6. Wenn reproduziert: Test schreiben, Fix, PR

## Lokale Tipps

- **Backend hot-reload**: `air` läuft im Container, Code-Änderungen → Auto-Restart
- **Frontend hot-reload**: Vite HMR via Caddy
- **DB-Migrations zurückrollen**: `goose -dir infra/migrations down`
- **Stack neu aufsetzen**: `make dev-reset && make dev`
- **Container-Logs eines Services**: `make dev-logs SERVICE=backend`
```

### 4. `docs/deployment.md` (Erweiterung)

Ergänze:
- Voraussetzungen (Provider-Account, Domain, age-Key)
- Erstmaliges Deployment-Schritt-für-Schritt:
  1. Terraform-Variables setzen
  2. `terraform apply` für staging
  3. DNS-Eintrag setzen
  4. Ansible-Inventory updaten
  5. Secrets befüllen + encrypten
  6. `ansible-playbook .../site.yml`
  7. Smoke-Test
- Folge-Deployments: `bash scripts/deploy.sh staging 0.1.x`
- Rollback-Prozedur
- Container-Registry-Setup (ghcr.io-Visibility)
- Branch-Protection-Setup (aus Session 8)

### 5. `docs/runbook.md` (vollständig)

Konkrete Szenarien mit Schritt-für-Schritt-Diagnose:

- "Backend antwortet nicht"
- "DB-Connections am Limit"
- "Externe API X gibt 500 zurück"
- "Wir wurden DDoS'd"
- "Disk fast voll"
- "Ein Service crashed loop"
- "Memory-Leak im Backend"
- "Worker-Job hängt"
- "Letzter Release ist kaputt — wie rolle ich zurück?"
- "Postgres-Restore aus Backup"

Pro Szenario: Symptome → Sofortmaßnahmen → Diagnose-Schritte → Fix → Postmortem-Hinweise.

### 6. ADRs ergänzen

`docs/adr/`:

**`0002-go-for-backend.md`** — Warum Go statt Node/Python/Rust?
- Performance, Concurrency-Modell, Self-Hosting-Footprint, Toolchain-Reife

**`0003-monorepo.md`** — Warum ein Repo statt Multi-Repo?
- Atomare Änderungen über Service-Grenzen, einfachere CI, gemeinsame Tooling

**`0004-compose-before-k3s.md`** — Warum nicht direkt Kubernetes?
- Solo-Maintainer, kein Cluster-Bedarf, Compose deckt 1-Host-Deployment ab,
  Wachstumspfad nach K3s ist offen aber nicht jetzt

**`0005-sops-for-secrets.md`** — Warum SOPS und nicht Vault/Doppler/...?
- Self-Hosting-Konsistenz, in Git versionierbar, kein zusätzlicher Service,
  age-Keys einfacher als GPG

Format MADR (siehe ADR-0001 als Vorlage).

### 7. `CONTRIBUTING.md`

```markdown
# Contributing

## Setup

Siehe README.md → Quickstart.

## Branching

- `main` ist immer deploybar
- Feature-Branches: `feat/short-description`
- Fix-Branches: `fix/issue-123-description`
- Niemals direkt auf `main` pushen

## Commits

Conventional Commits:
- `feat(scope): add X`
- `fix(scope): handle null in Y`
- `chore(deps): bump golangci-lint`
- `docs(adr): add 0006-i18n-strategy`

Scopes: `backend`, `frontend`, `pyworkers`, `infra`, `api`, `ci`, `deps`, `docs`.

## Pull Requests

- Eine PR = ein logisches Thema
- Lokale Tests müssen grün sein bevor PR aufgemacht wird:
  ```bash
  make lint
  make test
  ```
- Beschreibung enthält: Was, Warum, wie getestet
- Bei Schema-Änderung: `make gen` und `make gen-check` lokal grün
- Self-review vor Anfrage zum Review

## Code-Style

Wird durch Linter erzwungen. Bei Auto-Fix:
```bash
make fmt
```

## Reviews

- Reviewer prüft: Logik, Tests, Doku, Side-Effects
- Reviewee fragt nach, wenn Feedback unklar ist
- Niemand merged eigene PRs ohne Review (außer Solo-Maintainer-Phase)

## Reporting Bugs

GitHub Issues mit Template.

## Reporting Security Issues

Privat per E-Mail an security@worldweathernews.com (TBD: Email einrichten).
```

### 8. `CLAUDE.md` finales Update

Ergänze die "Wichtige Kommandos"-Sektion mit allen tatsächlichen Make-Targets,
die jetzt existieren. Aktualisiere die Liste der offenen Fragen — die meisten
sollten jetzt beantwortet sein. Ergänze Erfahrungswerte aus den Sessions, die
für nachfolgende Arbeit hilfreich sind.

### 9. Cleanup-Pass

Repo durchgehen:

**TODO-Marker** finden:
```bash
grep -rn "TODO\|FIXME\|XXX" --include="*.go" --include="*.py" --include="*.ts" --include="*.svelte" .
```

Jeden Marker bewerten:
- Konkret und kurzfristig fixbar? → Issue aufmachen, Marker mit Issue-Nr ergänzen:
  `// TODO(#42): Replace with proper validation`
- Technisches Schuldscheinchen für später? → in `docs/backlog.md` festhalten,
  Code-Marker entfernen
- Gar nicht mehr relevant? → Marker löschen

Liste der konvertierten/gelöschten Marker am Ende der Session ausgeben.

**Tote Imports und Dead Code**:
- `golangci-lint` mit `unused`-Linter
- `ruff check` zeigt unbenutzte Imports
- `svelte-check` und `eslint`

**Konsistenz**:
- README's pro Service vorhanden und aktuell?
- Alle Make-Targets dokumentiert?
- Alle ENV-Variablen in `.env.example`?
- Alle Container-Image-Namen konsistent?
- Alle "<org>"-Platzhalter ersetzt?

### 10. Lizenz-Datei

Wenn Maintainer entschieden hat:
- `LICENSE` im Root anlegen mit dem entsprechenden Text
- Falls Mixed-License: `LICENSE` für Standard, zusätzliche `LICENSE-*` für
  abweichende Verzeichnisse, Erwähnung in den jeweiligen READMEs

## Vorgehen (verbindlich)

1. Plan zeigen, **inkl. Liste der zu konvertierenden TODOs**
2. Lizenz-Frage mit Maintainer klären
3. Freigabe abwarten
4. Implementieren in Schritten:
   a) README.md überarbeiten
   b) Architektur-Doku mit Mermaid
   c) Development-, Deployment-, Runbook-Doku
   d) ADRs 0002-0005
   e) CONTRIBUTING.md
   f) CLAUDE.md-Update
   g) Cleanup: TODOs, Dead Code, Konsistenz, "<org>"-Replacements
   h) LICENSE-Datei
5. Diff zeigen, viele kleine Änderungen — am besten in mehreren Tranchen
6. **Liste der erstellten/zu erstellenden GitHub-Issues** ausgeben (du machst sie nicht
   selbst, der Maintainer)
7. Nicht committen

## Erfolgs-Kriterien

- [ ] Top-Level README enthält Quickstart, Tech-Stack, Verweise auf Docs
- [ ] `docs/architecture.md` mit Mermaid-Diagrammen
- [ ] `docs/development.md` mit Schritt-für-Schritt für häufige Tasks
- [ ] `docs/deployment.md` vollständig (Erst-Deployment, Folge-Deployments, Rollback)
- [ ] `docs/runbook.md` mit mindestens 8 konkreten Szenarien
- [ ] ADRs 0001-0005 vorhanden, MADR-konform
- [ ] `CONTRIBUTING.md` vorhanden
- [ ] `CLAUDE.md` aktualisiert mit aktuellen Kommandos und gelösten Fragen
- [ ] Alle "<org>"-Platzhalter ersetzt
- [ ] Alle TODO-Marker bewertet, Liste der Issues ausgegeben
- [ ] `LICENSE`-Datei vorhanden
- [ ] `make lint` grün
- [ ] `make test` grün
- [ ] `make build` grün

## Mögliche Stolpersteine

- **Mermaid in Markdown**: GitHub rendert es nativ. In manchen statischen-Site-
  Generatoren braucht man ein Plugin — irrelevant fürs Repo, dort ist GitHub
  der primäre Renderer.
- **Lizenz-Wahl**: nicht trivial, hat langfristige Folgen. Wenn unsicher: AGPL-3.0
  als Default, kann später relizenziert werden solange du allein Autor bist.
- **TODO-Issues**: schnell entstehen 20+ Issues. Pragmatisch in 5 Hauptthemen
  bündeln, nicht jeden TODO einzeln.
- **Doku-Drift**: nach dieser Session bleibt Doku nur aktuell, wenn sie bei
  Änderungen mitgezogen wird. CLAUDE.md macht das zur Regel.

## Was diese Session NICHT tut

- Keine Code-Änderungen außer Cleanup
- Keine neuen Features
- Keine Performance-Optimierungen
- Kein Sicherheits-Audit (separater Schritt vor Production-Launch)

## Suggested Commit-Message

Diese Session erzeugt mehrere kleinere Commits, idealerweise:

```
docs: write architecture, development, and deployment guides
docs: add ADRs 0002-0005
docs(contributing): add contribution guide
docs(claude): finalize CLAUDE.md with current commands and resolved questions
chore: cleanup TODOs and dead code
chore: add LICENSE
```

Oder, wenn als ein Block: `docs: finalize project documentation and update CLAUDE.md`.

---

## Nach Session 12 — Du bist hier fertig mit der Initial-Phase

Der nächste Schritt ist nicht mehr Infrastruktur, sondern Produkt:
- Erste Datenquelle (Open-Meteo) integrieren
- Locations-Suche real machen (Geocoding, DB-Schema, Endpoint, UI)
- Authentifizierung
- Erste Karten-Komponente
- ...

Diese Feature-Arbeit unterscheidet sich vom DevOps-Setup: kürzere Sessions,
TDD-orientiert, weniger Tooling-fokussiert. Eine eigene Session-Struktur dafür
bauen wir auf, wenn du da angekommen bist.

Bis dahin: respect für die Disziplin, ein solides Fundament zu legen, bevor
das eigentliche Bauen beginnt. Genau dieser Schritt entscheidet später,
ob das Projekt skaliert oder am eigenen technischen Schulden ersticken.
````
