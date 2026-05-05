"""Heartbeat-Job: bestätigt periodisch dass die Workers leben."""

import structlog
from opentelemetry import trace

from pyworkers.metrics import HEARTBEAT_TOTAL, measure_job

log = structlog.get_logger(__name__)
tracer = trace.get_tracer("pyworkers.jobs.heartbeat")


async def run() -> None:
    # Manueller Span — Auto-Instrumentation greift nur bei DB/HTTP-Calls,
    # und der Heartbeat macht beides nicht. Ohne diesen Span gäbe es keine
    # Trace-ID in den Logs und keinen Eintrag in Tempo.
    with tracer.start_as_current_span("heartbeat"):
        async with measure_job("heartbeat"):
            HEARTBEAT_TOTAL.labels(job="heartbeat").inc()
            log.info("heartbeat", message="alive")
