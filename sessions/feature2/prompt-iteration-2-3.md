# Iteration 2.3 — Stations-Map mit MapLibre

**Übergabe-Prompt für Claude Code auf wwn-dev**

---

## Verwendung

Diesen Prompt **als ersten Prompt einer neuen Claude-Code-Session**
auf wwn-dev (10.100.100.113) verwenden, im Repo-Root von
`worldweathernews`. Voraussetzung: Iteration 2.2 ist als v0.5.0 live
(PR #73, Squash `a21a6bb`), STATUS-Doku-PR #74 ist gemerged
(`docs/feature2-status-v0.5.0-live`, Squash `0a2debc`).

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

Wir starten Iteration 2.3 (Stations-Map mit MapLibre) auf
worldweathernews.com. Track 2 Iteration 2.2 (DWD-POI-Adapter) ist
v0.5.0 live: sechs Locations (Potsdam, Berlin, Hamburg, Brocken,
Zugspitze, Helgoland), zwei Datenquellen (DWD default, Open-Meteo via
`?source=open-meteo`), sechs Variablen inkl. Druck + Feuchte. 2.3
gibt diesen Daten die **erste räumliche Darstellung**: eine
interaktive Karte mit Markern für alle Stationen.

Lies bitte zuerst:

1. `CLAUDE.md` im Repo-Root — die zentralen Spielregeln. Besonders
   relevant für diese Iteration:
   - „Häufige Fallen" → **Standalone `apps/frontend/pnpm-lock.yaml`**
     (Docker-Build nutzt diesen, nicht den Workspace-Root — Re-Sync-
     Workflow für jedes Frontend-Dependency-Update)
   - „Häufige Fallen" → **Tailwind v4 mit `@theme`-Block**
     (keine `tailwind.config.js`, `@tailwindcss/vite`)
   - „Häufige Fallen" → **Self-hosting-Prinzip für neue Services**
     (Ausnahme: reine Edge-/Cache-Schichten)
   - „Wo finde ich was" → Frontend-Pfade
2. `sessions/feature2/STATUS.md` — Track-2-Stand, Tag-Roadmap
3. `sessions/feature2/plan-iteration-2-3.md` — Plan-Skizze mit
   Tile-Optionen-Analyse und Konzept-Fragen Q1–Q6
4. `sessions/feature1/feature-decisions.md` — insbesondere:
   - **A.19** (Self-hosting-Prinzip / Edge-Cache-Ausnahme)
   - **A.20** (OpenAPI ohne `nullable`) — falls Backend angefasst wird
   - **B.6** (Frontend-Position pro Feature)
5. Die 2.2-Commits zur Orientierung (gemergte PRs #73, #74) und die
   bestehende `/wetter`-Route als Referenz-Pattern

Sobald du diese gelesen hast, melde dich kurz mit einer Zusammen-
fassung von:

- Stack-Stand (v0.5.0 live, 6 Locations, 2 Quellen)
- Wie `/wetter` aktuell Daten lädt (Loader-Pattern, `ssr = false`)
- Der entschiedenen Tile-Quelle und warum (siehe unten)

## Feature-Phase-Modus

Wie üblich seit Track 1: **Claude Code committet nach expliziter
Freigabe.**

Workflow:

1. Branch anlegen (`feat/iteration-2-3-stations-map`)
2. Implementation in mehreren Commits
3. **Vor jedem Commit: mich um Freigabe fragen**
4. Bei "OK" / "commit" / "merge": committen oder mergen
5. Bei "warte" / "nochmal" / "anders": warten und nachbessern
6. Push zu GitHub: erst nach explizitem "push" oder "PR aufmachen"

## Bereits entschieden — nicht neu aufrollen

Diese Punkte sind vom Maintainer **vor** der Session entschieden
worden. Nicht erneut zur Diskussion stellen, einfach umsetzen:

### Tile-Quelle: T2 OpenFreeMap (fix)

- **Quelle**: OpenFreeMap Public Instance, Style **Liberty**
  (`https://tiles.openfreemap.org/styles/liberty`)
- **Begründung**: Vector-Tiles (GPU-Performance), frei, kein
  Account / kein API-Key, keine Cookies (explizit zugesichert) →
  **keine Cookie-Banner-Änderung nötig**. Tonal konsistent mit dem
  Projekt (frei, OSS, OSM-Daten).
- **Self-hosting-Spannung bewusst akzeptiert**: Tile-Serving ist
  client-seitig und **nicht** im backend-kritischen Pfad — die
  Plattform funktioniert ohne Karte weiter, die Karte degradiert nur.
  Fällt damit unter die Edge-/Cache-Ausnahme von A.19. Self-Hosting
  des OpenFreeMap-Stacks (Planetiler, Planet-Tiles, >100 GB) ist auf
  dem aktuellen Proxmox nicht tragbar und wäre eine eigene spätere
  Iteration (als Backlog-Backup-Pfad dokumentieren).
- **Pflicht-Umsetzung**: Style-/Source-URL als **eine zentrale
  Config-Konstante** halten (z. B. `apps/frontend/src/lib/config/map.ts`),
  damit ein späterer Wechsel auf self-hosted OpenFreeMap oder MapTiler
  ein Ein-Zeilen-Change ist. **Keine** URL hartkodiert in der
  Komponente.

### Konzept-Fragen Q1–Q6 (Defaults aus der Plan-Skizze, fixiert)

| Frage             | Entscheidung                                                                                                                                                      |
| ----------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Q1 — Navigation   | **N2**: Karte als Hero auf `/wetter`, oberhalb der bestehenden WeatherCards. Keine neue Route.                                                                    |
| Q2 — Marker       | **C**: Pin mit Temperatur-Wert als Label, Klick öffnet Detail-Popup.                                                                                              |
| Q3 — Marker-Daten | **Phase-1-Set**: Stadt-Name, Quelle, aktuelle Temperatur, Niederschlag jetzt, Wind (Speed + Richtung), Link zur Detail-Page. Keine Sparklines (eigene Iteration). |
| Q4 — Wind         | **W2**: SVG-Pfeil im Pin, CSS-Rotation nach Wind-Direction.                                                                                                       |
| Q5 — SSR          | **S1**: Karten-Komponente in `onMount()` initialisieren, Server rendert leeren Wrapper. `/wetter` bleibt `ssr = false` (B.6).                                     |
| Q6 — Bundle       | **Lazy-load**: `maplibre-gl` via dynamic `import()` in der Karten-Komponente, nicht im Default-Bundle.                                                            |

### Datenbeschaffung: kein neuer Backend-Endpoint

- Die Karte konsumiert **dieselben Loader-Daten wie die
  WeatherCards** auf `/wetter`. Keine separate API, kein N+1, kein
  OpenAPI-Schema-Change.
- **Voraussetzung verifizieren**: der `/wetter`-Loader liefert pro
  Location bereits aktuelle Observation (Temp, Niederschlag, Wind) +
  `latitude`/`longitude`. Falls die Koordinaten im Loader-Response
  fehlen, ist das die **einzige** erlaubte Backend-/Schema-Anpassung
  dieser Iteration (`latitude`/`longitude` zur Locations-Response
  ergänzen, A.20: ohne `nullable`, Pointer). Erst prüfen, dann
  vorschlagen.
- `/api/v1/map-overview` als gebündelter Endpoint ist **bewusst
  nicht** Teil von 2.3 → Backlog-Punkt, falls die N-fach-Detailfetches
  auf `/wetter` später Performance-Probleme machen.

## Was diese Iteration liefert

```
/wetter (ssr = false, B.6)
   │
   ├─ Hero: StationsMap.svelte  ← NEU
   │     maplibre-gl (lazy import)
   │     style: OpenFreeMap Liberty (zentrale Config-Konstante)
   │     6 Marker (Pin + Temp-Label + Wind-Pfeil)
   │     Klick → Popup (Stadt, Quelle, Temp, Niederschlag,
   │              Wind, Detail-Link)
   │
   └─ darunter: bestehende WeatherCards (unverändert)
         gespeist aus demselben Loader-Datensatz
```

## Iterations-Plan

### Schritt 1 — Branch + Plan + Verifikation

1. Branch anlegen: `feat/iteration-2-3-stations-map`
2. Verifikation:
   - `uname -n` zeigt `wwn-dev`
   - `git rev-parse --show-toplevel` ist Repo-Root
   - `git config --get user.email` = `hwr@relations4u.de`
3. Live-Stand + Loader-Datenform prüfen:
   ```
   curl -s https://api.research.worldweathernews.com/api/v1/locations | jq '.[0]'
   # Enthält der Datensatz latitude/longitude? Aktuelle Observation?
   ```
   Den `/wetter`-Loader (`+page.ts` / `+page.server.ts`) lesen und
   feststellen, ob Koordinaten + aktuelle Obs schon durchgereicht
   werden.
4. Plan-Vorschlag für Schritte 2-9, dann Freigabe abwarten.
   **Im Plan explizit benennen**, ob eine `latitude`/`longitude`-
   Backend-Ergänzung nötig ist (siehe „Datenbeschaffung" oben).

### Schritt 2 — maplibre-gl als Frontend-Dependency

1. `maplibre-gl` mit **exakter Version pinnen** (kein `latest`, kein
   Caret — aktuelle Stable-Major v5.x, exakte Patch-Version festlegen
   und im Commit begründen).
2. **Standalone-Lockfile-Re-Sync** (CLAUDE.md „Häufige Fallen"):
   - `pnpm-workspace.yaml` temporär verstecken
   - `pnpm install` (regeneriert `apps/frontend/pnpm-lock.yaml`)
   - Workspace zurück
   - `pnpm install` erneut (regeneriert Workspace-Lockfile)
   - **beide** Lockfiles committen
3. Das maplibre-gl-CSS (`maplibre-gl/dist/maplibre-gl.css`) muss
   mitgeladen werden — entweder Import in der Komponente oder via
   `app.css`. Tailwind v4 nicht damit kollidieren lassen
   (Lightning CSS, kein autoprefixer-Konflikt — separat verifizieren,
   dass der Build grün bleibt).

### Schritt 3 — Map-Config-Konstante

`apps/frontend/src/lib/config/map.ts`:

```ts
// Tile-Quelle für die Stations-Map.
// Wechsel auf self-hosted OpenFreeMap oder MapTiler ist hier
// ein Ein-Zeilen-Change. Siehe Iteration-2.3-Entscheidung.
export const MAP_STYLE_URL = "https://tiles.openfreemap.org/styles/liberty";

export const MAP_INITIAL_VIEW = {
  // grob zentriert auf die 6 DE-Stationen
  center: [10.0, 51.0] as [number, number],
  zoom: 5,
};
```

### Schritt 4 — StationsMap-Komponente

`apps/frontend/src/lib/components/StationsMap.svelte`:

- Props: die Locations-Liste aus dem `/wetter`-Loader (inkl.
  Koordinaten + aktuelle Obs)
- `onMount()` (S1): `const maplibre = await import("maplibre-gl")`,
  Map mit `MAP_STYLE_URL` initialisieren
- Server rendert nur einen leeren `<div>`-Wrapper mit fester Höhe +
  Loading-State; WebGL/DOM-Init erst client-side
- Aufräumen in `onDestroy()`: `map.remove()`
- Map-Höhe responsive (Mobile niedriger als Desktop), aber feste
  min-height gegen Layout-Shift

### Schritt 5 — Marker (Pin + Temp + Wind-Pfeil)

- Custom HTML-Marker pro Location (`new maplibregl.Marker(el)`)
- Pin trägt die **aktuelle Temperatur** als Label (Variante C)
- **SVG-Pfeil** im Pin, `transform: rotate()` nach
  `wind_direction` (W2). Meteorologische Konvention beachten:
  Windrichtung = woher der Wind kommt; Pfeil zeigt in Wind-Richtung
  konsistent mit der WeatherCard-Darstellung aus 2.1/2.2 (gleiche
  Konvention wie dort, nicht neu erfinden — bestehende
  Wind-Pfeil-Logik aus `WeatherCard.svelte` wiederverwenden).
- Marker-Element accessible halten (Button-Semantik, `aria-label`
  mit Stadt + Temperatur)

### Schritt 6 — Popup

- Klick auf Marker → `maplibregl.Popup`
- Inhalt (Phase-1-Set, Q3): Stadt-Name, Quellen-Badge (DWD /
  Open-Meteo), aktuelle Temperatur, Niederschlag jetzt, Wind
  (Speed + Richtung als Text), Link zur Detail-Page
  (`/wetter/<slug>` bzw. bestehende Detail-Route)
- Strings über Paraglide (DE + EN), bestehende Keys aus 2.2
  wiederverwenden wo möglich (`weather_source_dwd`, … ),
  neue nur wenn nötig

### Schritt 7 — Integration in /wetter

`apps/frontend/src/routes/wetter/+page.svelte`:

- `StationsMap` als **Hero oberhalb** der bestehenden WeatherCards
  (N2). Loader unverändert lassen, falls er schon Koordinaten +
  Obs liefert; sonst minimal ergänzen (siehe Schritt 1 / 4).
- WeatherCards darunter **unverändert** — keine Regression an der
  bestehenden Card-Darstellung
- Mobile: Karte zuerst, dann Cards untereinander; Desktop: Karte
  full-width Hero, Cards im bestehenden Grid darunter
- Touch/Pinch-Zoom auf Mobile testen

### Schritt 8 — Datenschutz- + Attribution-Doku

1. **Attribution**: MapLibre fügt die OpenFreeMap-Attribution
   automatisch ein — verifizieren, dass sie sichtbar ist
   („OpenFreeMap © OpenMapTiles Data from OpenStreetMap").
2. **Quellen-Attribution-Page** erweitern
   (`apps/frontend/src/content/pages/de/quellen-attribution.md` +
   `/en/source-attribution.md`): Block für OpenStreetMap +
   OpenFreeMap.
3. **Datenschutz-Page** ergänzen: kurzer Hinweis, dass die Karte
   OpenFreeMap/OSM-Tiles nachlädt und dabei die Client-IP an den
   Tile-Server geht, keine Cookies/kein Tracking.
4. `docs/backlog.md` ergänzen um:
   - Self-hosted OpenFreeMap-Stack als Backup-Pfad (Storage-Bedarf)
   - `/api/v1/map-overview` gebündelter Endpoint (falls N-Fetch
     auf `/wetter` später Performance-Thema wird)
   - Marker-Clustering (relevant erst bei vielen Stationen, Phase 2)

### Schritt 9 — Tests + Smoke-Checks + Doku

Frontend:

- Unit-Test für die Wind-Pfeil-Rotations-Berechnung (Vitest), falls
  Logik nicht 1:1 aus `WeatherCard` wiederverwendet
- `svelte-check` / `eslint` / `prettier` grün

E2E manuell:

- `/wetter` zeigt Karte als Hero, darunter 6 Cards
- 6 Marker an korrekten Koordinaten, Temp-Label + Wind-Pfeil
- Klick → Popup mit korrekten Daten + funktionierendem Detail-Link
- Mobile-Responsive (Pinch-Zoom, Touch-Pan)
- **Lighthouse**: Bundle-Size-Auswirkung von maplibre-gl messen;
  Karte darf nicht im Default-Bundle landen (Lazy-Import
  verifizieren — Network-Tab: maplibre-gl-Chunk lädt erst beim
  Mount der `/wetter`-Route)
- Keine Cookies von `tiles.openfreemap.org` (Network-/Application-Tab)

Doku:

- `docs/data-sources.md`: kurzer Block „Kartenbasis: OpenFreeMap
  (OSM-Daten)" — keine Wetter-Datenquelle, aber der Vollständigkeit
  halber
- `CLAUDE.md` „Wo finde ich was" um StationsMap-Komponente +
  `lib/config/map.ts` ergänzen
- `sessions/feature2/STATUS.md` auf 2.3-Done aktualisieren,
  Tag-Roadmap-Zeile v0.6.0 abhaken

## Akzeptanzkriterien

- [ ] Branch `feat/iteration-2-3-stations-map` angelegt
- [ ] `maplibre-gl` exakt gepinnt, **beide** Lockfiles
      (Workspace + standalone `apps/frontend/pnpm-lock.yaml`)
      konsistent committed
- [ ] Tile-Style-URL **ausschließlich** in
      `lib/config/map.ts`, nirgends hartkodiert
- [ ] Karte initialisiert client-side (S1), Server rendert leeren
      Wrapper, kein SSR-Crash, `/wetter` bleibt `ssr = false`
- [ ] 6 Marker an korrekten Koordinaten mit Temp-Label + Wind-Pfeil
      (W2, gleiche Wind-Konvention wie WeatherCard)
- [ ] Klick → Popup mit Phase-1-Set (Q3) + funktionierendem
      Detail-Link
- [ ] Karte als Hero oberhalb der bestehenden Cards (N2),
      Cards ohne Regression
- [ ] maplibre-gl lazy geladen — nicht im Default-Bundle
      (im Network-Tab verifiziert)
- [ ] Keine Cookies von der Tile-Quelle, OpenFreeMap-Attribution
      sichtbar
- [ ] Quellen-Attribution-Page + Datenschutz-Page erweitert (DE+EN)
- [ ] Falls Backend angefasst: nur `latitude`/`longitude`-Ergänzung,
      OpenAPI ohne `nullable` (A.20), `make gen-check` grün
- [ ] Mobile-Responsive (Pinch-Zoom, Touch-Pan)
- [ ] `make lint && make test` grün
- [ ] `docs/data-sources.md` / `docs/backlog.md` / `CLAUDE.md` /
      `sessions/feature2/STATUS.md` aktualisiert
- [ ] **Deploy auf wwn-prod ohne manuelle Schritte** (A.22) —
      nur relevant falls Migration für lat/lon nötig war
- [ ] PR-Erstellung erst nach finalem OK des Maintainers
- [ ] Tag **v0.6.0** als Iteration-2.3-Release

> **Tag-Numbering-Note:** Die Plan-Skizze nennt versehentlich
> v0.3.0. Korrekt ist **v0.6.0** — v0.1.0–v0.3.0 sind durch Track 1
> vergeben (siehe Tag-Roadmap in `sessions/feature2/STATUS.md`).

## Was du **noch nicht** baust

- **Marker-Clustering** → erst relevant bei vielen Stationen,
  Phase-2-Backlog
- **Forecast-Sparkline im Popup** → eigene Chart-Komponenten-Iteration
- **Alarm-/Warn-Indikator im Marker** → spätere Iteration
- **Wind-Pfeil als nativer MapLibre-Layer** (W3) → Phase 2, erst
  wenn die Karte größere Stations-Mengen zeigt
- **`/api/v1/map-overview` gebündelter Endpoint** → Backlog, nur
  falls N-Fetch auf `/wetter` zum Performance-Thema wird
- **Self-hosted Tile-Stack** → eigene spätere Iteration
  (Storage-Bedarf), als Backlog-Backup-Pfad dokumentiert
- **Eigene `/karte`-Route** → bewusst verworfen (N2 entschieden)
- **MapTiler / API-Key-basierte Quellen** → bewusst verworfen
  (Cookie-/Tracking-/Vendor-Lock-Gründe)

Wenn du verlockt bist, Clustering oder Sparklines gleich
mitzubauen — widerstehen, Iterations-Disziplin halten.

## Wenn etwas unklar ist

Frag mich. Insbesondere:

- **Loader-Datenform**: Liefert der `/wetter`-Loader schon
  Koordinaten + aktuelle Obs pro Location? Erst prüfen, dann
  vorschlagen, ob eine minimale Backend-Ergänzung nötig ist —
  nicht ungefragt das API-Schema anfassen.
- **Wind-Pfeil-Konvention**: bestehende Logik aus
  `WeatherCard.svelte` wiederverwenden, nicht neu definieren.
  Falls dort keine isolierbare Funktion existiert: vorschlagen,
  wie wir sie teilen, bevor du dupliziertst.
- **maplibre-gl-Patch-Version**: aktuelle Stable-v5.x ermitteln
  und die exakte Version zur Bestätigung nennen, bevor du sie pinnst.
- **CSS-Integration**: ob maplibre-gl-CSS in `app.css` oder
  Komponenten-scoped — Vorschlag mit Tailwind-v4-Verträglichkeit
  begründen.

Lass uns loslegen. Bestätige mir kurz, dass du die Dokumente
gelesen hast, und schlag den ersten Schritt vor.
