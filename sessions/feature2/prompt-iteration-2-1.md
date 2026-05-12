# Iteration 2.1 — Open-Meteo Hello World

**Übergabe-Prompt für Claude Code auf wwn-dev**

---

## Verwendung

Diesen Prompt **als ersten Prompt einer neuen Claude-Code-Session**
auf wwn-dev (10.100.100.113) verwenden, im Repo-Root von
`worldweathernews`. Voraussetzung: Security-Triage-PR
`chore/security-triage-post-v0-0-4` ist gemerged und v0.0.5 (oder
v0.0.4 plus alle Security-Fixes) läuft live.

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

Hallo Claude Code. Wir starten **Track 2** der Feature-Phase auf
worldweathernews.com. Track 1 ist bei Iteration 1.3a fertig, die
Plattform läuft auf v0.0.4 (oder v0.0.5 nach Security-Triage-PR)
mit Compliance-Pages, Hetzner Object Storage, mdsvex + Paraglide-
i18n und Sveltia-CMS. Bild-Upload (1.3b) ist bewusst aufgeschoben
bis Blog-Bedarf entsteht.

Track 2 bringt **echte Wetterdaten** ins System. Diese Iteration
ist das Hello World — ein **Worker → DB → API → Frontend-Pattern**,
am einfachsten Beispiel: Open-Meteo, drei Städte, vier Variablen.
Das ist Architektur-Härtetest, kein Featuristen-Reichtum.

Lies bitte zuerst:

1. `CLAUDE.md` im Repo-Root — die zentralen Spielregeln
2. `STATUS.md` und `sessions/STATUS.md` — letzter Stand der Sessions
3. `sessions/feature1/STATUS.md` — Track 1 Status
4. Externe Tracking-Dokumente (Pfad vom Maintainer, typischerweise
   `/home/hwr/wwn-handover/` oder ähnlich):
   - `feature-decisions.md` (insbesondere B.1, B.4)
   - `feature-roadmap.md` (Iteration 2.1)

Sobald du diese gelesen hast, melde dich kurz mit einer Zusammen-
fassung des Stack-Stands plus der vier Entscheidungs-Punkte aus B.1
(Locations, Variablen, Frequenzen, Storage), damit ich sicher bin
dass du den Kontext hast.

## Feature-Phase-Modus

**Wichtige Abweichung von der Setup-Phase:** In der Setup-Phase galt
"Maintainer committet selbst." In der Feature-Phase gilt **"Claude Code
committet nach expliziter Freigabe."**

Workflow:

1. Branch anlegen (`feat/iteration-2-1-open-meteo`)
2. Implementation in mehreren Commits auf dem Branch
3. **Vor jedem Commit: mich um Freigabe fragen**
4. Bei "OK" oder "commit" oder "merge": committen oder mergen
5. Bei "warte" oder "nochmal" oder "anders": warten und nachbessern
6. Push zu GitHub: erst nach explizitem "push" oder "PR aufmachen"

Kein eigenständiger Commit ohne Freigabe.

## Was diese Iteration liefert

**Eine durchgängige Daten-Pipeline:**

```
Open-Meteo API
     ↓
Python-Worker in apps/pyworkers (APScheduler oder systemd-timer)
     ↓
Postgres + TimescaleDB Hypertables
     ↓
Go-Backend-API: GET /api/v1/locations, GET /api/v1/observations/:slug
     ↓
SvelteKit-Frontend: Startseite zeigt aktuelles Wetter für 3 Städte
```

**Konkreter Scope** (siehe B.1 in `feature-decisions.md`):

- **3 Städte**: Potsdam, Berlin, Hamburg
- **4 Variablen**: Temperatur (°C), Niederschlag (mm), Wind-
  geschwindigkeit (km/h), Windrichtung (°)
- **2 Frequenzen**:
  - `current` (Single-Datapoint „jetzt")
  - `hourly` Forecast für die nächsten 24h
- **Keine Historie** in 2.1 (Forecast-First, Archive kommt mit
  Klima-Features)
- **Storage**: Postgres + TimescaleDB-Hypertables, kein S3

## Konzept-Hintergrund

**Warum Open-Meteo zuerst:**

- REST/JSON, kein Auth, kein Aggregations-Preprocessing
- CC-BY-4.0-Attribution (Footer-Snippet + `/quellen-attribution`)
- Einfachste Datenquelle zum Aufbau des Worker-Patterns
- DWD kommt in Iteration 2.2 mit erprobter Pipeline

**Architektur-Entscheidungen (siehe `feature-decisions.md`):**

- B.1: Open-Meteo zuerst, drei Städte, vier Variablen
- B.4: CC-BY-4.0-Attribution, Pattern „Footer-Link + Detail-Page"
- A.16: Self-hosted Proxmox (Pyworkers laufen auf wwn-prod)
- A.17: Paraglide-Strings, nicht hardcoded
- A.19: Self-hosting-Prinzip — falls Worker-Scheduling extern
  gehostet wäre, müssten wir das gegen A.19 abgleichen.
  APScheduler oder systemd-timer auf wwn-prod sind beide konform.

## Iterations-Plan

### Schritt 1 — Branch + Plan + Verifikation

1. Branch anlegen: `feat/iteration-2-1-open-meteo`
2. Verifikation: bist du auf `wwn-dev` (uname -n)? Bist du im richtigen
   Repo-Root (git rev-parse --show-toplevel)?
3. Maintainer-Identität prüfen (`git config --get user.email` muss
   `hwr@relations4u.de` zeigen)
4. v0.0.4+ live? `curl -I https://research.worldweathernews.com`
   und `curl https://api.research.worldweathernews.com/api/v1/ping`
   müssen 200 zeigen.
5. Sobald alles OK: kurzen Plan zeigen wie du Schritte 2-8 angehst,
   dann Freigabe vom Maintainer abwarten

### Schritt 2 — DB-Schema und Migration

`infra/db/migrations/` (goose) — neue Migration mit drei Tabellen:

```sql
-- 0NN_open_meteo_hello_world.sql

-- +goose Up

CREATE TABLE locations (
    id          BIGSERIAL PRIMARY KEY,
    slug        TEXT NOT NULL UNIQUE,        -- 'potsdam', 'berlin', 'hamburg'
    name        TEXT NOT NULL,                -- 'Potsdam', 'Berlin', 'Hamburg'
    country     TEXT NOT NULL DEFAULT 'DE',
    latitude    DOUBLE PRECISION NOT NULL,
    longitude   DOUBLE PRECISION NOT NULL,
    timezone    TEXT NOT NULL DEFAULT 'Europe/Berlin',
    source      TEXT NOT NULL,                -- 'open-meteo', 'dwd', ...
    active      BOOLEAN NOT NULL DEFAULT TRUE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Seed-Daten direkt in Migration für Hello World
-- (späteres Iterations-Pattern: Seeding über eigenes Skript)
INSERT INTO locations (slug, name, latitude, longitude, source) VALUES
    ('potsdam', 'Potsdam', 52.3906, 13.0645, 'open-meteo'),
    ('berlin',  'Berlin',  52.5200, 13.4050, 'open-meteo'),
    ('hamburg', 'Hamburg', 53.5511, 9.9937,  'open-meteo');

-- Observations: Hypertable für aktuelle Messwerte
CREATE TABLE observations (
    location_id  BIGINT NOT NULL REFERENCES locations(id) ON DELETE CASCADE,
    observed_at  TIMESTAMPTZ NOT NULL,
    temperature  DOUBLE PRECISION,             -- °C
    precipitation DOUBLE PRECISION,            -- mm
    wind_speed   DOUBLE PRECISION,             -- km/h
    wind_direction INTEGER,                    -- 0-360
    source       TEXT NOT NULL,                -- 'open-meteo'
    fetched_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (location_id, observed_at)
);

SELECT create_hypertable('observations', 'observed_at');
CREATE INDEX idx_observations_location_observed
    ON observations (location_id, observed_at DESC);

-- Forecasts: Hypertable für stündliche Vorhersagen
CREATE TABLE forecasts (
    location_id    BIGINT NOT NULL REFERENCES locations(id) ON DELETE CASCADE,
    forecast_for   TIMESTAMPTZ NOT NULL,       -- für welchen Zeitpunkt gilt der Forecast
    run_at         TIMESTAMPTZ NOT NULL,       -- wann wurde der Forecast generiert
    temperature    DOUBLE PRECISION,
    precipitation  DOUBLE PRECISION,
    wind_speed     DOUBLE PRECISION,
    wind_direction INTEGER,
    source         TEXT NOT NULL,
    fetched_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (location_id, forecast_for, run_at)
);

SELECT create_hypertable('forecasts', 'forecast_for');
CREATE INDEX idx_forecasts_location_forecast_for
    ON forecasts (location_id, forecast_for DESC);

-- +goose Down

DROP TABLE IF EXISTS forecasts;
DROP TABLE IF EXISTS observations;
DROP TABLE IF EXISTS locations;
```

Verifikation:

- `make migrate` auf wwn-dev läuft durch
- TimescaleDB-Hypertables existieren (`SELECT * FROM timescaledb_information.hypertables`)
- Seed-Daten für drei Städte sichtbar

### Schritt 3 — sqlc-Queries für Backend-Access

`apps/backend/internal/storage/queries/locations.sql` und
`observations.sql`:

```sql
-- name: ListActiveLocations :many
SELECT id, slug, name, country, latitude, longitude, timezone, source
FROM locations
WHERE active = TRUE
ORDER BY name;

-- name: GetLocationBySlug :one
SELECT id, slug, name, country, latitude, longitude, timezone, source
FROM locations
WHERE slug = $1 AND active = TRUE;

-- name: GetLatestObservation :one
SELECT observed_at, temperature, precipitation, wind_speed, wind_direction,
       source, fetched_at
FROM observations
WHERE location_id = $1
ORDER BY observed_at DESC
LIMIT 1;

-- name: GetForecastNext24h :many
SELECT forecast_for, temperature, precipitation, wind_speed, wind_direction,
       run_at
FROM forecasts
WHERE location_id = $1
  AND forecast_for > NOW()
  AND forecast_for <= NOW() + INTERVAL '24 hours'
  AND run_at = (SELECT MAX(run_at) FROM forecasts WHERE location_id = $1)
ORDER BY forecast_for;
```

Verifikation:

- `make sqlc-generate` läuft
- Generierte Go-Files erscheinen in `apps/backend/internal/storage/`

### Schritt 4 — Python-Worker (apps/pyworkers)

`apps/pyworkers/wwn_pyworkers/workers/open_meteo.py`:

```python
"""
Open-Meteo Worker.

Fetcht alle 10 Minuten 'current' Werte und alle 60 Minuten
'hourly' Forecast für alle aktiven Locations mit source='open-meteo'.

Open-Meteo API: https://open-meteo.com/en/docs
Attribution: Daten von Open-Meteo.com, CC BY 4.0
"""

import asyncio
import logging
from datetime import datetime, timezone

import httpx
import structlog
from sqlalchemy import select
from sqlalchemy.ext.asyncio import AsyncSession

from wwn_pyworkers.config import settings
from wwn_pyworkers.db import get_session
from wwn_pyworkers.models import Location, Observation, Forecast

logger = structlog.get_logger(__name__)

OPEN_METEO_BASE = "https://api.open-meteo.com/v1/forecast"
ATTRIBUTION = "Daten von Open-Meteo.com, CC BY 4.0"


async def fetch_current(location: Location) -> dict:
    """Fetch current weather for one location."""
    params = {
        "latitude": location.latitude,
        "longitude": location.longitude,
        "current": "temperature_2m,precipitation,wind_speed_10m,wind_direction_10m",
        "timezone": location.timezone,
    }
    async with httpx.AsyncClient(timeout=10.0) as client:
        response = await client.get(OPEN_METEO_BASE, params=params)
        response.raise_for_status()
        return response.json()


async def fetch_hourly_forecast(location: Location) -> dict:
    """Fetch hourly forecast for next 24h for one location."""
    params = {
        "latitude": location.latitude,
        "longitude": location.longitude,
        "hourly": "temperature_2m,precipitation,wind_speed_10m,wind_direction_10m",
        "forecast_days": 2,  # to ensure 24h ahead are covered
        "timezone": location.timezone,
    }
    async with httpx.AsyncClient(timeout=10.0) as client:
        response = await client.get(OPEN_METEO_BASE, params=params)
        response.raise_for_status()
        return response.json()


async def run_current_for_all_locations() -> None:
    """Iterate over all active open-meteo locations and persist current."""
    async with get_session() as session:
        result = await session.execute(
            select(Location).where(
                Location.active == True,
                Location.source == "open-meteo",
            )
        )
        locations = result.scalars().all()

    for loc in locations:
        try:
            data = await fetch_current(loc)
            await persist_current(loc, data)
            logger.info("open-meteo current persisted",
                        location=loc.slug,
                        temperature=data["current"]["temperature_2m"])
        except Exception as e:
            logger.exception("open-meteo current failed",
                             location=loc.slug, error=str(e))


# ... persist_current, run_hourly_for_all_locations, persist_forecast ...

# Scheduling via APScheduler oder systemd-timer
# Entscheidung im Implementation: APScheduler einfacher, eine Library mehr.
# systemd-timer braucht Service-File, robuster aber mehr Setup.
# Vorschlag: APScheduler für Iteration 2.1 (im selben Container)
```

**Scheduling-Frage** — bitte klären mit Maintainer:

- **Option W1: APScheduler im Worker-Container** — eine Library, alles
  im selben Python-Prozess. Beim Container-Restart starten Jobs neu.
  Trade-off: Cron-State im Memory, kein Persistent-State.
- **Option W2: Separater Cron-Container** (z. B. `mcuadros/ofelia`)
  oder systemd-Timer auf wwn-prod, der `docker exec` macht.
  Trade-off: zwei Tools.
- **Option W3: APScheduler mit PostgresJobStore** —
  Persistent-State, Failures überleben Restart.

Empfehlung Phase-1: **W1** (einfachste Option). Migration zu W3 möglich,
wenn Job-State-Verlust ein Problem wird.

### Schritt 5 — Backend-API-Endpoints

Neuer Handler in `apps/backend/internal/api/handlers/locations.go`:

```go
// GET /api/v1/locations
// Returns: [{slug, name, latitude, longitude}, ...]
// Lists all active locations.

// GET /api/v1/locations/{slug}
// Returns: location with latest observation and next 24h forecast
// {
//   "location": {slug, name, ...},
//   "current": {observed_at, temperature, ...},
//   "forecast": [{forecast_for, temperature, ...}, ...],
//   "attribution": "Daten von Open-Meteo.com, CC BY 4.0"
// }
```

OpenAPI-Schema in `packages/api-schema/openapi.yaml` ergänzen.
Code-Generierung über `make openapi-gen` (Go-Stubs + TS-Types).

Verifikation:

- `curl http://localhost:8080/api/v1/locations` zeigt drei Städte
- `curl http://localhost:8080/api/v1/locations/potsdam` zeigt
  Current + Forecast-Array
- `make test` grün

### Schritt 6 — Frontend-Integration

`apps/frontend/src/routes/+page.svelte` erweitern (oder neue Route
`apps/frontend/src/routes/wetter/+page.svelte` falls Hero-Page
Compliance-Banner-fokussiert bleiben soll — mit Maintainer klären):

- Loader holt `/api/v1/locations` (Liste) und für jede Location
  `/api/v1/locations/{slug}` (Details)
- WeatherCard-Component pro Stadt: Name, Temperatur jetzt,
  Niederschlag-Vorschau 24h, Wind (Geschwindigkeit + Richtung als
  Pfeil)
- Footer-Snippet: „Daten von Open-Meteo.com, CC BY 4.0" mit Link
  zu `/quellen-attribution`
- Paraglide-Strings für Variablen-Labels (de + en)
- Mobile-Responsive

WeatherCard-Component-Skizze:

```svelte
<!-- apps/frontend/src/lib/components/WeatherCard.svelte -->
<script lang="ts">
  import * as m from '$lib/paraglide/messages.js';
  export let location: Location;
  export let current: Observation;
  export let forecast: Forecast[];

  // Helper für Windrichtung → kompass-Pfeil-Rotation
</script>

<article class="rounded-lg border p-4">
  <h2>{location.name}</h2>
  <p>{m.weather_temperature_current()}: {current.temperature}°C</p>
  <!-- ... -->
</article>
```

### Schritt 7 — Quellen-Attribution-Page erweitern

`apps/frontend/src/content/pages/de/quellen-attribution.md` und
`/en/source-attribution.md` ergänzen um den Open-Meteo-Block:

```markdown
## Open-Meteo

Datenquelle für aktuelle Wetterdaten und stündliche Vorhersagen
für die initial unterstützten Locations (Potsdam, Berlin, Hamburg).

- API: open-meteo.com
- Lizenz: CC BY 4.0
- Attribution: „Daten von Open-Meteo.com, CC BY 4.0"
- Genutzte Variablen: temperature_2m, precipitation, wind_speed_10m,
  wind_direction_10m
```

### Schritt 8 — Tests und Smoke-Checks

Backend:

- Unit-Test für GetLocationBySlug und GetLatestObservation
- Integration-Test mit Test-Postgres (containerized)

Pyworkers:

- Test mit gemocktem httpx-Response (siehe httpx.MockTransport)
- Persistierung-Test gegen Test-Postgres
- Idempotenz: zweiter Fetch mit gleichem timestamp überschreibt nicht

E2E manuell:

- Worker einmal manuell ausführen: `make worker-open-meteo-current`
- API-Endpoint live: `curl https://api.research.worldweathernews.com/api/v1/locations`
- Frontend lokal: `pnpm dev` zeigt drei Wetter-Cards
- Lighthouse: keine Regression in Performance/Accessibility

### Schritt 9 — Doku

Neue oder erweiterte Files:

1. **`docs/data-sources.md`** neu anlegen oder erweitern:
   - Liste aller Datenquellen mit Lizenz
   - Open-Meteo-Spezifika (Endpoints, Rate-Limits, Attribution)
   - Worker-Frequenzen und -Architektur

2. **`docs/runbook.md`** erweitern:
   - Szenario „Open-Meteo-Worker liefert keine Daten"
   - Diagnose-Schritte (API-Down? DB voll? Schema-Drift?)

3. **`CLAUDE.md`** „Wo finde ich was"-Tabelle erweitern:
   - Pyworkers Open-Meteo-Modul
   - Backend-Locations-Handler
   - WeatherCard-Component

4. **`docs/backlog.md`** erweitern um:
   - Persistent-Job-State für Worker (W3)
   - Locations-Suche / Geocoding
   - Daily-Aggregate-Tabelle
   - Era5-Historie / Klima-Features

## Akzeptanzkriterien (komplette Liste)

- [ ] Branch `feat/iteration-2-1-open-meteo` angelegt
- [ ] Migration legt drei Tabellen + zwei Hypertables an, drei Städte
      seeded
- [ ] sqlc-Generierung grün
- [ ] Pyworkers Open-Meteo-Modul fetcht aus echtem Open-Meteo API
- [ ] Current-Worker speichert Observations in DB
- [ ] Hourly-Worker speichert Forecasts in DB
- [ ] APScheduler-Job läuft im Worker-Container
- [ ] Backend-Endpoints `/api/v1/locations` und
      `/api/v1/locations/{slug}` antworten korrekt
- [ ] OpenAPI-Schema erweitert, redocly-lint grün
- [ ] Frontend zeigt drei Wetter-Cards mit echten Daten
- [ ] Attribution-Footer-Snippet sichtbar mit Link zu
      /quellen-attribution
- [ ] Quellen-Attribution-Page enthält Open-Meteo-Block (DE + EN)
- [ ] Paraglide-Strings für alle UI-Labels (DE + EN)
- [ ] Mobile-Responsive, Lighthouse keine Regression
- [ ] `make lint && make test` grün
- [ ] `docs/data-sources.md` neu, Runbook + CLAUDE.md + Backlog
      erweitert
- [ ] PR-Erstellung erst nach finalem OK des Maintainers
- [ ] Tag v0.1.0 als erstes Track-2-Feature-Release

## Was du **noch nicht** baust

Diese Dinge sind explizit für spätere Iterationen:

- **DWD-Adapter** → Iteration 2.2 (mit erprobter Pipeline)
- **Wetterkarten** → 2.x, B.2-Diskussion noch offen (POSTPONED)
- **Locations-Suche / Geocoding** → eigene Iteration
- **Era5-Historie** → erste Klima-Iteration
- **MapLibre-Integration** → eigene Iteration nach B.2
- **Daily-Aggregate-Views** → kommt mit Klima-Features

Wenn du verlockt bist, diese Dinge einzubauen — widerstehen,
Iterations-Disziplin halten. „Hello World" heißt: **kleinste durchgängige
Pipeline**, nicht „erste schöne Feature".

## Wenn etwas unklar ist

Frag mich. Insbesondere:

- **Scheduling-Wahl (W1/W2/W3)**: ich entscheide nach kurzem Input
  von dir
- **Worker-Frequenzen** (10 Min current, 60 Min hourly)?
  Können wir diskutieren, falls Open-Meteo Rate-Limit-Probleme zeigt
- **OpenAPI-Schema-Änderungen**: erst Skizze zeigen, dann generieren
- **Frontend-Position** (Startseite vs. eigene Wetter-Route): mit
  mir abstimmen vor Implementation von Schritt 6

Lass uns loslegen. Bestätige mir kurz, dass du die Dokumente
gelesen hast, und schlag den ersten Schritt vor.
