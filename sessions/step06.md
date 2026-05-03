# Session 6 — Python-Workers-Skelett

**Phase**: B (Services)
**Geschätzte Dauer**: 1-2 Stunden
**Vorbedingung**: Sessions 4 und 5 abgeschlossen, Stack komplett lauffähig.

## Ziel

Ein Python-Worker-Service ist aufgesetzt mit:
- uv als Paketmanager
- pydantic-settings für Config
- structlog für strukturierte Logs
- prometheus-client für Metriken
- apscheduler für zeitgesteuerte Jobs
- Stub-Heartbeat-Job, der alle 30s aktiv loggt
- Tests via pytest
- Dockerfile mit non-root user

In späteren Sessions kommen GRIB-Parser, Wetterdienst-Aggregatoren etc. hinzu.

## Vor-Klärung

Falls noch nicht klar:
- **Scheduler-Bibliothek**: apscheduler ist solide aber nicht mehr aktiv weiterentwickelt
  in der gleichen Form wie früher. Alternative: `arq` (Redis-basiert, async-first) oder
  `taskiq`. Default-Empfehlung: apscheduler 4.x, weil es für simple Cron-Tasks gut
  passt und keine zusätzliche Infrastruktur braucht. Wenn der Maintainer Redis-basierte
  Queues lieber will: arq.

## Aufgaben

### 1. uv-Projekt initialisieren

In `apps/pyworkers/` (das `pyproject.toml` aus Session 2 ist Stub, jetzt füllen):

```bash
cd apps/pyworkers
uv init --no-readme --package
```

Falls schon ein pyproject.toml da ist (sollte): nur die fehlenden Felder ergänzen,
Package-Layout (`pyworkers/`-Subverzeichnis) sicherstellen.

### 2. Dependencies

`pyproject.toml`:

```toml
[project]
name = "wwn-pyworkers"
version = "0.0.1"
description = "Worker services for worldweathernews.com"
requires-python = ">=3.12"
dependencies = [
    "pydantic>=2.7",
    "pydantic-settings>=2.3",
    "structlog>=24.1",
    "httpx>=0.27",
    "asyncpg>=0.29",
    "redis>=5.0",
    "prometheus-client>=0.20",
    "apscheduler>=4.0.0a5",  # 4.x ist async-first, ggf. neueste a-Version
    "click>=8.1",
]

[project.scripts]
pyworkers = "pyworkers.__main__:main"

[dependency-groups]
dev = [
    "pytest>=8.0",
    "pytest-asyncio>=0.23",
    "pytest-cov>=5.0",
    "ruff>=0.5",
    "mypy>=1.10",
    "watchfiles>=0.22",  # für hot-reload in dev
]

[build-system]
requires = ["hatchling"]
build-backend = "hatchling.build"

[tool.hatch.build.targets.wheel]
packages = ["pyworkers"]

[tool.ruff]
line-length = 100
target-version = "py312"

[tool.ruff.lint]
select = ["E", "F", "I", "N", "UP", "B", "SIM", "RUF", "S", "C4"]
ignore = ["S101"]  # assert ok in tests

[tool.ruff.format]
quote-style = "double"

[tool.mypy]
python_version = "3.12"
strict = true
plugins = ["pydantic.mypy"]

[tool.pytest.ini_options]
asyncio_mode = "auto"
testpaths = ["tests"]
addopts = "--cov=pyworkers --cov-report=term-missing"
```

Wenn apscheduler 4.x noch zu instabil: Fallback auf 3.x mit `AsyncIOScheduler`.

### 3. Verzeichnisstruktur

```
apps/pyworkers/
├── pyworkers/
│   ├── __init__.py
│   ├── __main__.py
│   ├── cli.py
│   ├── config.py
│   ├── logging.py
│   ├── metrics.py
│   ├── version.py
│   ├── storage/
│   │   ├── __init__.py
│   │   ├── postgres.py
│   │   └── redis.py
│   └── jobs/
│       ├── __init__.py
│       └── heartbeat.py
├── tests/
│   ├── __init__.py
│   ├── conftest.py
│   └── test_heartbeat.py
├── pyproject.toml
├── uv.lock
├── README.md
├── Dockerfile
├── .dockerignore
└── Makefile
```

### 4. `pyworkers/config.py`

```python
from pydantic import Field, PostgresDsn, RedisDsn
from pydantic_settings import BaseSettings, SettingsConfigDict


class Settings(BaseSettings):
    model_config = SettingsConfigDict(
        env_prefix="WWN_PY_",
        env_file=".env",
        env_file_encoding="utf-8",
        extra="ignore",
    )

    # General
    environment: str = Field(default="production", pattern="^(dev|staging|production)$")
    log_level: str = Field(default="INFO", pattern="^(DEBUG|INFO|WARNING|ERROR)$")
    log_format: str = Field(default="json", pattern="^(json|text)$")

    # Storage
    database_url: PostgresDsn
    redis_url: RedisDsn

    # Metrics
    metrics_enabled: bool = True
    metrics_port: int = Field(default=9100, ge=1, le=65535)

    # Heartbeat (für unsere Stub-Job)
    heartbeat_interval_seconds: int = Field(default=30, ge=1)


def load_settings() -> Settings:
    return Settings()  # type: ignore[call-arg]
```

### 5. `pyworkers/logging.py`

```python
import logging
import sys
from typing import Any

import structlog


def configure_logging(level: str, fmt: str) -> None:
    log_level = getattr(logging, level.upper(), logging.INFO)

    structlog.configure(
        processors=_get_processors(fmt),
        wrapper_class=structlog.make_filtering_bound_logger(log_level),
        logger_factory=structlog.PrintLoggerFactory(file=sys.stdout),
        cache_logger_on_first_use=True,
    )

    # Stdlib logging an structlog koppeln
    logging.basicConfig(
        format="%(message)s",
        stream=sys.stdout,
        level=log_level,
    )


def _get_processors(fmt: str) -> list[Any]:
    common = [
        structlog.contextvars.merge_contextvars,
        structlog.processors.add_log_level,
        structlog.processors.TimeStamper(fmt="iso"),
        structlog.processors.StackInfoRenderer(),
    ]
    if fmt == "json":
        common.append(structlog.processors.JSONRenderer())
    else:
        common.append(structlog.dev.ConsoleRenderer())
    return common


def get_logger(name: str | None = None) -> structlog.stdlib.BoundLogger:
    return structlog.get_logger(name)
```

### 6. `pyworkers/metrics.py`

```python
import asyncio
from contextlib import asynccontextmanager
from typing import AsyncIterator

from prometheus_client import Counter, Histogram, start_http_server

# Metriken
HEARTBEAT_TOTAL = Counter(
    "wwn_heartbeat_total",
    "Total number of heartbeat ticks",
    ["job"],
)

JOB_DURATION = Histogram(
    "wwn_job_duration_seconds",
    "Job execution duration",
    ["job", "status"],
)

JOB_RUNS_TOTAL = Counter(
    "wwn_job_runs_total",
    "Total number of job runs",
    ["job", "status"],
)


def start_metrics_server(port: int) -> None:
    start_http_server(port)


@asynccontextmanager
async def measure_job(name: str) -> AsyncIterator[None]:
    """Context Manager der Job-Duration und -Status erfasst."""
    import time
    start = time.monotonic()
    status = "success"
    try:
        yield
    except Exception:
        status = "error"
        raise
    finally:
        duration = time.monotonic() - start
        JOB_DURATION.labels(job=name, status=status).observe(duration)
        JOB_RUNS_TOTAL.labels(job=name, status=status).inc()
```

### 7. `pyworkers/jobs/heartbeat.py`

```python
import structlog

from pyworkers.metrics import HEARTBEAT_TOTAL, measure_job

log = structlog.get_logger(__name__)


async def run() -> None:
    """Heartbeat-Job: bestätigt periodisch dass die Workers leben."""
    async with measure_job("heartbeat"):
        HEARTBEAT_TOTAL.labels(job="heartbeat").inc()
        log.info("heartbeat", message="alive")
```

### 8. `pyworkers/__main__.py`

```python
import asyncio
import signal
from contextlib import suppress

import structlog

from pyworkers.config import load_settings
from pyworkers.jobs import heartbeat
from pyworkers.logging import configure_logging
from pyworkers.metrics import start_metrics_server
from pyworkers.version import VERSION

log = structlog.get_logger(__name__)


async def main_async() -> None:
    settings = load_settings()
    configure_logging(settings.log_level, settings.log_format)

    log.info(
        "starting",
        version=VERSION,
        environment=settings.environment,
    )

    if settings.metrics_enabled:
        start_metrics_server(settings.metrics_port)
        log.info("metrics_started", port=settings.metrics_port)

    # Scheduler aufsetzen
    from apscheduler.schedulers.asyncio import AsyncIOScheduler
    from apscheduler.triggers.interval import IntervalTrigger

    scheduler = AsyncIOScheduler()
    scheduler.add_job(
        heartbeat.run,
        trigger=IntervalTrigger(seconds=settings.heartbeat_interval_seconds),
        id="heartbeat",
        max_instances=1,
        coalesce=True,
    )
    scheduler.start()

    # Auf Shutdown-Signal warten
    stop_event = asyncio.Event()
    loop = asyncio.get_running_loop()

    def request_shutdown() -> None:
        log.info("shutdown_requested")
        stop_event.set()

    for sig in (signal.SIGINT, signal.SIGTERM):
        loop.add_signal_handler(sig, request_shutdown)

    await stop_event.wait()

    log.info("shutting_down")
    scheduler.shutdown(wait=True)
    log.info("stopped")


def main() -> None:
    with suppress(KeyboardInterrupt):
        asyncio.run(main_async())


if __name__ == "__main__":
    main()
```

Hinweis: Bei apscheduler 4.x ist die API anders (Scheduler ist async-context).
Bei Bedarf entsprechend anpassen.

### 9. `pyworkers/version.py`

```python
import os

VERSION = os.environ.get("WWN_PY_VERSION", "dev")
COMMIT = os.environ.get("WWN_PY_COMMIT", "unknown")
BUILD_DATE = os.environ.get("WWN_PY_BUILD_DATE", "unknown")
```

(Wird beim Container-Build via ENV gesetzt — siehe Dockerfile.)

### 10. `pyworkers/storage/postgres.py` und `redis.py`

Nur Skeleton:

```python
# postgres.py
from typing import Any
import asyncpg
import structlog

log = structlog.get_logger(__name__)


async def create_pool(database_url: str) -> asyncpg.Pool:
    pool = await asyncpg.create_pool(
        database_url,
        min_size=2,
        max_size=10,
        command_timeout=30,
    )
    log.info("postgres_pool_created")
    return pool


async def health_check(pool: asyncpg.Pool) -> bool:
    try:
        async with pool.acquire() as conn:
            await conn.execute("SELECT 1")
        return True
    except Exception as e:
        log.warning("postgres_health_check_failed", error=str(e))
        return False
```

Analog für Redis. Werden in späteren Sessions tatsächlich genutzt.

### 11. Tests

`tests/conftest.py`:

```python
import pytest
from pyworkers.logging import configure_logging


@pytest.fixture(autouse=True)
def configure_test_logging():
    configure_logging("DEBUG", "text")
```

`tests/test_heartbeat.py`:

```python
import pytest
from prometheus_client import REGISTRY

from pyworkers.jobs import heartbeat
from pyworkers.metrics import HEARTBEAT_TOTAL


async def test_heartbeat_runs_and_increments_counter():
    before = HEARTBEAT_TOTAL.labels(job="heartbeat")._value.get()
    await heartbeat.run()
    after = HEARTBEAT_TOTAL.labels(job="heartbeat")._value.get()
    assert after == before + 1
```

(Genaue Counter-Inspektion ist fragwürdig wegen prometheus_client-Internals.
Falls das nicht stabil ist: alternative Test, der einfach prüft dass `run()` nicht
crasht.)

### 12. Dockerfile

```dockerfile
# syntax=docker/dockerfile:1.7

FROM python:3.12-slim AS builder
ARG VERSION=dev
ARG COMMIT=unknown

# uv installieren
COPY --from=ghcr.io/astral-sh/uv:latest /uv /usr/local/bin/uv

WORKDIR /app
COPY pyproject.toml uv.lock ./
COPY pyworkers ./pyworkers

ENV UV_COMPILE_BYTECODE=1 \
    UV_LINK_MODE=copy \
    UV_PROJECT_ENVIRONMENT=/opt/venv

RUN uv sync --frozen --no-dev


FROM python:3.12-slim
LABEL org.opencontainers.image.source="https://github.com/<org>/worldweathernews"
LABEL org.opencontainers.image.title="wwn-pyworkers"

ARG VERSION=dev
ARG COMMIT=unknown
ARG BUILD_DATE=unknown

ENV PYTHONDONTWRITEBYTECODE=1 \
    PYTHONUNBUFFERED=1 \
    PATH="/opt/venv/bin:$PATH" \
    WWN_PY_VERSION=$VERSION \
    WWN_PY_COMMIT=$COMMIT \
    WWN_PY_BUILD_DATE=$BUILD_DATE

RUN groupadd -r app && useradd -r -g app -d /home/app -s /sbin/nologin app && \
    mkdir -p /home/app && chown -R app:app /home/app

WORKDIR /app
COPY --from=builder /opt/venv /opt/venv
COPY --from=builder /app/pyworkers /app/pyworkers

USER app
EXPOSE 9100
HEALTHCHECK --interval=10s --timeout=3s --start-period=10s --retries=3 \
    CMD python -c "import urllib.request; urllib.request.urlopen('http://localhost:9100/').read()" || exit 1

ENTRYPOINT ["python", "-m", "pyworkers"]
```

### 13. `.dockerignore`

```
.venv
__pycache__
*.pyc
.pytest_cache
.mypy_cache
.ruff_cache
.coverage
htmlcov
.git
*.md
!README.md
tests
```

### 14. `Makefile`

```makefile
.PHONY: install dev test lint typecheck fmt clean

install:
	uv sync

dev:
	uv run watchfiles --filter python "python -m pyworkers" pyworkers

run:
	uv run python -m pyworkers

test:
	uv run pytest

lint:
	uv run ruff check
	uv run ruff format --check

typecheck:
	uv run mypy pyworkers

fmt:
	uv run ruff format
	uv run ruff check --fix

clean:
	rm -rf .venv .pytest_cache .mypy_cache .ruff_cache .coverage htmlcov
```

### 15. Compose-Integration

`infra/compose/compose.dev.yml`:

```yaml
pyworkers:
  build:
    context: ../../apps/pyworkers
    dockerfile: Dockerfile
  container_name: wwn-pyworkers
  command: sh -c "uv sync && uv run watchfiles --filter python 'python -m pyworkers' pyworkers"
  working_dir: /app
  volumes:
    - ../../apps/pyworkers:/app
    - pyworkers_venv:/app/.venv
  environment:
    WWN_PY_DATABASE_URL: ${WWN_PY_DATABASE_URL}
    WWN_PY_REDIS_URL: ${WWN_PY_REDIS_URL}
    WWN_PY_LOG_LEVEL: ${WWN_PY_LOG_LEVEL:-DEBUG}
    WWN_PY_LOG_FORMAT: ${WWN_PY_LOG_FORMAT:-text}
    WWN_PY_ENVIRONMENT: dev
    WWN_PY_HEARTBEAT_INTERVAL_SECONDS: 30
  expose:
    - "9100"
  depends_on:
    postgres:
      condition: service_healthy
    redis:
      condition: service_healthy
```

**Wichtig**: Im Dev-Container brauchen wir uv und Dev-Deps. Daher entweder
ein Dev-Stage im Dockerfile (multi-target) oder das Compose-Image baut ohne
`--no-dev`. Pragmatisch: Compose nutzt `python:3.12-slim` direkt und installiert
Dependencies via uv beim Start (langsamer, aber einfacher).

Alternative: separater Dev-Dockerfile-Stage. Frag mich falls unklar.

Volumes-Section:
```yaml
pyworkers_venv:
```

### 16. Top-Level Makefile

```makefile
pyworkers-dev: ## Python-Workers hot-reload
	$(MAKE) -C apps/pyworkers dev

pyworkers-test: ## Python-Workers Tests
	$(MAKE) -C apps/pyworkers test

pyworkers-lint: ## Python-Workers Lint
	$(MAKE) -C apps/pyworkers lint
```

### 17. README

`apps/pyworkers/README.md`:
- Stack-Übersicht
- Lokale Befehle
- ENV-Variablen
- Wie ein neuer Job hinzugefügt wird (Pattern-Beschreibung)
- Wie Metriken inspiziert werden (`curl localhost:9100`)

## Vorgehen (verbindlich)

1. Plan zeigen
2. Freigabe abwarten
3. Implementieren
4. **Lokal außerhalb von Docker** ausprobieren: `uv sync && uv run python -m pyworkers`
   — sieht der Heartbeat-Log nach 30s gut aus?
5. `uv run pytest` grün?
6. `uv run ruff check` grün?
7. `uv run mypy pyworkers` grün? (mypy strict ist anspruchsvoll, ggf. Annotations
   ergänzen oder gezielt mit `# type: ignore[reason]` annotieren)
8. Im Container: `docker compose up pyworkers` zeigt Heartbeat-Logs
9. `docker compose exec pyworkers wget -qO- http://localhost:9100/` — Prometheus-Metriken
10. Nicht committen

## Erfolgs-Kriterien

- [ ] `uv sync` erfolgreich, lock file aktuell
- [ ] `uv run python -m pyworkers` läuft und loggt heartbeats
- [ ] `uv run pytest` grün
- [ ] `uv run ruff check && uv run ruff format --check` grün
- [ ] `uv run mypy pyworkers` grün (oder dokumentierte Ausnahmen)
- [ ] Im Container: heartbeat-Logs erscheinen alle 30s
- [ ] Metrics-Endpoint zeigt `wwn_heartbeat_total` und `wwn_job_duration_seconds`
- [ ] Hot-Reload via watchfiles funktioniert: Änderung an `heartbeat.py` triggert Restart
- [ ] Container-Image-Größe (final stage): < 200 MB

## Mögliche Stolpersteine

- **apscheduler-Version**: 4.x ist async-first aber als Alpha. 3.x funktioniert
  auch, hat etwas andere API. Bei Inkompatibilitäten: melden und 3.x nutzen.
- **mypy strict + structlog**: structlog hat lockere Types. Eventuell sind in
  Mode strict ein paar `Any`-Returns nicht exakt typisierbar — pragmatisch
  mit explizitem `cast` oder gezielten `# type: ignore` lösen.
- **uv im Container für dev**: `uv sync` braucht Schreibrechte auf `.venv`. Wenn
  Container-User non-root ist und Volume-Permissions schief: in dev als root
  starten oder Permissions vorab setzen.
- **prometheus_client und Counter-Inspektion in Tests**: Internals sind nicht
  stabil. Lieber pragmatisch testen.
- **uv und Lock-File**: `uv.lock` muss committed werden für Reproduzierbarkeit.

## Was diese Session NICHT tut

- Kein GRIB/NetCDF-Parsing
- Keine echten Wetterdienst-Integrationen
- Kein DB-Schema oder Modelle
- Kein Job-Queue mit Persistenz
- Keine Authentifizierung gegenüber externen APIs

## Suggested Commit-Message

```
feat(pyworkers): add Python worker skeleton with scheduler and metrics
```

---

## Zwischenstand nach Session 6

Du hast jetzt einen kompletten lokalen Stack:
- PostgreSQL + Redis + Caddy + Mailhog (Session 3)
- Go-Backend mit Health/Ping/Metrics (Session 4)
- SvelteKit-Frontend mit Backend-Connectivity (Session 5)
- Python-Workers mit Heartbeat (Session 6)

Das ist ein guter Punkt für ein Tag `v0.1.0-alpha` und eine kurze Pause. Optional
kannst du jetzt schon ein erstes Manuelles "Ende-zu-Ende"-Testszenario fahren:
Frontend startet, ruft Backend, Backend antwortet, Workers laufen, Logs erscheinen
strukturiert, Metriken sind alle da.

Phase C (CI/CD) kommt als nächstes: OpenAPI, GitHub Actions, Releases.
