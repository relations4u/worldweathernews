# Plan-Skizze — Iteration 2.2 — DWD-Adapter

**Stand: 12. Mai 2026 — Plan-Skizze, post-2.1-Update**

Dieses Dokument ist eine konsolidierte Plan-Skizze für Iteration 2.2,
basierend auf den Lessons aus Iteration 2.1. Final wird daraus
`prompt-iteration-2-2.md` ausgearbeitet — siehe dort für den
Submission-Ready-Übergabe-Prompt an Claude Code.

---

## Ziel

Erweitere die in 2.1 etablierte Pipeline um **echte DWD-Daten**. Das
ist die zweite Quelle, die mit dem schon erprobten Worker → DB → API →
Frontend-Pattern integriert wird, plus erste Auseinandersetzung mit
deutschen meteorologischen Daten-Formaten und mehr Locations.

**Tag: v0.5.0** (fortgeführt vom 2.1-Schema, nicht v0.2.0 — siehe
A.22-Note in `../feature1/feature-decisions.md`)
**Geschätzte Dauer: 4-6 Tage** (DWD-Format-Komplexität)

---

## Was wir wissen über DWD-OpenData

DWD-OpenData ist ein direkter HTTPS-Server unter
`https://opendata.dwd.de/` mit Verzeichnis-Listings. Keine REST-API
mit JSON — File-Server mit verschiedenen Formaten je Daten-Typ.

Relevante Verzeichnisse für 2.2:

```
weather/weather_reports/poi/
  → *BEOB.csv pro Stationspunkt, halbstündliche aktuelle Werte
  → einfachstes Format, CSV-Standard, gute Wahl für Hello World

weather/local_forecasts/mos/
  → MOSMIX-Vorhersagen als KMZ-Files pro Station
  → für 2.2b — Forecasts kommen als eigene Folge-Iteration

weather/weather_reports/synoptic_reports/germany/
  → SYNOP-BUFR-Meldungen, komplex
  → in 2.2 NICHT nutzen, eventuell später
```

**Termsofuse:** GeoNutzV — funktional CC-BY-äquivalent.
Attribution-Wording: „Datenbasis: Deutscher Wetterdienst,
eigene Bearbeitung" (siehe B.4).

---

## Sechs DWD-Stationen für 2.2 (Maintainer-bestätigt 12. Mai)

Drei Stadt-Stationen (passen zu den drei Open-Meteo-Locations
aus 2.1) plus drei Klimakontraste:

| Slug        | Station-ID | Höhe   | Typ           |
| ----------- | ---------- | ------ | ------------- |
| `potsdam`   | 03342      | 81 m   | Stadt         |
| `berlin`    | 10384      | 48 m   | Stadt         |
| `hamburg`   | 10147      | 11 m   | Stadt (Küste) |
| `brocken`   | 00078      | 1134 m | Bergstation   |
| `zugspitze` | 05792      | 2964 m | Hochgebirge   |
| `helgoland` | 02115      | 4 m    | Insel/Küste   |

Slug-Verwendung:

- Stadt-Slugs (Potsdam/Berlin/Hamburg) **gleichen** Slugs in 2.1 —
  Open-Meteo + DWD landen unter derselben Location, unterschiedliche
  Daten-Quellen
- Brocken/Zugspitze/Helgoland sind **neue** Slugs mit
  `source='dwd'`-only (keine Open-Meteo-Daten dafür in 2.2)

Schema-Konsequenz: `locations`-Tabelle bekommt zwei neue Spalten:

```sql
ALTER TABLE locations
  ADD COLUMN dwd_station_id TEXT,
  ADD COLUMN altitude_m INTEGER;
```

Plus drei neue Seed-Einträge für die Klimakontrast-Stationen.

---

## Architektur-Skizze (post-2.1, mit Lessons)

```
┌─────────────────────────────────────────────────────┐
│  apps/pyworkers/wwn_pyworkers/workers/dwd.py        │
│                                                     │
│  fetch_poi_observations(station_id):                │
│    GET https://opendata.dwd.de/weather/weather_     │
│        reports/poi/{station_id}-BEOB.csv            │
│    parse_csv() → list of measurements              │
│                                                     │
│  Scheduling: W1 wie in 2.1                          │
│  (APScheduler in-memory)                            │
│  - DWD POI updated alle 30 Min → Worker alle 30 Min │
└─────────────────────────────────────────────────────┘
                       ↓
┌─────────────────────────────────────────────────────┐
│  Postgres + TimescaleDB                             │
│                                                     │
│  locations (erweitert: dwd_station_id, altitude_m;  │
│             6 neue/erweiterte Seed-Einträge)        │
│  observations (gleicher Hypertable wie 2.1,         │
│                 + pressure, + humidity Spalten)     │
│  (forecasts bleibt unverändert, MOSMIX in 2.2b)     │
└─────────────────────────────────────────────────────┘
                       ↓
┌─────────────────────────────────────────────────────┐
│  Backend-API                                        │
│                                                     │
│  GET /api/v1/locations                              │
│    erweitert: Antwort enthält 6 Locations           │
│    (3 mit beiden Quellen, 3 nur DWD)                │
│                                                     │
│  GET /api/v1/locations/{slug}?source=dwd            │
│    optional `source` Query-Param, default = "dwd"   │
│    wenn verfügbar, sonst Open-Meteo                 │
└─────────────────────────────────────────────────────┘
                       ↓
┌─────────────────────────────────────────────────────┐
│  Frontend                                           │
│                                                     │
│  /wetter zeigt jetzt 6 Cards statt 3                │
│  Jede Card mit kleinem Source-Label "DWD" /         │
│  "Open-Meteo"                                       │
│  Für Stadt-Karten: Default DWD, Toggle für          │
│  Open-Meteo-Anzeige                                 │
└─────────────────────────────────────────────────────┘
```

---

## Klärungen aus 2.1 (jetzt entschieden, keine offenen Fragen mehr)

### Q1 — DWD-Library oder eigener Parser?

**Entscheidung: Eigener httpx + CSV-Parser**, analog zum
2.1-Open-Meteo-Pattern.

Begründung:

- 2.1-Pattern hat sich bewährt (`httpx.AsyncClient` + Async-Generators)
- DWD POI-CSV ist Standard-Format, kein Parsen-Drama
- Externe Library (`wetterdienst`) wäre eine zusätzliche Dependency
  ohne klaren Mehrwert für die POI-Pfade
- Bei späteren komplexeren DWD-Pfaden (MOSMIX-KMZ, RADOLAN) kann
  Library-Frage neu gestellt werden

### Q2 — Welche Stationen?

**Entscheidung: 6 Stationen wie oben tabelliert.**

### Q3 — Welche Variablen?

**Entscheidung: Die 4 aus 2.1 plus Druck und Feuchte.**

Schema-Erweiterung der `observations`-Tabelle:

```sql
ALTER TABLE observations
  ADD COLUMN pressure DOUBLE PRECISION,    -- hPa, MSL pressure
  ADD COLUMN humidity DOUBLE PRECISION;    -- % relative humidity
```

Plus: Open-Meteo-Worker aus 2.1 sollte auch diese zwei Felder
liefern (sie sind in der API verfügbar) — kommt als kleiner
Backfill-Commit im 2.2-PR oder als Folge-Commit.

### Q4 — DB-Schema-Migration?

**Entscheidung: Wide-Format beibehalten, ALTER TABLE.**

Nicht „long format" (`variable_name, value` pro Row) — würde
Schema komplett umkrempeln. Wide-Format ist gut so.

### Q5 — Wie umgehen mit Open-Meteo + DWD parallel?

**Entscheidung: Option AB2 — Default DWD wenn verfügbar,
Open-Meteo als Fallback.**

Eine Quelle pro Card, transparent ausgewählt:

- Für Stadt-Locations (Potsdam/Berlin/Hamburg): Default DWD
  weil offizielle deutsche Quelle
- Open-Meteo bleibt im DB, wird über `?source=open-meteo`
  Query-Param erreichbar
- Frontend zeigt Source-Label auf jeder Card
- Quellen-Toggle ist Backlog-Punkt für 2.x-Folge-Iteration
  („Quellen-Vergleich")

### Q6 — MOSMIX im selben PR?

**Entscheidung: 2.2b als Folge-Iteration, nicht in 2.2.**

POI-Observations als 2.2 (v0.5.0), MOSMIX-Forecasts als 2.2b (v0.5.1)
im selben Track-Block. Begründung:

- 2.2 hat schon 6 neue Stationen + Schema-Migration + neue Variablen
- MOSMIX bringt KMZ/KML-Parsen dazu — eigener Komplexitäts-Bereich
- Zwei kleinere PRs reviewen sich leichter als einer großer

---

## Lessons aus Iteration 2.1, die in 2.2 einfließen

Diese gehören in den 2.2-Übergabe-Prompt als „Beachten" oder
direkt in die Schritt-Anweisungen:

### Deploy-Lesson — DB-Migration als Pflicht-Step

Ansible-App-Rolle staged goose-Binary und führt Migration vor
`docker compose up` aus. **Funktioniert ab v0.4.2 ohne Sonder-
Workflow** — keine manuellen `docker cp`/`docker exec`-Schritte
nötig in 2.2.

Akzeptanzkriterium für 2.2: Deploy auf wwn-prod muss **ohne**
manuelle Migrations-Schritte funktionieren.

### Frontend-Lesson — /wetter ist ssr=false

`/wetter` läuft aktuell client-only weil `PUBLIC_API_BASE_URL`
browser-orientiert ist. **Bleibt so in 2.2** — SSR-Upgrade ist
eigener Backlog-Punkt.

Akzeptanzkriterium: `/wetter` zeigt 6 Cards nach Deploy, ohne
Server-Side-Rendering.

### OpenAPI-Lesson — keine nullable

Neue Felder (pressure, humidity, dwd_station_id, altitude_m)
**ohne `nullable`-Marker**. `required: false` reicht.

### sqlc-Lesson — Schema regenerieren

Nach jeder neuen Migration: `make sqlc-schema` (Pre-Processing)
plus `make sqlc-generate`. `make gen-check` validiert Drift in CI.

### Worker-Lesson — Idempotenz beachten

DWD POI-CSV liefert halbstündliche Werte. Worker läuft alle 30 Min.
Wenn Worker zwei Mal denselben Zeitpunkt abruft (Race oder Retry):
**INSERT ... ON CONFLICT (location_id, observed_at) DO UPDATE**
oder **DO NOTHING** — Idempotenz im DB-Layer, nicht im Worker.

---

## Skizze der Implementations-Schritte (final in prompt-iteration-2-2.md)

1. Branch + Verifikation (analog zu 2.1)
2. DB-Migration:
   - `locations` erweitert um `dwd_station_id`, `altitude_m`
   - `observations` erweitert um `pressure`, `humidity`
   - Drei neue Seed-Locations (Brocken, Zugspitze, Helgoland)
   - Drei bestehende Stadt-Locations bekommen `dwd_station_id`-Update
3. sqlc-Schema und Queries regenerieren
4. DWD-Worker-Modul in pyworkers (POI-CSV Fetcher + Parser)
5. Worker-Scheduling integrieren (APScheduler, alle 30 Min)
6. Open-Meteo-Worker um pressure/humidity erweitern (Backfill)
7. Backend-Endpoint-Erweiterung:
   - `/api/v1/locations` antwortet mit 6 Locations
   - `/api/v1/locations/{slug}` unterstützt `?source=` Query
8. Frontend-Integration:
   - WeatherCard zeigt Source-Label
   - Default DWD für Stadt-Locations
   - 3 zusätzliche Cards für Klimakontraste
9. Quellen-Attribution-Page erweitern (DWD-Block, GeoNutzV)
10. Tests + Smoke-Checks (DWD-CSV-Fixtures)
11. Doku (data-sources.md, runbook, CLAUDE.md, Backlog)

---

## Refs

- B.1 (Open-Meteo, fertig): `../feature1/feature-decisions.md`
- B.4 (Lizenzen, Attribution-Pattern): `../feature1/feature-decisions.md`
- B.5 (Worker-Scheduling W1): `../feature1/feature-decisions.md`
- B.6 (Frontend-Position): `../feature1/feature-decisions.md`
- A.20 (OpenAPI ohne nullable): `../feature1/feature-decisions.md`
- A.21 (sqlc-Schema-Input): `../feature1/feature-decisions.md`
- A.22 (DB-Migration als Deploy-Step): `../feature1/feature-decisions.md`
- Track-2-Status: `STATUS.md`
- DWD-OpenData: https://opendata.dwd.de/weather/weather_reports/poi/
- DWD-FAQ: https://www.dwd.de/DE/leistungen/opendata/faqs_opendata.html
