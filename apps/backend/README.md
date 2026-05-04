# wwn-backend

Go-API für worldweathernews.com. Chi-Router, slog-Logging, Prometheus-Metriken,
PostgreSQL über pgxpool, Redis über go-redis.

## Lokal starten

Über den Compose-Stack (empfohlen):

```bash
make dev          # vom Repo-Root: bringt postgres, redis, caddy, mailhog, backend hoch
```

Direkt ohne Container (Postgres/Redis müssen separat laufen):

```bash
cd apps/backend
make run          # einmaliger Build+Run
make dev          # Hot-Reload via air
```

## Endpunkte

| Pfad               | Beschreibung                                      |
| ------------------ | ------------------------------------------------- |
| `GET /health`      | Liveness — JSON mit `status`, `version`, `uptime` |
| `GET /ready`       | Readiness — pingt DB+Redis, 200 oder 503          |
| `GET /metrics`     | Prometheus-Exposition (deaktivierbar)             |
| `GET /api/v1/ping` | Demo-Endpoint mit Trace-ID                        |

Beispiele:

```bash
curl http://api.localhost/health | jq
curl http://api.localhost/ready | jq
curl http://api.localhost/api/v1/ping | jq
curl -s http://api.localhost/metrics | head
```

## ENV-Variablen

Prefix `WWN_`, geschachtelt mit Unterstrich. Beispiele:

| Variable                       | Default                | Beschreibung                  |
| ------------------------------ | ---------------------- | ----------------------------- |
| `WWN_HTTP_PORT`                | `8080`                 | HTTP-Port                     |
| `WWN_HTTP_READTIMEOUT`         | `10s`                  | Read-Timeout                  |
| `WWN_HTTP_WRITETIMEOUT`        | `30s`                  | Write-Timeout                 |
| `WWN_HTTP_IDLETIMEOUT`         | `120s`                 | Idle-Timeout                  |
| `WWN_HTTP_SHUTDOWNTIMEOUT`     | `15s`                  | Graceful-Shutdown-Timeout     |
| `WWN_HTTP_CORSORIGINS`         | `http://app.localhost` | Erlaubte CORS-Origins         |
| `WWN_DATABASE_URL`             | —                      | **Pflicht.** Postgres-URL     |
| `WWN_DATABASE_MAXOPENCONNS`    | `25`                   | Pool max-conns                |
| `WWN_DATABASE_MAXIDLECONNS`    | `5`                    | Pool min-conns                |
| `WWN_DATABASE_CONNMAXLIFETIME` | `1h`                   | Conn-Lifetime                 |
| `WWN_REDIS_URL`                | —                      | **Pflicht.** Redis-URL        |
| `WWN_LOGGING_LEVEL`            | `info`                 | `debug`/`info`/`warn`/`error` |
| `WWN_LOGGING_FORMAT`           | `json`                 | `json` oder `text`            |
| `WWN_ENVIRONMENT`              | `production`           | `dev`/`staging`/`production`  |
| `WWN_METRICSENABLED`           | `true`                 | `/metrics` an/aus             |
| `WWN_CONFIG_FILE`              | _(leer)_               | Optional: YAML-Config-Pfad    |

In `dev` läuft das Backend auch ohne erreichbare DB/Redis weiter (mit Warnung);
in `staging`/`production` schlägt Boot hart fehl.

## Architektur (Stichworte)

- Chi-Router mit RequestID, RealIP, Recoverer, Timeout, CORS, Metrics-Middleware
- `/api/v1/*`-Routen: implementieren das `StrictServerInterface` aus
  `internal/api/api.gen.go` (siehe [ADR-0001](../../docs/adr/0001-openapi-as-source-of-truth.md));
  Handler in `internal/http/handler/api.go`
- System-Endpoints (`/health`, `/ready`, `/metrics`) bewusst außerhalb OpenAPI
- slog-Logger im Context; Request-Logger hängt `request_id` an
- Prometheus mit eigener Registry (4 Backend-Metriken + Standard-Go-Collectors)
- pgxpool für Postgres, go-redis für Redis — Health-Pings über `SELECT 1` / `PING`
- Graceful Shutdown über `signal.NotifyContext`
- DB-Queries kommen mit erstem Feature über sqlc
- OpenTelemetry-Tracing folgt in Session 10

## Generierter Code

`internal/api/api.gen.go` wird von `oapi-codegen` aus
`packages/api-schema/openapi.yaml` erzeugt. Schema-Änderungen → vom Repo-Root
`make gen`. Der File ist als `linguist-generated=true` markiert und aus
`golangci-lint` ausgeschlossen.
