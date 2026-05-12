"""Open-Meteo-Worker.

Holt periodisch aktuelle Werte und Stundenvorhersage für alle aktiven
Locations mit ``source = 'open-meteo'`` und persistiert sie in den
TimescaleDB-Hypertables ``observations`` und ``forecasts``.

Daten von Open-Meteo.com, CC BY 4.0
(https://open-meteo.com/en/docs).

Idempotenz: Beide Persistierungs-Statements verwenden
``ON CONFLICT DO NOTHING`` auf dem Primary Key, so dass wiederholte
Läufe (z. B. wegen current-API-15-min-Rastern bei 10-min-Polling) keine
Duplikate erzeugen.
"""

from __future__ import annotations

from datetime import UTC, datetime
from typing import Any, TypedDict
from zoneinfo import ZoneInfo

import asyncpg
import httpx
import structlog
from opentelemetry import trace

from pyworkers.metrics import OPEN_METEO_FETCHES_TOTAL, measure_job

log = structlog.get_logger(__name__)
tracer = trace.get_tracer("pyworkers.jobs.open_meteo")

OPEN_METEO_BASE = "https://api.open-meteo.com/v1/forecast"
SOURCE = "open-meteo"
ATTRIBUTION = "Daten von Open-Meteo.com, CC BY 4.0"

CURRENT_VARIABLES = "temperature_2m,precipitation,wind_speed_10m,wind_direction_10m"
HOURLY_VARIABLES = "temperature_2m,precipitation,wind_speed_10m,wind_direction_10m"

HTTP_TIMEOUT_SECONDS = 10.0
HOURLY_FORECAST_DAYS = 2  # deckt die nächsten 24 h sicher ab


class Location(TypedDict):
    """Felder, die der Worker aus der locations-Tabelle braucht."""

    id: int
    slug: str
    name: str
    latitude: float
    longitude: float
    timezone: str


ObservationRow = tuple[
    int,  # location_id
    datetime,  # observed_at
    float | None,  # temperature
    float | None,  # precipitation
    float | None,  # wind_speed
    int | None,  # wind_direction
    str,  # source
]

ForecastRow = tuple[
    int,  # location_id
    datetime,  # forecast_for
    datetime,  # run_at
    float | None,  # temperature
    float | None,  # precipitation
    float | None,  # wind_speed
    int | None,  # wind_direction
    str,  # source
]


# ----------------------------------------------------------------------------
# HTTP-Layer
# ----------------------------------------------------------------------------


async def fetch_current(client: httpx.AsyncClient, location: Location) -> dict[str, Any]:
    """Aktuellen Wert für eine Location abrufen."""
    params: dict[str, str | float | int] = {
        "latitude": location["latitude"],
        "longitude": location["longitude"],
        "current": CURRENT_VARIABLES,
        "timezone": location["timezone"],
    }
    response = await client.get(OPEN_METEO_BASE, params=params, timeout=HTTP_TIMEOUT_SECONDS)
    response.raise_for_status()
    data: dict[str, Any] = response.json()
    return data


async def fetch_hourly(client: httpx.AsyncClient, location: Location) -> dict[str, Any]:
    """Stundenvorhersage für eine Location abrufen."""
    params: dict[str, str | float | int] = {
        "latitude": location["latitude"],
        "longitude": location["longitude"],
        "hourly": HOURLY_VARIABLES,
        "forecast_days": HOURLY_FORECAST_DAYS,
        "timezone": location["timezone"],
    }
    response = await client.get(OPEN_METEO_BASE, params=params, timeout=HTTP_TIMEOUT_SECONDS)
    response.raise_for_status()
    data: dict[str, Any] = response.json()
    return data


# ----------------------------------------------------------------------------
# Parsing-Layer (pure, ohne HTTP/DB — testbar)
# ----------------------------------------------------------------------------


def parse_current(data: dict[str, Any], location: Location) -> ObservationRow:
    """Open-Meteo current-Response in ein observations-Row-Tuple wandeln."""
    current = data["current"]
    tz = ZoneInfo(location["timezone"])
    observed_at = datetime.fromisoformat(current["time"]).replace(tzinfo=tz)
    return (
        location["id"],
        observed_at,
        _as_float(current.get("temperature_2m")),
        _as_float(current.get("precipitation")),
        _as_float(current.get("wind_speed_10m")),
        _as_int(current.get("wind_direction_10m")),
        SOURCE,
    )


def parse_hourly(data: dict[str, Any], location: Location, run_at: datetime) -> list[ForecastRow]:
    """Open-Meteo hourly-Response in eine Liste forecast-Row-Tuples wandeln."""
    hourly = data["hourly"]
    tz = ZoneInfo(location["timezone"])
    times: list[str] = hourly["time"]
    temps: list[float | None] = hourly.get("temperature_2m", [None] * len(times))
    precs: list[float | None] = hourly.get("precipitation", [None] * len(times))
    winds: list[float | None] = hourly.get("wind_speed_10m", [None] * len(times))
    dirs: list[float | None] = hourly.get("wind_direction_10m", [None] * len(times))

    rows: list[ForecastRow] = []
    for i, t_str in enumerate(times):
        forecast_for = datetime.fromisoformat(t_str).replace(tzinfo=tz)
        rows.append(
            (
                location["id"],
                forecast_for,
                run_at,
                _as_float(temps[i]),
                _as_float(precs[i]),
                _as_float(winds[i]),
                _as_int(dirs[i]),
                SOURCE,
            )
        )
    return rows


def _as_float(v: Any) -> float | None:
    return None if v is None else float(v)


def _as_int(v: Any) -> int | None:
    return None if v is None else int(v)


# ----------------------------------------------------------------------------
# DB-Layer
# ----------------------------------------------------------------------------


OBSERVATION_INSERT = """
INSERT INTO observations
    (location_id, observed_at, temperature, precipitation,
     wind_speed, wind_direction, source, fetched_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())
ON CONFLICT (location_id, observed_at) DO NOTHING
"""

FORECAST_INSERT = """
INSERT INTO forecasts
    (location_id, forecast_for, run_at, temperature, precipitation,
     wind_speed, wind_direction, source, fetched_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW())
ON CONFLICT (location_id, forecast_for, run_at) DO NOTHING
"""


async def fetch_active_locations(pool: asyncpg.Pool) -> list[Location]:
    """Alle aktiven Open-Meteo-Locations aus der DB lesen."""
    rows = await pool.fetch(
        """
        SELECT id, slug, name, latitude, longitude, timezone
        FROM locations
        WHERE active = TRUE AND source = $1
        ORDER BY name
        """,
        SOURCE,
    )
    return [
        Location(
            id=r["id"],
            slug=r["slug"],
            name=r["name"],
            latitude=r["latitude"],
            longitude=r["longitude"],
            timezone=r["timezone"],
        )
        for r in rows
    ]


# ----------------------------------------------------------------------------
# Job-Entry-Points (vom Scheduler aufgerufen)
# ----------------------------------------------------------------------------


async def run_current(pool: asyncpg.Pool) -> None:
    """Aktuellen Wert für alle Locations holen und persistieren."""
    with tracer.start_as_current_span("open_meteo.run_current"):
        async with measure_job("open_meteo_current"):
            locations = await fetch_active_locations(pool)
            if not locations:
                log.info("open_meteo_current_no_locations")
                return
            async with httpx.AsyncClient() as client:
                for loc in locations:
                    await _process_current(client, pool, loc)


async def run_hourly(pool: asyncpg.Pool) -> None:
    """Stundenvorhersage für alle Locations holen und persistieren."""
    with tracer.start_as_current_span("open_meteo.run_hourly"):
        async with measure_job("open_meteo_hourly"):
            locations = await fetch_active_locations(pool)
            if not locations:
                log.info("open_meteo_hourly_no_locations")
                return
            run_at = datetime.now(UTC)
            async with httpx.AsyncClient() as client:
                for loc in locations:
                    await _process_hourly(client, pool, loc, run_at)


async def _process_current(client: httpx.AsyncClient, pool: asyncpg.Pool, loc: Location) -> None:
    try:
        data = await fetch_current(client, loc)
        row = parse_current(data, loc)
        await pool.execute(OBSERVATION_INSERT, *row)
        OPEN_METEO_FETCHES_TOTAL.labels(kind="current", status="ok").inc()
        log.info(
            "open_meteo_current_persisted",
            location=loc["slug"],
            observed_at=row[1].isoformat(),
            temperature=row[2],
        )
    except Exception as e:
        OPEN_METEO_FETCHES_TOTAL.labels(kind="current", status="error").inc()
        log.exception("open_meteo_current_failed", location=loc["slug"], error=str(e))


async def _process_hourly(
    client: httpx.AsyncClient,
    pool: asyncpg.Pool,
    loc: Location,
    run_at: datetime,
) -> None:
    try:
        data = await fetch_hourly(client, loc)
        rows = parse_hourly(data, loc, run_at)
        await pool.executemany(FORECAST_INSERT, rows)
        OPEN_METEO_FETCHES_TOTAL.labels(kind="hourly", status="ok").inc()
        log.info(
            "open_meteo_hourly_persisted",
            location=loc["slug"],
            rows=len(rows),
            run_at=run_at.isoformat(),
        )
    except Exception as e:
        OPEN_METEO_FETCHES_TOTAL.labels(kind="hourly", status="error").inc()
        log.exception("open_meteo_hourly_failed", location=loc["slug"], error=str(e))
