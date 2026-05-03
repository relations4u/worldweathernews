# worldweathernews.com

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

## Dokumentation

- [Architektur](docs/architecture.md)
- [Entwicklung](docs/development.md)
- [Deployment](docs/deployment.md)
- [Runbook](docs/runbook.md)
- [Spielregeln für Claude Code](CLAUDE.md)

## Lizenz

All rights reserved. Eine endgültige Lizenz-Entscheidung wird in Session 12
getroffen (siehe [`sessions/step12.md`](sessions/step12.md)).
