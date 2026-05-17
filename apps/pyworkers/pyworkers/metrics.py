"""Prometheus-Metriken und Mess-Helfer für Workers-Jobs."""

import time
from collections.abc import AsyncIterator
from contextlib import asynccontextmanager

from prometheus_client import Counter, Histogram, start_http_server

HEARTBEAT_TOTAL: Counter = Counter(
    "wwn_heartbeat_total",
    "Total number of heartbeat ticks",
    ["job"],
)

JOB_DURATION: Histogram = Histogram(
    "wwn_job_duration_seconds",
    "Job execution duration",
    ["job", "status"],
)

JOB_RUNS_TOTAL: Counter = Counter(
    "wwn_job_runs_total",
    "Total number of job runs",
    ["job", "status"],
)

OPEN_METEO_FETCHES_TOTAL: Counter = Counter(
    "wwn_open_meteo_fetches_total",
    "Total Open-Meteo API fetches, labeled by kind (current|hourly) and status (ok|error).",
    ["kind", "status"],
)

DWD_FETCHES_TOTAL: Counter = Counter(
    "wwn_dwd_fetches_total",
    "Total DWD POI fetches, labeled by status (ok|error|empty).",
    ["status"],
)

EUMETSAT_FETCHES_TOTAL: Counter = Counter(
    "wwn_eumetsat_fetches_total",
    "Total EUMETView WMS frame fetches, labeled by status (ok|error|skipped).",
    ["status"],
)


def start_metrics_server(port: int) -> None:
    """Startet den Prometheus-HTTP-Server auf dem angegebenen Port."""
    start_http_server(port)


@asynccontextmanager
async def measure_job(name: str) -> AsyncIterator[None]:
    """Context Manager der Job-Duration und -Status erfasst."""
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
