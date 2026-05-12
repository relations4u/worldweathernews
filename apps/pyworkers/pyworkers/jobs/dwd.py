"""DWD-POI-Worker.

Holt periodisch DWD-POI-Observations für alle aktiven Locations mit
``dwd_station_id IS NOT NULL`` aus opendata.dwd.de und persistiert sie
in die TimescaleDB-Hypertable ``observations``.

Datenbasis: Deutscher Wetterdienst, eigene Bearbeitung
(https://opendata.dwd.de/weather/weather_reports/poi/).

Format: halbstündliche CSV-Files pro Station,
``{station_id}-BEOB.csv``, semikolon-separiert, latin-1 encoded,
deutsches Dezimalkomma, ``---`` als NaN-Marker.

Idempotenz: ``ON CONFLICT (location_id, source, observed_at) DO UPDATE``,
so dass wiederholte Läufe (inkl. der ~48 historischen Rows pro CSV)
keine Duplikate erzeugen und DWD-Quality-Korrekturen propagiert werden.
"""

from __future__ import annotations

import csv
import io
from datetime import UTC, datetime
from typing import TypedDict

import asyncpg
import httpx
import structlog
from opentelemetry import trace

from pyworkers.metrics import DWD_FETCHES_TOTAL, measure_job

log = structlog.get_logger(__name__)
tracer = trace.get_tracer("pyworkers.jobs.dwd")

DWD_POI_BASE = "https://opendata.dwd.de/weather/weather_reports/poi"
SOURCE = "dwd"
USER_AGENT = "worldweathernews/0.5 (+https://worldweathernews.com)"
ATTRIBUTION = "Datenbasis: Deutscher Wetterdienst, eigene Bearbeitung"

# DWD-POI-Header hat drei Rows (Variablen-Namen / Units / deutsche Labels);
# Daten beginnen ab Row 4. Encoding ist latin-1 (Server liefert es als
# ``application/octet-stream``, aber Bytes sind ISO-Latin).
DWD_HEADER_ROWS = 3
DWD_NAN = "---"
DWD_CSV_ENCODING = "latin-1"

# Spalten-Mapping: DB-Feld → wortwörtlicher Spalten-Name aus DWD-Header-Row 1.
# Die Spaces in zwei Namen sind echte DWD-Eigenheiten — exakt so im Header,
# nicht "korrigieren".
DWD_COL_TEMPERATURE = "dry_bulb_temperature_at_2_meter_above_ground"
DWD_COL_PRECIPITATION = "precipitation_amount_last_hour"
DWD_COL_WIND_SPEED = "mean_wind_speed_during last_10_min_at_10_meters_above_ground"
DWD_COL_WIND_DIRECTION = "mean_wind_direction_during_last_10 min_at_10_meters_above_ground"
DWD_COL_PRESSURE = "pressure_reduced_to_mean_sea_level"
DWD_COL_HUMIDITY = "relative_humidity"

HTTP_TIMEOUT_SECONDS = 15.0


class DwdLocation(TypedDict):
    """Felder, die der DWD-Worker aus der locations-Tabelle braucht."""

    id: int
    slug: str
    name: str
    dwd_station_id: str


class DwdMeasurement(TypedDict):
    """Eine geparste POI-Datenzeile, ohne location_id (Caller füllt)."""

    observed_at: datetime
    temperature: float | None
    precipitation: float | None
    wind_speed: float | None
    wind_direction: int | None
    pressure: float | None
    humidity: float | None


# ----------------------------------------------------------------------------
# HTTP-Layer
# ----------------------------------------------------------------------------


async def fetch_poi_csv(client: httpx.AsyncClient, station_id: str) -> str:
    """POI-CSV für eine Station holen. Liefert den dekodierten Inhalt."""
    url = f"{DWD_POI_BASE}/{station_id}-BEOB.csv"
    headers = {"User-Agent": USER_AGENT, "Accept-Encoding": "gzip"}
    response = await client.get(url, headers=headers, timeout=HTTP_TIMEOUT_SECONDS)
    response.raise_for_status()
    return response.content.decode(DWD_CSV_ENCODING)


# ----------------------------------------------------------------------------
# Parsing-Layer (pure, ohne HTTP/DB — testbar)
# ----------------------------------------------------------------------------


def parse_poi_csv(raw_csv: str) -> list[DwdMeasurement]:
    """DWD-POI-CSV in eine Liste ``DwdMeasurement`` umwandeln.

    CSV-Layout:
      Zeile 1: Variablen-Namen (englisch, canonical). Index 0/1 sind
               Labels ("surface observations" / "Parameter description").
      Zeile 2: Units (Grad C, hPa, %, km/h, …). Index 0 ist Station-ID.
      Zeile 3: deutsche Beschreibungen.
      Zeile 4..N: Datenzeilen, absteigend nach Zeit (neueste zuerst).

    Datumsformat: ``DD.MM.YY`` in Spalte 0, ``HH:MM`` in Spalte 1, UTC.
    Decimal-Separator: deutsches Komma. NaN-Marker: ``---``.

    Zeilen mit nicht-parsbarem Timestamp werden geloggt und übersprungen.
    """
    reader = csv.reader(io.StringIO(raw_csv), delimiter=";")
    rows = list(reader)
    if len(rows) <= DWD_HEADER_ROWS:
        return []

    header = rows[0]
    col_idx: dict[str, int] = {name: i for i, name in enumerate(header)}

    out: list[DwdMeasurement] = []
    for data_row in rows[DWD_HEADER_ROWS:]:
        if len(data_row) < 2:
            continue
        try:
            observed_at = _parse_dwd_datetime(data_row[0], data_row[1])
        except ValueError:
            log.warning(
                "dwd_poi_row_skipped_bad_timestamp",
                date=data_row[0] if data_row else "",
                time=data_row[1] if len(data_row) > 1 else "",
            )
            continue
        out.append(
            DwdMeasurement(
                observed_at=observed_at,
                temperature=_get_float(data_row, col_idx, DWD_COL_TEMPERATURE),
                precipitation=_get_float(data_row, col_idx, DWD_COL_PRECIPITATION),
                wind_speed=_get_float(data_row, col_idx, DWD_COL_WIND_SPEED),
                wind_direction=_get_int(data_row, col_idx, DWD_COL_WIND_DIRECTION),
                pressure=_get_float(data_row, col_idx, DWD_COL_PRESSURE),
                humidity=_get_float(data_row, col_idx, DWD_COL_HUMIDITY),
            )
        )
    return out


def _parse_dwd_datetime(date_str: str, time_str: str) -> datetime:
    """``DD.MM.YY`` + ``HH:MM`` (UTC) → tz-aware datetime in UTC."""
    day, month, year_short = date_str.strip().split(".")
    hour, minute = time_str.strip().split(":")
    # POI publiziert nur 21.-Jh.-Daten — 2-stelliges Jahr direkt mappen.
    year = 2000 + int(year_short)
    return datetime(year, int(month), int(day), int(hour), int(minute), tzinfo=UTC)


def _get_float(row: list[str], idx: dict[str, int], col: str) -> float | None:
    pos = idx.get(col)
    if pos is None or pos >= len(row):
        return None
    raw = row[pos].strip()
    if not raw or raw == DWD_NAN:
        return None
    return float(raw.replace(",", "."))


def _get_int(row: list[str], idx: dict[str, int], col: str) -> int | None:
    val = _get_float(row, idx, col)
    if val is None:
        return None
    return round(val)


# ----------------------------------------------------------------------------
# DB-Layer
# ----------------------------------------------------------------------------


OBSERVATION_UPSERT = """
INSERT INTO observations
    (location_id, observed_at, temperature, precipitation,
     wind_speed, wind_direction, pressure, humidity, source, fetched_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW())
ON CONFLICT (location_id, source, observed_at) DO UPDATE SET
    temperature    = EXCLUDED.temperature,
    precipitation  = EXCLUDED.precipitation,
    wind_speed     = EXCLUDED.wind_speed,
    wind_direction = EXCLUDED.wind_direction,
    pressure       = EXCLUDED.pressure,
    humidity       = EXCLUDED.humidity,
    fetched_at     = NOW()
"""


async def fetch_active_locations(pool: asyncpg.Pool) -> list[DwdLocation]:
    """Alle aktiven Locations mit DWD-Station-ID aus der DB lesen."""
    rows = await pool.fetch(
        """
        SELECT id, slug, name, dwd_station_id
        FROM locations
        WHERE active = TRUE AND dwd_station_id IS NOT NULL
        ORDER BY name
        """,
    )
    return [
        DwdLocation(
            id=r["id"],
            slug=r["slug"],
            name=r["name"],
            dwd_station_id=r["dwd_station_id"],
        )
        for r in rows
    ]


# ----------------------------------------------------------------------------
# Job-Entry-Point (vom Scheduler aufgerufen)
# ----------------------------------------------------------------------------


async def run_poi(pool: asyncpg.Pool) -> None:
    """POI-CSV für alle DWD-Locations holen und persistieren."""
    with tracer.start_as_current_span("dwd.run_poi"):
        async with measure_job("dwd_poi"):
            locations = await fetch_active_locations(pool)
            if not locations:
                log.info("dwd_poi_no_locations")
                return
            async with httpx.AsyncClient() as client:
                for loc in locations:
                    await _process_location(client, pool, loc)


async def _process_location(
    client: httpx.AsyncClient,
    pool: asyncpg.Pool,
    loc: DwdLocation,
) -> None:
    try:
        raw = await fetch_poi_csv(client, loc["dwd_station_id"])
        measurements = parse_poi_csv(raw)
        if not measurements:
            log.warning(
                "dwd_poi_empty",
                location=loc["slug"],
                station=loc["dwd_station_id"],
            )
            DWD_FETCHES_TOTAL.labels(status="empty").inc()
            return
        rows = [
            (
                loc["id"],
                m["observed_at"],
                m["temperature"],
                m["precipitation"],
                m["wind_speed"],
                m["wind_direction"],
                m["pressure"],
                m["humidity"],
                SOURCE,
            )
            for m in measurements
        ]
        await pool.executemany(OBSERVATION_UPSERT, rows)
        DWD_FETCHES_TOTAL.labels(status="ok").inc()
        log.info(
            "dwd_poi_persisted",
            location=loc["slug"],
            station=loc["dwd_station_id"],
            rows=len(rows),
            latest_observed_at=measurements[0]["observed_at"].isoformat(),
        )
    except Exception as e:
        DWD_FETCHES_TOTAL.labels(status="error").inc()
        log.exception(
            "dwd_poi_failed",
            location=loc["slug"],
            station=loc["dwd_station_id"],
            error=str(e),
        )
