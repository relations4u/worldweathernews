# Session 4 — Go-Backend-Skelett

**Phase**: B (Services)
**Geschätzte Dauer**: 2-3 Stunden
**Vorbedingung**: Session 3 abgeschlossen, Compose-Stack läuft.

## Ziel

Das Go-Backend hat ein produktionstaugliches Skelett:

- Chi-Router mit Standard-Middleware
- Viper-basierte Config (ENV + Datei + Defaults)
- Strukturiertes Logging mit slog
- Prometheus-Metriken
- DB- und Redis-Connections (graceful)
- `/health`, `/ready`, `/metrics`, `/api/v1/ping` Endpunkte
- Hot-Reload via `air` lokal
- Multi-Stage-Dockerfile mit distroless final stage

Am Ende: `curl http://api.localhost/health` antwortet mit JSON.

## Vor-Klärung

Frag den Maintainer, falls noch nicht klar:

- **Go-Modul-Pfad**: `github.com/<org>/worldweathernews/apps/backend` — welcher `<org>`?
- **Validation-Library**: `github.com/go-playground/validator/v10` (verbreitet, aber Tags-Magic)
  oder manuelle Validation in der Config-Struct? Default-Empfehlung: manuell, weniger Magic.

## Aufgaben

### 1. Go-Modul initialisieren

`apps/backend/`:

```bash
cd apps/backend
go mod init github.com/<org>/worldweathernews/apps/backend
```

### 2. Verzeichnisstruktur

```
apps/backend/
├── cmd/
│   └── api/
│       └── main.go
├── internal/
│   ├── config/
│   │   └── config.go
│   ├── http/
│   │   ├── handler/
│   │   │   ├── health.go
│   │   │   ├── ping.go
│   │   │   └── handler_test.go
│   │   ├── middleware/
│   │   │   └── logging.go
│   │   └── router.go
│   ├── observability/
│   │   ├── logging.go
│   │   ├── metrics.go
│   │   └── tracing.go    (Stub für Session 10)
│   ├── storage/
│   │   ├── postgres.go
│   │   └── redis.go
│   └── version/
│       └── version.go
├── .air.toml
├── .dockerignore
├── Dockerfile
├── Makefile
├── README.md
├── config.example.yaml
├── go.mod
└── go.sum
```

### 3. Dependencies

```bash
go get github.com/go-chi/chi/v5
go get github.com/go-chi/chi/v5/middleware
go get github.com/go-chi/cors
go get github.com/spf13/viper
go get github.com/prometheus/client_golang/prometheus
go get github.com/prometheus/client_golang/prometheus/promhttp
go get github.com/jackc/pgx/v5
go get github.com/jackc/pgx/v5/pgxpool
go get github.com/redis/go-redis/v9
```

`log/slog` ist stdlib, keine externe Dependency.

### 4. `internal/version/version.go`

```go
package version

// Werte werden via -ldflags beim Build gesetzt
var (
    Version   = "dev"
    Commit    = "unknown"
    BuildDate = "unknown"
)

type Info struct {
    Version   string `json:"version"`
    Commit    string `json:"commit"`
    BuildDate string `json:"buildDate"`
}

func Get() Info {
    return Info{Version: Version, Commit: Commit, BuildDate: BuildDate}
}
```

### 5. `internal/config/config.go`

```go
package config

import (
    "fmt"
    "strings"
    "time"

    "github.com/spf13/viper"
)

type Config struct {
    HTTP        HTTPConfig
    Database    DatabaseConfig
    Redis       RedisConfig
    Logging     LoggingConfig
    Environment string        // dev | staging | production
    MetricsEnabled bool
}

type HTTPConfig struct {
    Port            int
    ReadTimeout     time.Duration
    WriteTimeout    time.Duration
    IdleTimeout     time.Duration
    ShutdownTimeout time.Duration
    CORSOrigins     []string
}

type DatabaseConfig struct {
    URL             string
    MaxOpenConns    int
    MaxIdleConns    int
    ConnMaxLifetime time.Duration
}

type RedisConfig struct {
    URL string
}

type LoggingConfig struct {
    Level  string // debug | info | warn | error
    Format string // json | text
}

func Load(configPath string) (*Config, error) {
    v := viper.New()

    // Defaults
    v.SetDefault("http.port", 8080)
    v.SetDefault("http.readTimeout", "10s")
    v.SetDefault("http.writeTimeout", "30s")
    v.SetDefault("http.idleTimeout", "120s")
    v.SetDefault("http.shutdownTimeout", "15s")
    v.SetDefault("http.corsOrigins", []string{"http://app.localhost"})
    v.SetDefault("database.maxOpenConns", 25)
    v.SetDefault("database.maxIdleConns", 5)
    v.SetDefault("database.connMaxLifetime", "1h")
    v.SetDefault("logging.level", "info")
    v.SetDefault("logging.format", "json")
    v.SetDefault("environment", "production")
    v.SetDefault("metricsEnabled", true)

    // Optional config file
    if configPath != "" {
        v.SetConfigFile(configPath)
        if err := v.ReadInConfig(); err != nil {
            return nil, fmt.Errorf("read config file: %w", err)
        }
    }

    // ENV overrides
    v.SetEnvPrefix("WWN")
    v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
    v.AutomaticEnv()

    var cfg Config
    if err := v.Unmarshal(&cfg); err != nil {
        return nil, fmt.Errorf("unmarshal config: %w", err)
    }

    if err := cfg.validate(); err != nil {
        return nil, err
    }

    return &cfg, nil
}

func (c *Config) validate() error {
    if c.Database.URL == "" {
        return fmt.Errorf("database.url (WWN_DATABASE_URL) is required")
    }
    if c.Redis.URL == "" {
        return fmt.Errorf("redis.url (WWN_REDIS_URL) is required")
    }
    if c.HTTP.Port < 1 || c.HTTP.Port > 65535 {
        return fmt.Errorf("http.port must be between 1 and 65535")
    }
    switch c.Logging.Level {
    case "debug", "info", "warn", "error":
    default:
        return fmt.Errorf("logging.level must be debug|info|warn|error")
    }
    switch c.Logging.Format {
    case "json", "text":
    default:
        return fmt.Errorf("logging.format must be json or text")
    }
    return nil
}
```

ENV-Mapping: `WWN_HTTP_PORT`, `WWN_DATABASE_URL`, `WWN_REDIS_URL`, `WWN_LOGGING_LEVEL` etc.

### 6. `internal/observability/logging.go`

slog-Setup:

- Format aus Config
- Trace-ID als Default-Attribut wenn vorhanden (über Context)
- Helper `FromContext(ctx)` und `WithLogger(ctx, log)`

### 7. `internal/observability/metrics.go`

Prometheus-Registry mit:

- `wwn_http_requests_total{method, path, status}` Counter
- `wwn_http_request_duration_seconds{method, path}` Histogram
- `wwn_db_pool_connections{state}` Gauge (von pgxpool gespeist)
- `wwn_redis_commands_total{cmd, result}` Counter

Helfer-Funktion zum Registrieren der Standard-Go-Collectors.

### 8. `internal/storage/postgres.go`

- `NewPool(ctx, cfg DatabaseConfig) (*pgxpool.Pool, error)`
- Pool-Konfiguration aus Config
- `HealthCheck(ctx) error` der `SELECT 1` ausführt
- Ping mit Timeout

### 9. `internal/storage/redis.go`

- `NewClient(ctx, cfg RedisConfig) (*redis.Client, error)`
- URL parsen mit `redis.ParseURL`
- `HealthCheck(ctx) error` mit PING

### 10. `internal/http/middleware/logging.go`

Middleware die jede Request loggt:

- Method, Path, Status, Duration, RemoteAddr, UserAgent, RequestID
- Slog-Output via Context-Logger

### 11. `internal/http/router.go`

Router-Setup:

```go
func NewRouter(cfg *config.Config, deps Deps) http.Handler {
    r := chi.NewRouter()

    // Standard middleware
    r.Use(middleware.RequestID)
    r.Use(middleware.RealIP)
    r.Use(loggingMiddleware(deps.Logger))
    r.Use(middleware.Recoverer)
    r.Use(middleware.Timeout(30 * time.Second))
    r.Use(corsHandler(cfg.HTTP.CORSOrigins))
    r.Use(metricsMiddleware())

    // System endpoints
    r.Get("/health", handler.Health(deps))
    r.Get("/ready", handler.Ready(deps))
    if cfg.MetricsEnabled {
        r.Handle("/metrics", promhttp.Handler())
    }

    // API v1
    r.Route("/api/v1", func(r chi.Router) {
        r.Get("/ping", handler.Ping(deps))
    })

    return r
}
```

### 12. Handler

**`/health`** → Liveness:

```json
{
  "status": "ok",
  "version": { "version": "...", "commit": "..." },
  "uptime": "..."
}
```

**`/ready`** → Readiness:

- Pingt DB und Redis
- 200 wenn beide ok, 503 sonst
- Body listet Status pro Komponente

**`/api/v1/ping`** →

```json
{ "message": "pong", "traceId": "<request-id>" }
```

### 13. `cmd/api/main.go`

Reihenfolge:

1. Config laden (Fehler → exit 1)
2. Logger initialisieren
3. Pool zu Postgres aufmachen — wenn dev: graceful (warnen, weiterlaufen),
   wenn staging/production: hart fail
4. Redis-Client öffnen — analog
5. Router bauen
6. HTTP-Server starten in Goroutine
7. Auf SIGTERM/SIGINT warten
8. Graceful Shutdown mit Timeout

### 14. `Dockerfile`

```dockerfile
# syntax=docker/dockerfile:1.7

FROM golang:1.23-alpine AS builder
ARG VERSION=dev
ARG COMMIT=unknown
ARG BUILD_DATE=unknown

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY . .

ENV CGO_ENABLED=0 GOOS=linux
RUN go build \
    -trimpath \
    -ldflags="-s -w \
      -X github.com/<org>/worldweathernews/apps/backend/internal/version.Version=${VERSION} \
      -X github.com/<org>/worldweathernews/apps/backend/internal/version.Commit=${COMMIT} \
      -X github.com/<org>/worldweathernews/apps/backend/internal/version.BuildDate=${BUILD_DATE}" \
    -o /out/api ./cmd/api

FROM gcr.io/distroless/static:nonroot
LABEL org.opencontainers.image.source="https://github.com/<org>/worldweathernews"
LABEL org.opencontainers.image.title="wwn-backend"
LABEL org.opencontainers.image.description="worldweathernews.com API"
COPY --from=builder /out/api /api
USER nonroot:nonroot
EXPOSE 8080
HEALTHCHECK --interval=10s --timeout=3s --start-period=5s --retries=3 \
    CMD ["/api", "healthcheck"] || exit 1
ENTRYPOINT ["/api"]
```

Hinweis: `HEALTHCHECK` braucht ggf. einen `healthcheck`-Subcommand im Binary. Falls
zu komplex für jetzt: Healthcheck im Compose-File als `wget`-Call definieren
(distroless hat kein wget, daher Compose-extern).

### 15. `.air.toml` für Hot-Reload

```toml
root = "."
tmp_dir = "tmp"

[build]
bin = "tmp/api"
cmd = "go build -o tmp/api ./cmd/api"
delay = 500
exclude_dir = ["tmp", "vendor", "testdata"]
include_ext = ["go", "yaml"]
stop_on_error = true

[log]
time = true
```

### 16. `apps/backend/Makefile`

```makefile
.PHONY: run dev test lint build tidy

run:
	go run ./cmd/api

dev:
	air

test:
	go test -race -cover ./...

lint:
	golangci-lint run ./...

build:
	CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o bin/api ./cmd/api

tidy:
	go mod tidy
```

### 17. Compose-Integration

`infra/compose/compose.dev.yml` erweitern:

```yaml
backend:
  build:
    context: ../../apps/backend
    dockerfile: Dockerfile
    target: builder # für dev mit air
  container_name: wwn-backend
  command: air
  volumes:
    - ../../apps/backend:/src
    - go_mod_cache:/go/pkg/mod
  working_dir: /src
  environment:
    WWN_DATABASE_URL: ${WWN_DATABASE_URL}
    WWN_REDIS_URL: ${WWN_REDIS_URL}
    WWN_LOG_LEVEL: ${WWN_LOG_LEVEL:-debug}
    WWN_LOG_FORMAT: ${WWN_LOG_FORMAT:-text}
    WWN_ENVIRONMENT: dev
    WWN_HTTP_PORT: 8080
  depends_on:
    postgres:
      condition: service_healthy
    redis:
      condition: service_healthy
  expose:
    - "8080"
```

`volumes`-Section ergänzen:

```yaml
go_mod_cache:
```

`Caddyfile` updaten:

```caddy
api.localhost {
    reverse_proxy backend:8080
}
```

### 18. Tests

Mindestens ein Handler-Test in `internal/http/handler/handler_test.go`:

- `/health` returns 200 with JSON containing version
- Ping returns 200 with traceId

### 19. Top-Level Makefile

```makefile
backend-dev: ## Backend lokal mit hot-reload
	$(MAKE) -C apps/backend dev

backend-test: ## Backend-Tests
	$(MAKE) -C apps/backend test

backend-lint: ## Backend-Lint
	$(MAKE) -C apps/backend lint
```

`test`, `lint` Targets entsprechend ergänzen damit sie das Backend-Subverzeichnis
sauber einbinden.

### 20. README

`apps/backend/README.md`:

- Wie lokal starten (`make dev` oder im Container `docker compose up backend`)
- Endpunkte mit Beispiel-curl
- Komplette ENV-Variablen-Tabelle
- Architektur-Stichworte (Chi, sqlc kommt später, etc.)

## Vorgehen (verbindlich)

1. Plan zeigen, **detailliert**, viele Files
2. Freigabe abwarten
3. Implementierung in sinnvollen Schritten:
   - a) Skeleton + Config + Logging
   - b) Storage-Layer
   - c) HTTP-Layer
   - d) Tests
   - e) Dockerfile + Compose-Integration
   - Nach jedem Schritt kurze Status-Meldung
4. Tests laufen lassen: `cd apps/backend && go test -race ./...`
5. golangci-lint laufen lassen: `golangci-lint run ./...`
6. **Selbst** `make dev` starten, `curl api.localhost/health` machen, Output zeigen
7. `curl api.localhost/api/v1/ping` zeigen
8. `curl api.localhost/ready` zeigen — sollte 200 sein wenn DB und Redis healthy
9. `curl api.localhost/metrics` — kurzer Auszug
10. Nicht committen

## Erfolgs-Kriterien

- [ ] `go build` erfolgreich
- [ ] `go test -race ./...` grün (mindestens 2 Tests)
- [ ] `golangci-lint run` grün
- [ ] `air` läuft im Container, reagiert auf Code-Änderungen
- [ ] `/health` 200 mit Version-JSON
- [ ] `/ready` 200 wenn DB+Redis healthy, sonst 503
- [ ] `/metrics` zeigt Prometheus-Format
- [ ] `/api/v1/ping` mit Trace-ID
- [ ] Container-Image-Größe (build mit --target distroless): < 30 MB
- [ ] Logs sind strukturiert (JSON in prod, text in dev)
- [ ] Caddy routet api.localhost zum Backend

## Mögliche Stolpersteine

- **air auf Apple Silicon**: Architektur-Mismatch wenn Image AMD64 ist. Plattform
  in Compose explizit setzen oder `goimports` direkt nutzen.
- **distroless HEALTHCHECK**: distroless hat kein `wget`/`curl`. Lösung: Compose
  definiert Healthcheck mit `wget` von außen, oder Subcommand im Binary.
- **pgxpool und context-Cancellation**: bei Shutdown sauber `pool.Close()` aufrufen.
- **CORS**: in dev sehr permissiv (`http://app.localhost`), in prod explizit auf
  Production-Domain.
- **slog-Format-Wechsel**: JSON in prod, text in dev. Nicht versehentlich beides
  mischen.

## Was diese Session NICHT tut

- Keine echten Wetterdaten
- Keine Authentifizierung
- Kein OpenTelemetry-Tracing (kommt Session 10)
- Keine sqlc-Queries (kommt mit erstem Feature)
- Keine API-Versionierung über v1 hinaus

## Suggested Commit-Message

```
feat(backend): add Go API skeleton with health, metrics, structured logging
```
