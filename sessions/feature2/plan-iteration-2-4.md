# Plan-Skizze — Iteration 2.4 — Satellitenbilder (EUMETSAT)

**Stand: 16. Mai 2026 — Plan-Skizze**

Dieses Dokument ist eine **frühe Skizze**, **nicht** der submission-ready
Übergabe-Prompt. Final wird `prompt-iteration-2-4.md` daraus
ausgearbeitet, nachdem die offenen Konzept-Fragen unten entschieden
sind. Die Konzept-Session (B.2/B.3/EUMETSAT, 16. Mai) ist die
Grundlage — siehe `sessions/feature1/feature-decisions.md` und
`STATUS.md` → Folge-Iterationen 2.4.

---

## Ziel

Satellitenbilder (Meteosat-RGB-Composites) als Layer auf der
bestehenden MapLibre-Karte (`/wetter`, aus 2.3) bzw. als dedizierte
Bild-Ansicht. Erste echte Satelliten-Visualisierung der Plattform.

**Tag: v0.7.0** (Roadmap: 2.4/2.5/2.6 nach Konzept-Session)
**Geschätzte Dauer: 3-4 Tage**

---

## Konzept-Session-Entscheidungen (fix, nicht neu aufrollen)

- **B.2 = K3 Hybrid**: EUMETSAT-Imagery selbst holen + servieren;
  Modell-Karten extern nur als Outbound-Link (kein Embed). K1
  (komplettes Selbst-Rendern) ist späterer Evolutionspfad (~2.6).
- **B.3 = A.13-Bucket**: vorhandenen Hetzner-OS-Bucket
  (`media-worldweathernews-prod`) wiederverwenden. < 1–2 GB
  rollierend. Kein neues Storage-Infra in 2.4.
- **EUMETSAT-Lizenz**: Meteosat-Bildprodukte kostenfrei/lizenzfrei,
  Attribution „© EUMETSAT". Account registriert; verschlüsselter
  Secret-Stub liegt unter `infra/secrets/production/eumetsat.env`
  (`WWN_PY_EUMETSAT_CONSUMER_KEY` / `_CONSUMER_SECRET`, Werte vom
  Maintainer).

---

## Was wir wissen — zwei EUMETSAT-Zugangspfade

### Pfad A — EUMETView (fertige RGB-Composites) — empfohlen für 2.4

```
Quelle:   eumetview.eumetsat.int (WMS + static-images)
          view.eumetsat.int (moderner Viewer/WMS)
Form:     FERTIG gerenderte RGB-Composites (PNG/JPEG), georeferenziert
Auth:     öffentlich (WMS-GetMap; Credentials ggf. nicht nötig — siehe Q4)
Frequenz: ~15 min (Meteosat-0°), „latest"-Dateinamen verfügbar
Reproj.:  WMS kann direkt in CRS=EPSG:3857 (WebMercator) liefern
          → MapLibre-Raster-Source ohne eigene Reprojektion
Aufwand:  niedrig — kein Satpy, keine Roh-Daten, kein pyresample
```

**Verdict:** ideal für K3/2.4. Ein pyworkers-Job holt periodisch das
WMS-GetMap (oder die latest-static-Images), legt es im A.13-Bucket ab,
das Frontend bindet es als MapLibre-Raster-Source ein. Reine
`httpx`-Last, keine schwere Geo-Dependency.

### Pfad B — Data Store + eumdac + Satpy (Roh-SEVIRI) — K1-Evolution

```
Quelle:   EUMETSAT Data Store, Produkt EO:EUM:DAT:MSG:HRSEVIRI (L1.5)
Client:   eumdac (offizieller Python-Client, Auth Consumer-Key/Secret)
Form:     ROH-Radianzen (12 Kanäle) → Composite selbst rendern
Render:   Satpy + pyresample + dask + numpy/xarray (+ ggf. PWLT-Decomp)
Aufwand:  hoch — schwerer Dependency-Stack, eigene Reprojektion
```

**Verdict:** das ist der **K1-Pfad** (volle Kontrolle, eigene
Composites/Channels). Bewusst **nicht** für 2.4 — kommt mit der
K1-Evolution (~2.6, zusammen mit ICON-Selbst-Rendern). Der
registrierte Account + Secret-Stub bleibt vorbereitet, schadet nicht.

---

## Architektur-Skizze (Pfad A)

```
EUMETView WMS / static-images (© EUMETSAT)
     │  apps/pyworkers/pyworkers/jobs/eumetsat.py  (APScheduler, ~15 min, W1)
     │  GetMap CRS=EPSG:3857, BBOX=Europa-Sektor, RGB-Produkt
     ▼
Hetzner Object Storage (A.13-Bucket, eigener Prefix z.B. sat/)
     │  rollierendes Fenster (N Frames), alte Frames löschen
     ▼
media.worldweathernews.com  (bestehender Caddy-Reverse-Proxy)
     ▼
Frontend  /wetter
     │  StationsMap.svelte: zusätzliche MapLibre raster-Source +
     │  raster-Layer (Opacity-Regler), Zeit-Slider über die Frames
     ▼
Attribution „© EUMETSAT" ergänzt Quellen-/Datenschutz-Seite
```

Kein Backend-/OpenAPI-/DB-Eingriff vorgesehen (Bilder liegen im
Bucket, Frontend lädt direkt über `media.`). Frame-Index (Liste der
verfügbaren Timestamps) entweder als kleines JSON im Bucket oder über
Bucket-Listing — siehe Q6.

---

## Offene Konzept-Fragen

### Q1 — Welches RGB-Produkt zuerst?

```
P1 — Natural Color RGB (Tag, intuitiv „wie aus dem All")
P2 — IR 10.8 (Tag+Nacht durchgehend, Wolkenoberkanten-Temp)
P3 — Beide (Natural Color Tag, IR Nacht — oder als Layer-Auswahl)
```

Bauchgefühl: **P3 light** — IR 10.8 als Default (24/7 verfügbar),
Natural Color als zweite wählbare Source. Schmaler Start: erst nur
IR 10.8, Natural Color als Folge-Layer.

### Q2 — Region / BBOX?

```
R1 — Europa-Sektor (konsistent mit den 6 DE-Stationen aus 2.2/2.3)
R2 — Full-Disk Meteosat-0° (ganze Erdscheibe)
R3 — Deutschland-eng (kleiner, schärfer, weniger Kontext)
```

Bauchgefühl: **R1 Europa-Sektor** — passt zur Stations-Karte, sinnvolle
Dateigröße, zoom-bar genug.

### Q3 — Frequenz + Retention?

```
F1 — 15-min-Pull, rollierendes 24-h-Fenster (~96 Frames)
F2 — stündlich, 48-h-Fenster
F3 — 15-min-Pull, nur „latest" (1 Frame, kein Loop)
```

Bauchgefühl: **F1** — 24-h-Loop ist der Mehrwert; bei R1/PNG bleibt
das < 1 GB (B.3-Annahme bestätigt sich).

### Q4 — EUMETView öffentlich oder Data-Store-Auth nötig?

EUMETView-WMS ist als öffentliche Image-Gallery dokumentiert. **Zu
verifizieren bei der Plan-Ausarbeitung:** ob das WMS-GetMap ohne
Credentials nutzbar ist. Falls ja → der `eumetsat.env`-Secret wird
für 2.4 **nicht** gebraucht (bleibt für Pfad B/K1 vorbereitet).
Falls nein → Consumer-Key/Secret in den pyworkers-Job (ENV aus SOPS).
Überschreibt nichts an der Konzept-Entscheidung, klärt nur den Auth-Weg.

### Q5 — Frontend-Darstellung?

```
D1 — Raster-Layer auf der bestehenden /wetter-MapLibre-Karte
     (Opacity-Slider, Zeit-Slider), Marker bleiben darüber
D2 — Eigene dedizierte Bild-Ansicht (eigener Abschnitt/Route)
D3 — Beides (Layer auf Karte + Vollbild-Loop separat)
```

Bauchgefühl: **D1** — nutzt den 2.3-Stack maximal, ein kohärentes
Karten-Erlebnis. B.6 (eigene Route pro Feature) spricht eher für eine
eigene `/satellit`-Route statt Überfrachtung von `/wetter` — das ist
die eigentliche Q5-Entscheidung (Layer-auf-/wetter vs. eigene Route).

### Q6 — Frame-Index?

```
I1 — pyworkers schreibt sat/index.json (Liste Timestamps→URL) in den Bucket
I2 — Frontend listet den Bucket-Prefix (CORS/Listing-Policy nötig)
```

Bauchgefühl: **I1** — explizites kleines JSON, kein Bucket-Listing,
deterministisch, kein CORS-Policy-Umbau.

### Q7 — SSR?

MapLibre ist bereits `ssr=false` auf `/wetter` (S1 aus 2.3). Eigene
`/satellit`-Route analog `ssr=false`. Konsistent mit B.6.

---

## Skizze der Implementations-Schritte (Pfad A)

1. Branch + Verifikation + Live-Check EUMETView-WMS (Q4 klären:
   GetMap mit/ohne Credentials, CRS=EPSG:3857 testen)
2. pyworkers-Job `jobs/eumetsat.py`: WMS-GetMap holen, in A.13-Bucket
   (Prefix `sat/`) ablegen, rollierendes Fenster, `sat/index.json`
   schreiben (W1-Scheduling, ~15 min, idempotent)
3. Bucket-Prefix + ggf. CORS für `media.` prüfen (read-only, wie
   bestehende Media-Assets — vermutlich kein Policy-Change)
4. Frontend: MapLibre raster-Source/-Layer in StationsMap **oder**
   neue `/satellit`-Route (Q5), Opacity- + Zeit-Slider
5. Attribution „© EUMETSAT" → `/quellen-attribution` (Karten-Sektion)
   - Datenschutz §5 (EUMETView als Bild-Host, analog OpenFreeMap-Eintrag)
6. Paraglide-Strings DE+EN (Layer-Name, Slider-Labels, Attribution)
7. Tests (Job-Parsing/Index-Logik), `make lint && make test`
8. Doku: `docs/data-sources.md` (EUMETSAT-Block aktiv),
   `docs/backlog.md` (K1/Satpy-Pfad als Folge), `CLAUDE.md`
   Externe-Datenquellen + Wo-finde-ich-was, `STATUS.md` 2.4-Done
9. Deploy v0.7.0, Smoke + Maintainer-Browser-Check

---

## Querbezüge

- **Datenschutz**: EUMETView ist ein weiterer Drittanbieter-Bild-Host
  (Client lädt Kacheln/Bilder → IP-Übertragung), analog zur
  OpenFreeMap-Behandlung in 2.3. **Aber**: wenn die Bilder über
  `media.worldweathernews.com` (eigener Caddy-Proxy) ausgeliefert
  werden statt direkt von EUMETSAT, entsteht **kein** neuer
  Drittanbieter-Client-Pfad — der Worker holt serverseitig, der
  Browser lädt nur vom eigenen Origin. Das ist A.19-konform und
  datenschutz-sauber (kein §5-Eintrag nötig). Diese Server-seitig-
  Holen-Architektur ist bewusst Pfad-A-Default.
- **Cookie-Banner**: unverändert (kein Drittanbieter-Client-Embed).
- **B.6**: Q5 (Layer auf `/wetter` vs. eigene `/satellit`-Route) ist
  die einzige echte Frontend-Architektur-Frage.

---

## Was an Recherche / Entscheidung noch fehlt

- [ ] Q1–Q7 vom Maintainer entscheiden (Bauchgefühle oben als Default)
- [ ] Q4 verifizieren: EUMETView-WMS GetMap ohne Credentials?
      CRS=EPSG:3857 + Europa-BBOX live testen
- [ ] Exakter WMS-Endpoint + Layer-Namen für IR 10.8 / Natural Color
- [ ] Frame-Größe real messen (PNG Europa-Sektor) → B.3-Annahme
      (< 1–2 GB rollierend) bestätigen
- [ ] `media.`-Bucket-Prefix-Konvention + ob CORS/Policy-Touch nötig
- [ ] Übergabe-Prompt `prompt-iteration-2-4.md` ausarbeiten

---

## Refs

- EUMETView Image Gallery / WMS: https://eumetview.eumetsat.int/static-images/
- EUMETSAT View (moderner Viewer/WMS): https://view.eumetsat.int/
- EUMETView User Guide: https://user.eumetsat.int/resources/user-guides/eumet-view-user-guide
- WMS User Guide (PDF): https://www-cdn.eumetsat.int/files/2020-04/pdf_pf_wms_ug.pdf
- EUMDAC (Data-Store-Client, Pfad B/K1): https://user.eumetsat.int/resources/user-guides/eumetsat-data-access-client-eumdac-guide
- HRSEVIRI L1.5 (Roh, Pfad B): https://navigator.eumetsat.int/product/EO:EUM:DAT:MSG:HRSEVIRI
- Konzept-Entscheidungen: `sessions/feature1/feature-decisions.md` (B.2/B.3/B.4)
- 2.3-Karten-Stack: `sessions/feature2/prompt-iteration-2-3.md`
