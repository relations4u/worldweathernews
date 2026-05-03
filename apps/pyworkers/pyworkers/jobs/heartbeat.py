"""Heartbeat-Job: bestätigt periodisch dass die Workers leben."""

import structlog

from pyworkers.metrics import HEARTBEAT_TOTAL, measure_job

log = structlog.get_logger(__name__)


async def run() -> None:
    async with measure_job("heartbeat"):
        HEARTBEAT_TOTAL.labels(job="heartbeat").inc()
        log.info("heartbeat", message="alive")
