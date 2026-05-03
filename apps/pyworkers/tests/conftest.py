"""Test-Setup. structlog wird für jeden Test in einen ruhigen Modus gebracht."""

from collections.abc import Iterator

import pytest

from pyworkers.logging import configure_logging


@pytest.fixture(autouse=True)
def _configure_test_logging() -> Iterator[None]:
    configure_logging("DEBUG", "text")
    yield
