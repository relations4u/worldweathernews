# Iteration 2.4 — Satellitenbilder (EUMETSAT)

**Übergabe-Prompt für Claude Code auf wwn-dev**

---

## Verwendung

Diesen Prompt **als ersten Prompt einer neuen Claude-Code-Session**
auf wwn-dev (10.100.100.113) verwenden, im Repo-Root von
`worldweathernews`. Voraussetzung: Iteration 2.3 ist als v0.6.0 live
(PR #76), Konzept-Session-2.4-Doku + Plan-Skizze sind gemerged
(PRs #79, #80).

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

Wir starten Iteration 2.4 (Satellitenbilder, EUMETSAT) auf
worldweathernews.com. Track 2 Iteration 2.3 (Stations-Map, MapLibre)
ist v0.6.0 live. 2.4 ergänzt Meteosat-RGB-Satellitenbilder als
zeitlich animierbaren Raster-Layer.

Lies bitte zuerst:

1. `CLAUDE.md` — zentrale Spielregeln. Besonders relevant:
   - „Häufige Fallen" → Standalone `apps/frontend/pnpm-lock.yaml`,
     Tailwind v4, Self-hosting-Prinzip (A.19)
   - „Externe Datenquellen", „Wo finde ich was"
2. `sessions/feature2/plan-iteration-2-4.md` — die Plan-Skizze
   (Pfad A/B, Architektur, Q1–Q7 mit Bauchgefühlen)
3. `sessions/feature1/feature-decisions.md` — **B.2** (K3 Hybrid),
   **B.3** (A.13-Bucket), **B.4** (EUMETSAT-Lizenz), **B.5** (W1),
   **B.6** (eigene Route pro Feature), **A.13** (Hetzner-OS-Bucket),
   **A.19** (Self-hosting)
4. `sessions/feature2/STATUS.md` + `docs/data-sources.md` +
   `docs/media-storage.md` (A.13-Bucket-Setup)
5. Die 2.3-Commits / `StationsMap.svelte` als MapLibre-Referenz

Sobald gelesen, melde dich mit Zusammenfassung von: Stack-Stand
(v0.6.0), Pfad-A-Architektur, und welcher Q4-Verifikationsschritt
zuerst nötig ist.

## Feature-Phase-Modus

Wie üblich: **Claude Code committet nach expliziter Freigabe.**

1. Branch `feat/iteration-2-4-satellite`
2. Implementation in mehreren Commits
3. **Vor jedem Commit: um Freigabe fragen**
4. „OK"/„commit"/„merge" → committen/mergen; „warte"/„anders" →
   nachbessern
5. Push/PR erst nach explizitem „push"/„PR aufmachen"

## Bereits entschieden — nicht neu aufrollen

Vom Maintainer vor der Session entschieden. Umsetzen, nicht
neu diskutieren:

### Architektur: Pfad A (EUMETView, fertige Composites)

- Quelle: **EUMETView** (`eumetview.eumetsat.int` / `view.eumetsat.int`),
  fertig gerenderte RGB-Composites via **WMS**, ~15 min, georeferenziert.
- **Kein Satpy / pyresample / Roh-SEVIRI.** Der Roh-Pfad
  (Data Store + eumdac + Satpy) ist der **K1-Evolutionspfad (~2.6)**,
  bewusst nicht 2.4.
- Server-seitig holen: pyworkers-Job zieht das WMS-Bild, legt es im
  **A.13-Bucket** (Hetzner OS, eigener Prefix `sat/`) ab; das Frontend
  lädt ausschließlich über das eigene `media.worldweathernews.com`
  (Caddy-Proxy). **Kein Drittanbieter-Client-Pfad** → A.19-konform,
  kein Datenschutz-§5-Eintrag (anders als OpenFreeMap).

### Q1–Q7 (Bauchgefühl-Defaults aus der Plan-Skizze, fixiert)

| Frage                   | Entscheidung                                                                                                                                                                                                                                                                  |
| ----------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Q1 — Produkt            | **IR 10.8 als Default** (24/7 verfügbar). Natural Color als zweite, wählbare Source — **schmaler Start: erst nur IR 10.8**, Natural Color als Folge-Layer in derselben Iteration nur wenn Zeit, sonst Backlog.                                                                |
| Q2 — Region             | **R1 Europa-Sektor** (konsistent mit den 6 DE-Stationen).                                                                                                                                                                                                                     |
| Q3 — Frequenz/Retention | **F1**: 15-min-Pull, rollierendes **24-h-Fenster** (~96 Frames), alte Frames löschen.                                                                                                                                                                                         |
| Q4 — Auth               | **In Schritt 1 verifizieren**: EUMETView-WMS-GetMap ohne Credentials nutzbar? Falls ja → `eumetsat.env`-Secret für 2.4 **nicht** gebraucht (bleibt für K1). Falls nein → Consumer-Key/Secret aus SOPS in den Job.                                                             |
| Q5 — Frontend           | **Eigene `/satellit`-Route** (nicht Layer auf `/wetter`). Begründung: **B.6 ist `[DECIDED]` „eigene Route pro Feature"** und überstimmt das weiche D1-Bauchgefühl. MapLibre-Raster-Layer + Opacity- + Zeit-Slider. _(Maintainer kann das im Prompt-Review kippen — dann D1.)_ |
| Q6 — Frame-Index        | **I1**: pyworkers schreibt `sat/index.json` (Timestamps→URL) in den Bucket. Kein Bucket-Listing, kein CORS-Policy-Umbau.                                                                                                                                                      |
| Q7 — SSR                | `/satellit` ist `ssr = false` (analog `/wetter`/S1, konsistent B.6).                                                                                                                                                                                                          |

### Datenbeschaffung

- **Kein Backend-/OpenAPI-/DB-Eingriff.** Bilder liegen im Bucket,
  Frontend liest `sat/index.json` + Frames direkt über `media.`.

## Flagged — vor Umsetzung vorschlagen, nicht still annehmen

- **S3-Client für pyworkers** (neue Dependency, CLAUDE.md-Regel):
  pyworkers schreibt bisher nur nach Postgres. Zum Bucket-Schreiben
  ist ein S3-kompatibler Client nötig (`boto3` / `aiobotocore` /
  `minio`). **Stack-Tabelle prüfen, Option mit Begründung
  vorschlagen, Maintainer fragen** — nicht einfach boto3 ziehen.
  Bucket-Credentials liegen in `infra/secrets/production/media-storage.env`
  (A.13). Alternative, die du mit abwägen sollst: Frames auf ein
  Shared-Volume schreiben, das Caddy direkt ausliefert (kein S3-Client)
  — Trade-off gegen die B.3-Entscheidung (A.13-Bucket) sauber benennen.
- **Exakter WMS-Endpoint + Layer-Name** für IR 10.8 / Europa-Sektor:
  in Schritt 1 live ermitteln (GetCapabilities), nicht raten.
- **CRS-Lieferung**: verifizieren, dass GetMap `EPSG:3857` direkt
  liefert (sonst `EPSG:4326` + MapLibre-Reprojektions-Fallback prüfen).

## Iterations-Plan

### Schritt 1 — Branch + Verifikation + EUMETView-Recherche

1. Branch `feat/iteration-2-4-satellite`
2. `uname -n` = wwn-dev, Repo-Root, `git config user.email` =
   `hwr@relations4u.de`
3. **EUMETView live prüfen** (kein Code): GetCapabilities abrufen,
   IR-10.8-Europa-Layer-Name, ob GetMap ohne Auth + `EPSG:3857`
   funktioniert (Q4). Ergebnis berichten.
4. Plan-Vorschlag Schritte 2–9 + S3-vs-Volume-Empfehlung, dann
   Freigabe abwarten.

### Schritt 2 — pyworkers-Job

`apps/pyworkers/pyworkers/jobs/eumetsat.py`:

- WMS-GetMap (IR 10.8, Europa-BBOX, `EPSG:3857`, PNG) via `httpx`
- Frame in Bucket-Prefix `sat/` ablegen (Dateiname mit
  UTC-Timestamp), rollierendes 24-h-Fenster, ältere löschen
- `sat/index.json` aktualisieren (sortierte Timestamp→URL-Liste)
- W1-Scheduling (APScheduler, ~15 min, idempotent — gleicher
  Timestamp überschreibt, kein Doppel-Frame)
- Storage-Zugriff gemäß Schritt-1-Entscheidung (S3-Client _oder_
  Volume)

### Schritt 3 — Storage/Serving verifizieren

- Frames + `index.json` über `https://media.worldweathernews.com/sat/…`
  erreichbar (read-only, wie bestehende Media-Assets)
- Prüfen ob CORS/Policy-Touch nötig (vermutlich nein —
  `infra/object-storage/{bucket-policy,cors}.json`)

### Schritt 4 — Frontend `/satellit`-Route

- Neue Route `apps/frontend/src/routes/satellit/` (`ssr = false`)
- MapLibre-Karte (OpenFreeMap-Base wie 2.3, zentrale
  `lib/config/map.ts` wiederverwenden) + Raster-Source aus den
  Bucket-Frames
- Opacity-Slider + Zeit-Slider (über die `index.json`-Frames),
  Play/Pause-Loop optional
- maplibre-gl lazy wie in `StationsMap.svelte` (Q6/S1-Muster aus 2.3)

### Schritt 5 — Navigation + Attribution

- Sitebar-/Nav-Eintrag „Satellit"
- Attribution **„© EUMETSAT"**: `/quellen-attribution` Karten-Sektion
  ergänzen. **Kein** Datenschutz-§5-Eintrag (server-seitig geholt,
  eigener Origin — explizit so begründen).

### Schritt 6 — Paraglide-Strings DE+EN

Layer-Name, Slider-Labels, Play/Pause, Attribution-Snippet.

### Schritt 7 — Tests

- pyworkers: Index-Logik (Fenster-Rotation, Sortierung,
  Idempotenz), WMS-URL-Bau — Fixture-basiert, kein Live-Call im Test
- Frontend: Slider-/Frame-Auswahl-Logik (Vitest)

### Schritt 8 — Doku

- `docs/data-sources.md`: EUMETSAT-Block aktiv (Pfad A)
- `docs/backlog.md`: K1/Satpy-Pfad als Folge (~2.6), Natural-Color-
  Layer falls in 2.4 nicht geschafft
- `CLAUDE.md`: „Externe Datenquellen" (EUMETSAT aktiv) +
  „Wo finde ich was" (eumetsat-Job, `/satellit`-Route)
- `sessions/feature2/STATUS.md`: 2.4-Done

### Schritt 9 — Deploy

`make lint && make test` grün, Tag **v0.7.0** (nach Maintainer-OK),
`scripts/deploy.sh production 0.7.0`, Smokes aus `docs/runbook.md`,
Maintainer-Browser-Check.

## Akzeptanzkriterien

- [ ] Branch `feat/iteration-2-4-satellite`
- [ ] EUMETView-Q4 in Schritt 1 verifiziert + berichtet
- [ ] S3-Client-vs-Volume in Schritt 1 vorgeschlagen + Maintainer-OK,
      neue Dependency (falls) gepinnt + begründet
- [ ] pyworkers-Job holt IR-10.8-Europa-Frame ~15 min, rollierendes
      24-h-Fenster, `sat/index.json` konsistent, idempotent (W1)
- [ ] Frames + index über `media.worldweathernews.com/sat/…` (read-only)
- [ ] `/satellit`-Route (`ssr=false`), MapLibre-Raster + Opacity- +
      Zeit-Slider, maplibre-gl lazy
- [ ] Nav-Eintrag, Attribution „© EUMETSAT" auf `/quellen-attribution`
- [ ] **Kein** Backend-/OpenAPI-/DB-Eingriff
- [ ] **Kein** Datenschutz-§5-Eintrag (server-seitig, begründet)
- [ ] Paraglide DE+EN, `make lint && make test` grün
- [ ] Doku + STATUS aktualisiert
- [ ] PR erst nach finalem Maintainer-OK; Tag **v0.7.0**

## Was du **noch nicht** baust

- **Roh-SEVIRI + Satpy** (eigene Composites) → K1-Pfad ~2.6
- **Natural Color RGB** als zweiter Layer → nur wenn Zeit übrig,
  sonst Backlog (IR 10.8 ist der 2.4-Scope)
- **Full-Disk** → Europa-Sektor genügt
- **Radar (2.5)** / **ICON (2.6)** → eigene Iterationen
- **Backend-Endpoint / `/api/v1/...` für Satelliten** → bewusst nicht,
  Bucket-direkt

## Wenn etwas unklar ist

Frag. Insbesondere:

- **WMS-Layer-Namen/CRS**: erst GetCapabilities, dann Mapping
  vorschlagen — nicht raten.
- **S3-Client-Wahl**: Stack-Tabelle prüfen, Optionen mit Begründung,
  Maintainer fragen. Volume-Alternative fair gegen B.3 abwägen.
- **Q5**: Default ist eigene `/satellit`-Route (B.6). Wenn der
  Maintainer Layer-auf-`/wetter` (D1) will, sagt er es im Review.
- **Bucket-Pfad/Retention-Kante**: Zeitzonen (UTC), Lückenframes
  (WMS liefert mal nichts) — robust handhaben, Vorschlag zeigen.

Bestätige kurz, dass du die Dokumente gelesen hast, und schlag
Schritt 1 vor.
