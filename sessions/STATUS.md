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
           golangci-lint als system-Hook (conditional auf apps/backend/*.go).
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
Status: ⏸ Pending

## Session 5 — SvelteKit-Frontend-Skelett
Status: ⏸ Pending

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
