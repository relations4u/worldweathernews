# Plan-Skizze — Iteration 2.3 — Stations-Map mit MapLibre

**Stand: 12. Mai 2026 — Plan-Skizze**

Dieses Dokument ist eine **frühe Skizze** für Iteration 2.3, **nicht**
der Submission-Ready-Übergabe-Prompt für Claude Code. Final wird
`prompt-iteration-2-3.md` daraus ausgearbeitet, nachdem Iteration 2.2
gemerged ist.

---

## Ziel

Visualisiere die Locations und ihre aktuellen Wetter-Daten **auf einer
interaktiven Karte**. Erste echte räumliche Darstellung der Plattform.

**Tag: v0.3.0**
**Geschätzte Dauer: 3-4 Tage**

---

## Was wir wissen über MapLibre + Tile-Optionen

**MapLibre GL JS** ist die etablierte open-source Alternative zu
Mapbox-GL-JS (BSD-3-Clause, aktiv gewartet, vector-tile-rendering
auf der GPU). Funktioniert nativ in SvelteKit ohne Hacks.

Tile-Quellen — drei realistische Optionen:

### Option T1 — OpenStreetMap Raster (direct)

```
URL:         https://tile.openstreetmap.org/{z}/{x}/{y}.png
Lizenz:      OSM ODbL, Attribution Pflicht
Kosten:      kostenlos
Limits:      offizielle Nutzungsbedingungen verbieten heavy use
             auf der OSM-Foundation-Infrastruktur
Cookies:     nein (Server setzt keine)
DSGVO:       OK (deutscher Server-Standort variiert)
Format:      Raster-Tiles 256px PNG

Verdict:     funktioniert zum Loslegen, aber unfair gegen die
             OSM-Foundation auf längere Sicht. Sollte nicht
             primäre Quelle für Production werden.
```

### Option T2 — OpenFreeMap (recommended für Production)

```
URL:         https://tiles.openfreemap.org/styles/liberty
             (plus weitere Styles: positron, bright, dark)
Lizenz:      MIT, Daten aus OSM
Kosten:      kostenlos, donation-finanziert
Limits:      keine — unlimitiert, kein Account, keine API-Keys
Cookies:     nein (explizit zugesichert)
DSGVO:       OK (Public Instance ist EU-basiert)
Format:      Vector-Tiles via MapLibre Style-JSON

Attribution: „OpenFreeMap © OpenMapTiles Data from OpenStreetMap"
             (MapLibre fügt sie automatisch ein)

Verdict:     ideal für eine Forschungs-Plattform.
             Vector-Tiles = bessere Performance und Skalierung.
             Public Instance hat die letzten 12 Monate stabil
             funktioniert (Single-Person-Projekt mit klarer
             Donor-Strategie).
```

### Option T3 — MapTiler (commercial Free Tier)

```
URL:         https://api.maptiler.com/maps/streets/style.json?key=KEY
Lizenz:      commercial, Free Tier mit 100k loads/Monat
Kosten:      ab 100k loads → bezahlt
Limits:      Account + API-Key nötig
Cookies:     ja (Tracking-Cookies möglich)
DSGVO:       complex — eigene DPA, EU-Datenschutz
Format:      Vector + Raster

Verdict:     wenn wir auf Map-bedingten Polish-Wert legen wollen
             (besseres Styling). Aber: API-Key in Frontend,
             Tracking-Implikationen, Cookie-Banner muss
             erweitert werden.
```

### Bauchgefühl

**Option T2 (OpenFreeMap)** für 2.3 — keine Account-Schritte,
keine Tracking-Implikationen, gleicher Tonalität wie das übrige Projekt
(self-hosted, frei, OSS). Falls die Public Instance Probleme bekommt,
ist Self-Hosting der OpenFreeMap-Stack die Backup-Option (eigene
Iteration). Falls echte Karten-Polish gebraucht wird, kommt MapTiler
später als Upgrade.

---

## Architektur-Skizze

```
┌─────────────────────────────────────────────────────┐
│  Frontend                                           │
│                                                     │
│  apps/frontend/src/routes/karte/+page.svelte        │
│    (oder Karten-Komponente auf /wetter)             │
│                                                     │
│  apps/frontend/src/lib/components/StationsMap.svelte│
│    + maplibre-gl als pnpm-Dep                       │
│    + style: OpenFreeMap Liberty                     │
│    + Markers für alle Locations                     │
│    + Popup bei Klick: Stadt, Temp, Wind, Source     │
│    + Optional: Pfeile für Wind-Richtung             │
└─────────────────────────────────────────────────────┘
                       ↓ (load on mount)
┌─────────────────────────────────────────────────────┐
│  Backend-API                                        │
│                                                     │
│  GET /api/v1/locations (existiert seit 2.1)         │
│    erweitert um aktuelles Wetter pro Location       │
│    (oder eigener Endpoint /api/v1/map-overview)     │
└─────────────────────────────────────────────────────┘
```

---

## Offene Konzept-Fragen

### Q1 — Wo lebt die Karte in der Navigation?

```
Option N1 — Eigene Route /karte
  Klare Trennung, separate URL
  Sitebar-Navigation: Übersicht | Karte | ...

Option N2 — Karte auf /wetter, oberhalb der Cards
  Eine Wetter-Page, Karte als Hero, Cards darunter
  Mehr Aufmerksamkeit für die Karte als Default

Option N3 — Karte als Tab auf der bestehenden Route
  Karte | Liste — User wählt Darstellung
  Mehr Flexibilität
```

Bauchgefühl: **Option N2** — Karte als Hero auf der Wetter-Page.
Sofort sichtbar, ohne Tab-Klick.

### Q2 — Marker-Style und Interaktivität?

```
Marker-Variante A — Pin mit Temperatur-Wert direkt drin
  Marker zeigt direkt "18°C" → schnelle Übersicht
  Größerer Marker, im Cluster-Fall schwerer

Marker-Variante B — kleine Pins, Klick → Popup
  Karten-Übersicht clean, Detail on demand
  Klick erforderlich

Marker-Variante C — Hybrid: Pin mit Wert + Klick öffnet
                    Detail-Popup mit Forecast-Chart
  Beste Übersicht plus drill-down
  Etwas mehr Code
```

Bauchgefühl: **Variante C** — fast wie A, aber mit Popup. Pin trägt
Temperatur als Beschriftung, Popup hat mehr Details. Phase 1 ist die
Stations-Liste klein (3-6 Marker), Cluster braucht's noch nicht.

### Q3 — Welche Daten pro Marker?

```
Phase 1 (mit Backend wie 2.1):
  - Name der Stadt
  - Quelle (Open-Meteo / DWD)
  - Aktuelle Temperatur
  - Niederschlag jetzt
  - Wind: Geschwindigkeit + Richtung
  - Link zur Detail-Page

Spätere Erweiterungen:
  - Forecast-Sparkline (24h, mini-Chart)
  - Alarm-Indikator (falls Warnung)
  - Aktualität (vor wie viel Min)
```

Bauchgefühl: Phase-1-Set ist genug für 2.3. Sparklines kommen mit
eigener Iteration (Chart-Komponente).

### Q4 — Wind-Richtung visuell?

```
Option W1 — Nur als Zahl im Popup (z.B. "275° West")
  Einfach, semantisch klar
  Wenig visuell

Option W2 — Pfeil-Icon im Marker, Drehung nach
            Wind-Richtung
  Visuell sofort verständlich
  Mehr Code, kompass-Logik

Option W3 — Pfeil-Layer in MapLibre (eigene Layer pro
            Location mit rotation property)
  Native MapLibre-Lösung
  Skaliert mit Zoom, sieht professionell aus
  Komplexer
```

Bauchgefühl: **Option W2** als Phase-1-Lösung — SVG-Pfeil im Pin,
CSS-Rotation nach Wind-Direction. Phase 2 könnte zu W3 wechseln, wenn
die Karte größere Stations-Mengen zeigen soll.

### Q5 — SSR vs Client-Only?

MapLibre braucht Browser-DOM und WebGL, geht nicht im SvelteKit-SSR.
Optionen:

```
Option S1 — Karten-Komponente nur client-side mounten
  if (browser) { ... import + render ... }
  → SvelteKit-Standard für Browser-only Code

Option S2 — Eigene Route, nur CSR (no SSR)
  Modul-Level `export const ssr = false`
  → Karten-Page läuft komplett client-side

Option S3 — Statisches Fallback im SSR
  Server rendert leeren Container mit Loading-State,
  client mountet Karte
```

Bauchgefühl: **Option S1** als Standard — Karten-Komponente in
`onMount()` initialisieren, Server rendert leeren Wrapper. Sauber
und SvelteKit-konform.

### Q6 — Performance + Bundle-Size?

```
maplibre-gl ist ~600 KB minified, ~140 KB gzipped.
Das ist groß, aber unvermeidbar für GPU-vector-rendering.

Strategien:
  - Lazy-load auf der Karten-Route (dynamic import)
  - Code-splitting via SvelteKit-Routes
  - Karten-Komponente erst on user interaction laden
    (Klick auf „Karte zeigen"-Button)
```

Bauchgefühl: **Lazy-load via dynamic import** — die Karte ist nicht
auf jeder Page, also kein Default-Bundle. Acceptable user perception.

---

## Was an Recherche noch fehlt

- [ ] OpenFreeMap Public Instance Robustheit prüfen
      (Status-Page, Outage-Historie der letzten Monate)
- [ ] MapLibre-Version pinning-Pattern für SvelteKit
- [ ] Bundle-Size-Auswirkung in unserer Build-Konfiguration messen
- [ ] Map-Privacy: Mit DSGVO-Beauftragten (oder eRecht24) klären,
      ob OpenFreeMap-Public-Instance Datenschutz-Beratung braucht
      (vermutlich nicht, weil keine Cookies/Tracking)
- [ ] Custom Marker mit SVG-Icon: how-to in MapLibre v5.x
- [ ] Wie zeigt man Pfeil-Rotation in HTML-Marker-Variante?

---

## Skizze der Implementations-Schritte

1. Branch + Verifikation
2. maplibre-gl als Frontend-Dependency
3. StationsMap-Komponente, OpenFreeMap-Style
4. Marker-Komponente mit Pin + Temperatur + Wind-Pfeil
5. Popup-Komponente mit Detail-Anzeige
6. Backend-Endpoint-Anpassung (falls /api/v1/map-overview neu)
7. Integration in /wetter-Page (oder /karte-Route)
8. Performance-Test (Lighthouse, Bundle-Size)
9. Mobile-Responsive testen (Pinch-Zoom, Touch)
10. Tests + Smoke-Checks
11. Doku-Updates

---

## Querbezüge

**Cookie-Banner**: OpenFreeMap setzt keine Cookies — keine Erweiterung
nötig. Falls später MapTiler oder andere Tracker-Quellen genutzt
werden, muss der Cookie-Banner erweitert werden.

**Datenschutz-Page**: kurzer Hinweis ergänzen — „Diese Karte nutzt
OpenFreeMap und OpenStreetMap-Daten. Keine personenbezogenen Daten
werden an die Karten-Server übertragen außer der IP-Adresse beim
Tile-Request."

**Quellen-Attribution-Page**: neuer Block für OpenStreetMap und
OpenFreeMap.

---

## Refs

- B.2 (Wetterkarten — POSTPONED): `../feature1/feature-decisions.md`
- Track-2-Status: `STATUS.md`
- MapLibre GL JS: https://maplibre.org/
- OpenFreeMap: https://openfreemap.org/
- OpenStreetMap Tile Usage Policy:
  https://operations.osmfoundation.org/policies/tiles/
- Free basemap roundup (2024):
  https://medium.com/@go2garret/free-basemap-tiles-for-maplibre-18374fab60cb
