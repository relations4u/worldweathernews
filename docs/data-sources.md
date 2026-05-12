# Datenquellen

Liste aller externen Datenquellen, die worldweathernews.com einbindet —
mit Lizenz, Attribution-Pflicht, technischer Anbindung und Status.

Pflege diese Datei mit, wenn eine Quelle hinzukommt, ihren Status
wechselt oder ihre Attribution sich ändert. Quellen-Status bestimmt,
ob die Quelle auf `/quellen-attribution` als „aktiv" oder „geplant"
gelistet wird.

Stand: Mai 2026 (Iteration 2.1 live).

---

## Aktive Quellen

### Open-Meteo

| Feld               | Wert                                                                      |
| ------------------ | ------------------------------------------------------------------------- |
| API                | `https://api.open-meteo.com/v1/forecast`                                  |
| Doku               | https://open-meteo.com/en/docs                                            |
| Lizenz             | CC BY 4.0 (https://creativecommons.org/licenses/by/4.0/)                  |
| Attribution        | „Daten von Open-Meteo.com, CC BY 4.0"                                     |
| Auth               | keine (öffentliche API, kein Key)                                         |
| Rate-Limit (Free)  | 10 000 Calls / Tag (großzügig)                                            |
| Eingebunden seit   | Iteration 2.1 (Mai 2026)                                                  |
| Genutzte Variablen | `temperature_2m`, `precipitation`, `wind_speed_10m`, `wind_direction_10m` |
| Polling-Frequenzen | `current` alle 10 min, `hourly` (24-h-Vorhersage) alle 60 min             |

**Architektur:**

```
Open-Meteo API
     │
     │  apps/pyworkers/pyworkers/jobs/open_meteo.py
     ▼
APScheduler-Jobs (in-process, Container wwn-pyworkers)
     │  current  →  observations (TimescaleDB-Hypertable auf observed_at)
     │  hourly   →  forecasts    (TimescaleDB-Hypertable auf forecast_for)
     ▼
Postgres (TimescaleDB-HA-Image, im Compose-Stack)
     │
     │  sqlc-Queries unter apps/backend/internal/storage/queries/
     ▼
Backend-API
     │  GET /api/v1/locations           → Liste aller aktiven Locations
     │  GET /api/v1/locations/{slug}    → Location + current + 24h-Forecast
     ▼
Frontend
     │  /wetter zeigt WeatherCard pro Stadt
```

**Idempotenz:** Beide INSERTs nutzen `ON CONFLICT DO NOTHING` auf dem
Primary Key (`(location_id, observed_at)` bzw. `(location_id,
forecast_for, run_at)`), so dass wiederholte Worker-Läufe — etwa weil
Open-Meteo current-Werte in 15-min-Rastern liefert, der Worker aber
alle 10 min pollt — keine Duplikate erzeugen.

**Locations (Iteration 2.1):**

| Slug      | Name    | Latitude | Longitude | Timezone      |
| --------- | ------- | -------- | --------- | ------------- |
| `potsdam` | Potsdam | 52.3906  | 13.0645   | Europe/Berlin |
| `berlin`  | Berlin  | 52.5200  | 13.4050   | Europe/Berlin |
| `hamburg` | Hamburg | 53.5511  | 9.9937    | Europe/Berlin |

Locations sind in der `locations`-Tabelle gespeichert (per
goose-Migration `0001_open_meteo_hello_world.sql` geseedet, `active =
TRUE`, `source = 'open-meteo'`). Weitere Locations können später
hinzukommen — entweder per Folge-Migration oder per eigenständiges
Seeding-Skript (CLAUDE.md → spätere Iterationen).

---

## Geplante Quellen

Die folgenden Quellen sind für zukünftige Iterationen vorgesehen.
Endgültige Lizenz- und Attribuierungs-Hinweise folgen bei der
jeweiligen produktiven Anbindung.

| Quelle                              | Iteration | Lizenz / Status                         |
| ----------------------------------- | --------- | --------------------------------------- |
| Deutscher Wetterdienst (DWD)        | 2.2       | CC-BY-4.0 (GeoNutzV), Stations + MOSMIX |
| NOAA / National Weather Service     | später    | Public Domain (US-Gov)                  |
| Met Office (UK), JMA (JP), Météo-FR | später    | jeweilige Open-Data-Lizenzen            |
| EUMETSAT                            | 2.4       | Lizenzbedingungen je nach Produkt       |
| USGS                                | später    | Public Domain (Erdbeben)                |
| NOAA Space Weather                  | später    | Public Domain (Aurora)                  |

Die Roadmap und Plan-Skizzen liegen unter `sessions/feature2/`.

---

## Worker-Konfiguration

Frequenzen sind ENV-konfigurierbar (`WWN_PY_OPEN_METEO_*`,
siehe `.env.example`). Defaults in `apps/pyworkers/pyworkers/config.py`:

| Variable                                     | Default | Bedeutung                             |
| -------------------------------------------- | ------- | ------------------------------------- |
| `WWN_PY_OPEN_METEO_ENABLED`                  | `true`  | Worker aktivieren / deaktivieren      |
| `WWN_PY_OPEN_METEO_CURRENT_INTERVAL_SECONDS` | `600`   | Polling-Intervall für `current`       |
| `WWN_PY_OPEN_METEO_HOURLY_INTERVAL_SECONDS`  | `3600`  | Polling-Intervall für `hourly` (24-h) |

**Metriken** (Prometheus, Counter `wwn_open_meteo_fetches_total`,
gelabelt nach `kind={current|hourly}` und `status={ok|error}`) werden
auf Port 9100 des Worker-Containers exponiert. Auch
`wwn_job_runs_total` und `wwn_job_duration_seconds` (vom
`measure_job`-Wrapper) gelten für die zwei Open-Meteo-Jobs
(`open_meteo_current`, `open_meteo_hourly`).

**Tracing:** Jeder Job startet einen eigenen Span
(`open_meteo.run_current` / `open_meteo.run_hourly`). HTTP-Calls
(httpx) und DB-Inserts (asyncpg) werden durch die OpenTelemetry-
Auto-Instrumentation als Kind-Spans erfasst.

---

## Attribution auf der Plattform

Quellen-Attribution wird an zwei Stellen ausgespielt:

1. **Footer-Snippet auf jedem Wetter-relevanten Endpoint**:
   Backend liefert `attribution: "Daten von Open-Meteo.com, CC BY
4.0"` in jedem `/api/v1/locations`- und
   `/api/v1/locations/{slug}`-Response. Frontend zeigt den String
   plus Link zu `/quellen-attribution` (siehe
   `apps/frontend/src/routes/wetter/+page.svelte`).

2. **Detail-Seite `/quellen-attribution`**: pro Quelle eine eigene
   Karte mit API-Link, Lizenz-Link, Attribution-Pflichtsatz,
   Variablen-Liste, Polling-Frequenzen und Einbindungs-Iteration
   (siehe `apps/frontend/src/routes/quellen-attribution/+page.svelte`).

---

## Code-Stellen

| Anliegen              | Datei                                                                              |
| --------------------- | ---------------------------------------------------------------------------------- |
| Open-Meteo-Worker     | `apps/pyworkers/pyworkers/jobs/open_meteo.py`                                      |
| Worker-Tests          | `apps/pyworkers/tests/test_open_meteo.py`                                          |
| Worker-Scheduling     | `apps/pyworkers/pyworkers/__main__.py` (`scheduler.add_job(...)`)                  |
| DB-Migration          | `infra/migrations/0001_open_meteo_hello_world.sql`                                 |
| sqlc-Queries          | `apps/backend/internal/storage/queries/{locations,observations,forecasts}.sql`     |
| Backend-Handler       | `apps/backend/internal/http/handler/api.go` (`ListLocations`, `GetLocationDetail`) |
| OpenAPI-Schema        | `packages/api-schema/openapi.yaml`                                                 |
| Frontend-Route        | `apps/frontend/src/routes/wetter/`                                                 |
| WeatherCard-Component | `apps/frontend/src/lib/components/WeatherCard.svelte`                              |
| Attribution-Page      | `apps/frontend/src/routes/quellen-attribution/+page.svelte`                        |
