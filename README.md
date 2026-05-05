# worldweathernews.com

![CI Backend](https://github.com/relations4u/worldweathernews/actions/workflows/ci-backend.yml/badge.svg)
![CI Frontend](https://github.com/relations4u/worldweathernews/actions/workflows/ci-frontend.yml/badge.svg)
![CI PyWorkers](https://github.com/relations4u/worldweathernews/actions/workflows/ci-pyworkers.yml/badge.svg)
![Release](https://github.com/relations4u/worldweathernews/actions/workflows/release.yml/badge.svg)
[![Latest Release](https://img.shields.io/github/v/release/relations4u/worldweathernews?include_prereleases&sort=semver)](https://github.com/relations4u/worldweathernews/releases)

Wetter- und Klima-Plattform mit Community-Features. Self-hosted Monorepo.

> **Status: WIP — initial setup phase.**
> Befehle und Services werden schrittweise eingeführt
> (siehe [`sessions/`](sessions/STATUS.md)).

## Quickstart (Vorschau)

Nicht alle Befehle existieren bereits. Sie werden in den Setup-Sessions 1–12
implementiert.

```bash
mise install        # Toolchain (Go, Node, Python, ...) installieren
make bootstrap      # Erst-Setup nach Repo-Clone
make dev            # Lokale Entwicklung starten
```

`mise` ist Voraussetzung. Installation:

- macOS: `brew install mise`
- Linux: `curl https://mise.run | sh`

Danach `eval "$(mise activate zsh)"` (bzw. `bash`) in die Shell-Init aufnehmen.

## API-Schema

Die HTTP-API ist als OpenAPI 3.1 in [`packages/api-schema/openapi.yaml`](packages/api-schema/openapi.yaml)
definiert (siehe [ADR-0001](docs/adr/0001-openapi-as-source-of-truth.md)).
Aus dem Schema werden Go-Server-Stubs und TypeScript-Types generiert:

```bash
make gen          # regeneriert beide Outputs
make gen-check    # CI: prüft, dass das Generierte aktuell ist
```

## Dokumentation

- [Architektur](docs/architecture.md)
- [Entwicklung](docs/development.md)
- [Deployment](docs/deployment.md)
- [Runbook](docs/runbook.md)
- [Architecture Decision Records](docs/adr/)
- [Spielregeln für Claude Code](CLAUDE.md)

## Releases

Container-Images werden auf [ghcr.io](https://github.com/relations4u/worldweathernews/pkgs/container/wwn-backend)
veröffentlicht. Ein neuer Release wird durch einen Tag im Format `v*` ausgelöst:

```bash
make release      # interaktiv: bump auswählen, signed Tag erzeugen, pushen
```

Die [Release-Pipeline](.github/workflows/release.yml) baut alle drei Service-Images,
signiert sie mit cosign (keyless via Sigstore), erzeugt SBOMs (Syft), scannt mit
Trivy und legt einen GitHub-Release mit auto-generierten Notes (git-cliff) an.

## Lizenz

All rights reserved. Eine endgültige Lizenz-Entscheidung wird in Session 12
getroffen (siehe [`sessions/step12.md`](sessions/step12.md)).
