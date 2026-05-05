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

Status: ✅ Done
Datum: 2026-05-03
Commit: <SHA>
Notizen: apscheduler 3.x (AsyncIOScheduler) statt 4.x-Alpha.
mypy strict — Overrides für asyncpg + apscheduler (kein py.typed).
Heartbeat alle 30s im Container, Metriken auf :9100.
Production-Image 250 MB (über 200 MB-Soft-Target; Optimierung später).
Compose nutzt python:3.12-slim + pip install uv beim Start.

## Session 7 — OpenAPI-Schema und Type-Generation

Status: ✅ Done
Datum: 2026-05-03
Commit: <SHA>
Notizen: OpenAPI 3.1 als SoT in packages/api-schema/.
oapi-codegen v2.4.1 (warnt bei 3.1, funktioniert), openapi-typescript 7.10.
pnpm-workspace ergänzt um packages/api-schema; allowBuilds true für
core-js/protobufjs; .npmrc mit verify-deps-before-run=false (pnpm-11-Quirks).
/api/v1/locations validiert required-q automatisch (400).
ADR-0001, scripts/check-generated.sh für CI.

## Session 8 — GitHub Actions CI-Workflows

Status: ✅ Done
Datum: 2026-05-04
Commit: <SHA>
Notizen: Workflows in zwei Vorab-Commits gelandet (fac75c9, 90c4078); diese
Session ergänzt das Drumherum (.commitlintrc.yaml, README-Badges) und fixt
zwei Lücken aus früheren Sessions, die in CI auffliegen würden:
fehlende lint/test-Scripts in apps/frontend/package.json (Session 5) und
fehlendes pip-audit-Dev-Dep in apps/pyworkers (Session 6).
Frontend- und PyWorkers-CI-Jobs fassen Schritte zusammen statt separater
install/lint/check/test-Jobs aus step08.md — bewusste Optimierung,
Caching greift trotzdem über setup-node/setup-uv.
Branch-Protection wird direkt in der GitHub-UI gepflegt, nicht im Repo
dokumentiert (docs/deployment.md bleibt deshalb beim TODO-Platzhalter).
Live-Verifikation der Workflows erst nach Push einer Test-PR möglich.

## Session 9 — Release-Workflow und Container-Registry

Status: ✅ Done
Datum: 2026-05-05
Commit: 0c39b83 (initial pipeline) + edb42b8 (SARIF-fix)
Notizen: Tag-getriggerte Release-Pipeline für alle drei Services. amd64-only
(arm64 dropped — kein ARM-Hosting geplant). Cosign keyless via Sigstore,
Syft-SBOM, Trivy-Scan; SBOMs als Release-Asset, Trivy-SARIF als Workflow-
Artefakt (NICHT zur Code-Scanning-Tab — Repo ist privat ohne GHAS, der
github/codeql-action/upload-sarif-Endpoint ist gesperrt). git-cliff für
auto-generierte Release-Notes, softprops/action-gh-release als Publisher.
Caddy `rate_limit` weggelassen (nicht in caddy:2-alpine, TODO im Caddyfile).
Live-Test mit v0.0.1-rc1 (rot, SARIF-Upload-Fehler), Fix in PR #19, danach
v0.0.1-rc2 komplett grün: 3 signierte Images, 3 SBOMs, GitHub-Release als
prerelease (auto-detected via "-"). cosign verify nicht lokal getestet —
GHCR-Pakete sind privat, Pull-Time-Verifikation gehört zu Session 11
(Ansible). Erfahrungswert: kompletter Release-Run dauert ~3 Min.

## Session 10 — Observability-Stack lokal

Status: ⏸ Pending

## Session 11 — Ansible, SOPS, Terraform-Skelett

Status: ⏸ Pending

## Session 12 — Dokumentation finalisieren

Status: ⏸ Pending
