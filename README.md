# worldweathernews.com

![CI Backend](https://github.com/relations4u/worldweathernews/actions/workflows/ci-backend.yml/badge.svg)
![CI Frontend](https://github.com/relations4u/worldweathernews/actions/workflows/ci-frontend.yml/badge.svg)
![CI PyWorkers](https://github.com/relations4u/worldweathernews/actions/workflows/ci-pyworkers.yml/badge.svg)
![Release](https://github.com/relations4u/worldweathernews/actions/workflows/release.yml/badge.svg)
[![Latest Release](https://img.shields.io/github/v/release/relations4u/worldweathernews?include_prereleases&sort=semver)](https://github.com/relations4u/worldweathernews/releases)
[![License: AGPL v3](https://img.shields.io/badge/License-AGPL_v3-blue.svg)](LICENSE)

Eine self-hosted Plattform für Wetter- und Klima-Beobachtungen, die Daten aus
nationalen Wetterdiensten weltweit aggregiert, Klimadaten visualisiert und eine
Community von Beobachtern, Citizen Scientists und Wetter-Interessierten verbindet.

> **Status:** Forschungs-Phase. Plattform-Skelett und Infrastruktur stehen,
> das Bauen der eigentlichen Features beginnt jetzt.
> Live unter <https://research.worldweathernews.com>.

## Quickstart

Voraussetzungen: Docker, Docker Compose und [mise](https://mise.jdx.dev/) für
die Toolchain (Go 1.25, Node 22 + pnpm 9, Python 3.12 + uv).

```bash
git clone git@github.com:relations4u/worldweathernews.git
cd worldweathernews
mise install         # Toolchain installieren
make bootstrap       # Pre-commit-Hooks, Workspace-Deps, Codegen
cp apps/backend/.env.example apps/backend/.env  # ggf. anpassen
make dev             # App-Stack ohne Monitoring
make dev-full        # App-Stack inkl. Prometheus/Grafana/Loki/Tempo
```

Danach erreichbar:

- App: <http://app.localhost>
- API: <http://api.localhost>
- Mailhog: <http://mail.localhost>
- Grafana: <http://grafana.localhost> (mit `make dev-full`)

`mise` einrichten falls noch nicht vorhanden:

- macOS: `brew install mise`
- Linux: `curl https://mise.run | sh`

…dann `eval "$(mise activate zsh)"` (bzw. `bash`) in die Shell-Init aufnehmen.
Details siehe [`vm-setup.md`](vm-setup.md).

## Tech-Stack

| Schicht       | Technologie                              |
| ------------- | ---------------------------------------- |
| Backend       | Go 1.25, Chi, sqlc, pgx                  |
| Workers       | Python 3.12, asyncio, asyncpg, structlog |
| Frontend      | SvelteKit, TypeScript, Tailwind, shadcn  |
| Datenbank     | PostgreSQL 16 + PostGIS + TimescaleDB    |
| Cache         | Redis 7                                  |
| Reverse Proxy | Caddy 2                                  |
| Container     | Docker + Docker Compose (lokal & prod)   |
| Monitoring    | Prometheus + Grafana + Loki + Tempo      |
| Errors/Trace  | OpenTelemetry → Tempo                    |
| CI/CD         | GitHub Actions, ghcr.io                  |
| Deployment    | Ansible, SOPS+age, Terraform             |

## Repo-Struktur

```
apps/
  backend/        Go-API (Chi + sqlc + pgx)
  frontend/       SvelteKit + Svelte 5 Runes
  pyworkers/      Python-Workers (asyncio)
packages/
  api-schema/     OpenAPI 3.1 — Single Source of Truth
  shared-types/   Generierte TS-Types
infra/
  compose/        Docker-Compose-Files (dev & prod)
  caddy/          Reverse-Proxy-Config (Caddyfile)
  monitoring/     Prometheus/Grafana/Loki/Tempo (dev)
  ansible/        Server-Konfiguration & Deployment
  terraform/      VM-Provisionierung (Proxmox jetzt, Hetzner-Stub)
  secrets/        SOPS-verschlüsselte ENV-Files
  migrations/     goose-DB-Migrations
  deploy/         Stand-alone Deploy-Scripts (Caddy)
docs/             Architektur, Runbook, ADRs
sessions/         Setup-Session-Dokumente
.github/workflows/  CI/CD
```

## API-Schema

Die HTTP-API ist als OpenAPI 3.1 in
[`packages/api-schema/openapi.yaml`](packages/api-schema/openapi.yaml) definiert
(siehe [ADR-0001](docs/adr/0001-openapi-as-source-of-truth.md)). Aus dem Schema
werden Go-Server-Stubs und TypeScript-Types generiert:

```bash
make gen          # regeneriert beide Outputs
make gen-check    # CI: prüft, dass das Generierte aktuell ist
```

## Releases

Container-Images werden auf
[ghcr.io](https://github.com/relations4u/worldweathernews/pkgs/container/wwn-backend)
veröffentlicht. Ein neuer Release wird durch einen Tag im Format `v*` ausgelöst:

```bash
make release      # interaktiv: bump auswählen, signed Tag erzeugen, pushen
```

Die [Release-Pipeline](.github/workflows/release.yml) baut alle drei
Service-Images, signiert sie mit cosign (keyless via Sigstore), erzeugt
SBOMs (Syft), scannt mit Trivy und legt einen GitHub-Release mit
auto-generierten Notes (git-cliff) an.

Production-Deploys laufen via Ansible:

```bash
bash scripts/deploy.sh production 0.1.0
```

Details: [`docs/deployment.md`](docs/deployment.md).

## Dokumentation

- [Architektur](docs/architecture.md) — System-Diagramm, Service-Verantwortlichkeiten, Datenflüsse
- [Entwicklung](docs/development.md) — Wie ich Feature X hinzufüge, Migrations, Workers
- [Deployment](docs/deployment.md) — Erst-Deployment, Folge-Deploys, Rollback
- [Runbook](docs/runbook.md) — Was tun wenn X kaputt ist?
- [Secrets](docs/secrets.md) — SOPS-Workflow
- [Architecture Decision Records](docs/adr/)
- [Contributing](CONTRIBUTING.md) — Setup, Branching, Commits, Reviews
- [CLAUDE.md](CLAUDE.md) — Spielregeln für KI-gestützte Entwicklung

## Mitwirken

Pull Requests sind willkommen. Vor dem ersten PR: kurz
[`CONTRIBUTING.md`](CONTRIBUTING.md) lesen — der Workflow ist
trunk-basiert, mit signierten Commits und grünem CI als Gate. Sicherheits-
Probleme bitte privat melden, nicht über öffentliche Issues.

## Lizenz

[GNU Affero General Public License v3.0](LICENSE) (AGPL-3.0).

Wenn du eine modifizierte Version dieser Software über ein Netzwerk
betreibst (z. B. als gehosteter Service), musst du die geänderten
Quellen zugänglich machen. Wer den Code privat oder lokal nutzt, ist
nicht verpflichtet, ihn zu veröffentlichen.
