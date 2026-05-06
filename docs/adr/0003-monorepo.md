# 3. Monorepo statt Multi-Repo

Date: 2026-05-03
Status: Accepted

## Context

Die Plattform besteht aus mindestens drei Services (Backend, Frontend,
Pyworkers), zwei geteilten Paketen (api-schema, shared-types) und
Infrastruktur-Code (Compose, Caddy, Ansible, Terraform, Migrations).
Diese Komponenten sind eng gekoppelt: ein API-Schema-Change führt
gleichzeitig zu Codegen im Backend und Frontend, ein
Migrations-Update braucht ggf. eine Backend-Anpassung.

## Decision

Wir halten alles in **einem Repo** (`relations4u/worldweathernews`).
Layout siehe README → Repo-Struktur. Toolchain-Pins liegen in der
Wurzel (`.mise.toml`), pnpm und Go nutzen Workspaces.

## Consequences

**Positiv**:

- **Atomare Cross-Service-Änderungen** — eine PR kann Schema, Backend-
  Handler, Frontend-Aufruf und Test gleichzeitig ändern. Keine
  Koordinations-Probleme zwischen N Repos.
- Einfachere CI-Setups — eine Workflow-Datei pro Service, gemeinsame
  Tools (commitlint, prettier, ruff, mise).
- Versions-Pinning lebt an einer Stelle. Drift zwischen Services ist
  sichtbar.
- Generated Code (`api.gen.go`, `types.gen.ts`) lebt neben seiner
  Quelle (`openapi.yaml`).
- Code-Review umfasst die ganze Änderung, nicht nur einen Teil.
- Docs/ADRs/Runbook leben mit dem Code.

**Negativ**:

- Repo wächst über Zeit; CI-Trigger müssen pfad-basiert filtern
  (Frontend-CI nicht für Backend-Änderungen feuern lassen) — bereits
  implementiert via `paths:` in den Workflows.
- Globaler `git log` zeigt Änderungen aller Services. Per-Service-
  History ist via `git log -- apps/<service>/` erreichbar.
- Repo-Klon-Größe wächst. Aktuell unter 50 MB, perspektivisch noch
  unkritisch. Bei Bedarf: shallow clone in CI.
- Externer Zugriff auf einzelne Services schwerer steuerbar
  (Visibility ist all-or-nothing pro Repo). Aktuell privat, also
  nicht relevant.

## Alternatives Considered

- **Multi-Repo pro Service** (klassisch) — entkoppelt Versionierung,
  aber Cross-Service-Änderungen werden 3 PRs in 3 Repos mit manueller
  Reihenfolge. Für ein Solo-Setup reines Reibungswachstum.
- **Polyrepo + Submodules** — verbindet die Welten, aber Submodules
  sind ein eigener Schmerz und hilft kaum bei atomaren Änderungen.
- **Monorepo + nx/Turborepo** — die Tools wären nice-to-have für
  inkrementelle Builds, aber unsere Build-Zeiten sind im
  Single-Digit-Minutenbereich, plus die Komplexität dieser Tools
  zahlt sich erst bei vielen Paketen aus.
