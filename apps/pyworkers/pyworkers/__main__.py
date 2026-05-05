"""Einstiegspunkt: ``python -m pyworkers`` startet Scheduler + Heartbeat."""

import asyncio
import signal
from contextlib import suppress

import structlog
from apscheduler.schedulers.asyncio import AsyncIOScheduler
from apscheduler.triggers.interval import IntervalTrigger

from pyworkers.config import load_settings
from pyworkers.jobs import heartbeat
from pyworkers.logging import configure_logging
from pyworkers.metrics import start_metrics_server
from pyworkers.observability import init_tracing, instrument_libraries
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

    scheduler = AsyncIOScheduler()
    scheduler.add_job(
        heartbeat.run,
        trigger=IntervalTrigger(seconds=settings.heartbeat_interval_seconds),
        id="heartbeat",
        max_instances=1,
        coalesce=True,
    )
    scheduler.start()
    log.info(
        "scheduler_started",
        heartbeat_interval_seconds=settings.heartbeat_interval_seconds,
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
    log.info("stopped")


def main() -> None:
    with suppress(KeyboardInterrupt):
        asyncio.run(main_async())


if __name__ == "__main__":
    main()
