# Plan-Skizze — Iteration 2.2 — DWD-Adapter

**Stand: 12. Mai 2026 — Plan-Skizze**

Dieses Dokument ist eine **frühe Skizze** für Iteration 2.2, **nicht**
der Submission-Ready-Übergabe-Prompt für Claude Code. Final wird
`prompt-iteration-2-2.md` daraus ausgearbeitet, nachdem Iteration 2.1
gemerged ist und ihre Erkenntnisse einfließen können.

---

## Ziel

Erweitere die in 2.1 etablierte Pipeline um **echte DWD-Daten**. Das
ist die zweite Quelle, die mit dem schon erprobten Worker → DB → API →
Frontend-Pattern integriert wird, plus erste Auseinandersetzung mit
deutschen meteorologischen Daten-Formaten.

**Tag: v0.2.0**
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

weather/weather_reports/synoptic_reports/germany/
  → SYNOP-BUFR-Meldungen, komplex
  → in 2.2 NICHT nutzen, eventuell später

weather/local_forecasts/
  → MOSMIX-Vorhersagen als KML-Files pro Station
  → relevant für Forecast-Vergleich (DWD vs Open-Meteo)
```

**Termsofuse:** GeoNutzV — funktional CC-BY-äquivalent.
Attribution-Wording: „Datenbasis: Deutscher Wetterdienst,
eigene Bearbeitung" (siehe B.4).

---

## Architektur-Skizze

```
┌─────────────────────────────────────────────────────┐
│  apps/pyworkers/wwn_pyworkers/workers/dwd.py        │
│                                                     │
│  fetch_poi_observations(station_id):                │
│    GET https://opendata.dwd.de/weather/weather_     │
│        reports/poi/{station_id}-BEOB.csv            │
│    parse_csv() → list of measurements              │
│                                                     │
│  fetch_mosmix_forecast(station_id):                 │
│    GET https://opendata.dwd.de/weather/local_       │
│        forecasts/mos/MOSMIX_S/all_stations/         │
│        kml/MOSMIX_S_LATEST_240.kmz                  │
│    parse_kml() → forecast time series              │
└─────────────────────────────────────────────────────┘
                       ↓
┌─────────────────────────────────────────────────────┐
│  Postgres + TimescaleDB                             │
│                                                     │
│  locations (erweitert: source='dwd', dwd_station_id)│
│  observations (gleicher Hypertable wie 2.1,         │
│                 zusätzliche source='dwd' Zeilen)    │
│  forecasts (gleicher Hypertable wie 2.1,            │
│              zusätzliche source='dwd' Zeilen)       │
└─────────────────────────────────────────────────────┘
                       ↓
┌─────────────────────────────────────────────────────┐
│  Backend-API                                        │
│                                                     │
│  GET /api/v1/locations/{slug}/observations?         │
│      source=dwd                                     │
│  GET /api/v1/locations/{slug}/observations?         │
│      source=open-meteo                              │
│  (default = beide, Default-Quelle wird im           │
│   Response markiert)                                │
└─────────────────────────────────────────────────────┘
                       ↓
┌─────────────────────────────────────────────────────┐
│  Frontend                                           │
│                                                     │
│  WeatherCard zeigt Quelle als kleines Label         │
│  Optional: "Vergleich"-Modus mit beiden Quellen     │
└─────────────────────────────────────────────────────┘
```

---

## Offene Konzept-Fragen (vor Implementations-Start zu klären)

### Q1 — DWD-Library oder eigener Parser?

```
Option A — Eigener CSV/KML-Parser
  + Volle Kontrolle, keine Library-Abhängigkeit
  + Lerneffekt mit DWD-Format
  - Mehr Code, mehr Edge-Cases
  - KML-Parsing braucht xml/lxml-Stack

Option B — wetterdienst (PyPI)
  + Fertige Abstraktion für POI, MOSMIX, CDC
  + Aktiv gewartet (Nachfolger von dwdweather2)
  + Reichweite: stündlich, täglich, station-search
  - Externe Dependency, eigene Versions-Policy
  - Möglicherweise mehr als wir brauchen

Option C — Hybrid
  wetterdienst für Station-Discovery / Listing
  Eigener Parser für die Hot-Path Polling-Loop
  + Best of both: Library hilft bei der Erschließung,
    eigener Parser im Production-Hot-Path
  - Mehr Aufwand initial
```

Bauchgefühl: **Option B** als Default — wetterdienst macht den Einstieg
deutlich einfacher. Wenn wir später die Library-Abhängigkeit als Risiko
sehen, können wir den eigenen Parser nachziehen (Option C).

### Q2 — Welche Stationen?

```
DWD hat ~5000 Stationen, davon ~600 voll automatisiert.
Für Iteration 2.2 brauchen wir die DREI bestehenden
Locations aus 2.1 (Potsdam, Berlin, Hamburg) plus ggf.
ein paar mehr.

Konkrete Vorschläge je Location:
  Potsdam:  03987 (Potsdam-Stadt) oder 03342 (Potsdam-Telegrafenberg)
  Berlin:   10384 (Berlin-Tempelhof) oder 00433 (Berlin-Tegel)
  Hamburg:  10147 (Hamburg-Fuhlsbüttel)

Frage an Maintainer: welche Stationen pro Stadt?
Tempelhof ist klassisch, Tegel ist umstritten (geschlossen),
Fuhlsbüttel ist Hamburg-Standard.

Plus ggf. zusätzliche Stationen, weil DWD nicht-Stadt-zentrische
Locations hat (z.B. Bergstation, Küstenstation als Klimakontrast):
  Brocken (00078)
  Zugspitze (05792)
  Helgoland (02115)
```

Bauchgefühl: **drei DWD-Stationen für die drei Open-Meteo-Locations.**
Maintainer wählt die exakten Station-IDs. Bergstation/Küste als
optionale Erweiterung für 2.2b, nicht im initial Scope.

### Q3 — Welche Variablen?

DWD POI-CSV liefert deutlich mehr als Open-Meteo:

```
DWD POI-Daten (alle automatisch verfügbar):
  - Temperature_2m
  - Dew_point_2m
  - Relative_humidity
  - Mean_wind_speed_10m
  - Maximum_wind_gust_10m
  - Wind_direction_10m
  - Total_precipitation_hour
  - Surface_pressure
  - Mean_sea_level_pressure
  - Total_cloud_cover
  - Visibility
  - Sunshine_duration
  - Soil_temperature
  - Snow_depth
  ... weitere
```

Bauchgefühl: für 2.2 die **GLEICHEN 4 Variablen wie 2.1** plus
**Druck und Feuchte** zusätzlich (klassische Wetter-Anzeige). Spätere
Erweiterungen einfach ohne Schema-Migration.

### Q4 — DB-Schema-Migration?

Wenn ja, was muss erweitert werden?

```
locations:
  + dwd_station_id TEXT NULL  (nur für source='dwd')
  + dwd_station_name TEXT NULL
  + altitude_m INTEGER NULL  (DWD liefert das, Open-Meteo nicht)

observations:
  + pressure DOUBLE PRECISION NULL  (in 2.1 nicht da)
  + humidity DOUBLE PRECISION NULL  (in 2.1 nicht da)
  + ... weitere optionale Felder

  ODER ALTERNATIV:

  Generisches Schema:
  observations (location_id, observed_at, variable_name,
                value, unit, source, fetched_at)
  → pro Datenpunkt eine Zeile, „long format"
  → vereinfacht Schema-Migrationen, komplizierter Queries
```

Bauchgefühl: **bei „wide format" bleiben** (so wie in 2.1 angefangen),
neue Variablen als zusätzliche Spalten mit `NULL`-Default. „Long format"
wäre eleganter aber bedeutet komplette Schema-Umstellung — das machen
wir, wenn die Variablen-Liste explodiert (Track-2 später).

### Q5 — Wie umgehen mit Open-Meteo + DWD parallel?

```
Option AB1 — Beide Quellen pro Location, User wählt
  In WeatherCard: kleines Source-Label oben rechts,
  default DWD (für deutsche Stationen)
  Plus „Quelle wechseln"-Toggle

Option AB2 — Default DWD wenn verfügbar, Open-Meteo als Fallback
  Nur eine Quelle wird angezeigt, transparent ausgewählt
  Quelle als kleine Info im Footer der Card

Option AB3 — Beide Quellen nebeneinander zeigen
  „DWD: 18°C  |  Open-Meteo: 17.8°C"
  Zeigt Konsens und Abweichungen, didaktisch wertvoll
  Mehr Platz pro Card
```

Bauchgefühl: **Option AB2** für 2.2 — eine Quelle pro Card, default
DWD wenn verfügbar. Begründet im Quellen-Attribution-Block. Option AB3
wäre eine spätere Iteration „Quellen-Vergleich" für interessierte
Nutzer (kommt mit Forecast-Iterationen, wo Vergleich wertvoller wird).

### Q6 — MOSMIX im selben PR oder eigene Folge-Iteration?

```
Option M1 — MOSMIX zusammen mit POI-Observations in 2.2
  Klare Iteration: alle DWD-Funktionalität in einer Stelle
  Risiko: mehr Komplexität in einem PR

Option M2 — MOSMIX als eigene Iteration 2.2b oder 2.3
  Kleinere PRs
  POI-Observations stabil testen, dann Forecast nachziehen
  Vortrags-Tag-Roadmap würde v0.2.0 (POI) und v0.2.1 (MOSMIX)
  bedeuten
```

Bauchgefühl: **Option M2** — POI-Observations in 2.2 (v0.2.0),
MOSMIX-Forecasts als 2.2b (v0.2.1) im selben Track-Block, aber als
eigenes PR. Stations-Map (heutige 2.3) wird dann 2.3 mit Stand
v0.3.0. Mehr Releases, jeder kleiner, in 4-6 Tagen Gesamt-Aufwand.

---

## Was an Recherche noch fehlt

Bevor `prompt-iteration-2-2.md` ausgearbeitet wird:

- [ ] Konkrete Station-IDs pro Stadt mit Maintainer abstimmen
      (Tempelhof? Tegel? Babelsberg? Fuhlsbüttel?)
- [ ] `wetterdienst` Library prüfen: aktive Wartung? Python-Versions-
      Kompat? Lizenz? Größe?
- [ ] DWD-POI-CSV-Format genau ansehen (Encoding, Trennzeichen, NULL-
      Markierung, Zeitzonen)
- [ ] DWD-Update-Frequenz für POI klären (alle 30 Min? 60 Min?)
- [ ] Robotex-Policy von opendata.dwd.de prüfen (User-Agent setzen?)
- [ ] HTTP-If-Modified-Since-Caching nutzbar? (für Worker-Effizienz)

---

## Skizze der 9 Implementations-Schritte (final in prompt-iteration-2-2.md)

1. Branch + Verifikation
2. DB-Migration: locations-Spalten erweitern, ggf. neue Variablen-
   Spalten in observations
3. DWD-Worker-Modul in pyworkers (POI-CSV fetcher + parser)
4. Worker-Scheduling integrieren (passend zu 2.1-Pattern)
5. Backend-Endpoint-Erweiterung: source-Parameter
6. Frontend-Integration: Source-Label in WeatherCard
7. Quellen-Attribution-Page erweitern (DWD-Block, GeoNutzV)
8. Tests + Smoke-Checks (echte DWD-CSV-Fixtures)
9. Doku (data-sources.md, runbook, CLAUDE.md, Backlog)

---

## Refs

- B.1 (Open-Meteo Hello World, fertig): `../feature1/feature-decisions.md`
- B.4 (Lizenzen, Attribution-Pattern): `../feature1/feature-decisions.md`
- Track-2-Status: `STATUS.md`
- DWD-OpenData: https://opendata.dwd.de/weather/weather_reports/poi/
- DWD-FAQ: https://www.dwd.de/DE/leistungen/opendata/faqs_opendata.html
- wetterdienst (PyPI): https://pypi.org/project/wetterdienst/
- wetterdienst Successor von dwdweather2: aktiv gewartet 2026
