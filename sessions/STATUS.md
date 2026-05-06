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

## Session 10b — Caddy live auf wwn-prod

Status: ✅ Done
Datum: 2026-05-06
Commit: 54f1bf5 + 5c41fea (direkt auf main)
Notizen: Caddy als eigenständiger Compose-Stack unter `/srv/wwn/caddy`
mit `network_mode: host` deployed. Nicht Teil des App-Stacks
(`compose.prod.yml`) — bewusste Trennung für unverfälschte Client-IPs
und unabhängige Lifecycle. Vier Let's-Encrypt-Zertifikate ausgestellt
(Apex, www, research, api.research) via HTTP-01-Challenge. HSTS
`max-age=31536000` ohne `includeSubDomains` (Begründung: zukünftige
interne Subdomains evtl. lange ohne TLS, Lockout-Risiko). Stub-Antworten
via `respond` 200 — Cutover auf `reverse_proxy` ist Phase von Session 11a.
Deploy-Pfad: `infra/deploy/deploy-caddy.sh` (rsync + ssh docker compose).
Snapshot `caddy-online` gesetzt nach grünem End-to-End-Test.
Stolperstein dokumentiert: `sudo` über plain SSH (BatchMode) braucht
Terminal — Skript prüft jetzt Verzeichnis-Existenz statt `sudo install`
direkt zu rufen. Zielverzeichnis `/srv/wwn/caddy` einmalig manuell mit
`ssh -t` vorzubereiten.

## Session 11 — Ansible, SOPS, Terraform-Skelett

Status: ✅ Done (Skelett gemerged; Server-Deployment via Session 11a)
Datum: 2026-05-06
Commit: PR #22 (Skelett) + PR #23 (Caddy-Block-Cleanup)
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
Folge-PR #23 hat den verwaisten Caddy-Block aus
`infra/compose/compose.prod.yml` entfernt.

## Session 11a — Komplettes Deployment auf wwn-prod und wwn-mon

Status: ✅ Done
Datum: 2026-05-06
Commits: PR #25 (bind-mount migration), PR #26 (monitoring-stack role),
PR #27 (ansible.cfg yaml callback), PR #28 (site.yml role tags),
PR #29 (Caddy cutover, UFW LAN-scrape, Grafana 0.0.0.0),
PR #30 (agent scope, deploy plumbing, restart handler),
PR #31 (frontend healthcheck IPv4), PR #32 (Grafana provisioning perms,
backend CORS env, Caddy OPTIONS pass-through, release.yml frontend
build-arg), Tag v0.0.1-rc4.

Notizen: Sechs Phasen abgearbeitet. Phase 1 — Ansible-Bootstrap auf
wwn-prod und wwn-mon (deploy-User, SSH-Hardening, ufw, Docker, GHCR-Login).
Phase 2 — Caddy von Docker-Named-Volume zu Bind-Mount unter
`/srv/wwn/caddy/data` migriert; alle vier LE-Certs überlebten unverändert
(Tar-Backup als Sicherheitsnetz, `notBefore`-Diff als Verifikation).
Phase 3 — Neue Rolle `monitoring-stack` deployt Prometheus/Loki/Tempo/
Grafana auf wwn-mon. Phase 4 — App-Stack v0.0.1-rc2 → später rc3 → rc4
auf wwn-prod via `playbooks/deploy.yml`. Phase 5 — Caddy-Cutover von
Stub-`respond` auf `reverse_proxy` 127.0.0.1:{3000,8080}; Cert-Dates
unverändert. Phase 6 — End-to-End-Verifikation grün.

Stolpersteine, die als Code-Fix im Repo gelandet sind:
(a) Fine-grained PAT lacked Org-level Packages: Read → ghcr.io 403;
auf Classic PAT mit `read:packages` umgestellt.
(b) Single-File-Bind-Mount-Inode-Falle bei Caddy und monitoring-stack:
rsync/copy macht atomic-rename, Container liest weiterhin den alten
Inode → `docker compose restart` nach Config-Update nötig (jetzt im
deploy-caddy.sh und als Ansible-Handler in monitoring-stack).
(c) UFW blockierte Prometheus-Scrapes (9101) → neue
`monitoring_scrape_ports`-Liste in der common-Rolle, LAN-only.
(d) monitoring-agent-Rolle kollidierte auf wwn-mon mit dem
Stack-eigenen Promtail (gleicher Container-Name) → site.yml limitiert
den Agent jetzt auf `app`-Hosts.
(e) Grafana (uid 472) konnte sein provisioning-Verzeichnis nicht
traversieren — `grafana/` und `grafana/provisioning/` mussten explizit
mit Mode 0755 in den Directory-Loop, sonst implizite Eltern-Dirs mit
umask-Default 0750.
(f) Backend-CORS war auf Default `http://app.localhost`; Apex-Frontend
bekam „Backend offline / Failed to fetch" — `WWN_HTTP_CORSORIGINS`
in backend.env gesetzt für die drei aktiven Origins.
(g) Caddy fing OPTIONS-Preflights mit bare 204 ohne CORS-Header ab →
chi-cors auf Backend-Seite kam nicht zum Zug. Caddy-Block entfernt.
(h) Release-Workflow gab `PUBLIC_API_BASE_URL` nicht als build-arg an
den Frontend-Build → JS-Bundle hatte Dockerfile-Default
`http://api.localhost` eingebacken. Per-Service `build_args` in der
Matrix in release.yml ergänzt.

scripts/deploy.sh läuft jetzt mit `-e ansible_user=hwr` (deploy-User
hat nur docker-NOPASSWD), monitoring-stack-Rolle hat einen
`Restart monitoring stack`-Handler bei Config-Änderungen. Public-Smoke
gegen apex/www/research/api.research alle 200, Grafana-Dashboards sind
unter `worldweathernews/`-Folder provisioniert.

Bekannte offene Punkte (nicht Session-blockierend, in
`prometheus.yml` als Comment dokumentiert):

- Backend-/Pyworkers-`/metrics`-Ports binden 127.0.0.1 only —
  Prometheus auf wwn-mon kann sie nicht scrapen. Entscheidung
  zwischen LAN-Bind+ufw vs. Push-Sidecar steht aus.
- node-exporter für wwn-mon nicht im Stack — als Folge-PR, wenn
  Host-Metriken für Grafana benötigt werden.

## Session 12 — Dokumentation finalisieren

Status: ✅ Done
Datum: 2026-05-06
Commits: PR #33 (README polish, CONTRIBUTING, status updates),
PR #34 (architecture.md mit Mermaid + development.md),
PR #35 (deployment.md + runbook.md mit 10 Szenarien),
PR #36 (ADRs 0002–0005, AGPL-3.0 LICENSE, docs/backlog.md, TODO-Triage).

Notizen: Vier Tranchen statt einem Block — README/CONTRIBUTING zuerst,
dann die zwei Stub-Files (architecture/development), dann die zwei
Operations-Files (deployment/runbook), zuletzt ADRs+License+Backlog.

Inhaltliche Schwerpunkte: alle Docs reflektieren den **Ist-Zustand
nach Session 11a**, nicht den theoretischen Endzustand. runbook.md hat
das 4-Optionen-Diagnose-Pattern für „Backend offline / Failed to
fetch" als Szenario 2 — exakt die Reihenfolge, die Session 11a
abgearbeitet hat (CORS-env, Caddy-OPTIONS, Frontend-Bundle-URL,
Browser-Cache), damit beim nächsten Mal kein Re-Discovery nötig ist.
deployment.md dokumentiert die Bind-Mount-Inode-Falle und den
deploy-User-NOPASSWD-Scope expliziert. architecture.md hat ein
Mermaid-Diagramm mit den drei Caveats (Backend-Metrics 127.0.0.1-only,
mon-node-exporter-Lücke, SvelteKit-`PUBLIC_*`-Build-Time-Pinning).

ADR-Set komplett auf MADR (matching 0001):
0002 Go-für-Backend, 0003 Monorepo, 0004 Compose-vor-K3s,
0005 SOPS+age. Jeweils mit Alternatives-Considered-Block.

LICENSE: kanonischer AGPL-3.0-Text von gnu.org. Begründung im
README + CONTRIBUTING. Maintainer ist Sole-Author während der
Setup-Phase, Re-Lizenzierung später möglich.

TODO-Triage: zwei in-context-TODOs auf docs/backlog.md
umgeschrieben (i18n-Library-Wahl, Monitoring-Stack-UFW-Folge-PR);
vier Marker behalten, weil der Text die Begründung der Stelle ist
(tracing.go AlwaysSample, Hetzner-Migrations-Stub, Caddy-Admin-
Metrics-Stub, Backend-Metrics-Scrape-Lücke). Alle haben einen
Eintrag in docs/backlog.md als Folge-Tracker.

step12.md-Erfolgs-Kriterien: alle ✅. Damit ist die Initial-
Setup-Phase formal abgeschlossen — der nächste Schritt ist
Feature-Arbeit (Open-Meteo, Locations-Suche, Auth, Maps), nicht
mehr Infrastruktur.
