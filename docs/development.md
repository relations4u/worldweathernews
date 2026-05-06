# Development Guide

So erweitert man worldweathernews.com lokal. Dieses Dokument geht davon
aus, dass das [README-Quickstart](../README.md#quickstart) durchgelaufen
ist (mise + make bootstrap + make dev).

## Make-Targets im Überblick

```bash
make help            # Liste aller Targets

# Lokaler Stack
make dev             # App-Stack ohne Monitoring
make dev-full        # … inkl. Prometheus/Grafana/Loki/Tempo
make dev-monitoring  # nur Monitoring (Apps separat)
make dev-down        # Stack stoppen
make dev-reset       # Stack stoppen + Volumes löschen
make dev-logs SERVICE=backend
make dev-psql        # psql-Shell auf DB
make dev-redis       # redis-cli auf Cache

# Pro Service
make backend-dev      # Hot-Reload via air
make backend-test
make backend-lint
make frontend-dev     # Hot-Reload via vite
make frontend-test
make frontend-lint
make frontend-check   # svelte-check
make pyworkers-dev    # Hot-Reload via watchfiles
make pyworkers-test
make pyworkers-lint
make pyworkers-typecheck

# Code-Hygiene
make lint             # alle Linter via pre-commit
make fmt              # Auto-Format
make test             # alle Tests
make build            # alle Container bauen

# Schema/Codegen
make gen              # OpenAPI → Go-Stubs + TS-Types
make gen-check        # CI-Verifikation, dass aktuell

# Migrations und Release
make migrate          # goose up
make release          # interaktiv: bump → signed Tag → push
make clean            # Aufräumen
```

## Wie ich einen neuen API-Endpoint hinzufüge

OpenAPI ist Single Source of Truth. Reihenfolge:

1. **Schema erweitern** in `packages/api-schema/openapi.yaml`:

   ```yaml
   /api/v1/weather/{locationId}:
     get:
       operationId: getWeather
       parameters:
         - in: path
           name: locationId
           required: true
           schema: { type: string, format: uuid }
       responses:
         "200":
           description: Current weather observation
           content:
             application/json:
               schema:
                 $ref: "#/components/schemas/WeatherObservation"
   ```

2. `make gen` regeneriert Go-Server-Stubs und TS-Client-Types.
3. **Backend-Handler** in `apps/backend/internal/http/handler/weather.go`:

   ```go
   func (h *APIHandler) GetWeather(
       ctx context.Context,
       req api.GetWeatherRequestObject,
   ) (api.GetWeatherResponseObject, error) {
       row, err := h.queries.GetWeatherByLocation(ctx, req.LocationId)
       if err != nil {
           return nil, err
       }
       return api.GetWeather200JSONResponse(toAPI(row)), nil
   }
   ```

4. **Test schreiben** — Handler-Test mit `httptest`,
   Integration-Test gegen die echte DB (Compose-Stack muss laufen).
5. **Frontend-Aufruf** in `apps/frontend/src/lib/api/client.ts` ist
   schon generiert; einfach importieren:

   ```ts
   import { getWeather } from "$lib/api/client";
   const weather = await getWeather(locationId);
   ```

6. PR aufmachen. Commitlint erzwingt Conventional Commits, Branch-
   Protection blockt Merges ohne grüne CI.

## Wie ich eine neue DB-Migration mache

`goose` ist sprachunabhängig, Migrations leben in `infra/migrations/`:

```bash
goose -dir infra/migrations create add_weather_observations sql
```

Beide Richtungen schreiben (Up + Down). Lokal anwenden:

```bash
make migrate
```

Down-Migration testen vor Commit. CI hat keinen automatischen
Down-Test, aber Rollback ist im Runbook und Pflicht für jede
Migration.

## Wie ich einen neuen Worker-Job hinzufüge

1. Datei `apps/pyworkers/pyworkers/jobs/my_job.py` mit
   `async def run(...)`-Funktion anlegen.
2. In `apps/pyworkers/pyworkers/__main__.py` als Scheduler-Job
   registrieren:

   ```python
   scheduler.add_job(
       my_job.run,
       trigger=IntervalTrigger(minutes=15),
       id="my_job",
       max_instances=1,
   )
   ```

3. Test in `tests/test_my_job.py` (pytest, asyncio_mode="auto").
4. Falls relevant: Prometheus-Metriken über das `metrics`-Modul
   ergänzen.
5. Manueller OpenTelemetry-Span, wenn der Job keine HTTP/DB/Redis-
   Calls macht (Auto-Instrumentation greift sonst nicht — der
   Heartbeat-Job ist das Beispiel).

## Wie ich einen Bug debugge

```bash
# 1. Reproduzieren
make dev-full

# 2. Logs in Grafana → Loki, gefiltert nach Service + Level
#    http://grafana.localhost mit `make dev-full`
#    {service="wwn-backend"} |= "error"

# 3. Trace-ID aus dem Log → Tempo Search → Span-Tree

# 4. DB-Stand prüfen
make dev-psql

# 5. Cache prüfen
make dev-redis

# 6. Container-Logs eines Services
make dev-logs SERVICE=backend
```

Wenn reproduziert: Test schreiben, der den Bug zeigt, dann fixen.
Test muss vor dem Fix rot sein.

## Wie ich Konfiguration ändere

Drei Quellen pro Service: Defaults im Code, optionale Config-Datei,
ENV-Variablen mit Service-Prefix:

| Service   | Prefix    | Loader                                       |
| --------- | --------- | -------------------------------------------- |
| Backend   | `WWN_`    | Viper, `internal/config/`                    |
| Pyworkers | `WWN_PY_` | pydantic-settings                            |
| Frontend  | `PUBLIC_` | SvelteKit `$env/static/public` (build-time!) |

**Falle**: SvelteKit-`PUBLIC_*`-Vars werden zur **Build-Zeit** ins JS-
Bundle gebacken. Runtime-ENV in Compose hat keinen Effekt aufs gebaute
Bundle. Wer das Frontend für eine andere Umgebung baut, muss
`PUBLIC_API_BASE_URL` als `--build-arg` übergeben (siehe
`.github/workflows/release.yml` und `apps/frontend/Dockerfile`).

## Versions-Pinning

Vier Quellen müssen dieselbe Major.Minor zeigen — sonst CI-Fail mit
verwirrenden Symptomen. Ausführlich in
[CLAUDE.md → Versions-Pinning](../CLAUDE.md#versions-pinning--verbindlich).

1. `apps/<service>/<lang>.mod` (oder `pyproject.toml`)
2. `.mise.toml`
3. `.github/workflows/ci-*.yml`
4. `apps/<service>/Dockerfile`

## Lokale Tipps

- **Backend Hot-Reload**: `air` läuft im Container, Code-Änderungen
  triggern Auto-Restart.
- **Frontend Hot-Reload**: Vite HMR durch Caddy hindurch — Caddy
  erlaubt explizit den WebSocket-Upgrade dafür (`compose.dev.yml`).
- **DB-Migrations zurückrollen**: `goose -dir infra/migrations down`.
- **Stack neu aufsetzen**: `make dev-reset && make dev` — räumt Volumes
  weg.
- **OpenAPI-Linting**: `redocly lint --config packages/api-schema/redocly.yaml`.
  Stilistische Findings (operation-summary, info-license, unused-components)
  sind aktuell Warnings, kein Block.
- **Pre-commit-Cache rebuild**: nach Änderung an
  `.pre-commit-config.yaml` einmal `pre-commit clean && pre-commit
install --install-hooks`.

## Pull Request-Checkliste

Bevor du Review anforderst:

- [ ] `make lint` grün
- [ ] `make test` grün
- [ ] Bei Schema-Änderung: `make gen && make gen-check` grün
- [ ] Conventional-Commit-Subject, Body-Zeilen ≤ 100 Zeichen
- [ ] Signierter Commit (SSH key)
- [ ] PR-Beschreibung enthält Was/Warum/Test plan
- [ ] Self-Review durch
