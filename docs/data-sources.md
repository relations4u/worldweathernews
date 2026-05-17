# Datenquellen

Liste aller externen Datenquellen, die worldweathernews.com einbindet —
mit Lizenz, Attribution-Pflicht, technischer Anbindung und Status.

Pflege diese Datei mit, wenn eine Quelle hinzukommt, ihren Status
wechselt oder ihre Attribution sich ändert. Quellen-Status bestimmt,
ob die Quelle auf `/quellen-attribution` als „aktiv" oder „geplant"
gelistet wird.

Stand: Mai 2026 (Iteration 2.2 live).

---

## Aktive Quellen

### Open-Meteo

| Feld               | Wert                                                                                                              |
| ------------------ | ----------------------------------------------------------------------------------------------------------------- |
| API                | `https://api.open-meteo.com/v1/forecast`                                                                          |
| Doku               | https://open-meteo.com/en/docs                                                                                    |
| Lizenz             | CC BY 4.0 (https://creativecommons.org/licenses/by/4.0/)                                                          |
| Attribution        | „Daten von Open-Meteo.com, CC BY 4.0"                                                                             |
| Auth               | keine (öffentliche API, kein Key)                                                                                 |
| Rate-Limit (Free)  | 10 000 Calls / Tag (großzügig)                                                                                    |
| Eingebunden seit   | Iteration 2.1 (Mai 2026)                                                                                          |
| Genutzte Variablen | `temperature_2m`, `precipitation`, `wind_speed_10m`, `wind_direction_10m`, `pressure_msl`, `relative_humidity_2m` |
| Polling-Frequenzen | `current` alle 10 min, `hourly` (24-h-Vorhersage, T/Niederschlag/Wind ohne Druck/Feuchte) alle 60 min             |
| Idempotenz         | `INSERT … ON CONFLICT (location_id, source, observed_at) DO NOTHING` — kein Update, OM-Werte sind nicht-revidiert |

**Druck-Variable:** bewusst `pressure_msl` (auf Meereshöhe reduziert)
gewählt statt `surface_pressure`, damit die Werte mit DWDs MSL-
reduziertem Druck konsistent sind. `surface_pressure` würde für
Berlin/Potsdam um ~5 hPa abweichen (Höhe-Reduktions-Differenz).

### Deutscher Wetterdienst (DWD) — POI-Observations

| Feld               | Wert                                                                                                                                        |
| ------------------ | ------------------------------------------------------------------------------------------------------------------------------------------- |
| Endpoint           | `https://opendata.dwd.de/weather/weather_reports/poi/{station_id}-BEOB.csv`                                                                 |
| Format             | CSV, semikolon-separiert, latin-1 encoded; 3 Header-Reihen + Datenzeilen absteigend; deutsche Komma-Decimals; `---` als NaN                 |
| Lizenz             | GeoNutzV (funktional CC-BY-äquivalent, https://www.dwd.de/DE/service/copyright/copyright_node.html)                                         |
| Attribution        | „Datenbasis: Deutscher Wetterdienst, eigene Bearbeitung"                                                                                    |
| Auth               | keine (öffentlicher File-Server)                                                                                                            |
| Rate-Limit         | nicht offiziell dokumentiert; halbstündliches Polling pro Station ist innerhalb der DWD-„fair use"-Erwartung                                |
| Eingebunden seit   | Iteration 2.2 (Mai 2026)                                                                                                                    |
| Genutzte Variablen | Lufttemperatur (2 m), Niederschlag (letzte Stunde), Windgeschwindigkeit + Windrichtung (10-min-Mittel, 10 m), Druck (MSL), rel. Luftfeuchte |
| Polling-Frequenz   | `current` alle 30 min (POI-Files werden ~30 min nach Beobachtung aktualisiert)                                                              |
| Idempotenz         | `INSERT … ON CONFLICT (location_id, source, observed_at) DO UPDATE` — Quality-Korrekturen aus DWD-Reruns werden übernommen                  |

**Variablen-Mapping**: Spaltennamen aus DWD-Header-Reihe 1
(canonical, englisch). Achtung — zwei Spalten haben echte
embedded-space-Quirks im Namen, die exakt so verwendet werden
müssen: `mean_wind_speed_during last_10_min_at_10_meters_above_ground`
und `mean_wind_direction_during_last_10 min_at_10_meters_above_ground`.

**Druck bei Hochgebirgsstationen:** DWD reduziert den MSL-Druck
nicht für Stationen oberhalb von ~2000 m (Zugspitze, 2964 m). Der
Parser liefert dann `pressure=None`, das Frontend lässt die
Druck-Zeile in der WeatherCard weg.

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

**Idempotenz** (übergreifend, seit Iteration 2.2 mit zwei Quellen):
der `observations`-Primary-Key wurde in Migration 0002 auf
`(location_id, source, observed_at)` erweitert, damit beide Quellen
parallel für dieselbe Location und Beobachtungszeit speichern können.
Forecasts bleiben bei `(location_id, forecast_for, run_at)` — bisher
nur Open-Meteo, DWD-MOSMIX kommt mit 2.2b.

**Locations (Iteration 2.2, 6 Stationen):**

| Slug        | Name                   | Lat     | Lon     | Höhe   | DWD-ID | Quellen          |
| ----------- | ---------------------- | ------- | ------- | ------ | ------ | ---------------- |
| `berlin`    | Berlin (Tempelhof)     | 52.5200 | 13.4050 | 48 m   | 10384  | DWD + Open-Meteo |
| `brocken`   | Brocken                | 51.7991 | 10.6178 | 1134 m | 10454  | DWD              |
| `hamburg`   | Hamburg (Fuhlsbüttel)  | 53.5511 | 9.9937  | 11 m   | 10147  | DWD + Open-Meteo |
| `helgoland` | Helgoland              | 54.1827 | 7.8868  | 4 m    | 10015  | DWD              |
| `potsdam`   | Potsdam (Telegrafenb.) | 52.3906 | 13.0645 | 81 m   | 10379  | DWD + Open-Meteo |
| `zugspitze` | Zugspitze              | 47.4209 | 10.9854 | 2964 m | 10961  | DWD              |

Die drei Stadt-Slugs sind in Migration 0001 (2.1) als OM-Locations
angelegt und in Migration 0002 (2.2) um `dwd_station_id` + `altitude_m`
erweitert. Brocken / Helgoland / Zugspitze sind in 0002 als
DWD-only-Locations dazugekommen. `locations.source` bleibt Legacy-Feld
aus 2.1 — die API leitet `availableSources` aus den tatsächlich
gespeicherten observations-Rows ab.

**Default-Quelle pro Location:** Backend wählt für
`GET /api/v1/locations/{slug}` ohne `?source=`-Parameter die
DWD-Variante, wenn die Location eine `dwd_station_id` hat — sonst
Open-Meteo. Frontend zeigt entsprechend für alle 6 Stadt-Slugs
DWD-Daten und nutzt den Open-Meteo-Pfad nur, wenn der Caller explizit
`?source=open-meteo` setzt. Quellen-Toggle-UI ist Backlog-Punkt.

**WMO-Synop-IDs**, **nicht** DWD-Legacy/CDC-IDs: der POI-Endpoint
adressiert ausschließlich über 5-stellige WMO-Synop-Kennungen
(z. B. Potsdam-Telegrafenberg ist `10379`, nicht das CDC-übliche
`03342`). Vor der Aufnahme einer neuen Station gegen
https://opendata.dwd.de/weather/weather_reports/poi/ prüfen, dass
`{station_id}-BEOB.csv` 200 liefert.

### EUMETSAT — Meteosat-Satellitenbilder (EUMETView WMS)

| Feld             | Wert                                                                                                          |
| ---------------- | ------------------------------------------------------------------------------------------------------------- |
| Dienst           | `https://view.eumetsat.int/geoserver/wms` (EUMETView WMS 1.3.0)                                               |
| Produkt / Layer  | Meteosat 0° High Rate SEVIRI IR 10.8 µm — Layer `msg_fes:ir108`                                               |
| Form             | Fertig gerenderte PNG-Composites, GetMap `CRS=EPSG:3857`, Europa-BBOX, 1024×1024                              |
| Lizenz           | EUMETSAT-Datenpolitik — Meteosat-Bildprodukte kostenfrei + lizenzfrei                                         |
| Attribution      | „© EUMETSAT"                                                                                                  |
| Auth             | **keine** (öffentliches WMS, Q4 live verifiziert — der `eumetsat.env`-Secret wird für Pfad A NICHT gebraucht) |
| Eingebunden seit | Iteration 2.4 (Mai 2026)                                                                                      |
| Polling-Frequenz | `WWN_PY_EUMETSAT_INTERVAL_SECONDS` (Default 900 = 15 min), rollierendes 24-h-Fenster                          |
| Idempotenz       | Frame-Key = auf das Intervall gerundeter UTC-Slot → Re-Run im selben Slot überschreibt; keine Doppel-Frames   |

**Pfad A (Konzept-Session 2.4, B.2=K3):** Der `pyworkers`-EUMETSAT-Worker
holt die Composites **server-seitig**, legt sie in den A.13-Hetzner-
Object-Storage-Bucket (Prefix `sat/ir108`) und schreibt `sat/index.json`.
Das Frontend (`/satellit`) lädt ausschließlich über das eigene
`media.worldweathernews.com` — **kein** Drittanbieter-Client-Pfad
(A.19-konform, daher **kein** Datenschutz-§5-Eintrag, anders als
OpenFreeMap). Roh-SEVIRI + Satpy (eigene Composites) ist bewusst NICHT
Teil von 2.4 — das ist der K1-Evolutionspfad (~Iteration 2.6, siehe
`docs/backlog.md`).

---

## Geplante Quellen

Die folgenden Quellen sind für zukünftige Iterationen vorgesehen.
Endgültige Lizenz- und Attribuierungs-Hinweise folgen bei der
jeweiligen produktiven Anbindung.

| Quelle                              | Iteration | Lizenz / Status                                         |
| ----------------------------------- | --------- | ------------------------------------------------------- |
| DWD MOSMIX / ICON-Forecasts         | 2.2b      | GeoNutzV — Forecast-Pfad ergänzend zu POI               |
| NOAA / National Weather Service     | später    | Public Domain (US-Gov)                                  |
| Met Office (UK), JMA (JP), Météo-FR | später    | jeweilige Open-Data-Lizenzen                            |
| EUMETSAT                            | 2.4       | Lizenzbedingungen je nach Produkt                       |
| USGS                                | später    | Public Domain (Erdbeben)                                |
| NOAA Space Weather                  | später    | Public Domain (Aurora)                                  |
| DWD-CDC-Archiv                      | Klima     | GeoNutzV — historische Reihen für Klima-Anomalien (B.3) |

Die Roadmap und Plan-Skizzen liegen unter `sessions/feature2/`.

---

## Kartenbasis (Geodaten)

Keine Wetter-Datenquelle, der Vollständigkeit halber dokumentiert:
Die interaktive Stationskarte auf `/wetter` (Iteration 2.3) nutzt
**OpenFreeMap** (Liberty-Style, `tiles.openfreemap.org`) für die
Vektor-Kartenkacheln.

| Aspekt       | Wert                                                      |
| ------------ | --------------------------------------------------------- |
| Quelle       | OpenFreeMap Public Instance (Vector-Tiles)                |
| Lizenz       | OpenFreeMap MIT; Kartendaten OpenStreetMap ODbL           |
| Attribution  | „OpenFreeMap © OpenMapTiles Data from OpenStreetMap"      |
| Account/Key  | keiner — kostenfrei, kein API-Key                         |
| Cookies      | keine, kein Tracking (client-seitiger Tile-Abruf, nur IP) |
| Wechselpunkt | `apps/frontend/src/lib/config/map.ts` (eine zentrale URL) |
| Eingebunden  | Iteration 2.3 (Mai 2026)                                  |

Entscheidung und Self-hosting-Einordnung (A.19): siehe
`sessions/feature2/prompt-iteration-2-3.md` und `docs/backlog.md`
(„Self-hosted OpenFreeMap-Stack").

---

## Worker-Konfiguration

Frequenzen sind ENV-konfigurierbar (`WWN_PY_OPEN_METEO_*`,
siehe `.env.example`). Defaults in `apps/pyworkers/pyworkers/config.py`:

| Variable                                     | Default | Bedeutung                                                           |
| -------------------------------------------- | ------- | ------------------------------------------------------------------- |
| `WWN_PY_OPEN_METEO_ENABLED`                  | `true`  | Open-Meteo-Worker aktivieren / deaktivieren                         |
| `WWN_PY_OPEN_METEO_CURRENT_INTERVAL_SECONDS` | `600`   | Polling-Intervall für `current`                                     |
| `WWN_PY_OPEN_METEO_HOURLY_INTERVAL_SECONDS`  | `3600`  | Polling-Intervall für `hourly` (24-h)                               |
| `WWN_PY_DWD_ENABLED`                         | `true`  | DWD-POI-Worker aktivieren / deaktivieren                            |
| `WWN_PY_DWD_POI_INTERVAL_SECONDS`            | `1800`  | Polling-Intervall für DWD-POI (entspricht der DWD-Veröffentlichung) |
| `WWN_PY_EUMETSAT_ENABLED`                    | `true`  | EUMETSAT-Satelliten-Worker aktivieren / deaktivieren                |
| `WWN_PY_EUMETSAT_INTERVAL_SECONDS`           | `900`   | Pull-Intervall (EUMETView ~15 min)                                  |
| `WWN_PY_EUMETSAT_WINDOW_HOURS`               | `24`    | Rollierendes Frame-Fenster                                          |

S3-Ziel (A.13-Bucket): die drei Pflicht-Werte werden als
`WWN_PY_S3_ENDPOINT` / `WWN_PY_S3_ACCESS_KEY` / `WWN_PY_S3_SECRET_KEY`
in die `infra/secrets/production/pyworkers.env` (SOPS) eingetragen —
ein 1:1-Prefix-Mapping der `S3_*`-Namen aus dem media-storage-Secret,
kein Ansible-/Compose-Code-Change nötig (der pyworkers-Container
`env_file`d die Datei bereits). `s3_bucket`/`s3_region`/`sat_prefix`/
`media_base_url` haben Defaults. Fehlt die S3-Config, **skippt** der
Job sauber (`status="skipped"`), statt den Worker zu crashen.

Der DWD-Job läuft beim Container-Start sofort einmal (`next_run_time
= now` in `__main__.py`), damit nach einem Deploy nicht bis zu
30 Min gewartet wird, bis die ersten Daten in der DB liegen. OM-Jobs
warten dagegen das volle Intervall ab (sind seit 24 h+ kontinuierlich
gelaufen, brauchen kein Initial-Catch-up).

**Metriken** (Prometheus, Port 9100 des Worker-Containers):

- `wwn_open_meteo_fetches_total{kind={current|hourly},status={ok|error}}`
- `wwn_dwd_fetches_total{status={ok|error|empty}}`
- `wwn_eumetsat_fetches_total{status={ok|error|skipped}}`
- `wwn_job_runs_total{job,status}` und `wwn_job_duration_seconds`
  (Wrapper `measure_job`, deckt `heartbeat`, `open_meteo_current`,
  `open_meteo_hourly`, `dwd_poi`, `eumetsat` ab).

**Tracing:** Jeder Job startet einen eigenen Span
(`open_meteo.run_current`, `open_meteo.run_hourly`, `dwd.run_poi`).
HTTP-Calls (httpx) und DB-Inserts (asyncpg) werden durch die
OpenTelemetry-Auto-Instrumentation als Kind-Spans erfasst.

---

## Attribution auf der Plattform

Quellen-Attribution wird an zwei Stellen ausgespielt:

1. **Footer-Snippet auf jedem Wetter-relevanten Endpoint**:
   Backend liefert das `attribution`-Feld dynamisch je nach gewählter
   Quelle — für `GET /api/v1/locations/{slug}` ist es die Attribution
   der tatsächlich ausgelieferten Quelle (DWD oder Open-Meteo). Für
   `GET /api/v1/locations` (Liste) ist es ein zusammengesetzter Text
   (`„Datenbasis: Deutscher Wetterdienst … · Daten von Open-Meteo.com …"`).
   Frontend zeigt zusätzlich pro Card ein Source-Badge mit dem
   Kurz-Label der jeweiligen Quelle.

2. **Detail-Seite `/quellen-attribution`**: pro Quelle eine eigene
   Karte mit API-Link, Lizenz-Link, Attribution-Pflichtsatz,
   Variablen-Liste, Polling-Frequenzen und Einbindungs-Iteration
   (siehe `apps/frontend/src/routes/quellen-attribution/+page.svelte`).

---

## Code-Stellen

| Anliegen                 | Datei                                                                                                                 |
| ------------------------ | --------------------------------------------------------------------------------------------------------------------- |
| Open-Meteo-Worker        | `apps/pyworkers/pyworkers/jobs/open_meteo.py`                                                                         |
| DWD-POI-Worker           | `apps/pyworkers/pyworkers/jobs/dwd.py`                                                                                |
| EUMETSAT-Worker          | `apps/pyworkers/pyworkers/jobs/eumetsat.py`                                                                           |
| Satelliten-Route         | `apps/frontend/src/routes/satellit/` + `src/lib/components/SatelliteMap.svelte`                                       |
| Satelliten-Config/Helper | `apps/frontend/src/lib/config/satellite.ts` (Index-URL/Typ), `src/lib/satellite.ts` (pure Helpers)                    |
| Worker-Tests             | `apps/pyworkers/tests/test_open_meteo.py`, `test_dwd.py`, `test_eumetsat.py`                                          |
| Worker-Scheduling        | `apps/pyworkers/pyworkers/__main__.py` (`scheduler.add_job(...)`)                                                     |
| DB-Migrationen           | `infra/migrations/0001_open_meteo_hello_world.sql`, `infra/migrations/0002_dwd_poi_stations_and_variables.sql`        |
| sqlc-Queries             | `apps/backend/internal/storage/queries/{locations,observations,forecasts}.sql` (inkl. `GetLatestObservationBySource`) |
| Backend-Handler          | `apps/backend/internal/http/handler/api.go` (`ListLocations`, `GetLocationDetail`, `resolveSource`)                   |
| Backend-Tests            | `apps/backend/internal/http/handler/{handler_test,api_internal_test}.go`                                              |
| OpenAPI-Schema           | `packages/api-schema/openapi.yaml`                                                                                    |
| Frontend-Route           | `apps/frontend/src/routes/wetter/`                                                                                    |
| WeatherCard-Component    | `apps/frontend/src/lib/components/WeatherCard.svelte`                                                                 |
| Attribution-Page         | `apps/frontend/src/routes/quellen-attribution/+page.svelte`                                                           |
