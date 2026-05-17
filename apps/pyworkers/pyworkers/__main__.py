"""Einstiegspunkt: ``python -m pyworkers`` startet Scheduler + Heartbeat."""

import asyncio
import signal
from contextlib import suppress
from datetime import UTC, datetime

import structlog
from apscheduler.schedulers.asyncio import AsyncIOScheduler
from apscheduler.triggers.interval import IntervalTrigger

from pyworkers.config import load_settings
from pyworkers.jobs import dwd, eumetsat, heartbeat, open_meteo
from pyworkers.logging import configure_logging
from pyworkers.metrics import start_metrics_server
from pyworkers.observability import init_tracing, instrument_libraries
from pyworkers.storage.postgres import create_pool
from pyworkers.version import COMMIT, VERSION

log = structlog.get_logger(__name__)


async def main_async() -> None:
    settings = load_settings()
    configure_logging(settings.log_level, settings.log_format)

    # Tracing-Setup VOR den Auto-Instrumentations-Aufrufen — sonst greift
    # der TracerProvider noch nicht und die Spans gehen ins Leere.
    init_tracing(
        enabled=settings.tracing_enabled,
        endpoint=settings.tracing_endpoint,
        service_name="wwn-pyworkers",
        environment=settings.environment,
        version=VERSION,
    )
    if settings.tracing_enabled:
        instrument_libraries()
        log.info("tracing_enabled", endpoint=settings.tracing_endpoint)

    log.info(
        "starting",
        version=VERSION,
        commit=COMMIT,
        environment=settings.environment,
    )

    if settings.metrics_enabled:
        start_metrics_server(settings.metrics_port)
        log.info("metrics_started", port=settings.metrics_port)

    pool = await create_pool(str(settings.database_url))

    scheduler = AsyncIOScheduler()
    scheduler.add_job(
        heartbeat.run,
        trigger=IntervalTrigger(seconds=settings.heartbeat_interval_seconds),
        id="heartbeat",
        max_instances=1,
        coalesce=True,
    )

    if settings.open_meteo_enabled:
        scheduler.add_job(
            open_meteo.run_current,
            args=[pool],
            trigger=IntervalTrigger(seconds=settings.open_meteo_current_interval_seconds),
            id="open_meteo_current",
            max_instances=1,
            coalesce=True,
        )
        scheduler.add_job(
            open_meteo.run_hourly,
            args=[pool],
            trigger=IntervalTrigger(seconds=settings.open_meteo_hourly_interval_seconds),
            id="open_meteo_hourly",
            max_instances=1,
            coalesce=True,
        )

    if settings.dwd_enabled:
        # DWD-POI publiziert alle 30 Min — Worker zieht halbstündlich nach.
        # next_run_time=now: erster Lauf beim Container-Start, damit nach
        # einem Deploy nicht bis zu 30 Min gewartet wird, bis Daten in der
        # DB sind. APScheduler in-memory (B.5), bei Restart vergessen.
        scheduler.add_job(
            dwd.run_poi,
            args=[pool],
            trigger=IntervalTrigger(seconds=settings.dwd_poi_interval_seconds),
            id="dwd_poi",
            max_instances=1,
            coalesce=True,
            next_run_time=datetime.now(UTC),
        )

    if settings.eumetsat_enabled:
        # EUMETView publiziert ~alle 15 Min — Worker zieht im selben
        # Takt nach. next_run_time=now: erster Frame direkt nach
        # Container-Start, damit der Loop nach einem Deploy nicht erst
        # leer ist. APScheduler in-memory (B.5).
        scheduler.add_job(
            eumetsat.run,
            args=[settings],
            trigger=IntervalTrigger(seconds=settings.eumetsat_interval_seconds),
            id="eumetsat",
            max_instances=1,
            coalesce=True,
            next_run_time=datetime.now(UTC),
        )

    scheduler.start()
    log.info(
        "scheduler_started",
        heartbeat_interval_seconds=settings.heartbeat_interval_seconds,
        open_meteo_enabled=settings.open_meteo_enabled,
        open_meteo_current_interval_seconds=settings.open_meteo_current_interval_seconds,
        open_meteo_hourly_interval_seconds=settings.open_meteo_hourly_interval_seconds,
        dwd_enabled=settings.dwd_enabled,
        dwd_poi_interval_seconds=settings.dwd_poi_interval_seconds,
        eumetsat_enabled=settings.eumetsat_enabled,
        eumetsat_interval_seconds=settings.eumetsat_interval_seconds,
    )

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
    await pool.close()
    log.info("stopped")


def main() -> None:
    with suppress(KeyboardInterrupt):
        asyncio.run(main_async())


if __name__ == "__main__":
    main()
