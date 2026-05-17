# Track 2 — Wetterdaten-Import — Status

Pflege diese Datei am Ende jeder Iteration. Format analog zu
`sessions/feature1/STATUS.md`.

Status-Legende: ✅ Done · 🟡 In Progress · ⏳ Geplant · ❌ Blocked · ⏭ Skipped

Stand: 2026-05-17 (2.3 v0.6.0 live; 2.4 lokal fertig, PR/Deploy v0.7.0 ausstehend)

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

Status: ✅ Done — gemerged (PR #73, Squash `a21a6bb`), **v0.5.0** live auf wwn-prod
Datum: 2026-05-12
Tag: **v0.5.0** (fortgeführt vom 2.1-Schema, siehe Tag-Note unten)
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

**Post-Deploy-Verlauf (12. Mai 2026):**

1. **Tag v0.5.0** gesetzt + signiert direkt auf dem Squash-Commit
   `a21a6bb`; Release-Pipeline grün (~3 Min, 6 Jobs: meta, build
   backend/frontend/pyworkers/cms-auth, GitHub-Release).
2. **`bash scripts/deploy.sh production 0.5.0`** sauber durchgelaufen:
   12 OK / 6 changed / 0 failed. Migration 0002 wurde **automatisch**
   im postgres-Container ausgeführt (A.22-Akzeptanz erfüllt — kein
   manueller `docker cp`-Schritt mehr nötig). Alle vier wwn-Container
   `healthy` auf `:0.5.0`.
3. **Production-Smoke** (https://api.research.worldweathernews.com):
   - `/api/v1/locations` liefert 6 Einträge mit altitudeM, dwdStationId,
     availableSources
   - `/locations/berlin` default → DWD: T=7.1 P=1007.2 H=93
   - `/locations/berlin?source=open-meteo` → OM mit 24h-Forecast
   - `/locations/zugspitze` → T=-11.7 P=null (>2000m) H=95 altM=2964
   - Frontend `/wetter` 200

**Transient-Issue post-deploy (nicht eskaliert):**

Die jüngste OM-Row für Berlin/Hamburg/Potsdam wurde **vor dem Deploy**
vom alten Worker geschrieben (mit `pressure=NULL, humidity=NULL`).
`ON CONFLICT (location_id, source, observed_at) DO NOTHING` blockiert
das Update auf den gleichen `observed_at`. `?source=open-meteo` zeigt
deshalb für ~10-15 Min nach Deploy `pressure=null, humidity=null` — bis
die nächste 15-min-OM-Boundary einen frischen Insert mit den neuen
Feldern auslöst. Auf dem Default-Pfad (DWD) merkt der User davon nichts.
Self-resolving. Wenn das bei Folge-Iterationen mit Schema-Updates für
existierende Hypertables wieder auftritt: entweder warten oder die
jüngsten N Rows der betroffenen `(location_id, source)`-Paare einmal
löschen.

### Iteration 2.3 — Stations-Map mit MapLibre

Status: ✅ Done — gemerged (PR #76, Squash `530e0e4`), **v0.6.0** live auf wwn-prod
Datum: 2026-05-15
Plan-Skizze: `plan-iteration-2-3.md`
Übergabe-Prompt: `prompt-iteration-2-3.md`
Tag: **v0.6.0** (nicht v0.3.0 — die v0.1.0–v0.3.0-Tags sind
durch Track 1 vergeben, siehe Tag-Roadmap unten)

**Commits auf dem Branch `feat/iteration-2-3-stations-map` (6):**

1. `d035b98` — build(frontend): maplibre-gl@5.24.0 exakt gepinnt,
   beide Lockfiles via Standalone-Re-Sync-Workflow synchron
2. `bf7fea4` — refactor(frontend): `lib/config/map.ts` (zentrale
   OpenFreeMap-Style-URL) + `lib/wind.ts` (compass extrahiert +
   `windArrowRotationDeg` +180°), WeatherCard nutzt geteilten Helper
3. `69a40f5` — feat(frontend): `StationsMap.svelte` — lazy
   maplibre-gl+CSS in onMount (Q6/S1), leerer SSR-Wrapper, Marker
   mit Temp-Label + Wind-Pfeil, Popup Phase-1-Set (DOM-API, XSS-safe),
   4 neue Paraglide-Keys DE+EN
4. `a5bcc2b` — feat(frontend): StationsMap als Hero auf `/wetter`
   (N2), Loader unverändert; Lazy-Chunk verifiziert (maplibre als
   separater 1003 KB Chunk, nicht im Entry)
5. `061d8ed` — docs(frontend): OpenFreeMap-Attribution +
   Datenschutz-§5 + backlog.md (3 Einträge) + data-sources.md
6. `220fa8f` — test(2.3): wind-Helper Vitest (12 Cases) +
   CLAUDE.md/STATUS-Updates + eslint resolve()-Fix für interne Links

**Getroffene Implementations-Entscheidungen:**

- **Kein Backend-/Schema-Eingriff**: `/api/v1/locations` liefert
  `latitude`/`longitude` bereits typisiert; der `/wetter`-Loader
  reicht `details: LocationDetail[]` mit `.location` + `.current`
  durch. Die Karte konsumiert denselben Datensatz — kein N+1 über
  das Card-Maß hinaus, kein `/map-overview` (Backlog).
- **Detail-Link**: kein `/wetter/[slug]`-Route vorhanden; Popup-Link
  springt zum In-Page-Anker `#weather-card-<slug>` (von WeatherCard
  gesetzt) — passt zu N2 (Karte Hero, Cards darunter).
- **Wind-Pfeil**: keine wiederverwendbare Pfeil-Logik in WeatherCard
  vorhanden (nur `compass()`-Text). Neuer Helper `windArrowRotationDeg`
  (+180°, vom Maintainer bestätigt), `compass()` mitgeteilt nach
  `lib/wind.ts`.
- **maplibre-gl-Pin**: exakt `5.24.0` (kein Caret), bewusste
  Abweichung von der `^`-Konvention der Nachbar-Deps für
  deterministisches GPU-Rendering.
- **Lazy-Verifikation**: maplibre landet als separater ~1 MB Chunk,
  nicht im Entry-/Default-Bundle (Q6 erfüllt, im Build verifiziert).

**Akzeptanzkriterien:** alle erfüllt außer (a) Lighthouse-Run und
(b) Mobile-Pinch/Touch-Smoke — beides Maintainer-Browser-Tasks,
analog zum 2.1/2.2-Muster, noch offen. `svelte-check` 0/0, `lint`
grün, 12/12 Vitest grün, Build grün.

**Post-Deploy-Verlauf (15. Mai 2026):**

1. **Tag v0.6.0** signiert auf dem Squash-Commit `530e0e4` gesetzt,
   Release-Pipeline gebaut/published (4 Images).
2. **`bash scripts/deploy.sh production 0.6.0`** durchgelaufen.
3. **Authoritative Live-Check** (`docker ps` auf wwn-prod
   10.100.100.21): alle vier wwn-Container auf `:0.6.0`, alle
   `healthy` (backend, frontend, pyworkers, cms-auth). Keine
   DB-Migration in 2.3 (kein Schema-Eingriff) — A.22 nicht berührt.
4. **Public-Smoke:**
   - `https://research.worldweathernews.com/wetter` → 200
   - `https://api.research.worldweathernews.com/api/v1/locations`
     → JSON mit den 6 Locations (lat/lon, availableSources)

**Offen (Maintainer-Browser-Tasks, kein Blocker):** Lighthouse-Run
auf `/wetter` und Mobile-Pinch/Touch-Smoke der Karte — visuelle
Bestätigung Marker/Wind-Pfeil/Popup, Lazy-Chunk im Network-Tab.

**Voraussetzungen:**

- [x] Iteration 2.2 gemerged und v0.5.0 live (PR #73, STATUS-PR #74)
- [x] Tile-Quelle entschieden: **T2 OpenFreeMap** (Liberty-Style,
      Vector-Tiles, frei, kein Account/Key, cookiefrei). T1 OSM-Raster
      und T3 MapTiler bewusst verworfen — Begründung in der
      Plan-Skizze + Tile-Optionen-Analyse vom 15. Mai.
- [x] Cookie-Banner-Implikationen geprüft: **keine Änderung nötig** —
      OpenFreeMap setzt keine Cookies (explizit zugesichert), kein
      Tracking. Datenschutz-Page bekommt nur einen IP-Hinweis-Block.
- [x] Übergabe-Prompt ausgearbeitet (`prompt-iteration-2-3.md`,
      T2 fix + Q1–Q6-Defaults eingearbeitet)

**Entscheidungs-Notiz (15. Mai 2026):**

Tile-Quelle als entscheidungsreifer Vergleich (T1/T2/T3) aufbereitet,
Maintainer hat **T2 OpenFreeMap** gewählt. Self-hosting-Spannung
(A.19) bewusst akzeptiert: Tile-Serving ist client-seitig und nicht
backend-kritisch (Karte degradiert nur, Plattform läuft weiter) →
fällt unter die Edge-/Cache-Ausnahme. Self-hosted OpenFreeMap-Stack
als Backup-Pfad in `docs/backlog.md` (Storage-Bedarf, eigene spätere
Iteration). Style-URL wird als zentrale Config-Konstante gehalten,
damit ein späterer Wechsel ein Ein-Zeilen-Change ist. Q1–Q6 mit den
Plan-Skizze-Defaults fixiert (N2 / Marker-C / Phase-1-Set / Wind-W2 /
SSR-S1 / Lazy-Bundle).

---

## Folge-Iterationen

### Iteration 2.4 — Satellitenbilder

Status: ✅ Done — **v0.7.2 live**, `/satellit` verifiziert
Datum: 2026-05-17
Plan-Skizze: `plan-iteration-2-4.md` · Übergabe-Prompt: `prompt-iteration-2-4.md`
Tag: **v0.7.2** (v0.7.0 = #82, v0.7.1 = #83 waren unvollständige
Zwischen-Tags; nur v0.7.2 enthält #82 + #83 + #84 und wurde deployed)

**Commits auf dem Branch `feat/iteration-2-4-satellite` (5):**

1. `d041461` — feat(pyworkers): EUMETSAT-Worker (EUMETView WMS
   `msg_fes:ir108` → A.13-Bucket, rollierendes 24-h-Fenster +
   `index.json`, `aiobotocore`, W1/`next_run_time=now`, skippt
   sauber ohne S3-Config)
2. `2c0553a` — feat(infra): `sat/*` public-read in
   `bucket-policy.json` + `media-storage.md` (5. Prefix, nur
   Worker-befüllt)
3. `a5dfd3e` — feat(frontend): `/satellit`-Route (`ssr=false`),
   `SatelliteMap.svelte` (lazy maplibre, image-Source aus Frames,
   Zeit-/Deckkraft-Slider, Play/Pause), `lib/config/satellite.ts`
4. `e590652` — feat(frontend): Nav „Satellit" +
   EUMETSAT-Attribution-Artikel; stale „geplant"-Zeile entfernt
5. `cae76f9` — test(2.4): pyworkers EUMETSAT (9) + Frontend
   Satelliten-Helpers (`lib/satellite.ts` extrahiert, 8 Vitest)

**Getroffene Implementations-Entscheidungen:**

- **Pfad A** statt Roh-SEVIRI: EUMETView WMS liefert fertige
  Composites; Q4 **live verifiziert** — `view.eumetsat.int/geoserver/wms`
  ist **public, kein Auth**, `EPSG:3857` direkt → keine Reprojektion,
  kein Satpy. → `eumetsat.env`-Secret wird für 2.4 **nicht** gebraucht
  (bleibt für K1 ~2.6 vorbereitet).
- **S3-Client = `aiobotocore`** (`>=2.13,<3`, löst 2.26.0) — async-
  nativ zum pyworkers-Stack; Volume-Alternative gegen B.3 abgewogen
  und verworfen.
- **Q5 = eigene `/satellit`-Route** (B.6 `[DECIDED]` überstimmt das
  weiche D1-Bauchgefühl).
- **Kein Backend-/OpenAPI-/DB-Eingriff**; **kein Datenschutz-§5**
  (server-seitig geholt, eigener `media.`-Origin — A.19).
- Plan-Skizze-Annahme „kein Policy-Change" war **falsch**: `sat/`
  brauchte einen Bucket-Policy-Eintrag (in Commit 2 erledigt).

**Folge-PRs nach #82:**

- **#83** `fix(pyworkers)` — S3-Config-Namen auf `S3_*`-Schema
  (`s3_access_key`/`s3_secret_key`), macht das Secret-Wiring zu einem
  sauberen 1:1-Prefix-Mapping; fixt zugleich den
  `media-storage.sops.env`→`media-storage.env`-Doku-Bug.
- **#84** `chore(secrets)` — `WWN_PY_S3_*` in `pyworkers.env` (SOPS).

**Maintainer-Tasks — alle erledigt:**

- [x] **Bucket-Policy auf den Live-Bucket angewendet** (nicht
      auto-deployed; Credentials aus `media-storage.env`-SOPS):

  ```
  aws s3api put-bucket-policy --bucket media-worldweathernews-prod \
    --policy file://infra/object-storage/bucket-policy.json \
    --endpoint-url "$S3_ENDPOINT"
  ```

- [x] **`WWN_PY_S3_ENDPOINT`/`_ACCESS_KEY`/`_SECRET_KEY`/`_REGION`/
      `_BUCKET`** in `infra/secrets/production/pyworkers.env` (SOPS) —
      1:1-Prefix-Mapping der `S3_*`-Werte aus `media-storage.env`, **kein**
      Ansible-/Compose-Change (Container `env_file`d die Datei bereits).
      Zwei sops-Korrektur-Runden nötig (erst fehlten die Keys, dann
      doppeltes `S3_`-Segment) — Verifikation lief ausschließlich über
      Key-Namen, nie über Werte.
- [x] `eumetsat.env`: für 2.4 **nicht** gebraucht — nichts zu tun.
- [ ] Lighthouse/Mobile-Smoke `/satellit` (Browser, optional, kein Blocker).

**Verifikation (CI):** pyworkers ruff/mypy-strict clean, **31 pytest**;
Frontend `svelte-check` 0/0, lint+build grün, **17 Vitest**;
maplibre lazy (separater Chunk, nicht im Entry).

**Verifikation (live, nach v0.7.2-Deploy 2026-05-17):**

- `https://research.worldweathernews.com/satellit` → HTTP 200
- `media.worldweathernews.com/sat/ir108/index.json` → 1 Frame,
  `source=eumetsat`, `© EUMETSAT`, BBOX Europa (−17.97..40.42°E /
  29.97..69.79°N)
- Frame-PNG → HTTP 200, `image/png`, 1.14 MB, 1024×1024 RGB
- Worker greift (`eumetsat_frame_persisted`, kein silent skip);
  rollierendes 15-Min-Fenster läuft.

Damit ist die Pfad-A-Kette pyworkers → A.13-Bucket → `media.` →
`/satellit` produktiv. Nächster Evolutionspfad ist K1 (Roh-SEVIRI +
Satpy, eigene Composites, ~Iteration 2.6) — siehe `docs/backlog.md`.

### Iteration 2.5 — Radar

Status: ⏳ Konzept offen — B.2 ist mit der 16.-Mai-Session entschieden
(K3-Linie gilt analog), offen bleibt die DWD-Radolan-Recherche
(Format, Update-Frequenz, Reprojektion)

### Iteration 2.6 — ICON-Modelle (komplette Modellläufe)

Status: ⏳ Konzept offen — hier landet (a) der **K1-Evolutionspfad**
aus B.2 (komplettes Modellfeld-Rendering, ICON+Cartopy) und (b) die
finale **B.3-Big-Data-Storage-Entscheidung** (MinIO-VM vs. Hetzner
Storage Box, mehrere GB pro Modelllauf)

---

## Tag-Roadmap

```
v0.0.5      Security-Triage post-v0.0.4               ✅ 2026-05-12
                ↓
v0.4.0      Iteration 2.1 (Open-Meteo Hello World)    ✅ 2026-05-12
v0.4.1      Ansible-migrate + Logo (PR #70)           ⚠️ partial (cleanup-fail)
v0.4.2      Hotfix docker-exec -u 0 (PR #71)          ✅ 2026-05-12 live
                ↓
v0.5.0      Iteration 2.2 (DWD-POI-Adapter, PR #73)   ✅ 2026-05-12 live
                ↓
v0.6.0      Iteration 2.3 (Stations-Map, PR #76)      ✅ 2026-05-15 live
                ↓
Konzept-Session (16. Mai): B.2=K3 / B.3=A.13-Bucket /
  EUMETSAT-Lizenz geklärt — alle DECIDED
                ↓
v0.7.0      Iteration 2.4 (Satellitenbilder, EUMETSAT)  🟡 lokal fertig, PR/Deploy offen
                ↓
v0.8.0+     2.5 (Radar) / 2.6 (ICON, K1-Pfad)           ⏳ später
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
