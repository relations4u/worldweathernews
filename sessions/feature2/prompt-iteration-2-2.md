# Iteration 2.2 — DWD-Adapter (POI-Observations)

**Übergabe-Prompt für Claude Code auf wwn-dev**

---

## Verwendung

Diesen Prompt **als ersten Prompt einer neuen Claude-Code-Session**
auf wwn-dev (10.100.100.113) verwenden, im Repo-Root von
`worldweathernews`. Voraussetzung: Iteration 2.1 ist als v0.4.2 live,
Konsolidierungs-Doku-PR ist gemerged (CLAUDE.md, feature-decisions.md,
feature-roadmap.md, STATUS.md auf 12.-Mai-Stand).

```
ssh hwr@10.100.100.113
cd ~/repos/worldweathernews
git checkout main && git pull
claude code
# → kompletten Inhalt unten als ersten Prompt einfügen
```

---

## Prompt für Claude Code (Copy-Paste ab hier)

---

Wir starten Iteration 2.2 (DWD-Adapter — POI-Observations) auf
worldweathernews.com. Track 2 Iteration 2.1 (Open-Meteo Hello World)
ist v0.4.2 live, das Worker → DB → API → Frontend-Pattern ist
erprobt. 2.2 erweitert die Pipeline um eine zweite Datenquelle
und drei neue Klimakontrast-Stationen.

Lies bitte zuerst:

1. `CLAUDE.md` im Repo-Root — die zentralen Spielregeln und neue
   Lessons aus 2.1 ("Häufige Fallen" wurde post-2.1 ergänzt)
2. `sessions/feature1/STATUS.md` und `sessions/feature2/STATUS.md`
3. `sessions/feature1/feature-decisions.md` — insbesondere die
   neuen Punkte aus Tranche 8:
   - B.5 (W1-Scheduling)
   - B.6 (Frontend-Position)
   - A.20 (OpenAPI ohne nullable)
   - A.21 (sqlc-Schema-Input)
   - A.22 (DB-Migrations als Deploy-Step)
4. `sessions/feature2/plan-iteration-2-2.md` — Plan-Skizze mit
   allen 2.1-Lessons eingearbeitet
5. Die 2.1-Commits zur Orientierung (`git log --oneline` für
   Branch `feat/iteration-2-1-open-meteo` oder gemergte PRs
   #69, #70, #71)

Sobald du diese gelesen hast, melde dich kurz mit einer Zusammen-
fassung von:

- Stack-Stand (v0.4.2 live, 3 Locations mit Open-Meteo)
- Den 6 DWD-Stationen, die in 2.2 dazukommen
- Den 5 Architektur-Lessons aus Tranche 8

## Feature-Phase-Modus

Wie üblich seit Track 1: **Claude Code committet nach expliziter
Freigabe.**

Workflow:

1. Branch anlegen (`feat/iteration-2-2-dwd-poi`)
2. Implementation in mehreren Commits
3. **Vor jedem Commit: mich um Freigabe fragen**
4. Bei "OK" / "commit" / "merge": committen oder mergen
5. Bei "warte" / "nochmal" / "anders": warten und nachbessern
6. Push zu GitHub: erst nach explizitem "push" oder "PR aufmachen"

## Was diese Iteration liefert

**Eine zweite Datenquelle in der bestehenden Pipeline:**

```
DWD OpenData API (POI-CSV)
     ↓
Python-Worker in apps/pyworkers (APScheduler, alle 30 Min)
     ↓
Postgres + TimescaleDB (gleiche Hypertable wie 2.1,
                        plus pressure + humidity Spalten)
     ↓
Go-Backend-API: bestehende Endpoints, source-Parameter
     ↓
SvelteKit-Frontend: /wetter zeigt jetzt 6 Cards mit Source-Label
```

**Konkreter Scope:**

- **6 DWD-Stationen**:
  - Stadt (passend zu 2.1-Locations):
    - Potsdam-Telegrafenberg (03342, 81 m)
    - Berlin-Tempelhof (10384, 48 m)
    - Hamburg-Fuhlsbüttel (10147, 11 m)
  - Klimakontraste (neue Locations):
    - Brocken (00078, 1134 m, Bergstation)
    - Zugspitze (05792, 2964 m, Hochgebirge)
    - Helgoland (02115, 4 m, Insel)
- **6 Variablen** (4 aus 2.1 plus 2 neu):
  - Temperatur, Niederschlag, Wind-Speed, Wind-Direction
  - **Druck (hPa, MSL)** — neu
  - **Luftfeuchte (%)** — neu
- **Frequenz**: current (DWD POI updated alle 30 Min)
- **Forecasts** in dieser Iteration NICHT — MOSMIX-KML kommt in 2.2b
- **Default-Quelle** für Frontend-Cards mit beiden Quellen: DWD
  (Open-Meteo bleibt verfügbar via `?source=open-meteo` Query)

## Konzept-Hintergrund

**Warum DWD jetzt:**

- 2.1-Pattern erprobt → DWD läuft sauberer durch
- Offizielle deutsche Datenquelle → bessere Plattform-Authentizität
- DWD POI-CSV ist Standard-Format, kein Aufwand wie GRIB/KMZ
- Klimakontrast-Stationen (Brocken/Zugspitze/Helgoland) machen die
  Plattform interessanter für 2.3 (Stations-Map)

**Lessons aus 2.1, die hier wichtig sind:**

- **A.22 (Deploy)**: DB-Migration läuft automatisch beim Deploy.
  Akzeptanzkriterium für 2.2: `make deploy` (oder Ansible-Playbook)
  muss **ohne manuelle Schritte** durchlaufen — kein `docker cp`
  oder ähnliches. v0.4.2-Pattern ist stabil.
- **A.20 (OpenAPI)**: neue Felder `pressure`, `humidity`,
  `dwd_station_id`, `altitude_m` **ohne `nullable`-Marker**. Sie
  bleiben `required: false` und werden in Go als Pointer generiert.
- **A.21 (sqlc-Schema)**: nach Migration → `make sqlc-schema` →
  `make sqlc-generate` → `make gen-check`. Generated `schema.sql`
  wird mit committed.
- **B.5 (Scheduling)**: APScheduler in-memory wie 2.1. Idempotenz
  im DB-Layer per `INSERT ... ON CONFLICT DO UPDATE`, nicht im
  Worker.
- **B.6 (Frontend)**: `/wetter` bleibt eine Route, bleibt
  `ssr = false`. SSR-Upgrade ist Backlog-Punkt.

## Iterations-Plan

### Schritt 1 — Branch + Plan + Verifikation

1. Branch anlegen: `feat/iteration-2-2-dwd-poi`
2. Verifikation:
   - `uname -n` zeigt `wwn-dev`
   - `git rev-parse --show-toplevel` ist Repo-Root
   - `git config --get user.email` = `hwr@relations4u.de`
3. Live-Stand prüfen:
   ```
   curl -s https://api.research.worldweathernews.com/api/v1/locations
   # Sollte 3 Locations zurückgeben (Potsdam, Berlin, Hamburg)
   ```
4. Plan-Vorschlag für Schritte 2-11, dann Freigabe abwarten

### Schritt 2 — DB-Migration

Neue Migration `infra/migrations/NNN_dwd_poi_stations_and_variables.sql`:

```sql
-- +goose Up

-- locations: dwd_station_id + altitude_m
ALTER TABLE locations
  ADD COLUMN dwd_station_id TEXT,
  ADD COLUMN altitude_m INTEGER;

CREATE INDEX idx_locations_dwd_station_id
  ON locations (dwd_station_id)
  WHERE dwd_station_id IS NOT NULL;

-- observations: pressure + humidity
ALTER TABLE observations
  ADD COLUMN pressure DOUBLE PRECISION,
  ADD COLUMN humidity DOUBLE PRECISION;

-- Bestehende Stadt-Locations mit DWD-Station-ID + Altitude befüllen
UPDATE locations SET dwd_station_id = '03342', altitude_m = 81
  WHERE slug = 'potsdam';
UPDATE locations SET dwd_station_id = '10384', altitude_m = 48
  WHERE slug = 'berlin';
UPDATE locations SET dwd_station_id = '10147', altitude_m = 11
  WHERE slug = 'hamburg';

-- Drei neue Klimakontrast-Locations (DWD-only)
INSERT INTO locations
  (slug, name, country, latitude, longitude, timezone, source,
   dwd_station_id, altitude_m)
VALUES
  ('brocken', 'Brocken', 'DE', 51.7991, 10.6178,
   'Europe/Berlin', 'dwd', '00078', 1134),
  ('zugspitze', 'Zugspitze', 'DE', 47.4209, 10.9854,
   'Europe/Berlin', 'dwd', '05792', 2964),
  ('helgoland', 'Helgoland', 'DE', 54.1827, 7.8868,
   'Europe/Berlin', 'dwd', '02115', 4);

-- +goose Down

DELETE FROM locations
  WHERE slug IN ('brocken', 'zugspitze', 'helgoland');

UPDATE locations
  SET dwd_station_id = NULL, altitude_m = NULL
  WHERE slug IN ('potsdam', 'berlin', 'hamburg');

ALTER TABLE observations
  DROP COLUMN humidity,
  DROP COLUMN pressure;

DROP INDEX IF EXISTS idx_locations_dwd_station_id;

ALTER TABLE locations
  DROP COLUMN altitude_m,
  DROP COLUMN dwd_station_id;
```

Verifikation lokal:

```
make migrate  # oder goose direkt
psql -c "SELECT slug, dwd_station_id, altitude_m FROM locations ORDER BY slug;"
# Sollte 6 Zeilen zeigen, alle mit dwd_station_id
```

### Schritt 3 — sqlc-Schema + Queries regenerieren

```
make sqlc-schema       # Pre-Processing aus goose-Migrations
make sqlc-generate     # sqlc liest schema.sql + queries
make gen-check         # CI-Drift-Check
```

Eventuelle neue Queries hinzufügen, falls Backend sie braucht
(siehe Schritt 5).

### Schritt 4 — Python-Worker (DWD)

`apps/pyworkers/wwn_pyworkers/workers/dwd.py`:

```python
"""
DWD Worker.

Fetcht alle 30 Minuten POI-Observations für alle aktiven DWD-Locations.

DWD POI: https://opendata.dwd.de/weather/weather_reports/poi/
Format:  CSV mit Header, semikolon-separiert, latin-1 encoded
Attribution: Datenbasis: Deutscher Wetterdienst, eigene Bearbeitung
"""

import asyncio
from datetime import datetime, timezone

import httpx
import structlog
from sqlalchemy import select
from sqlalchemy.ext.asyncio import AsyncSession

from wwn_pyworkers.config import settings
from wwn_pyworkers.db import get_session
from wwn_pyworkers.models import Location, Observation

logger = structlog.get_logger(__name__)

DWD_POI_BASE = "https://opendata.dwd.de/weather/weather_reports/poi"
USER_AGENT = "worldweathernews/0.5 (+https://worldweathernews.com)"
ATTRIBUTION = "Datenbasis: Deutscher Wetterdienst, eigene Bearbeitung"


async def fetch_poi_csv(station_id: str) -> str:
    """Fetch the POI CSV for one station. Returns raw text."""
    url = f"{DWD_POI_BASE}/{station_id}-BEOB.csv"
    headers = {"User-Agent": USER_AGENT, "Accept-Encoding": "gzip"}
    async with httpx.AsyncClient(timeout=15.0) as client:
        response = await client.get(url, headers=headers)
        response.raise_for_status()
        # DWD POI ist latin-1, nicht utf-8
        return response.content.decode("latin-1")


def parse_poi_csv(raw_csv: str) -> list[dict]:
    """
    Parse DWD POI CSV.

    Format:
    - Erste Zeile: Header (Variablen-Namen)
    - Zweite Zeile: Einheiten
    - Dritte Zeile: Beschreibungen
    - Ab vierter Zeile: Datenzeilen
    - Trennzeichen: ";"
    - Datumsformat: TT.MM.YY;HH:MM
    - NaN-Marker: "---"
    """
    # Implementation: csv.DictReader mit semicolon, latin-1 already decoded
    # observed_at aus erstem + zweitem Feld (Datum + Uhrzeit)
    # → tz-aware UTC datetime
    # Variable-Mapping: TT → temperature, RR1c → precipitation_1h,
    #                    FF → wind_speed, DD → wind_direction,
    #                    PPPP → pressure (MSL), Td → dew_point,
    #                    RH → humidity (oder aus T+Td berechnet)
    # Genaue Variablen-Namen aus DWD-Header lesen, mapping in
    # einem Dict am Modul-Top.
    raise NotImplementedError  # zur Klarheit in der Übergabe


async def run_dwd_poi_for_all_stations() -> None:
    """Iterate over all DWD-source locations, fetch and persist."""
    async with get_session() as session:
        result = await session.execute(
            select(Location).where(
                Location.active == True,
                Location.dwd_station_id.isnot(None),
            )
        )
        locations = result.scalars().all()

    for loc in locations:
        try:
            raw = await fetch_poi_csv(loc.dwd_station_id)
            measurements = parse_poi_csv(raw)
            await persist_observations(loc, measurements)
            logger.info("dwd-poi persisted",
                        location=loc.slug,
                        station=loc.dwd_station_id,
                        rows=len(measurements))
        except Exception as e:
            logger.exception("dwd-poi failed",
                             location=loc.slug,
                             station=loc.dwd_station_id,
                             error=str(e))


# persist_observations mit ON CONFLICT (location_id, observed_at)
# DO UPDATE SET ... (idempotent bei Re-Runs)
```

Verifikation:

```
# Eine Station manuell holen und parsen
docker compose exec pyworkers python -c "
import asyncio
from wwn_pyworkers.workers.dwd import fetch_poi_csv, parse_poi_csv
csv = asyncio.run(fetch_poi_csv('10384'))
print(csv[:500])
"
```

### Schritt 5 — Worker-Scheduling integrieren

In `apps/pyworkers/wwn_pyworkers/__main__.py`:

```python
scheduler.add_job(
    run_dwd_poi_for_all_stations,
    "interval",
    minutes=30,
    next_run_time=datetime.now(timezone.utc),  # initial run on startup
    id="dwd-poi",
    name="DWD POI Observations",
)
```

(Open-Meteo-Jobs aus 2.1 bleiben unverändert.)

### Schritt 6 — Open-Meteo um pressure + humidity erweitern (Backfill)

Damit die WeatherCard für `?source=open-meteo` auch Druck/Feuchte
zeigt:

In `workers/open_meteo.py` die `current`-Variables-Liste erweitern:

```python
"current": "temperature_2m,precipitation,wind_speed_10m,wind_direction_10m,surface_pressure,relative_humidity_2m",
```

Plus `persist_current` um `pressure` und `humidity` aus dem
Response-JSON ergänzen.

Test lokal, dass alte 2.1-Locations weiterhin saubere Daten haben.

### Schritt 7 — Backend-API erweitern

`apps/backend/internal/http/handler/locations.go`:

**GET /api/v1/locations:**

- Antwortet jetzt mit **allen 6 aktiven** Locations
- Antwort-Schema erweitert um:
  - `altitudeM` (optional, Pointer)
  - `dwdStationId` (optional, Pointer)
  - `availableSources` (Array von Strings, z.B. `["dwd"]` oder
    `["dwd", "open-meteo"]`)

**GET /api/v1/locations/{slug}:**

- Optional Query: `source=dwd` | `source=open-meteo`
- Default-Strategie: DWD wenn `dwd_station_id IS NOT NULL`,
  sonst Open-Meteo
- Response-Schema: Source-Field auf jeder Observation/Forecast
  ist schon da (aus 2.1)
- Wenn beide Quellen verfügbar: gewählte Quelle wird zurückgegeben,
  `availableSources` listet beide

OpenAPI-Schema in `packages/api-schema/openapi.yaml` ergänzen.
Code-Generierung über `make openapi-gen` (Go-Stubs + TS-Types).

**Wichtig (A.20):** neue optionale Felder **ohne `nullable`** —
nur `required: false`. oapi-codegen erzeugt Pointer.

### Schritt 8 — Frontend-Integration

`apps/frontend/src/routes/wetter/+page.ts` und `+page.svelte`:

- Loader holt `/api/v1/locations` → 6 Locations
- Pro Location ein zweiter Fetch auf `/api/v1/locations/{slug}` —
  nutzt Default-Source
- Optional: für Stadt-Locations zweiter Fetch mit
  `?source=open-meteo` für Comparison (kann weglassen, Vergleich
  ist 2.x-Folge)

`WeatherCard.svelte` erweitern:

- Source-Label oben rechts: „DWD" oder „Open-Meteo" als kleines
  Badge
- Höhen-Anzeige unter dem Stadt-Namen (z.B. „Zugspitze · 2964 m"),
  nur wenn `altitudeM > 100`
- Druck-Anzeige (falls vorhanden): hPa-Wert
- Luftfeuchte-Anzeige (falls vorhanden): %-Wert

Mobile-Responsive testen: 6 Cards sollen auf Mobile sauber
untereinander, auf Desktop in 2-3 Spalten.

Paraglide-Strings (DE + EN):

- `weather_source_dwd`
- `weather_source_open_meteo`
- `weather_pressure`
- `weather_humidity`
- `weather_altitude`

### Schritt 9 — Quellen-Attribution-Page erweitern

`apps/frontend/src/content/pages/de/quellen-attribution.md` (und
`/en/source-attribution.md`):

```markdown
## Deutscher Wetterdienst (DWD)

Aktuelle Stationsbeobachtungen für Standorte in Deutschland.

- API: opendata.dwd.de/weather/weather_reports/poi/
- Lizenz: GeoNutzV (CC-BY-äquivalent für offene Geodaten)
- Attribution: „Datenbasis: Deutscher Wetterdienst, eigene Bearbeitung"
- Genutzte Stationen: Potsdam-Telegrafenberg (03342),
  Berlin-Tempelhof (10384), Hamburg-Fuhlsbüttel (10147),
  Brocken (00078), Zugspitze (05792), Helgoland (02115)
- Genutzte Variablen: Lufttemperatur, Niederschlag,
  Windgeschwindigkeit, Windrichtung, Luftdruck (MSL),
  relative Luftfeuchte
- Aktualisierung: alle 30 Minuten
```

### Schritt 10 — Tests + Smoke-Checks

Pyworkers:

- Test für `parse_poi_csv` mit Fixture-CSV aus echtem DWD-Sample
- Test für `persist_observations`-Idempotenz (zwei Runs für gleichen
  observed_at überschreiben nicht widerholt)

Backend:

- Test für neuen `source`-Query-Param
- Test für „beide Quellen verfügbar" → Default-DWD-Auswahl

E2E manuell:

- Worker einmal manuell ausführen
- API liefert 6 Locations
- /wetter zeigt 6 Cards
- Mobile-Responsive
- Source-Label sichtbar
- Lighthouse keine Regression

### Schritt 11 — Doku

1. **`docs/data-sources.md`** erweitern:
   - DWD-Block (Endpoints, Attribution, Rate-Limits)
   - Worker-Frequenzen aktualisiert

2. **`docs/runbook.md`** erweitern:
   - Szenario „DWD-Worker liefert keine Daten"
   - Diagnose: HTTP-Status, CSV-Parse, DB-Insert

3. **`CLAUDE.md`** „Wo finde ich was"-Tabelle ergänzen falls nötig

4. **`docs/backlog.md`** ergänzen um:
   - MOSMIX-KML-Forecast (Iteration 2.2b)
   - Quellen-Vergleich-UI (eigene Folge-Iteration)
   - Stationsdaten-Backfill aus DWD-CDC-Archiv (Klima-Iteration)

5. **`sessions/feature2/STATUS.md`** aktualisieren mit 2.2-Done

## Akzeptanzkriterien

- [ ] Branch `feat/iteration-2-2-dwd-poi` angelegt
- [ ] Migration legt 6 Locations + 2 neue Spalten an
- [ ] sqlc-Generierung grün, `make gen-check` grün
- [ ] DWD-Worker fetcht aus echter opendata.dwd.de
- [ ] Worker speichert Observations idempotent (ON CONFLICT)
- [ ] Open-Meteo-Worker erweitert um pressure + humidity
- [ ] Backend-Endpoints liefern 6 Locations mit source-Parameter
- [ ] OpenAPI-Schema erweitert ohne `nullable`-Marker
- [ ] Frontend zeigt 6 WeatherCards mit Source-Label
- [ ] Höhen-Anzeige für Klimakontrast-Stationen (>100m)
- [ ] Druck + Feuchte in Card sichtbar
- [ ] Quellen-Attribution-Page enthält DWD-Block (DE + EN)
- [ ] Paraglide-Strings für alle neuen UI-Labels (DE + EN)
- [ ] Mobile-Responsive
- [ ] **Deploy auf wwn-prod läuft ohne manuelle Migrations-Schritte**
      durch (A.22-Akzeptanz)
- [ ] `make lint && make test` grün
- [ ] `docs/data-sources.md` + Runbook + Backlog aktualisiert
- [ ] `sessions/feature2/STATUS.md` reflektiert 2.2-Done
- [ ] PR-Erstellung erst nach finalem OK des Maintainers
- [ ] Tag v0.5.0 als Iteration-2.2-Release

## Was du **noch nicht** baust

- **MOSMIX-Forecasts** → Iteration 2.2b (KMZ-Parsing eigene
  Komplexität)
- **Quellen-Vergleich-UI** (DWD nebeneinander mit Open-Meteo) →
  eigene Folge-Iteration
- **DWD-Historie** (CDC-Archive) → erste Klima-Iteration
- **MapLibre-Karte** mit allen 6 Stationen → Iteration 2.3
- **SSR für /wetter** → eigener Backlog-Punkt

Wenn du verlockt bist, MOSMIX-Forecasts oder Quellen-Vergleich
gleich mitzubauen — widerstehen, Iterations-Disziplin halten.

## Wenn etwas unklar ist

Frag mich. Insbesondere:

- **DWD-CSV-Variablen-Namen**: die Spalten heißen evtl. anders als
  Open-Meteo (TT, RR1c, FF, DD, PPPP, RH oder T+Td). Erst CSV
  inspizieren, dann Mapping vorschlagen.
- **DWD-NaN-Marker**: vermutlich `---`, aber bestätigen lassen
- **Idempotenz-Strategie**: ON CONFLICT DO UPDATE oder DO NOTHING?
  → Vorschlag: DO UPDATE für Resilienz bei Quality-Korrekturen
- **Frontend-Layout** mit 6 Cards: Wie soll die Mobile-Reihenfolge
  sein? Vorschlag alphabetisch nach Stadt-Name.

Lass uns loslegen. Bestätige mir kurz, dass du die Dokumente
gelesen hast, und schlag den ersten Schritt vor.
