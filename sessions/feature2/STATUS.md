# Track 2 — Wetterdaten-Import — Status

Pflege diese Datei am Ende jeder Iteration. Format analog zu
`sessions/feature1/STATUS.md`.

Status-Legende: ✅ Done · 🟡 In Progress · ⏳ Geplant · ❌ Blocked · ⏭ Skipped

Stand: 2026-05-12 (Iteration 2.2 lokal fertig, PR offen)

---

## Konzept-Phase

Status: ✅ Done (für Iteration 2.1)
Datum: 2026-05-11

Vier B-Punkte aus `sessions/feature1/feature-decisions.md` Abschnitt B:

- **B.1 (Erste Datenquelle)** ✅ DECIDED 2026-05-11
  Open-Meteo zuerst. 3 Städte (Potsdam, Berlin, Hamburg),
  4 Variablen (Temperatur, Niederschlag, Windgeschwindigkeit,
  Windrichtung), 2 Frequenzen (current + hourly 24h). Keine Historie
  in 2.1. Storage in Postgres + TimescaleDB-Hypertables.
- **B.2 (Wetterkarten)** ⏳ POSTPONED 2026-05-11
  Eigene Konzept-Diskussion nach Iteration 2.1. Drei Optionen
  (K1 selbst rendern / K2 extern / K3 hybrid) dokumentiert.
- **B.3 (Storage für Datasets)** ⏳ POSTPONED 2026-05-11
  Für 2.1 nicht relevant. Wiederaufnahme bei Iterationen 2.4/2.5
  (Satellitenbilder, Radar).
- **B.4 (Daten-Lizenzen)** ✅ DECIDED 2026-05-11
  Attribution-Pattern: Footer-Link auf jeder Page + Detail-Page
  `/quellen-attribution`. Strings in Paraglide-Messages.

---

## Geplante Iterationen

### Iteration 2.1 — Open-Meteo Hello World

Status: ✅ Done — gemerged (PR #69), v0.4.0 → v0.4.2 live auf wwn-prod
Datum: 2026-05-12
Tags: **v0.4.0** (Iteration-Release) → **v0.4.1** (Folge-PR, partial deploy) → **v0.4.2** (Hotfix, vollständig live).

**Commits auf dem Branch (9 plus Branch-Setup):**

1. `82fab20` — feat(db): erste Migration für Open-Meteo Locations,
   Observations, Forecasts (Schritt 2)
2. `6925b12` — feat(backend): sqlc-Pipeline für locations/
   observations/forecasts (Schritt 3)
3. `0abb927` — feat(pyworkers): Open-Meteo-Worker, current 10 min
   und hourly 60 min (Schritt 4)
4. `3822113` — feat(backend): /api/v1/locations list + /{slug}
   detail mit current + forecast (Schritt 5)
5. `b1cf7db` — feat(frontend): /wetter route mit WeatherCard für
   drei Open-Meteo-Städte (Schritt 6)
6. `7f1d65a` — docs(frontend): Open-Meteo-Block auf
   /quellen-attribution (Schritt 7)
7. `d17c8e3` — test(pyworkers): tests für open_meteo parser +
   HTTP-Layer (Schritt 8)
8. (folgt) — docs(2.1): data-sources + runbook + CLAUDE +
   backlog + STATUS-Updates (Schritt 9)

**Getroffene Implementations-Entscheidungen:**

- **Scheduling**: W1 (APScheduler im Worker-Container, in-Memory-
  State). W3-Migration (PostgresJobStore) im Backlog.
- **Frontend-Position**: eigene Route `/wetter` (CSR-only via
  `ssr=false`; SSR-Upgrade über separaten Internal-API-Hostname im
  Backlog).
- **Search-Endpoint-Konflikt**: bisheriger Search-Stub auf
  `/api/v1/locations` durch das neue List-All ersetzt. Search wird
  eigene Iteration (siehe `docs/backlog.md` → Locations-Suche).
- **OpenAPI-Nullable-Frage**: oapi-codegen v2.4.1 kennt 3.1's
  type-Array nicht, redocly verbietet 3.0's `nullable` in 3.1-Specs.
  Lösung: kein Nullable-Marker; optionale Felder bleiben
  required:false → oapi-codegen erzeugt `*float32`-Pointer.
- **sqlc-Schema-Input**: Pre-Processing-Skript
  `scripts/build-sqlc-schema.py` baut `apps/backend/internal/
storage/schema.sql` aus den goose-Up-Sections in
  `infra/migrations/`. Generated, committed, von `make gen-check`
  validiert.

**Tag-Konflikt-Note:** Der ursprüngliche Übergabe-Prompt schlug
`v0.1.0` als Tag-Namen vor. `v0.1.0` ist aber bereits am 8. Mai für
Track 1 Iteration 1.1 vergeben („first feature-phase release").
Maintainer-Entscheidung 12. Mai: **v0.4.0**, weil das Tag-Schema
nach Track-1-Iter-1.3a auf v0.3.1 steht und sich für Track-2-
Iterationen sinnvoll fortsetzt.

**Akzeptanzkriterien:** siehe `prompt-iteration-2-1.md` — alle
erfüllt außer (a) Lighthouse-Run (Maintainer-Task im Browser).

**Post-Merge-Verlauf (12. Mai 2026):**

1. **v0.4.0** (PR #69, Iteration 2.1) released; Deploy auf wwn-prod
   ergab 500 auf `/api/v1/locations` (`relation "locations" does not
exist`). Ursache: Ansible-Deploy hatte keinen `goose up`-Step
   — die neue Migration kam nie auf prod an. Manueller Fix:
   `docker cp` von goose-Binary + Migrations in `wwn-postgres`,
   `docker exec` der Migration als Hot-Fix.

2. **PR #70** (Folge-PR, chore): zwei Themen gebündelt:
   (a) Ansible-Deploy-Migrationen automatisiert (App-Rolle staged
   goose-Binary, läuft im postgres-Container);
   (b) Runbook §13 dokumentiert das post-deploy-500-Pattern;
   (c) Site-Logo + Favicon-Set (eagle-PNG, 256x256 im Header +
   favicon.ico/16/32/180/192/512 im static/).
   Gemerged als v0.4.1.

3. **v0.4.1-Deploy scheiterte** am cleanup-rm: timescaledb-ha:pg16
   attached `docker exec` als `postgres`-User; der konnte das
   root-owned `/tmp/goose` (per `docker cp` gestaged) wegen
   /tmp-sticky-bit nicht löschen → Ansible-Task failed → Start-
   full-stack-Task übersprungen → App-Container blieben auf v0.4.0.
   Migration selbst war successful (`goose: migrated to version: 1`).

4. **PR #71** (Hotfix): `docker exec -u 0` für Goose + cleanup-rm.
   Gemerged als v0.4.2, deployed erfolgreich. Container alle v0.4.2,
   `/api/v1/locations` antwortet mit 3 Locations, Berlin liefert
   frisches `current` (9°C @ 09:30Z), Favicons + Logo live.

**Lessons learned in Memory + CLAUDE.md übernommen:**

- `docker cp` + `docker exec`-Pattern: ALWAYS `-u 0`, weil docker
  exec den Container-Default-User attached und root-owned Staged-
  Files in /tmp wegen sticky-bit nicht löschbar sind. (Memory:
  `feedback_docker_exec_default_user.md`.)

**Bekannt-offen, in `docs/backlog.md` dokumentiert:**

- W3 Persistent-Job-Store für APScheduler
- SSR-Upgrade für `/wetter` (Internal-API-Hostname)
- Daily-Aggregate-Tabelle + Era5-Historie (Klima-Iteration)
- testcontainers-Postgres für Backend-Handler-Tests
- Lighthouse-CI für `/wetter`
- mdsvex-Konvertierung der hardcoded Compliance-Pages
- EN-Übersetzung von `/quellen-attribution`

### Iteration 2.2 — DWD-Adapter (POI-Observations)

Status: 🟡 Lokal fertig, PR ausstehend
Datum: 2026-05-12
Geplanter Tag: **v0.5.0** (fortgeführt vom 2.1-Schema, siehe Tag-Note unten)
Plan-Skizze: `plan-iteration-2-2.md`
Übergabe-Prompt: `prompt-iteration-2-2.md`

**Commits auf dem Branch `feat/iteration-2-2-dwd-poi` (9):**

1. `afc3a63` — feat(db): migration 0002 für DWD-POI stations +
   observations-Spalten (locations +dwd_station_id/altitude_m, observations
   +pressure/humidity, PK auf (location_id, source, observed_at), 3 neue
   Klimakontrast-Locations Brocken/Zugspitze/Helgoland)
2. `6484b13` — feat(backend): sqlc-Queries + Schema für DWD-POI +
   Druck/Feuchte (neue Query `GetLatestObservationBySource`,
   `available_sources` als COALESCE-Array)
3. `0ee3e52` — feat(pyworkers): DWD-POI-Worker mit 6 Stationen
   (fetch + parse + persist mit `ON CONFLICT … DO UPDATE`)
4. `9fee0fb` — feat(pyworkers): DWD-POI Scheduler-Job alle 30 Min
   mit initial-run-on-startup
5. `9101fea` — feat(pyworkers): Open-Meteo um Druck + Feuchte
   erweitern, PK-Anpassung (pressure_msl + relative_humidity_2m,
   ON CONFLICT auf neue PK)
6. `f94ce88` — feat(backend): API erweitert um source-Param,
   altitudeM, availableSources, pressure, humidity
7. `54306a1` — feat(frontend): /wetter zeigt 6 Cards mit
   Source-Badge, Druck, Feuchte, Höhe (Paraglide DE+EN)
8. `0b57697` — docs(frontend): /quellen-attribution erweitert um
   DWD-Block (GeoNutzV)
9. `6952209` — test(2.2): DWD-Parser, backend resolveSource,
   OM-Tests an neue Tuple-Shape angepasst
10. (folgt) — docs(2.2): data-sources + runbook + backlog + STATUS

**Getroffene Implementations-Entscheidungen:**

- **DWD-Station-IDs** sind 5-stellige **WMO-Synop-Kennungen**, nicht
  DWD-Legacy/CDC-IDs. Die Plan-Skizze hatte CDC-IDs gelistet (03342,
  00078, 05792, 02115) — die POI-Endpoint kennt sie nicht. Korrekte
  POI-IDs: Potsdam 10379, Berlin 10384, Hamburg 10147, Brocken 10454,
  Zugspitze 10961, Helgoland 10015. Migration 0002 wurde noch vor
  PR-Erstellung per `git commit --amend` korrigiert.
- **`observations`-PK** erweitert auf `(location_id, source,
observed_at)`. Sonst hätten DWD und Open-Meteo sich gegenseitig
  überschrieben. TimescaleDB akzeptiert die neue PK, weil
  `observed_at` weiterhin enthalten ist.
- **Open-Meteo `pressure_msl` statt `surface_pressure`** für
  MSL-Konsistenz mit DWD. Ohne diese Korrektur klafften Berlin-Werte
  um ~5 hPa zwischen den Quellen.
- **`is_skippable` im sqlc-Schema-Builder** filtert jetzt auch
  UPDATE-Statements (vorher nur INSERT/DELETE). Migration 0002 hat
  UPDATE-Statements für die Stadt-Locations, die ohne den Fix in
  `apps/backend/internal/storage/schema.sql` geleakt wären.
- **Source-Default-Logik**: `?source=`-Param gewinnt; ohne Param
  ist DWD Default, wenn `dwd_station_id` an der Location hängt;
  sonst Open-Meteo. Frontend zeigt entsprechend die DWD-Variante
  für alle Stadt-Slugs.
- **Forecast-Pfad**: nur Open-Meteo. DWD-MOSMIX kommt als 2.2b.
- **Zugspitze hat keinen MSL-Druck** (DWD reduziert nicht für
  Stationen > ~2000 m). Parser liefert `pressure=None`,
  WeatherCard lässt die Druck-Zeile weg.

**Tag-Konflikt-Note (analog zu 2.1):** Plan-Skizze schlug `v0.2.0`,
das ist aber bereits für Track-1-Iter-1.2 vergeben. Neue Schiene
ab v0.4.0 (2.1) → v0.5.0 (2.2).

**Test-Stand:**

- 22 pyworkers-Tests grün (8 OM + 12 DWD + 2 heartbeat)
- Backend handler-Tests grün inkl. neuer resolveSource-Coverage
- Frontend vitest 5/5 grün, svelte-check 0 Errors
- E2E live gegen Dev-Stack: `http://api.localhost` liefert 6
  Locations mit allen neuen Feldern; `/locations/berlin` default
  → DWD; `?source=open-meteo` → OM + 24h-Forecast; brocken-only-DWD
  hat `current=null` für `?source=open-meteo`; zugspitze liefert
  `pressure` field omitted; unknown-slug → 404.

**Bekannt-offen** (in `docs/backlog.md` dokumentiert):

- MOSMIX-Forecast-Pfad als Iteration 2.2b (v0.5.1)
- Quellen-Vergleich-UI (Toggle pro Card / Side-by-side)
- DWD-CDC-Historie für Klima-Iteration
- Automatic Multi-Source-Failover für `current`
- OpenAPI-Strict-Enum-Validation (`?source=foo` liefert 200 statt 400)
- DWD-Station-Liste dynamisch importieren

**Browser-Smoke verbleibt auf Maintainer:** Mobile-Layout der 6 Cards,
Lighthouse-Score, visuelle Bestätigung Source-Badge / Höhe-Anzeige /
Druck-Feuchte-Zeilen.

### Iteration 2.3 — Stations-Map mit MapLibre

Status: ⏳ Geplant (Plan-Skizze fertig, Übergabe-Prompt nach 2.2)
Plan-Skizze: `plan-iteration-2-3.md`
Geschätzte Dauer: 3-4 Tage
Geplanter Tag: **v0.3.0**

**Voraussetzungen:**

- [ ] Iteration 2.2 gemerged und v0.5.0 live
- [ ] Tile-Quelle entschieden (siehe Plan-Skizze: OSM / Stadiamaps /
      MapTiler)
- [ ] Cookie-Banner-Implikationen für externe Tile-Quelle geprüft
- [ ] Übergabe-Prompt ausgearbeitet (`prompt-iteration-2-3.md`)

---

## Folge-Iterationen (Konzept-Diskussion ausstehend)

### Iteration 2.4 — Satellitenbilder

Status: ⏳ Konzept offen — braucht B.2-Wiederaufnahme + EUMETSAT-
Lizenz-Bestätigung

### Iteration 2.5 — Radar

Status: ⏳ Konzept offen — braucht B.2-Wiederaufnahme + DWD-Radolan-
Recherche

### Iteration 2.6 — ICON-Modelle (komplette Modellläufe)

Status: ⏳ Konzept offen — braucht B.3-Wiederaufnahme (Storage für
GRIB-Dateien, mehrere GB pro Modelllauf)

---

## Tag-Roadmap

```
v0.0.5      Security-Triage post-v0.0.4               ✅ 2026-05-12
                ↓
v0.4.0      Iteration 2.1 (Open-Meteo Hello World)    ✅ 2026-05-12
v0.4.1      Ansible-migrate + Logo (PR #70)           ⚠️ partial (cleanup-fail)
v0.4.2      Hotfix docker-exec -u 0 (PR #71)          ✅ 2026-05-12 live
                ↓
v0.5.0      Iteration 2.2 (DWD-POI-Adapter)           🟡 lokal fertig, PR offen
                ↓
v0.6.0      Iteration 2.3 (Stations-Map)              ⏳ nach 2.2
                ↓
Konzept-Session vor Track-2-Fortsetzung:
  - B.2 Wetterkarten-Strategie
  - B.3 Storage für große Datasets
  - EUMETSAT-Lizenz-Status für Phase 1
                ↓
v0.7.0+     2.4 / 2.5 / 2.6 nach Konzept-Session       ⏳ später
```

Tag-Numbering-Note: ursprünglicher Prompt schlug v0.1.0–v0.3.0 für
Track 2 vor. Diese Tags sind aber bereits durch Track 1 vergeben
(v0.1.0 für 1.1, v0.2.0 für 1.2, v0.3.0 für 1.3a — siehe
`git tag --list`). Neue Reihe fortgeführt ab v0.4.0.

Daten sind Schätzungen, kein Commitment. Iteration startet wenn
Voraussetzungen erfüllt sind, nicht nach Kalender.

---

## Querschnitt-Themen

### 1.3b — Image-Pipeline (Track 1, ausgesetzt)

Status: ⏭ Skipped bis Blog-Bedarf entsteht (Iteration 1.4)
Begründung: keine bildbedürftige Page in Sicht, Pipeline ohne
Use-Case wäre theoretisch. Wird mit 1.4 (Blog) zusammen
gebündelt oder unmittelbar davor implementiert.

---

## Refs

- Übergeordnete Decisions: `../feature1/feature-decisions.md` Abschnitt B
- Übergeordnete Roadmap: `../feature1/feature-roadmap.md`
- Track-1-Status: `../feature1/STATUS.md`
- Setup-Phase-Status: `../STATUS.md`
