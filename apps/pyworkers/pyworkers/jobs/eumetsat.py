"""EUMETSAT-Satelliten-Worker (EUMETView WMS, IR 10.8, Europa).

Pfad A der Konzept-Session 2.4: fertig gerenderte Composites werden
**server-seitig** vom öffentlichen EUMETView-WMS gezogen, in den
A.13-Hetzner-Object-Storage-Bucket (Prefix ``sat_prefix``) gelegt und
nur über ``media.worldweathernews.com`` ausgeliefert. Kein
Drittanbieter-Client-Pfad (A.19-konform), keine Auth nötig
(Q4 live verifiziert: ``view.eumetsat.int/geoserver/wms`` ist public).

Attribution: © EUMETSAT.

Idempotenz/Retention: pro Lauf ein Frame mit auf das Intervall
gerundetem UTC-Zeitstempel als Key — ein erneuter Lauf im selben Slot
überschreibt denselben Key (kein Doppel-Frame). Frames älter als
``window_hours`` werden gelöscht (rollierendes Fenster). ``index.json``
listet die verbleibenden Frames sortiert für das Frontend.

Roh-SEVIRI + Satpy (eigene Composites) ist bewusst NICHT hier — das
ist der K1-Evolutionspfad (~Iteration 2.6).
"""

from __future__ import annotations

import json
import math
from datetime import UTC, datetime, timedelta

import aiobotocore.session
import httpx
import structlog
from opentelemetry import trace

from pyworkers.config import Settings
from pyworkers.metrics import EUMETSAT_FETCHES_TOTAL, measure_job

log = structlog.get_logger(__name__)
tracer = trace.get_tracer("pyworkers.jobs.eumetsat")

WMS_ENDPOINT = "https://view.eumetsat.int/geoserver/wms"
WMS_LAYER = "msg_fes:ir108"  # FES = 0°/Full Earth Scan, deckt Europa
SOURCE = "eumetsat"
ATTRIBUTION = "© EUMETSAT"
USER_AGENT = "worldweathernews/0.7 (+https://worldweathernews.com)"

# Europa-Ausschnitt in EPSG:3857 (Q4 live getestet — liefert 200/PNG).
# WMS 1.3.0 + EPSG:3857 ⇒ Achsenreihenfolge x,y (Easting,Northing).
BBOX_3857 = (-2_000_000.0, 3_500_000.0, 4_500_000.0, 11_000_000.0)
IMG_WIDTH = 1024
IMG_HEIGHT = 1024
HTTP_TIMEOUT_SECONDS = 30.0

_WEB_MERCATOR_R = 6_378_137.0


def _mercator_to_lonlat(x: float, y: float) -> tuple[float, float]:
    """EPSG:3857 (m) → (lon, lat) in Grad."""
    lon = (x / _WEB_MERCATOR_R) * (180.0 / math.pi)
    lat = (2.0 * math.atan(math.exp(y / _WEB_MERCATOR_R)) - math.pi / 2.0) * (180.0 / math.pi)
    return lon, lat


def _geographic_bbox() -> dict[str, float]:
    """Geo-BBOX (lon/lat) für die MapLibre-Image-Source im Frontend."""
    lon_min, lat_min = _mercator_to_lonlat(BBOX_3857[0], BBOX_3857[1])
    lon_max, lat_max = _mercator_to_lonlat(BBOX_3857[2], BBOX_3857[3])
    return {
        "lonMin": round(lon_min, 5),
        "latMin": round(lat_min, 5),
        "lonMax": round(lon_max, 5),
        "latMax": round(lat_max, 5),
    }


def _slot(now: datetime, interval_seconds: int) -> datetime:
    """``now`` auf die letzte Intervall-Grenze (UTC) abrunden."""
    epoch = int(now.timestamp())
    floored = epoch - (epoch % interval_seconds)
    return datetime.fromtimestamp(floored, tz=UTC)


def _frame_key(prefix: str, ts: datetime) -> str:
    return f"{prefix}/{ts:%Y%m%dT%H%M}Z.png"


def _ts_from_key(prefix: str, key: str) -> datetime | None:
    name = key.removeprefix(f"{prefix}/").removesuffix(".png")
    try:
        return datetime.strptime(name, "%Y%m%dT%H%MZ").replace(tzinfo=UTC)
    except ValueError:
        return None


async def _fetch_frame(client: httpx.AsyncClient) -> bytes:
    params = {
        "service": "WMS",
        "version": "1.3.0",
        "request": "GetMap",
        "layers": WMS_LAYER,
        "styles": "",
        "format": "image/png",
        "transparent": "true",
        "crs": "EPSG:3857",
        "width": str(IMG_WIDTH),
        "height": str(IMG_HEIGHT),
        "bbox": ",".join(str(v) for v in BBOX_3857),
    }
    resp = await client.get(
        WMS_ENDPOINT,
        params=params,
        headers={"User-Agent": USER_AGENT},
        timeout=HTTP_TIMEOUT_SECONDS,
    )
    resp.raise_for_status()
    ctype = resp.headers.get("content-type", "")
    if not ctype.startswith("image/"):
        # EUMETView liefert Fehler als XML mit 200 — defensiv prüfen.
        raise RuntimeError(f"unexpected content-type from WMS: {ctype!r}")
    return resp.content


async def run(settings: Settings) -> None:
    """Einen IR-10.8-Frame holen, in den Bucket legen, Fenster pflegen."""
    async with measure_job("eumetsat"):
        s3_configured = (
            settings.s3_endpoint and settings.s3_access_key_id and settings.s3_secret_access_key
        )
        if not s3_configured:
            log.warning("eumetsat_skipped_no_s3_config")
            EUMETSAT_FETCHES_TOTAL.labels(status="skipped").inc()
            return

        prefix = settings.sat_prefix
        now = datetime.now(UTC)
        slot = _slot(now, settings.eumetsat_interval_seconds)
        cutoff = now - timedelta(hours=settings.eumetsat_window_hours)

        try:
            async with httpx.AsyncClient() as http_client:
                frame = await _fetch_frame(http_client)

            session = aiobotocore.session.get_session()
            async with session.create_client(
                "s3",
                endpoint_url=settings.s3_endpoint,
                region_name=settings.s3_region,
                aws_access_key_id=settings.s3_access_key_id,
                aws_secret_access_key=settings.s3_secret_access_key,
            ) as s3:
                await s3.put_object(
                    Bucket=settings.s3_bucket,
                    Key=_frame_key(prefix, slot),
                    Body=frame,
                    ContentType="image/png",
                    CacheControl="public, max-age=900",
                )

                # Bestehende Frames listen, alte löschen, Index bauen.
                kept: list[datetime] = []
                paginator = s3.get_paginator("list_objects_v2")
                async for page in paginator.paginate(
                    Bucket=settings.s3_bucket, Prefix=f"{prefix}/"
                ):
                    for obj in page.get("Contents", []):
                        key = obj["Key"]
                        if key.endswith("/index.json"):
                            continue
                        ts = _ts_from_key(prefix, key)
                        if ts is None:
                            continue
                        if ts < cutoff:
                            await s3.delete_object(Bucket=settings.s3_bucket, Key=key)
                        else:
                            kept.append(ts)

                kept.sort()
                index = {
                    "layer": "ir108",
                    "source": SOURCE,
                    "attribution": ATTRIBUTION,
                    "bbox": _geographic_bbox(),
                    "frames": [
                        {
                            "time": ts.isoformat().replace("+00:00", "Z"),
                            "url": (f"{settings.media_base_url}/{_frame_key(prefix, ts)}"),
                        }
                        for ts in kept
                    ],
                }
                await s3.put_object(
                    Bucket=settings.s3_bucket,
                    Key=f"{prefix}/index.json",
                    Body=json.dumps(index).encode("utf-8"),
                    ContentType="application/json",
                    CacheControl="public, max-age=60",
                )

            EUMETSAT_FETCHES_TOTAL.labels(status="ok").inc()
            log.info(
                "eumetsat_frame_persisted",
                slot=slot.isoformat(),
                frames_kept=len(kept),
                bytes=len(frame),
            )
        except Exception as e:
            EUMETSAT_FETCHES_TOTAL.labels(status="error").inc()
            log.exception("eumetsat_failed", error=str(e))
