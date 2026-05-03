"""Structlog-Setup für JSON oder Text-Output."""

import logging
import sys
from typing import Any

import structlog


def configure_logging(level: str, fmt: str) -> None:
    """Initialisiert structlog. `level` z.B. ``DEBUG``, `fmt` ``json`` oder ``text``."""
    log_level = getattr(logging, level.upper(), logging.INFO)

    structlog.configure(
        processors=_get_processors(fmt),
        wrapper_class=structlog.make_filtering_bound_logger(log_level),
        logger_factory=structlog.PrintLoggerFactory(file=sys.stdout),
        cache_logger_on_first_use=True,
    )

    logging.basicConfig(
        format="%(message)s",
        stream=sys.stdout,
        level=log_level,
    )


def _get_processors(fmt: str) -> list[Any]:
    common: list[Any] = [
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


def get_logger(name: str | None = None) -> Any:
    return structlog.get_logger(name)
