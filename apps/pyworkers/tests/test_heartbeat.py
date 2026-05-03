"""Heartbeat-Smoke + Logging-Setup-Test (robust ggü. prometheus_client-Internals)."""

import logging

import structlog

from pyworkers.jobs import heartbeat
from pyworkers.logging import configure_logging


async def test_heartbeat_run_does_not_crash() -> None:
    # Direkt ausführen — wir prüfen nur, dass der Job sauber durchläuft.
    await heartbeat.run()


def test_configure_logging_sets_root_level() -> None:
    configure_logging("WARNING", "text")
    # Stdlib-Root-Logger muss auf den neuen Level konfiguriert sein.
    assert logging.getLogger().level == logging.WARNING

    # structlog-Logger ist nutzbar und akzeptiert Felder.
    log = structlog.get_logger("test")
    log.warning("ok", field=1)
