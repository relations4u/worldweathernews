# Runbook

<!-- TODO: Wird in einer späteren Session strukturiert befüllt (siehe sessions/step12.md). -->

## Wo finde ich was im Monitoring?

Der Observability-Stack startet nur unter dem Compose-Profile `monitoring`,
also über `make dev-full` (oder `make dev-monitoring`, falls die Apps schon
laufen). `make dev` lässt ihn bewusst aussen vor, damit der schnelle
Inner-Loop schlank bleibt.

| Service      | URL                   | Anmeldung                                  |
| ------------ | --------------------- | ------------------------------------------ |
| Grafana      | http://localhost:3000 | `admin` / `admin` (Anonymous Viewer aktiv) |
| Prometheus   | http://localhost:9090 | —                                          |
| Loki direkt  | http://localhost:3100 | (UI ist Grafana → Explore)                 |
| Tempo direkt | http://localhost:3200 | (UI ist Grafana → Explore)                 |
| OTLP gRPC    | `tempo:4317` (intern) | für Backend/Pyworkers                      |

Drei vorprovisionierte Dashboards liegen unter `worldweathernews/`:
**WWN Backend Overview**, **WWN Pyworkers Overview**, **WWN Infra Overview**.

### Häufige Aufgaben

**Backend hat hohe Latency**

1. Grafana → _WWN Backend Overview_ → Panel "HTTP Latency p50/p95/p99".
2. Wenn auffällig: Tempo → Search → Service `wwn-backend`, sortiert nach
   Duration, langsame Traces öffnen.
3. Trace-ID notieren, in Loki Explore filtern: `{service="backend"} |= "<trace-id>"`.

**Worker-Job ist fehlgeschlagen**

1. Grafana → _WWN Pyworkers Overview_ → Panel "Job Run Rate (by status)".
2. Loki: `{service="pyworkers"} | json | level="error"`.
3. Aus dem Log-Eintrag heraus auf das Trace-ID-DerivedField klicken — springt
   nach Tempo, Auto-Instrumentation für asyncpg/redis/httpx zeigt DB- und
   HTTP-Calls als Sub-Spans.

**Errors über mehrere Services hinweg verfolgen**

- Trace-ID kopieren (z. B. aus einem Backend-Log) → Tempo → Search → Trace-ID
  einfügen. Service-Map zeigt den Aufruf-Zusammenhang. Über _Logs for this span_
  lassen sich die zugehörigen Loki-Logs öffnen.

### Bekannte Lücken (TODO)

- **Caddy-Metriken** sind in `prometheus.yml` auskommentiert — Caddy-Admin
  ist im Caddyfile nicht freigegeben.
- **Container-CPU/Memory** (cAdvisor / Docker-Stats-Exporter) und
  **postgres_exporter / redis_exporter** sind noch nicht gesetzt — das
  _Infra Overview_-Dashboard hat dafür einen Stub.
- **Sampler** ist in Dev `AlwaysSample`. Vor Production auf
  `ParentBased(TraceIDRatioBased(0.1))` umstellen (TODO in `tracing.go`).
- **Tempo-Storage** ist lokal-filesystem. Für echte Production gehört S3 /
  MinIO dahinter.

## Bekannte Auffälligkeiten / Migrations-TODOs

- **Mailhog läuft als linux/amd64-Image** (kein arm64-Build verfügbar). Auf
  Apple Silicon läuft das via Rosetta — funktional ok, nicht ideal. Drop-in-
  Alternative: [`axllent/mailpit`](https://github.com/axllent/mailpit), multi-arch
  und aktiver gewartet, bietet kompatibles SMTP (Port 1025) und Web-UI (Port
  8025). Ablösung steht offen — nicht kritisch genug für sofortige Migration.
