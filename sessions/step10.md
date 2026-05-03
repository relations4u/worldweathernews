# Session 10 — Observability-Stack lokal

**Phase**: D (Ops)
**Geschätzte Dauer**: 2 Stunden
**Vorbedingung**: Sessions 4-6 abgeschlossen, Services laufen mit Metrics-Endpoints.

## Ziel

Ein lokaler Observability-Stack (Prometheus, Grafana, Loki, Promtail, Tempo) ist
als optionales Compose-Profile lauffähig. Backend und Pyworkers werden mit
OpenTelemetry instrumentiert und schicken Traces an Tempo. Logs werden von
Promtail aus Container-Stdout gesammelt und in Loki gespeichert.

Drei vorprovisionierte Grafana-Dashboards zeigen Backend-, Pyworkers- und
Infra-Metriken. Trace-IDs verbinden Logs (Loki) und Traces (Tempo).

`make dev` bleibt schlank (ohne Monitoring), `make dev-full` startet alles.

## Aufgaben

### 1. Compose-Profile `monitoring`

`infra/compose/compose.dev.yml` erweitern. Profile-Mechanismus von Compose nutzen:

```yaml
prometheus:
  image: prom/prometheus:latest
  container_name: wwn-prometheus
  profiles: [monitoring]
  command:
    - "--config.file=/etc/prometheus/prometheus.yml"
    - "--storage.tsdb.retention.time=7d"
    - "--web.enable-lifecycle"
  volumes:
    - ../monitoring/prometheus/prometheus.yml:/etc/prometheus/prometheus.yml:ro
    - prometheus_data:/prometheus
  ports:
    - "127.0.0.1:9090:9090"

grafana:
  image: grafana/grafana:latest
  container_name: wwn-grafana
  profiles: [monitoring]
  environment:
    GF_SECURITY_ADMIN_USER: admin
    GF_SECURITY_ADMIN_PASSWORD: admin
    GF_AUTH_ANONYMOUS_ENABLED: "true"
    GF_AUTH_ANONYMOUS_ORG_ROLE: Viewer
    GF_USERS_DEFAULT_THEME: dark
    GF_INSTALL_PLUGINS: ""
  volumes:
    - ../monitoring/grafana/provisioning:/etc/grafana/provisioning:ro
    - ../monitoring/grafana/dashboards:/var/lib/grafana/dashboards:ro
    - grafana_data:/var/lib/grafana
  ports:
    - "127.0.0.1:3000:3000"
  depends_on: [prometheus, loki, tempo]

loki:
  image: grafana/loki:latest
  container_name: wwn-loki
  profiles: [monitoring]
  command: -config.file=/etc/loki/local-config.yaml
  volumes:
    - ../monitoring/loki/local-config.yaml:/etc/loki/local-config.yaml:ro
    - loki_data:/loki
  ports:
    - "127.0.0.1:3100:3100"

promtail:
  image: grafana/promtail:latest
  container_name: wwn-promtail
  profiles: [monitoring]
  command: -config.file=/etc/promtail/config.yaml
  volumes:
    - ../monitoring/promtail/config.yaml:/etc/promtail/config.yaml:ro
    - /var/lib/docker/containers:/var/lib/docker/containers:ro
    - /var/run/docker.sock:/var/run/docker.sock:ro
  depends_on: [loki]

tempo:
  image: grafana/tempo:latest
  container_name: wwn-tempo
  profiles: [monitoring]
  command: -config.file=/etc/tempo.yaml
  volumes:
    - ../monitoring/tempo/tempo.yaml:/etc/tempo.yaml:ro
    - tempo_data:/var/tempo
  ports:
    - "127.0.0.1:3200:3200" # Tempo HTTP
    - "127.0.0.1:4317:4317" # OTLP gRPC
    - "127.0.0.1:4318:4318" # OTLP HTTP
```

Volumes ergänzen:

```yaml
prometheus_data:
grafana_data:
loki_data:
tempo_data:
```

### 2. Prometheus-Config

`infra/monitoring/prometheus/prometheus.yml`:

```yaml
global:
  scrape_interval: 15s
  evaluation_interval: 15s
  external_labels:
    cluster: wwn-dev
    env: dev

scrape_configs:
  - job_name: prometheus
    static_configs:
      - targets: ["localhost:9090"]

  - job_name: backend
    metrics_path: /metrics
    static_configs:
      - targets: ["backend:8080"]
        labels:
          service: backend

  - job_name: pyworkers
    static_configs:
      - targets: ["pyworkers:9100"]
        labels:
          service: pyworkers

  - job_name: caddy
    static_configs:
      - targets: ["caddy:2019"]
        labels:
          service: caddy
    # Caddy-Admin liefert /metrics, falls aktiviert
```

(Caddy-Admin müsste in `Caddyfile` aktiviert werden — TODO-Kommentar setzen,
Caddy-Metriken sind nice-to-have, nicht kritisch.)

### 3. Loki-Config

`infra/monitoring/loki/local-config.yaml`:

```yaml
auth_enabled: false

server:
  http_listen_port: 3100

common:
  path_prefix: /loki
  storage:
    filesystem:
      chunks_directory: /loki/chunks
      rules_directory: /loki/rules
  replication_factor: 1
  ring:
    instance_addr: 127.0.0.1
    kvstore:
      store: inmemory

schema_config:
  configs:
    - from: 2024-01-01
      store: tsdb
      object_store: filesystem
      schema: v13
      index:
        prefix: index_
        period: 24h

limits_config:
  retention_period: 168h # 7 Tage
  allow_structured_metadata: true
```

### 4. Promtail-Config

`infra/monitoring/promtail/config.yaml`:

```yaml
server:
  http_listen_port: 9080
  grpc_listen_port: 0

positions:
  filename: /tmp/positions.yaml

clients:
  - url: http://loki:3100/loki/api/v1/push

scrape_configs:
  - job_name: docker
    docker_sd_configs:
      - host: unix:///var/run/docker.sock
        refresh_interval: 5s
    relabel_configs:
      - source_labels: ["__meta_docker_container_name"]
        regex: "/(wwn-.*)"
        target_label: container
      - source_labels: ["__meta_docker_container_log_stream"]
        target_label: stream
      - source_labels:
          ["__meta_docker_container_label_com_docker_compose_service"]
        target_label: service
    pipeline_stages:
      - docker: {}
      - json:
          expressions:
            level: level
            time: time
            msg: msg
            trace_id: trace_id
            request_id: request_id
      - labels:
          level:
          trace_id:
```

### 5. Tempo-Config

`infra/monitoring/tempo/tempo.yaml`:

```yaml
server:
  http_listen_port: 3200

distributor:
  receivers:
    otlp:
      protocols:
        grpc:
          endpoint: 0.0.0.0:4317
        http:
          endpoint: 0.0.0.0:4318

ingester:
  trace_idle_period: 10s
  max_block_bytes: 1_000_000
  max_block_duration: 5m

compactor:
  compaction:
    block_retention: 168h # 7 Tage

storage:
  trace:
    backend: local
    local:
      path: /var/tempo/blocks
    wal:
      path: /var/tempo/wal
```

### 6. Grafana-Provisioning

`infra/monitoring/grafana/provisioning/datasources/datasources.yml`:

```yaml
apiVersion: 1

datasources:
  - name: Prometheus
    type: prometheus
    access: proxy
    url: http://prometheus:9090
    isDefault: true
    jsonData:
      timeInterval: 15s

  - name: Loki
    type: loki
    access: proxy
    url: http://loki:3100
    jsonData:
      derivedFields:
        - name: TraceID
          matcherRegex: 'trace_id="([a-f0-9]+)"'
          url: "$${__value.raw}"
          datasourceUid: tempo

  - name: Tempo
    type: tempo
    access: proxy
    url: http://tempo:3200
    uid: tempo
    jsonData:
      tracesToLogsV2:
        datasourceUid: loki
        spanStartTimeShift: -1m
        spanEndTimeShift: 1m
        tags: ["service"]
        filterByTraceID: true
      serviceMap:
        datasourceUid: prometheus
```

`infra/monitoring/grafana/provisioning/dashboards/dashboards.yml`:

```yaml
apiVersion: 1

providers:
  - name: wwn
    orgId: 1
    folder: worldweathernews
    type: file
    disableDeletion: false
    updateIntervalSeconds: 30
    allowUiUpdates: true
    options:
      path: /var/lib/grafana/dashboards
```

### 7. Grafana-Dashboards (JSON)

Drei Dashboards in `infra/monitoring/grafana/dashboards/`:

**`backend-overview.json`** — Panels:

- HTTP Request Rate (`rate(wwn_http_requests_total[1m])` by status)
- HTTP Latency p50/p95/p99 (`histogram_quantile(...)`)
- Error Rate (5xx / total)
- Active Goroutines (`go_goroutines`)
- Memory (`go_memstats_alloc_bytes`)
- DB Pool Stats (`wwn_db_pool_connections`)

**`pyworkers-overview.json`** — Panels:

- Heartbeat Rate (`rate(wwn_heartbeat_total[5m])`)
- Job Run Rate by Status (`rate(wwn_job_runs_total[1m]) by (status)`)
- Job Duration p50/p95 (`histogram_quantile(...)`)
- Process CPU + Memory (Standard-Python-Metriken)

**`infra-overview.json`** — Panels:

- Container CPU (cAdvisor wird hier nicht installiert; nutze stattdessen
  Docker-Stats via `docker_container_*` falls verfügbar, sonst Panel mit
  TODO-Hinweis)
- Postgres `pg_stat_database` Connections (braucht postgres_exporter — als
  TODO markieren oder einfach Stub-Panels)
- Redis Ops/sec — analog

**Pragmatischer Hinweis**: Dashboards von Hand zu JSON-en ist mühsam. Pragmatisch:
schlanke Stubs erzeugen mit den wichtigsten Panels für Backend und Pyworkers,
Infra-Dashboard auf später vertagen mit klar kommentierten Lücken. Frag den
Maintainer, falls volle Auto-Generation gewünscht — alternativ in Grafana per
Hand bauen, dann JSON exportieren und committen.

### 8. Backend-Instrumentierung mit OpenTelemetry

`apps/backend/`:

Dependencies:

```bash
go get go.opentelemetry.io/otel
go get go.opentelemetry.io/otel/sdk
go get go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc
go get go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp
go get go.opentelemetry.io/contrib/instrumentation/github.com/go-chi/chi/otelchi
```

`internal/observability/tracing.go`:

```go
package observability

import (
    "context"
    "fmt"

    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
    "go.opentelemetry.io/otel/propagation"
    "go.opentelemetry.io/otel/sdk/resource"
    sdktrace "go.opentelemetry.io/otel/sdk/trace"
    semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

type TracingConfig struct {
    Enabled     bool
    Endpoint    string  // z.B. "tempo:4317"
    ServiceName string
    Environment string
    Version     string
}

func InitTracing(ctx context.Context, cfg TracingConfig) (func(context.Context) error, error) {
    if !cfg.Enabled {
        return func(context.Context) error { return nil }, nil
    }

    exp, err := otlptracegrpc.New(ctx,
        otlptracegrpc.WithEndpoint(cfg.Endpoint),
        otlptracegrpc.WithInsecure(),
    )
    if err != nil {
        return nil, fmt.Errorf("create OTLP exporter: %w", err)
    }

    res, err := resource.New(ctx,
        resource.WithAttributes(
            semconv.ServiceName(cfg.ServiceName),
            semconv.ServiceVersion(cfg.Version),
            semconv.DeploymentEnvironmentName(cfg.Environment),
        ),
    )
    if err != nil {
        return nil, fmt.Errorf("create resource: %w", err)
    }

    tp := sdktrace.NewTracerProvider(
        sdktrace.WithBatcher(exp),
        sdktrace.WithResource(res),
        sdktrace.WithSampler(sdktrace.AlwaysSample()),
    )

    otel.SetTracerProvider(tp)
    otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
        propagation.TraceContext{},
        propagation.Baggage{},
    ))

    return tp.Shutdown, nil
}
```

In `cmd/api/main.go` initialisieren, vor Router-Setup. Shutdown via defer.

In `internal/http/router.go` die otelhttp-Middleware vor allen anderen einhängen
oder otelchi-Middleware verwenden:

```go
r.Use(otelchi.Middleware(cfg.ServiceName, otelchi.WithChiRoutes(r)))
```

Trace-ID in slog mitloggen — Helper in der Logging-Middleware:

```go
spanCtx := trace.SpanContextFromContext(ctx)
if spanCtx.IsValid() {
    log = log.With(
        slog.String("trace_id", spanCtx.TraceID().String()),
        slog.String("span_id", spanCtx.SpanID().String()),
    )
}
```

Config-Erweiterung in `internal/config`:

```go
type ObservabilityConfig struct {
    TracingEnabled  bool
    TracingEndpoint string  // default "tempo:4317"
}
```

### 9. Pyworkers-Instrumentierung

Dependencies:

```toml
"opentelemetry-api>=1.25",
"opentelemetry-sdk>=1.25",
"opentelemetry-exporter-otlp-proto-grpc>=1.25",
"opentelemetry-instrumentation>=0.46b0",
"opentelemetry-instrumentation-asyncpg>=0.46b0",
"opentelemetry-instrumentation-redis>=0.46b0",
"opentelemetry-instrumentation-httpx>=0.46b0",
```

`pyworkers/observability/tracing.py`:

```python
from opentelemetry import trace
from opentelemetry.exporter.otlp.proto.grpc.trace_exporter import OTLPSpanExporter
from opentelemetry.sdk.resources import Resource
from opentelemetry.sdk.trace import TracerProvider
from opentelemetry.sdk.trace.export import BatchSpanProcessor
from opentelemetry.semconv.resource import ResourceAttributes


def init_tracing(
    *,
    enabled: bool,
    endpoint: str,
    service_name: str,
    environment: str,
    version: str,
) -> None:
    if not enabled:
        return

    resource = Resource.create({
        ResourceAttributes.SERVICE_NAME: service_name,
        ResourceAttributes.SERVICE_VERSION: version,
        ResourceAttributes.DEPLOYMENT_ENVIRONMENT: environment,
    })

    provider = TracerProvider(resource=resource)
    exporter = OTLPSpanExporter(endpoint=endpoint, insecure=True)
    provider.add_span_processor(BatchSpanProcessor(exporter))

    trace.set_tracer_provider(provider)
```

In `__main__.py` aufrufen, Auto-Instrumentation für asyncpg/redis/httpx aktivieren.

Trace-Context in structlog binden:

```python
from opentelemetry import trace as otel_trace
import structlog


def add_trace_context(_, __, event_dict):
    span = otel_trace.get_current_span()
    ctx = span.get_span_context()
    if ctx.is_valid:
        event_dict["trace_id"] = format(ctx.trace_id, "032x")
        event_dict["span_id"] = format(ctx.span_id, "016x")
    return event_dict
```

In `configure_logging` als Processor einreihen.

### 10. ENV-Variablen erweitern

`.env.example` und Compose-File:

```bash
# Backend
WWN_OBSERVABILITY_TRACING_ENABLED=true
WWN_OBSERVABILITY_TRACING_ENDPOINT=tempo:4317

# Pyworkers
WWN_PY_TRACING_ENABLED=true
WWN_PY_TRACING_ENDPOINT=tempo:4317
```

### 11. `Makefile`-Updates

```makefile
dev-full: ## Lokaler Stack inkl. Monitoring
	docker compose --profile monitoring up -d
	docker compose --profile monitoring logs -f --tail=20

dev-monitoring: ## Nur Monitoring-Services
	docker compose --profile monitoring up -d prometheus grafana loki promtail tempo
```

### 12. Runbook-Einträge

`docs/runbook.md` erweitern:

```markdown
## Wo finde ich was im Monitoring?

- **Grafana**: http://localhost:3000 (admin / admin in dev)
- **Prometheus**: http://localhost:9090
- **Loki Direct**: http://localhost:3100 (UI ist Grafana)
- **Tempo Direct**: http://localhost:3200

### Häufige Aufgaben

**Backend hat hohe Latency**

1. Grafana → Backend Overview → p99-Panel anschauen
2. Wenn auffällig: Tempo öffnen, nach langsamen Traces suchen
3. Trace-ID notieren, in Loki nach passenden Logs filtern

**Worker-Job ist fehlgeschlagen**

1. Grafana → Pyworkers Overview → Job Run Rate by Status
2. Loki: `{service="pyworkers"} |= "error"`
3. Trace verfolgen, Auto-Instrumentation zeigt DB-/HTTP-Calls

**Errors über mehrere Services hinweg verfolgen**

- Trace-ID kopieren, in Tempo Search-Eingabe einfügen
- Service-Map zeigt Aufruf-Zusammenhang
```

## Vorgehen (verbindlich)

1. Plan zeigen
2. Freigabe abwarten
3. Implementieren in Schritten:
   a) Compose-Profile + Configs (Prometheus, Loki, Promtail, Tempo)
   b) Grafana-Provisioning + initiale Dashboards (zumindest backend + pyworkers)
   c) Backend-Instrumentierung (otelchi, OTLP-Export, Trace in slog)
   d) Pyworkers-Instrumentierung (Auto-Instrumentation, Trace in structlog)
   e) Runbook-Einträge
4. **Selbst** `make dev-full` starten, alle Services healthy?
5. Backend `/api/v1/ping` mehrfach aufrufen
6. In Grafana:
   - Backend-Dashboard zeigt Request-Rate
   - Pyworkers-Dashboard zeigt Heartbeats
   - Loki-Explore: Logs erscheinen mit `service`-Label
   - Tempo-Search: erste Traces vorhanden
   - Trace-IDs verbinden Logs und Traces
7. Backend- und Pyworkers-Tests grün halten
8. Nicht committen

## Erfolgs-Kriterien

- [ ] `make dev-full` startet alle 5 Monitoring-Services + Apps
- [ ] Grafana erreichbar auf 127.0.0.1:3000
- [ ] Drei provisionierte Datasources (Prometheus, Loki, Tempo)
- [ ] Mindestens zwei Dashboards (backend, pyworkers) zeigen Daten
- [ ] Loki erhält Container-Logs, parst JSON
- [ ] Tempo erhält Traces vom Backend
- [ ] Tempo erhält Traces von Pyworkers
- [ ] Trace-ID in Backend-Log entspricht Trace-ID in Tempo
- [ ] Trace-IDs in Loki-Logs sind klickbar (Derived Field konfiguriert)
- [ ] `make dev` (ohne `-full`) startet Monitoring-Services NICHT
- [ ] Bestehende Tests bleiben grün

## Mögliche Stolpersteine

- **Promtail Docker-Socket**: braucht Read-Access. Auf manchen Systemen
  Permissions-Probleme — entweder Promtail als root oder Socket-Group anpassen.
- **JSON-Log-Parsing**: nur Logs, die echtes JSON sind, werden gut indexiert.
  In dev mit `text`-Format ist das nicht der Fall — Pipeline-Stages müssen
  damit umgehen oder Empfehlung: dev auch auf `json` umstellen.
- **OTel-Sampler**: `AlwaysSample` ist für dev OK, in prod auf
  `ParentBased(TraceIDRatioBased(0.1))` umstellen — TODO im Code.
- **Tempo-Storage**: lokal-filesystem ist OK für dev, in prod S3/MinIO oder
  ähnlich. Nicht für jetzt.
- **OTel-Library-Versionen**: das Ökosystem entwickelt sich schnell, Breaking
  Changes zwischen Minor-Versionen kommen vor. Versionen pinnen.
- **Grafana-Dashboard-JSON**: nicht von Hand schön zu schreiben. Pragmatischer
  Ansatz: minimale JSONs erzeugen, dann in Grafana feinschleifen, exportieren,
  committen. Falls beim Erst-Build zu viel Aufwand: ein einfaches Dashboard
  reicht, Rest als TODO.

## Was diese Session NICHT tut

- Kein Sentry (separater Schritt, später)
- Kein Uptime-Kuma (separater Schritt, im prod-Setup)
- Keine Alert-Rules (kommen wenn echte SLOs definiert sind)
- Keine Frontend-Instrumentierung (Browser-OTel, später)
- Kein Postgres-Exporter (kann später nachgerüstet werden)

## Suggested Commit-Message

```
feat(infra): add observability stack with prometheus, grafana, loki, tempo
```
