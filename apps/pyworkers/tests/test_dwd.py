"""Tests für den DWD-POI-Worker.

Drei Schichten getestet (analog zu test_open_meteo.py):
- Parser (pure, ohne HTTP/DB): parse_poi_csv gegen eine inline Fixture-CSV,
  inkl. NaN-Marker ``---``, deutsches Dezimalkomma und tz-aware UTC.
- HTTP-Layer: fetch_poi_csv über ``httpx.MockTransport`` — verifiziert URL-
  Aufbau, User-Agent-Header und Latin-1-Decoding.
- Helpers: _parse_dwd_datetime mit Edge-Cases (gültig / kaputt).

DB-Persistierung und Idempotenz sind live gegen die Dev-DB in Schritt 4
verifiziert (Re-Run → row-count delta=0).
"""

from __future__ import annotations

from datetime import UTC, datetime

import httpx
import pytest

from pyworkers.jobs import dwd

# Minimale, aber strukturgleiche POI-CSV-Fixture. Header-Reihe 1 enthält
# die kanonischen englischen Variablen-Namen, exakt wie DWD sie liefert —
# inklusive der zwei Spalten-Namen mit eingebetteten Spaces. Reihen 2/3
# (Units, deutsche Labels) werden vom Parser nicht verwendet, sind aber
# strukturell präsent. Datenreihen folgen ab Reihe 4, neueste zuerst.
FIXTURE_CSV = (
    "surface observations;Parameter description;"
    "dry_bulb_temperature_at_2_meter_above_ground;"
    "precipitation_amount_last_hour;"
    "mean_wind_speed_during last_10_min_at_10_meters_above_ground;"
    "mean_wind_direction_during_last_10 min_at_10_meters_above_ground;"
    "pressure_reduced_to_mean_sea_level;"
    "relative_humidity\n"
    "10384;Unit;Grad C;mm;km/h;Grad;hPa;%\n"
    "Datum;Uhrzeit (UTC);Temperatur (2m);Niederschlag (letzte Stunde);"
    "Windgeschwindigkeit;Windrichtung;Druck (auf Meereshoehe);Relative Feuchte\n"
    "12.05.26;17:00;7,5;0;19;280;1006,7;79\n"
    "12.05.26;16:00;---;0;---;---;1006,5;80\n"
    "not-a-date;junk;1;2;3;4;5;6\n"
)


# ----------------------------------------------------------------------------
# parse_poi_csv
# ----------------------------------------------------------------------------


def test_parse_poi_csv_extracts_full_measurement() -> None:
    measurements = dwd.parse_poi_csv(FIXTURE_CSV)
    # Erste Daten-Row wird komplett geparsed, kaputte Zeile (Row 6) wird
    # übersprungen → 2 valide Rows.
    assert len(measurements) == 2
    first = measurements[0]
    assert first["observed_at"] == datetime(2026, 5, 12, 17, 0, tzinfo=UTC)
    assert first["temperature"] == 7.5
    assert first["precipitation"] == 0.0
    assert first["wind_speed"] == 19.0
    assert first["wind_direction"] == 280
    assert first["pressure"] == 1006.7
    assert first["humidity"] == 79.0


def test_parse_poi_csv_handles_nan_marker() -> None:
    measurements = dwd.parse_poi_csv(FIXTURE_CSV)
    # Zweite Daten-Row hat '---' bei temperature/wind_speed/wind_direction.
    second = measurements[1]
    assert second["observed_at"] == datetime(2026, 5, 12, 16, 0, tzinfo=UTC)
    assert second["temperature"] is None
    assert second["wind_speed"] is None
    assert second["wind_direction"] is None
    # Andere Werte bleiben vorhanden.
    assert second["precipitation"] == 0.0
    assert second["pressure"] == 1006.5
    assert second["humidity"] == 80.0


def test_parse_poi_csv_skips_unparsable_timestamp() -> None:
    """Row mit kaputtem Datum/Uhrzeit wird ohne Exception übersprungen."""
    measurements = dwd.parse_poi_csv(FIXTURE_CSV)
    timestamps = [m["observed_at"] for m in measurements]
    # `not-a-date;junk;...` darf nicht im Ergebnis erscheinen
    assert all(ts.year == 2026 for ts in timestamps)


def test_parse_poi_csv_handles_empty_input() -> None:
    assert dwd.parse_poi_csv("") == []


def test_parse_poi_csv_handles_header_only_input() -> None:
    """Drei Header-Reihen ohne Datenreihen → leere Liste."""
    header_only = (
        "surface observations;Parameter description;dry_bulb_temperature_at_2_meter_above_ground\n"
        "10384;Unit;Grad C\n"
        "Datum;Uhrzeit (UTC);Temperatur (2m)\n"
    )
    assert dwd.parse_poi_csv(header_only) == []


def test_parse_poi_csv_returns_none_for_missing_columns() -> None:
    """Wenn eine Mapping-Spalte gar nicht im Header steht, ist das Feld None."""
    # Header ohne pressure-Spalte
    no_pressure = (
        "a;b;dry_bulb_temperature_at_2_meter_above_ground\n"
        "10384;Unit;Grad C\n"
        "Datum;Uhrzeit (UTC);Temperatur\n"
        "12.05.26;17:00;7,5\n"
    )
    measurements = dwd.parse_poi_csv(no_pressure)
    assert len(measurements) == 1
    assert measurements[0]["temperature"] == 7.5
    assert measurements[0]["pressure"] is None
    assert measurements[0]["humidity"] is None


def test_parse_poi_csv_is_deterministic() -> None:
    """Gleicher Input → identischer Output (Idempotenz-Voraussetzung für DB-Upsert)."""
    assert dwd.parse_poi_csv(FIXTURE_CSV) == dwd.parse_poi_csv(FIXTURE_CSV)


# ----------------------------------------------------------------------------
# _parse_dwd_datetime
# ----------------------------------------------------------------------------


def test_parse_dwd_datetime_two_digit_year_maps_to_2000s() -> None:
    # YY=26 muss zu 2026 werden, nicht 1926
    assert dwd._parse_dwd_datetime("12.05.26", "17:00") == datetime(2026, 5, 12, 17, 0, tzinfo=UTC)


def test_parse_dwd_datetime_strips_whitespace() -> None:
    assert dwd._parse_dwd_datetime(" 01.01.30 ", "  00:30  ") == datetime(
        2030, 1, 1, 0, 30, tzinfo=UTC
    )


def test_parse_dwd_datetime_raises_on_garbage() -> None:
    with pytest.raises(ValueError):
        dwd._parse_dwd_datetime("not-a-date", "17:00")


# ----------------------------------------------------------------------------
# HTTP-Layer
# ----------------------------------------------------------------------------


async def test_fetch_poi_csv_builds_correct_url_and_headers() -> None:
    captured: dict[str, object] = {}

    def handler(request: httpx.Request) -> httpx.Response:
        captured["url"] = str(request.url)
        captured["ua"] = request.headers.get("User-Agent")
        return httpx.Response(200, content=FIXTURE_CSV.encode("latin-1"))

    transport = httpx.MockTransport(handler)
    async with httpx.AsyncClient(transport=transport) as client:
        body = await dwd.fetch_poi_csv(client, "10384")

    assert captured["url"] == "https://opendata.dwd.de/weather/weather_reports/poi/10384-BEOB.csv"
    # User-Agent verweist auf das Projekt — DWD-Höflichkeit, kein Pflichtfeld
    assert "worldweathernews" in str(captured["ua"])
    # Latin-1-Decoding hat geklappt (CSV-Header beginnt mit ASCII)
    assert body.startswith("surface observations;")


async def test_fetch_poi_csv_propagates_http_errors() -> None:
    """404 vom DWD-File-Server wird als httpx.HTTPStatusError weitergereicht."""

    def handler(request: httpx.Request) -> httpx.Response:
        return httpx.Response(404)

    transport = httpx.MockTransport(handler)
    async with httpx.AsyncClient(transport=transport) as client:
        with pytest.raises(httpx.HTTPStatusError):
            await dwd.fetch_poi_csv(client, "99999")
