# wwn-pyworkers

Asynchroner Python-Worker-Service. Aktuell nur ein Heartbeat-Job; künftige
Sessions ergänzen Wetterdienst-Aggregatoren, GRIB-Parser, etc.

## Stack

- **uv** als Paketmanager
- **pydantic-settings** für Config (`WWN_PY_*`-ENVs)
- **structlog** für strukturierte Logs (JSON in prod, text in dev)
- **prometheus-client** für `/metrics`-Exposition (Port 9100)
- **apscheduler 3.x** (`AsyncIOScheduler`) für zeitgesteuerte Jobs
- **asyncpg** und **redis.asyncio** als Storage-Clients (Skeletons)
- **pytest** mit asyncio + cov; **mypy strict**

## Lokal starten

Über den Compose-Stack (empfohlen):

```bash
make dev    # vom Repo-Root
docker compose logs -f pyworkers
```

Direkt ohne Container (DB+Redis müssen parallel laufen, z.B. via `make dev`):

```bash
cd apps/pyworkers
uv sync
uv run python -m pyworkers
```

## Befehle

| Kommando         | Was                                      |
| ---------------- | ---------------------------------------- |
| `uv sync`        | venv anlegen + Dependencies installieren |
| `make run`       | einmaliger Run                           |
| `make dev`       | Hot-Reload via watchfiles                |
| `make test`      | pytest + Coverage                        |
| `make lint`      | ruff check + ruff format --check         |
| `make typecheck` | mypy strict                              |
| `make fmt`       | ruff format + ruff check --fix           |

## ENV-Variablen

Alle mit Prefix `WWN_PY_`:

| Variable                            | Default      | Wirkung                          |
| ----------------------------------- | ------------ | -------------------------------- |
| `WWN_PY_DATABASE_URL`               | —            | **Pflicht.** Postgres-DSN        |
| `WWN_PY_REDIS_URL`                  | —            | **Pflicht.** Redis-URL           |
| `WWN_PY_ENVIRONMENT`                | `production` | `dev`/`staging`/`production`     |
| `WWN_PY_LOG_LEVEL`                  | `INFO`       | `DEBUG`/`INFO`/`WARNING`/`ERROR` |
| `WWN_PY_LOG_FORMAT`                 | `json`       | `json` oder `text`               |
| `WWN_PY_METRICS_ENABLED`            | `true`       | `/metrics`-Server an/aus         |
| `WWN_PY_METRICS_PORT`               | `9100`       | Prom-Port                        |
| `WWN_PY_HEARTBEAT_INTERVAL_SECONDS` | `30`         | Heartbeat-Tick-Intervall         |

## Neuen Job hinzufügen

1. Modul anlegen: `pyworkers/jobs/<name>.py` mit:

   ```python
   import structlog
   from pyworkers.metrics import measure_job

   log = structlog.get_logger(__name__)

   async def run() -> None:
       async with measure_job("<name>"):
           # ... Job-Logik
           log.info("<name>_done")
   ```

2. In `pyworkers/__main__.py` einen `scheduler.add_job(...)`-Aufruf ergänzen
   (siehe Heartbeat als Vorlage).
3. Test in `tests/test_<name>.py` mit Smoke-Run.

## Metriken inspizieren

Im Container:

```bash
docker compose exec pyworkers python -c "import urllib.request; print(urllib.request.urlopen('http://localhost:9100/').read().decode())" | head -30
```

Lokal:

```bash
curl -s http://localhost:9100/ | grep -E '^wwn_'
```

Aktuell exportiert: `wwn_heartbeat_total`, `wwn_job_duration_seconds`,
`wwn_job_runs_total` (jeweils mit Job-Label).
