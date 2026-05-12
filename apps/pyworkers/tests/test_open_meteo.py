"""Tests für den Open-Meteo-Worker.

Drei Schichten getestet:
- Parser (pure, ohne HTTP/DB): parse_current und parse_hourly mit
  vollständigen sowie null-Werten.
- HTTP-Layer: fetch_current und fetch_hourly mit httpx.MockTransport
  (Query-Params, forecast_days=2 für hourly, Timezone-Param).

DB-Persistierung ist nicht getestet — würde testcontainers oder eine
Test-Postgres-Fixture brauchen, was über den Hello-World-Scope von
Iteration 2.1 hinausgeht. Idempotenz ist über `ON CONFLICT DO NOTHING`
in der DB realisiert und wurde in Schritt 4 live gegen die lokale DB
verifiziert (3 Rows nach zweimaligem run_current).
"""

from __future__ import annotations

from datetime import UTC, datetime
from zoneinfo import ZoneInfo

import httpx

from pyworkers.jobs import open_meteo


def _make_location() -> open_meteo.Location:
    return open_meteo.Location(
        id=1,
        slug="potsdam",
        name="Potsdam",
        latitude=52.3906,
        longitude=13.0645,
        timezone="Europe/Berlin",
    )


# ----------------------------------------------------------------------------
# parse_current
# ----------------------------------------------------------------------------


def test_parse_current_extracts_observation_with_local_time() -> None:
    loc = _make_location()
    data = {
        "current": {
            "time": "2026-05-12T09:00",
            "temperature_2m": 12.3,
            "precipitation": 0.0,
            "wind_speed_10m": 11.5,
            "wind_direction_10m": 245,
            "pressure_msl": 1013.2,
            "relative_humidity_2m": 67,
        }
    }
    row = open_meteo.parse_current(data, loc)
    assert row == (
        1,
        datetime(2026, 5, 12, 9, 0, tzinfo=ZoneInfo("Europe/Berlin")),
        12.3,
        0.0,
        11.5,
        245,
        1013.2,
        67.0,
        "open-meteo",
    )


def test_parse_current_handles_null_values() -> None:
    loc = _make_location()
    data = {
        "current": {
            "time": "2026-05-12T09:00",
            "temperature_2m": None,
            "precipitation": None,
            "wind_speed_10m": None,
            "wind_direction_10m": None,
            "pressure_msl": None,
            "relative_humidity_2m": None,
        }
    }
    row = open_meteo.parse_current(data, loc)
    assert row[2:8] == (None, None, None, None, None, None)


def test_parse_current_handles_missing_optional_fields() -> None:
    """Quelle liefert nur die 4 Kern-Variablen — Druck/Feuchte fehlen als Keys."""
    loc = _make_location()
    data = {
        "current": {
            "time": "2026-05-12T09:00",
            "temperature_2m": 7.4,
            "precipitation": 0.0,
            "wind_speed_10m": 9.7,
            "wind_direction_10m": 243,
        }
    }
    row = open_meteo.parse_current(data, loc)
    # pressure und humidity bleiben None, wenn die Keys fehlen
    assert row[6] is None
    assert row[7] is None


def test_parse_current_is_deterministic() -> None:
    loc = _make_location()
    data = {
        "current": {
            "time": "2026-05-12T09:00",
            "temperature_2m": 7.4,
            "precipitation": 0.0,
            "wind_speed_10m": 9.7,
            "wind_direction_10m": 243,
            "pressure_msl": 1013.2,
            "relative_humidity_2m": 67,
        }
    }
    assert open_meteo.parse_current(data, loc) == open_meteo.parse_current(data, loc)


# ----------------------------------------------------------------------------
# parse_hourly
# ----------------------------------------------------------------------------


def test_parse_hourly_emits_one_row_per_hour() -> None:
    loc = _make_location()
    run_at = datetime(2026, 5, 12, 7, 0, tzinfo=UTC)
    data = {
        "hourly": {
            "time": ["2026-05-12T00:00", "2026-05-12T01:00", "2026-05-12T02:00"],
            "temperature_2m": [10.0, 9.8, 9.5],
            "precipitation": [0.0, 0.0, 0.1],
            "wind_speed_10m": [5.0, 5.5, 6.0],
            "wind_direction_10m": [180, 190, 200],
        }
    }
    rows = open_meteo.parse_hourly(data, loc, run_at)
    assert len(rows) == 3
    # erste Zeile
    assert rows[0] == (
        1,
        datetime(2026, 5, 12, 0, 0, tzinfo=ZoneInfo("Europe/Berlin")),
        run_at,
        10.0,
        0.0,
        5.0,
        180,
        "open-meteo",
    )
    # alle Zeilen tragen dasselbe run_at — das ist die Idempotenz-Garantie
    # für (location_id, forecast_for, run_at) als PK.
    assert all(r[2] == run_at for r in rows)


def test_parse_hourly_with_null_values() -> None:
    loc = _make_location()
    run_at = datetime(2026, 5, 12, 7, 0, tzinfo=UTC)
    data = {
        "hourly": {
            "time": ["2026-05-12T00:00"],
            "temperature_2m": [None],
            "precipitation": [None],
            "wind_speed_10m": [None],
            "wind_direction_10m": [None],
        }
    }
    rows = open_meteo.parse_hourly(data, loc, run_at)
    assert rows[0][3:7] == (None, None, None, None)


# ----------------------------------------------------------------------------
# HTTP-Layer
# ----------------------------------------------------------------------------


async def test_fetch_current_passes_correct_query_params() -> None:
    loc = _make_location()
    captured: dict[str, httpx.URL] = {}

    def handler(request: httpx.Request) -> httpx.Response:
        captured["url"] = request.url
        return httpx.Response(
            200,
            json={
                "current": {
                    "time": "2026-05-12T09:00",
                    "temperature_2m": 1.0,
                    "precipitation": 0.0,
                    "wind_speed_10m": 1.0,
                    "wind_direction_10m": 0,
                }
            },
        )

    transport = httpx.MockTransport(handler)
    async with httpx.AsyncClient(transport=transport) as client:
        data = await open_meteo.fetch_current(client, loc)

    url = str(captured["url"])
    assert "open-meteo.com" in url
    assert "latitude=52.3906" in url
    assert "longitude=13.0645" in url
    # current-Liste umfasst die vier 2.1-Variablen plus pressure_msl +
    # relative_humidity_2m (Schritt 6). pressure_msl statt surface_pressure
    # für Konsistenz mit DWDs MSL-reduziertem Druck.
    assert "current=" in url
    assert "pressure_msl" in url
    assert "relative_humidity_2m" in url
    assert "timezone=Europe%2FBerlin" in url
    assert data["current"]["temperature_2m"] == 1.0


async def test_fetch_hourly_requests_forecast_days_2() -> None:
    loc = _make_location()
    captured: dict[str, httpx.URL] = {}

    def handler(request: httpx.Request) -> httpx.Response:
        captured["url"] = request.url
        return httpx.Response(
            200,
            json={
                "hourly": {
                    "time": [],
                    "temperature_2m": [],
                    "precipitation": [],
                    "wind_speed_10m": [],
                    "wind_direction_10m": [],
                }
            },
        )

    transport = httpx.MockTransport(handler)
    async with httpx.AsyncClient(transport=transport) as client:
        await open_meteo.fetch_hourly(client, loc)

    url = str(captured["url"])
    assert "hourly=temperature_2m%2Cprecipitation%2Cwind_speed_10m%2Cwind_direction_10m" in url
    assert "forecast_days=2" in url
    assert "timezone=Europe%2FBerlin" in url
