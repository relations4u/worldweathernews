# Session-Status

Pflege diese Datei am Ende jeder Session. Format pro Session:

```
## Session N — Titel
Status: 🟡 In Progress | ✅ Done | ❌ Blocked | ⏭ Skipped
Datum: YYYY-MM-DD
Commit: <kurzer SHA>
Notizen: <was lief gut, was offen blieb, blockierende Fragen>
```

---

## Session 1 — Repo-Skelett und Tooling

Status: ✅ Done
Datum: 2026-05-03
Commit: <SHA>
Notizen: README.md und STATUS.md von Root nach sessions/ verschoben.
sqlc + goose über mise ubi:-Backend (experimental aktiv).
Lizenz weiterhin offen (Session 12).

## Session 2 — Pre-commit, Makefile, lokale Workflows

Status: ✅ Done
Datum: 2026-05-03
Commit: <SHA>
Notizen: pre-commit via pipx in .mise.toml.
prettier-Hook auskommentiert, wird in Session 5 aktiviert.
golangci-lint als system-Hook (conditional auf apps/backend/\*.go).
pre-session-checklist.md vom EOF-fixer mitgepatcht.

## Session 3 — Compose-Stack mit DB und Redis

Status: ✅ Done
Datum: 2026-05-03
Commit: <SHA>
Notizen: TimescaleDB-HA-Image enthält PostGIS+TimescaleDB out-of-the-box.
Caddyfile musste auf http://-Schema umgestellt werden, sonst
bindet Caddy mit auto_https off auf :443.
mailhog läuft als amd64 unter Rosetta (mailpit als Alternative).
Stack-Smoke-Tests alle grün.

## Session 4 — Go-Backend-Skelett

Status: ✅ Done
Datum: 2026-05-03
Commit: <SHA>
Notizen: Modul github.com/relations4u/worldweathernews/apps/backend.
go.mod auto-bumped auf 1.25 (Viper transitive). Dockerfile-Builder
entsprechend angepasst.
golangci-lint v2 — .golangci.yml migriert.
Viper AutomaticEnv brauchte leere Defaults für DB/Redis-URL.
distroless-final-Image: 26.2 MB. Alle Endpunkte grün.

## Session 5 — SvelteKit-Frontend-Skelett

Status: ✅ Done
Datum: 2026-05-03
Commit: <SHA>
Notizen: SvelteKit 2 + Svelte 5 Runes + Tailwind v3 + shadcn-svelte (Badge
manuell, CLI hätte interaktiv gehängt).
Frontend-Container muss als root laufen (node:22-alpine läuft
default als uid=1000, kein Schreibrecht auf /usr/local/lib).
Production-Image bleibt non-root.
Prettier-Hook im pre-commit aktiviert.
Lokal alle Checks grün, Container liefert Hero unter app.localhost.

## Session 6 — Python-Workers-Skelett

Status: ⏸ Pending

## Session 7 — OpenAPI-Schema und Type-Generation

Status: ⏸ Pending

## Session 8 — GitHub Actions CI-Workflows

Status: ⏸ Pending

## Session 9 — Release-Workflow und Container-Registry

Status: ⏸ Pending

## Session 10 — Observability-Stack lokal

Status: ⏸ Pending

## Session 11 — Ansible, SOPS, Terraform-Skelett

Status: ⏸ Pending

## Session 12 — Dokumentation finalisieren

Status: ⏸ Pending
