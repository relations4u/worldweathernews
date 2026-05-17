"""Tests für den EUMETSAT-Satelliten-Worker.

Zwei Schichten (analog zu test_dwd.py):
- Pure Helpers ohne HTTP/S3: Slot-Rundung, Key-Bau/Parsing,
  Mercator→lon/lat und die Geo-BBOX.
- HTTP-Layer: _fetch_frame über ``httpx.MockTransport`` — WMS-Param-
  Aufbau, Content-Type-Guard, Fehler-Propagation.

S3-Persistierung/Fenster-Rotation läuft über aiobotocore und wird live
beim Deploy verifiziert (kein Mock hier — niedriger Wert, hoher Aufwand).
"""

from __future__ import annotations

from datetime import UTC, datetime

import httpx
import pytest

from pyworkers.jobs import eumetsat


def test_slot_floors_to_interval_boundary() -> None:
    now = datetime(2026, 5, 17, 13, 7, 23, tzinfo=UTC)
    assert eumetsat._slot(now, 900) == datetime(2026, 5, 17, 13, 0, 0, tzinfo=UTC)
    now2 = datetime(2026, 5, 17, 13, 52, 1, tzinfo=UTC)
    assert eumetsat._slot(now2, 900) == datetime(2026, 5, 17, 13, 45, 0, tzinfo=UTC)


def test_frame_key_format() -> None:
    ts = datetime(2026, 5, 17, 13, 45, tzinfo=UTC)
    assert eumetsat._frame_key("sat/ir108", ts) == "sat/ir108/20260517T1345Z.png"


def test_ts_from_key_roundtrips() -> None:
    ts = datetime(2026, 5, 17, 13, 45, tzinfo=UTC)
    key = eumetsat._frame_key("sat/ir108", ts)
    assert eumetsat._ts_from_key("sat/ir108", key) == ts


def test_ts_from_key_returns_none_for_garbage() -> None:
    assert eumetsat._ts_from_key("sat/ir108", "sat/ir108/index.json") is None
    assert eumetsat._ts_from_key("sat/ir108", "sat/ir108/not-a-date.png") is None


def test_mercator_origin_is_null_island() -> None:
    lon, lat = eumetsat._mercator_to_lonlat(0.0, 0.0)
    assert lon == pytest.approx(0.0, abs=1e-9)
    assert lat == pytest.approx(0.0, abs=1e-9)


def test_geographic_bbox_covers_europe() -> None:
    bbox = eumetsat._geographic_bbox()
    # Aus BBOX_3857 = (-2e6, 3.5e6, 4.5e6, 11e6) — grobe Europa-Box.
    assert bbox["lonMin"] == pytest.approx(-17.97, abs=0.1)
    assert bbox["lonMax"] == pytest.approx(40.42, abs=0.1)
    assert bbox["latMin"] == pytest.approx(29.92, abs=0.1)
    assert bbox["latMax"] == pytest.approx(69.83, abs=0.1)


@pytest.mark.asyncio
async def test_fetch_frame_builds_wms_request_and_returns_bytes() -> None:
    seen: dict[str, str] = {}

    def handler(request: httpx.Request) -> httpx.Response:
        seen.update(dict(request.url.params))
        seen["__ua"] = request.headers["user-agent"]
        return httpx.Response(200, content=b"\x89PNGfake", headers={"content-type": "image/png"})

    transport = httpx.MockTransport(handler)
    async with httpx.AsyncClient(transport=transport) as client:
        data = await eumetsat._fetch_frame(client)

    assert data == b"\x89PNGfake"
    assert seen["request"] == "GetMap"
    assert seen["layers"] == eumetsat.WMS_LAYER
    assert seen["crs"] == "EPSG:3857"
    assert seen["format"] == "image/png"
    assert seen["__ua"].startswith("worldweathernews/")


@pytest.mark.asyncio
async def test_fetch_frame_rejects_non_image_content_type() -> None:
    def handler(request: httpx.Request) -> httpx.Response:
        # EUMETView liefert Fehler als XML mit 200 — muss erkannt werden.
        return httpx.Response(
            200, content=b"<ServiceException/>", headers={"content-type": "application/xml"}
        )

    transport = httpx.MockTransport(handler)
    async with httpx.AsyncClient(transport=transport) as client:
        with pytest.raises(RuntimeError, match="unexpected content-type"):
            await eumetsat._fetch_frame(client)


@pytest.mark.asyncio
async def test_fetch_frame_propagates_http_errors() -> None:
    def handler(request: httpx.Request) -> httpx.Response:
        return httpx.Response(500)

    transport = httpx.MockTransport(handler)
    async with httpx.AsyncClient(transport=transport) as client:
        with pytest.raises(httpx.HTTPStatusError):
            await eumetsat._fetch_frame(client)
