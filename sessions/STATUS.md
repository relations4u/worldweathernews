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

Status: ✅ Done
Datum: 2026-05-05
Commit: <SHA>
Notizen: Compose-Profile `monitoring` mit pinned Images
(prom/prometheus:v2.55.1, grafana/grafana:11.3.0, grafana/loki:3.3.2,
grafana/promtail:3.3.2, grafana/tempo:2.6.1) — `make dev-full` startet alles,
`make dev` lässt Monitoring bewusst aussen vor. OTel mit konkreten Pins:
Backend otel/sdk@v1.34.0 + otelchi@v0.12.3 + contrib/otelhttp@v0.59.0;
Pyworkers opentelemetry-{api,sdk,exporter-otlp-proto-grpc}~=1.29.0 +
instrumentation-{asyncpg,redis,httpx}~=0.50b0. Trace-IDs landen in slog/
structlog, Loki-DerivedField verlinkt auf Tempo. Tracing graceful-degraded:
fehlt Tempo, retrydet OTLP still im Hintergrund; App läuft ohne Trace-Sink
weiter. Heartbeat-Job bekam manuellen Span — sonst keine Traces, weil
Auto-Instrumentation nur DB/HTTP/Redis-Calls greift. Live-Verifikation:
Tempo sieht Backend- und Pyworkers-Traces, Loki indexiert JSON-Logs mit
trace_id-Label, Cross-Lookup Loki→Tempo via Trace-ID erfolgreich. Volle
3 Dashboards (Backend, Pyworkers, Infra-Stub mit TODO-Roadmap).
.env-Format auf `json` umgestellt (Default), text-Empfehlung im Comment.
mypy strict erforderte separaten Override für pyworkers.observability.tracing
(disallow_untyped_calls=false), weil OTel-Instrumentor-Constructor untyped sind.

## Session 11 — Ansible, SOPS, Terraform-Skelett

Status: ✅ Done
Datum: 2026-05-06
Commit: <SHA>
Notizen: Skelett, kein Server-Provisioning. Ablauf gegenüber step11.md
angepasst, weil das Hosting auf Proxmox (wwn-prod 10.100.100.21,
wwn-mon 10.100.100.22) statt Hetzner liegt — Hetzner-Modul bleibt als
Migrations-Stub. `bpg/proxmox` ~> 0.66 für die aktive Pipeline.
SOPS via age, Public-Key in `.sops.yaml`; sechs verschlüsselte Demo-
Secret-Files unter `infra/secrets/production/` (backend, frontend,
pyworkers, postgres, ghcr, proxmox). Pre-commit-Hook
`forbid-unencrypted-secrets` blockt Plaintext-Files unter `infra/secrets/`.
Ansible: vier Rollen (`common`, `docker`, `app`, `monitoring-agent`),
drei Playbooks (`site`, `deploy`, `rollback`). `app`-Rolle deployt
Postgres/Redis/Backend/Frontend/Pyworkers — **Caddy bewusst NICHT**;
Caddy bleibt unter `/srv/wwn/caddy` mit `infra/deploy/deploy-caddy.sh`
gemanaged. Inventory startet mit `ansible_user=deploy`; Bootstrap-Pfad
für wwn-prod via `-e ansible_user=hwr` dokumentiert. Tooling-Pins
ergänzt: `terraform`, `ubi:getsops/sops`, `pipx:ansible-core`,
`pipx:ansible-lint`. Validierung lokal: ansible-lint, --syntax-check,
terraform fmt -check, terraform validate -backend=false. Hausaufgabe
für Maintainer (in README dokumentiert): `terraform import` der zwei
manuell erstellten VMs, bevor jemals `terraform apply` läuft.
Caddy-Block in `infra/compose/compose.prod.yml` wird in eigener
Folge-PR entfernt (separat von Session 11).

## Session 12 — Dokumentation finalisieren

Status: ⏸ Pending
