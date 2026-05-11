# Feature-Phase — Roadmap und Vorgehens-Schritte

Stand: 7. Mai 2026
Maintainer: Heinz W. Richter <hwr@relations4u.de>

Dieses Dokument hält fest, **wie wir vorgehen**. Konkrete Schritte,
Reihenfolge, Akzeptanzkriterien. Entscheidungen separat in
`feature-decisions.md`.

---

## Übergeordnete Strategie

Drei parallele Tracks, sequenziell gestartet:

```
Track 1 — Frontend & CMS  →  Track 2 — Wetterdaten  →  Track 3 — KI-Agenten
   (Erst sichtbar)            (Dann inhaltlich)         (Dann intelligent)
```

Track 2 startet, sobald Track 1 Iteration 1.3 (Sveltia funktional) durch ist.
Track 3 startet, sobald Track 2 die erste Datenquelle (Open-Meteo) live hat.

**Begründung der Reihenfolge:**

- Frontend ohne Inhalte = leere Plattform, schnell deprimierend
- Wetterdaten ohne Frontend = unsichtbare API
- Agenten ohne Datenbasis = keine sinnvollen Outputs

---

## Track 1 — Frontend, statische Seiten, CMS

### Iteration 1.1 — Hardcoded-Skelett mit Compliance (2-3 Tage Arbeit)

**Ziel:** Live-Seite mit allen rechtlich nötigen Pages und Cookie-Banner,
ohne CMS.

**Schritte:**

1. SvelteKit-Routes für statische Seiten anlegen
   - `apps/frontend/src/routes/+page.svelte` (Startseite)
   - `apps/frontend/src/routes/impressum/+page.svelte`
   - `apps/frontend/src/routes/datenschutz/+page.svelte`
   - `apps/frontend/src/routes/barrierefreiheit/+page.svelte`
   - `apps/frontend/src/routes/about/+page.svelte`
   - `apps/frontend/src/routes/kontakt/+page.svelte`
   - `apps/frontend/src/routes/quellen-attribution/+page.svelte`
   - `apps/frontend/src/routes/cookie-einstellungen/+page.svelte`

2. Forschungs-Phase-Banner-Komponente
   - `apps/frontend/src/lib/components/ResearchBanner.svelte`
   - Sticky oben, schließbar via localStorage
   - Text: kurz, mit Link zu /methodik

3. Cookie-Banner-Komponente (TTDSG-konform)
   - `apps/frontend/src/lib/components/CookieBanner.svelte`
   - Granulare Wahl (Essenziell/Funktional/Analytics/Marketing)
   - „Ablehnen" so prominent wie „Akzeptieren"
   - Keine Vorab-Häkchen
   - Speichert Consent in localStorage mit Versions-Stamp
   - Re-Check bei Settings-Änderungen
   - Phase 1: Banner zeigt nur „Essenziell" als gesetzt

4. Cookie-Settings-Page
   - `/cookie-einstellungen` für nachträgliche Anpassung
   - Footer-Link permanent erreichbar
   - Liste aller Kategorien mit Erklärung

5. Layout-Komponenten
   - Header mit Logo und Navigation
   - Footer mit allen Pflicht-Links:
     Impressum, Datenschutz, Barrierefreiheit, Cookie-Einstellungen,
     Quellen-Attribution, Kontakt
   - Konsistente Tailwind-Styles

6. Inhalts-Drafts erstellen
   - **Impressum** gemäß § 5 DDG:
     Diensteanbieter (Heinz W. Richter, Anschrift), Kontakt,
     berufsrechtliche Angaben falls relevant
   - **Datenschutz** für Forschungs-Phase:
     Server-Logs (IP, User-Agent, Zeitstempel — Speicherdauer dokumentieren)
     Cookies (nur essenziell in Phase 1)
     Drittland-Transfers (Cloudflare USA-Hosting → SCC-Hinweis)
     Betroffenen-Rechte
     Verantwortliche Person mit Kontakt
   - **Barrierefreiheit** Erklärung gemäß BFSG-Standard
   - **Quellen-Attribution** mit aktueller Quellen-Liste
   - **About-Text** mit Mission und Maintainer-Vorstellung
   - **Startseite-Hero** mit klarem Wertversprechen

7. Tests und Smoke-Checks
   - Alle Routes erreichbar (build + serve lokal)
   - Mobile-Responsive
   - Lighthouse-Check (Accessibility, Performance, Best-Practices)
   - Cookie-Banner-Verhalten testen (Akzeptieren/Ablehnen/Re-Open)

8. Juristische Abnahme
   - Impressum + Datenschutz von Anwalt prüfen lassen
   - Erfahrungsgemäß 1-2 Iterationen
   - eRecht24-Generator als Ausgangsbasis legal nutzbar

**Akzeptanzkriterien:**

- [ ] Alle 8 Routes live unter https://research.worldweathernews.com
- [ ] Forschungs-Banner sichtbar, schließbar, schließt persistent
- [ ] Cookie-Banner TTDSG-konform (Ablehnen ≥ Akzeptieren prominent)
- [ ] Cookie-Settings-Page funktioniert
- [ ] Lighthouse Performance > 90, Accessibility > 95
- [ ] Mobile- und Desktop-Layout sauber
- [ ] Impressum und Datenschutz juristisch geprüft und freigegeben
- [ ] Quellen-Attribution-Liste enthält alle Datenquellen die später
      eingebunden werden sollen (vorausschauend)
- [ ] Beide Sprachen funktional: /de/_ und /en/_ Routes erreichbar
- [ ] Locale-Wechsler in Header funktioniert
- [ ] Default-Redirect von `/` auf `/de/` (oder Sprach-Detection via
      Browser-Header)

**Deployment:** über bestehende `release.yml` als v0.1.0
(erstes Feature-Release, nicht mehr Setup-Versions-Schema)

**Offene Punkte:**

- A.17 i18n-Library entscheiden (vor Iteration 1.2)
- Banner-Text-Wording final
- Datenschutz-Detail-Konfiguration mit Anwalt

---

### Iteration 1.1b — Hetzner Object Storage einrichten (1 Tag, parallel zu 1.1)

**Ziel:** S3-kompatibler Bucket bei Hetzner bereit, bevor Sveltia-Bild-
Upload getestet wird.

**Vorbedingung:** A.13 entschieden (Hetzner Object Storage Falkenstein)

**Schritte:**

1. Hetzner Cloud Account vorbereiten
   - Account erstellen oder existierenden nutzen
   - Cloud Console → Object Storage aktivieren
   - Projekt anlegen oder bestehendes verwenden

2. Bucket erstellen
   - Region: Falkenstein (FSN1) — DSGVO-konform
   - Name: `media-worldweathernews-prod`
   - Erst privat anlegen, dann selektiv public read

3. S3-Credentials generieren
   - Cloud Console → Object Storage → S3 Credentials
   - Access Key und Secret Key sicher speichern
   - In SOPS-File `infra/secrets/production/media-storage.sops.env`:
     ```
     S3_ENDPOINT=https://fsn1.your-objectstorage.com
     S3_REGION=fsn1
     S3_ACCESS_KEY=...
     S3_SECRET_KEY=...
     S3_BUCKET=media-worldweathernews-prod
     ```

4. CORS-Konfiguration
   - Erlaubte Origins: `https://research.worldweathernews.com`,
     `https://worldweathernews.com`, `https://www.worldweathernews.com`
   - Methods: GET, HEAD, PUT (für Sveltia-Upload)
   - Headers: Content-Type, Authorization, x-amz-\*

5. Bucket-Struktur anlegen
   - `media/blog/<slug>/` für Blog-Bilder
   - `media/pages/<slug>/` für Page-Bilder
   - `media/team/` für Team-Fotos
   - `media/site/` für Logo, Favicons, OG-Bilder
   - Public-Read aktivieren mit IAM-Policy

6. CDN-Domain einrichten
   - Cloudflare DNS: CNAME `media.worldweathernews.com` →
     `media-worldweathernews-prod.fsn1.your-objectstorage.com`
   - HTTPS via Cloudflare Universal SSL (Proxy AN für diese Subdomain)
     ODER direkt vom Hetzner Endpoint mit eigener Subdomain-Config
   - Test: `media.worldweathernews.com/test.jpg` lädt nach Test-Upload

7. Test-Upload manuell
   - Mit `aws-cli` oder `s3cmd` ein Test-Bild uploaden
   - Über `media.worldweathernews.com/test.jpg` abrufen
   - HTTPS, CORS prüfen via Browser-DevTools

8. Sveltia-Worker-Anpassung
   - Sveltia-CMS unterstützt direkten S3-Upload via Pre-Signed-URLs
   - Cloudflare Worker für Pre-Signed-URL-Generierung
   - Worker-Secrets: S3-Credentials aus SOPS

**Akzeptanzkriterien:**

- [ ] Bucket erreichbar unter media.worldweathernews.com
- [ ] HTTPS funktioniert
- [ ] CORS für Frontend-Origin aktiv
- [ ] SOPS-Secrets eingerichtet und committed (verschlüsselt)
- [ ] Test-Upload via aws-cli erfolgreich
- [ ] Initial-Asset-Set hochgeladen (Logo, Favicon, OG-Default-Bild)

**Kosten-Erwartung:**

- Initial: €4.99/Monat (innerhalb 1 TB Storage, 1 TB Egress)
- Wachstum: linear, transparent

---

### Iteration 1.2 — mdsvex-Pipeline mit Paraglide-i18n (2 Tage)

**Ziel:** Methodik-Seite als mdsvex, gerendert zur Build-Zeit, mit
DE/EN-Versionen via Paraglide-i18n.

**Vorbedingung:** Iteration 1.1 fertig

**Schritte:**

1. mdsvex installieren und konfigurieren
   - `pnpm --filter frontend add -D mdsvex`
   - `svelte.config.js` Preprocessor erweitern
   - `extensions: ['.svelte', '.svx', '.md']`

2. Paraglide-i18n einrichten
   - `pnpm --filter frontend add @inlang/paraglide-sveltekit`
   - `npx @inlang/paraglide-js init`
   - Sprachen konfigurieren: `de` (Default), `en`
   - Paraglide-Plugin in `vite.config.ts` aktivieren
   - Locale-Routing via `paraglideMiddleware` in `hooks.server.ts`
   - `messages/de.json` und `messages/en.json` initial befüllen

3. Locale-Switcher-Component
   - `apps/frontend/src/lib/components/LocaleSwitcher.svelte`
   - Wechsel zwischen DE/EN
   - Speichert Präferenz im Cookie (essenziell, kein Banner-Zwang)
   - Im Header sichtbar

4. Content-Verzeichnis-Struktur (für Sveltia)
   - `apps/frontend/src/content/pages/de/methodik.md`
   - `apps/frontend/src/content/pages/en/methodology.md`
   - Frontmatter-Schema definieren:
     ```yaml
     ---
     title: "Unsere Methodik"
     slug: methodik
     lang: de
     translations:
       en: methodology
     lead: "Wie wir Wetterdaten zusammenführen"
     updated_at: 2026-05-08
     status: published
     ---
     ```

5. Dynamische Routes mit Paraglide-Routing
   - Paraglide-Routing handhabt `/de/...` und `/en/...` automatisch
   - `apps/frontend/src/routes/[slug]/+page.svelte`
   - `+page.ts` Loader mit lang-und-slug-Resolving:
     liest passendes MD aus `src/content/pages/<lang>/<slug>.md`
   - Locale-Switcher nutzt `goto` mit Sprach-Pfad-Replacement

6. Erste Content-Components für mdsvex
   - `apps/frontend/src/lib/content-components/`
   - `DataSourceCard.svelte` — Card-Anzeige für Datenquellen
   - `Callout.svelte` — Hinweis-Boxen (info, warning, note)
   - Plus: später Live-Diagramme, Map-Embed
   - Components in mdsvex registrieren via `globalScript`

7. Erste Methodik-Seiten schreiben (DE und EN)
   - DE: Wer steht hinter der Plattform, welche Datenquellen,
     was bedeutet Forschungs-Phase, Methodik-Versprechen
   - EN: Übersetzung mit gleicher Struktur
   - Beide nutzen mindestens einmal `<DataSourceCard>` als Beweis-
     stück für mdsvex-Components

8. SEO-Grundlagen mit i18n
   - Meta-Description aus Frontmatter
   - OpenGraph-Tags
   - hreflang-Tags für Sprach-Alternativen (Paraglide hilft)
   - sitemap.xml-Generator (alle Sprachen)

9. Components-Library-Doku
   - `docs/cms.md` erweitern um Component-Referenz
   - Wie nutze ich `<DataSourceCard>` etc. im MD
   - Beispiele zum Copy-Paste

**Akzeptanzkriterien:**

- [ ] `/de/methodik` zeigt MD-Inhalt korrekt gerendert
- [ ] `/en/methodology` zeigt englische Version
- [ ] Frontmatter wird zu Meta-Tags
- [ ] Build-Zeit-Render funktioniert (kein SSR-Hit pro Request)
- [ ] mdsvex-Components funktionieren in MD
- [ ] hreflang-Tags korrekt gesetzt
- [ ] Locale-Switcher funktioniert ohne Flicker
- [ ] Paraglide-Compile läuft im Dev-Server automatisch
- [ ] CI prüft, dass alle Translation-Keys in beiden Sprachen existieren

---

### Iteration 1.3 — Sveltia CMS einbinden (1 Tag)

**Ziel:** Co-Autor:innen können /methodik via Browser editieren.

**Schritte:**

1. GitHub OAuth-App anlegen
   - https://github.com/settings/applications/new
   - Application name: `worldweathernews-cms`
   - Homepage URL: `https://research.worldweathernews.com`
   - Authorization callback URL: später Cloudflare-Worker-URL
   - Client ID und Secret notieren (Secret in 1Password/Bitwarden)

2. Cloudflare Worker für OAuth-Proxy aufsetzen
   - Sveltias offiziellen `sveltia-cms-auth` clonen
   - Wrangler installieren, Worker deployen
   - Client-ID und Secret als Worker-Secrets setzen
   - Callback-URL in GitHub-OAuth-App eintragen
   - Test: Worker antwortet auf /auth-Endpoint

3. Sveltia-Loader im Frontend
   - `apps/frontend/static/admin/index.html` anlegen
   - Script-Tag für `@sveltia/cms` Module
   - Minimal HTML-Skelett

4. Sveltia-Konfiguration
   - `apps/frontend/static/admin/config.yml`
   - Backend: github-Provider mit Worker-URL
   - Collections: pages plus blog
   - Field-Definitionen pro Collection

5. Test-Edit-Cycle
   - Login als Maintainer
   - /methodik via Sveltia editieren
   - Commit erfolgt mit signiertem Author
   - Build-Pipeline läuft an
   - Live-Site zeigt Änderung

6. Dokumentation für Co-Autor:innen
   - `docs/cms.md` schreiben
   - Login-Anleitung, Markdown-Basics, Bild-Upload
   - Wie speichere ich, was passiert nach Commit
   - Wann ist meine Änderung live

**Akzeptanzkriterien:**

- [ ] /admin lädt Sveltia-UI
- [ ] GitHub-OAuth-Login funktioniert
- [ ] Edit von /methodik landet als Commit im Repo
- [ ] Live-Site zeigt Änderung nach Build (3-5 Min)
- [ ] CMS-Doku in docs/cms.md verfügbar
- [ ] Decap-Fallback dokumentiert (1-Zeilen-Swap im Worst Case)

---

### Iteration 1.4 — Blog-Collection und erste Artikel (laufend)

**Ziel:** Blog-System für regelmäßige Inhalts-Posts.

**Schritte:**

1. Blog-Routes
   - `apps/frontend/src/routes/blog/+page.svelte` (Listen-Page)
   - `apps/frontend/src/routes/blog/[slug]/+page.svelte` (Artikel)
   - Loader liest alle MD-Files aus `src/content/blog/`

2. Sveltia-Collection für Blog
   - Eigene Collection in config.yml
   - Slug-Schema: `{{year}}-{{month}}-{{day}}-{{slug}}`
   - Felder: title, date, author, tags, lead, cover, body

3. Listen-Seite
   - Sortiert nach Datum desc
   - Pagination ab 10+ Artikeln
   - Tag-Filter (Phase 1 simpel)

4. Artikel-Seite
   - Frontmatter-Header (Titel, Autor, Datum, Tags)
   - MD-Body
   - Schema.org Article-Markup für SEO
   - Social-Share-Links (kein Tracking)

5. RSS-Feed
   - `/blog/feed.xml` Endpoint
   - Build-Zeit generiert
   - Standard-Atom-Format

6. Erste echte Artikel
   - Launch-Artikel: „worldweathernews.com — was, warum, wer"
   - Methodik-Artikel: „Wie wir Wetterdaten zusammenführen"
   - Forschungs-Artikel: „Was diese Plattform anders macht"

**Akzeptanzkriterien:**

- [ ] /blog Listen-Page funktioniert
- [ ] /blog/[slug] zeigt Artikel
- [ ] Sveltia kann neue Artikel anlegen
- [ ] RSS-Feed valide (W3C-Validator)
- [ ] Mindestens 3 echte Artikel veröffentlicht

---

### Iteration 1.5 — Mehrere Editoren onboarden (sobald 1.4 stabil)

**Ziel:** 2-3 Co-Autor:innen können produktiv arbeiten.

**Schritte:**

1. GitHub-Repo-Zugriff
   - Co-Autor:innen als Members in `relations4u`-Org
   - Schreibrechte auf `worldweathernews`-Repo
   - Branch-Protection erlaubt PR-Workflow

2. Onboarding-Termin pro Person (30 Min)
   - GitHub-Account-Setup
   - Sveltia-Login-Flow zeigen
   - Erste Test-Änderung gemeinsam
   - docs/cms.md durchgehen

3. Editorial-Konvention
   - Style-Guide: Tonalität, Längen, Bilder
   - Wer macht Final-Review
   - Wie wird kollaboriert (PR-Review, oder Maintainer-Pass)

4. Feedback-Schleife
   - Nach 2 Wochen: was klappt, was nicht
   - Welche Sveltia-Bugs sind aufgetreten
   - Workflow-Anpassungen

**Akzeptanzkriterien:**

- [ ] Mindestens 2 Co-Autor:innen produktiv
- [ ] docs/cms.md ist nach erstem Onboarding aktualisiert
- [ ] Editorial-Konvention dokumentiert

---

## Track 2 — Wetterdaten-Import

### Iteration 2.1 — Open-Meteo als Hello World (3-4 Tage)

**Vorbedingung:** Track 1 Iteration 1.3 abgeschlossen
(Pipeline-Setup als Pattern bewiesen)

**Schritte:** TBD nach Track 2 Detail-Diskussion

**Offene Punkte (siehe feature-decisions.md):**

- B.1 Open-Meteo zuerst oder direkt DWD
- B.3 Storage-Strategie für große Datasets
- B.4 Lizenz-Bestätigung

---

### Iteration 2.2 — DWD Stations-Beobachtungen

**Schritte:** TBD

---

### Iteration 2.3 — DWD MOSMIX-Vorhersagen

**Schritte:** TBD

---

### Iteration 2.4 — Satellitenbilder

**Schritte:** TBD

---

### Iteration 2.5 — Radar (RADOLAN)

**Schritte:** TBD

---

### Iteration 2.6 — ICON-Modelldaten

**Schritte:** TBD

---

## Track 3 — KI-Agenten-Netzwerk

### Iteration 3.1 — Erstes Agent-Skelett

**Vorbedingung:** Track 2 Iteration 2.1 abgeschlossen
(erste echte Daten verfügbar)

**Schritte:** TBD nach Track 3 Detail-Diskussion

**Offene Punkte (siehe feature-decisions.md):**

- C.1 Welche 3 Agent-Rollen Phase 1
- C.2 Mensch-in-the-Loop-Niveau
- C.3 LLM-Anbieter
- C.4 DSGVO-Pipeline
- C.5 Budget-Rahmen

---

## Cross-Cutting Operations

### Aus Setup-Phase übernommen (siehe docs/backlog.md im Repo)

Diese Punkte aus dem Setup-Backlog laufen parallel zur Feature-Phase und
sollten nicht vergessen werden:

- Backend-/Pyworkers-Metriken-Scrape-Lücke (Prometheus sieht aktuell
  keine Application-Metriken)
- node-exporter für wwn-mon
- Automatische Postgres-Backups
- Webcam (10.100.100.213) ins IoT-VLAN umziehen
- DMARC `p=reject` nach 1-2 Wochen rua-Reports
- HSTS `includeSubDomains` aktivieren wenn alle Subdomains TLS haben
- cosign-verify im Deploy

---

## Status-Tracking

### Track 1

- [ ] Iteration 1.1 — Hardcoded-Skelett mit Compliance
- [ ] Iteration 1.1b — Object Storage einrichten (parallel möglich)
- [ ] Iteration 1.2 — Markdown-Pipeline
- [ ] Iteration 1.3 — Sveltia CMS einbinden
- [ ] Iteration 1.4 — Blog-Collection
- [ ] Iteration 1.5 — Editor-Onboarding

### Track 2

- [ ] Iteration 2.1 — Open-Meteo
- [ ] Iteration 2.2 — DWD Stations
- [ ] Iteration 2.3 — DWD MOSMIX
- [ ] Iteration 2.4 — Satellitenbilder
- [ ] Iteration 2.5 — Radar
- [ ] Iteration 2.6 — ICON-Modelle

### Track 3

- [ ] Iteration 3.1 — Erstes Agent-Skelett (TBD nach Diskussion)

---

## Changelog dieses Dokuments

- 2026-05-07 — Initiale Version. Track 1 detailliert (Iteration 1.1 bis 1.5),
  Track 2 und 3 als Skelett-Platzhalter mit offenen Entscheidungs-
  Verweisen auf feature-decisions.md.
- 2026-05-07 (Tranche 2) — Iteration 1.1 erweitert um vollständige
  Compliance-Pages (Barrierefreiheit, Cookie-Banner, Cookie-Settings-
  Page, Quellen-Attribution). Cookie-Banner-Anforderungen TTDSG-konform
  detailliert. Iteration 1.1b ergänzt für Object-Storage-Setup
  (kann parallel zu 1.1 laufen, Voraussetzung für Sveltia-Bild-Upload
  in 1.3).
- 2026-05-07 (Tranche 3) — Roadmap auf entschiedene Architektur
  umgestellt:
  - Iteration 1.1b komplett auf **Hetzner Object Storage Falkenstein**
    spezialisiert: konkrete Bucket-Namen, S3-Endpoint, CORS-Config,
    SOPS-Schema, Pricing-Erwartung dokumentiert
  - Iteration 1.2 erweitert auf **mdsvex + DE/EN parallel**: zwei Tage
    statt einem, Content-Components-Library eingeplant, hreflang/SEO
    für i18n, Frontmatter-Schema mit translations-Mapping
  - Iteration 1.1 Akzeptanzkriterien um i18n-Routes ergänzt
  - Offene Punkte verschoben: A.17 (i18n-Library) muss vor 1.2
    entschieden sein
- 2026-05-07 (Tranche 5) — Track 1 vollständig spezifiziert.
  Iteration 1.2 mit Paraglide-i18n als finaler Library aktualisiert
  (vorher generisch "i18n-Library"). Konkrete Setup-Schritte:
  `@inlang/paraglide-sveltekit` Adapter, `messages/de.json` und
  `messages/en.json` als zentrale Translation-Files (Strategy A),
  Paraglide-Routing für `/de/...` und `/en/...`. Locale-Switcher
  als eigene Component, CI-Check für Translation-Key-Konsistenz.
