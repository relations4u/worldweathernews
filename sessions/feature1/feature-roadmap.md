# Feature-Phase — Roadmap und Vorgehens-Schritte

Stand: 12. Mai 2026 (post-Iteration-2.1, v0.4.2 live)
Maintainer: Heinz W. Richter <hwr@relations4u.de>

Dieses Dokument hält fest, **wie wir vorgehen**. Konkrete Schritte,
Reihenfolge, Akzeptanzkriterien. Entscheidungen separat in
`feature-decisions.md`. Live-Status pro Iteration in `STATUS.md`.

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

**Aktueller Stand (Stand 12. Mai 2026):**

Track 1 ist bei Iteration 1.3a fertig — Sveltia läuft mit Text-Edit,
Image-Upload bewusst pausiert bis Iteration 1.3b (Image-Pipeline).
Track 2 Iteration 2.1 (Open-Meteo Hello World) ist v0.4.2 live —
drei Locations mit current + 24h-Forecast auf `/wetter`. Track 2
Iteration 2.2 (DWD-Adapter) ist als nächstes dran.

---

## Track 1 — Frontend, statische Seiten, CMS

### Iteration 1.1 — Hardcoded-Skelett mit Compliance (2-3 Tage Arbeit)

**Status: ✅ Done 2026-05-07 (PR #45, Squash 80fc6ec)**

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

**Status: ✅ Done 2026-05-08 (PR #44, Squash 55af7e3)**

**Ziel:** S3-kompatibler Bucket bei Hetzner bereit, bevor Sveltia-Bild-
Upload getestet wird.

**Vorbedingung:** A.13 entschieden (Hetzner Object Storage Falkenstein)

**Realisierte Schritte (mit Lessons aus der Implementation):**

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
     S3_PUBLIC_URL=https://media.worldweathernews.com
     ```

4. CORS-Konfiguration
   - Erlaubte Origins: `https://research.worldweathernews.com`,
     `https://worldweathernews.com`, `https://www.worldweathernews.com`,
     `http://localhost:5173`
   - Methods: GET, HEAD, PUT (für späteren Sveltia-Upload)
   - Headers: Content-Type, Authorization, x-amz-\*

5. Bucket-Struktur anlegen
   - `media/blog/<slug>/` für Blog-Bilder
   - `media/pages/<slug>/` für Page-Bilder
   - `media/team/` für Team-Fotos
   - `media/site/` für Logo, Favicons, OG-Bilder
   - Public-Read-Policy in `infra/object-storage/bucket-policy.json`

6. **Caddy-Reverse-Proxy** (Realisierung weicht vom ursprünglichen
   Plan ab, siehe Lesson unten):
   - Caddy-Block für `media.worldweathernews.com` in
     `infra/caddy/prod/Caddyfile`
   - Forward auf Bucket-Endpoint mit Host-Header-Rewrite
   - DNS via Cloudflare: CNAME `media → home.worldweathernews.com`,
     Proxy **AUS** (DNS-only, graue Wolke)
   - HTTPS via Caddy + Let's-Encrypt (HTTP-01-Challenge)
   - Deploy via `bash infra/deploy/deploy-caddy.sh` (macht `restart`,
     nicht `up -d`, wegen Bind-Mount-Inode-Falle)

7. Test-Upload
   - aws-cli auf wwn-dev oder Maintainer-Mac
   - Test-Datei in `s3://.../site/test.txt`
   - Abruf via `curl -I https://media.worldweathernews.com/site/test.txt`

8. Doku
   - `docs/media-storage.md` neu
   - `docs/runbook.md` erweitert um Szenario „media nicht erreichbar"
   - `CLAUDE.md` „Wo finde ich was"-Tabelle erweitert
   - `docs/backlog.md` ergänzt um CDN/Edge-Cache, Lifecycle-Policies,
     Bucket-Backup, Image-Optimierung

**Lesson aus Iteration 1.1b (relevant für Folge-Arbeit):**

Ursprünglicher Plan war direkter CNAME `media → bucket-endpoint` mit
Cloudflare-Proxy AN für TLS und CDN-Effekt. Das **funktioniert nicht**,
weil Hetzner S3 **host-basiert routet**: kommt der Request mit Host-Header
`media.worldweathernews.com`, antwortet der Bucket mit `400 BadRequest`,
weil der Bucket-Name nicht im Host steht.

Lösung: Caddy auf wwn-prod als Reverse-Proxy. Er terminiert TLS, schreibt
den Host-Header beim Forward auf `media-worldweathernews-prod.fsn1.your-objectstorage.com`
um, und liefert das Bucket-Object zurück. **Read-Only-Proxy** — Schreib-
zugriffe (Sveltia-Upload in 1.3b) gehen direkt zum Bucket-Endpoint über
Pre-Signed-URLs, nicht durch Caddy.

Parallel dazu wurde geprüft, ob ein **Cloudflare-Worker** als Edge-Cache
vor dem Caddy-Proxy sinnvoll wäre. Idee verworfen, weil:

- Cloudflare-Workers-Menüs im Account `hwr-06e` ausgegraut sind
  (Subscription-Status unklar)
- Hinzukommend: A.19 etabliert Self-hosting-Prinzip — weniger
  Cloudflare-Compute, nicht mehr

Beide Punkte (Edge-Cache + Subscription) liegen im Backlog.

**Akzeptanzkriterien (alle erfüllt):**

- [x] Bucket erreichbar unter media.worldweathernews.com
- [x] HTTPS funktioniert
- [x] CORS für Frontend-Origins aktiv
- [x] SOPS-Secrets eingerichtet und committed (verschlüsselt)
- [x] Test-Upload via aws-cli erfolgreich
- [x] Caddy-Block deployed, Let's-Encrypt-Cert ausgestellt

**Kosten-Erwartung:**

- Initial: €6.49/Monat (innerhalb 1 TB Storage, 1 TB Egress, post-April-2026)
- Wachstum: linear, transparent

---

### Iteration 1.2 — mdsvex-Pipeline mit Paraglide-i18n (2 Tage)

**Status: ✅ Done 2026-05-08 (PR #46 + Follow-Up #32f571d)**

**Ziel:** Methodik-Seite als mdsvex, gerendert zur Build-Zeit, mit
DE/EN-Versionen via Paraglide-i18n.

**Lesson (Follow-Up-Commit #32f571d):**

`apps/frontend/pnpm-lock.yaml` ist ein **standalone** Lockfile, den
der Docker-Build-Context nutzt (nicht der Workspace-Root-Lockfile).
pnpm regeneriert ihn standardmäßig **nicht**, weil der Workspace-Root
ihn überschattet. Workaround dokumentiert: `pnpm-workspace.yaml`
temporär verstecken, dann `pnpm install`, dann wieder zurück. Dieses
Pattern kommt bei jedem Frontend-Dependency-Update wieder, taucht
auch in der Dependabot-Triage am 2026-05-11 erneut auf (PR #62).

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

### Iteration 1.3a — Sveltia-Text-Edit + OAuth-Proxy (1 Tag → 3 Tage tatsächlich, in zwei Stufen)

**Status: ✅ Done 2026-05-08 (PRs #47/#48), Re-Build des OAuth-Proxys
2026-05-11 (PRs #58/#59, supersedes A.4 — siehe unten)**

**Ziel:** Co-Autor:innen können /methodik via Browser editieren.
Bild-Upload bewusst pausiert (kommt in 1.3b).

**Realisierte Schritte:**

1. GitHub OAuth-App anlegen
   - https://github.com/settings/applications/new
   - Application name: `worldweathernews-cms`
   - Homepage URL: `https://research.worldweathernews.com`
   - Authorization callback URL: zunächst auf CF-Worker, ab 2026-05-11
     auf self-hosted `cms-auth.worldweathernews.com/callback`
   - Client ID und Secret notieren (Secret in 1Password/Bitwarden)

2. OAuth-Proxy aufsetzen
   - **Initial 2026-05-08**: Cloudflare-Worker `wwn-cms-auth` im Account
     `hwr-06e`, Sveltias offizieller `sveltia-cms-auth` portiert
   - **Re-Build 2026-05-11**: self-hosted Go-Service `apps/cms-auth/`
     im App-Compose-Stack auf wwn-prod (siehe A.4 SUPERSEDED + A.19)
   - Chi-Router + slog + distroless-Image
   - Bind `127.0.0.1:8090`, Caddy proxied unter
     `cms-auth.worldweathernews.com`
   - ~170 LOC, 1:1-Logik vom CF-Worker

3. Sveltia-Loader im Frontend
   - `apps/frontend/static/admin/index.html` mit `@sveltia/cms` Script-Tag

4. Sveltia-Konfiguration
   - `apps/frontend/static/admin/config.yml`
   - Backend: github-Provider mit OAuth-Proxy-URL
   - Collections: pages plus blog
   - `publish_mode: editorial_workflow` (jeder Edit → PR)
   - `media_folder: ""` (Bild-Upload aus bis 1.3b)

5. Test-Edit-Cycle
   - Login als Maintainer am 2026-05-08, danach final am 2026-05-11
     mit self-hosted Proxy
   - /methodik via Sveltia editiert, PR erstellt, gemerged, live

6. Dokumentation
   - `docs/cms.md` mit Login-Anleitung, Markdown-Basics, Edit-Workflow
   - Decap-Fallback dokumentiert (1-Zeilen-Swap im config.yml)
   - „Maintainer-Aufgaben für Erst-Aktivierung" für den OAuth-Proxy

**Akzeptanzkriterien (alle erfüllt):**

- [x] /admin lädt Sveltia-UI
- [x] GitHub-OAuth-Login funktioniert (Cutover auf self-hosted geschafft)
- [x] Edit von /methodik landet als PR im Repo
- [x] Live-Site zeigt Änderung nach Build (3-5 Min)
- [x] CMS-Doku in docs/cms.md verfügbar
- [x] Decap-Fallback dokumentiert

**Lesson aus Re-Build (PR #58):** `gh pr merge` verweigerte teilweise
auf Workflow-Files ohne `workflow`-Scope auf dem PAT. Retry hat geklappt
(Race oder lazy-loaded Scope-Check). Falls wieder: `gh auth refresh -s workflow`
oder Web-UI-Merge.

---

### Iteration 1.3b — Image-Pipeline (Geplant, kein Termin)

**Status: ⏳ Geplant**

**Ziel:** Sveltia-Bild-Upload via Pre-Signed-URL S3 + WebP-Konvertierung

- responsive Sizes + EXIF-Strip.

**Vorbedingung:** Mindestens eine bildbedürftige Page in Sicht (Blog 1.4) —
keine Eile vor diesem Trigger. Aktuell sind alle Live-Pages text-only.

**Architektur-Entscheidung (A.19):** Image-Pipeline läuft als
self-hosted Go-Service `apps/cms-media-upload/`, **nicht** als
Cloudflare Worker. Analog zu cms-auth aus 1.3a.

**Schritte (Skizze, finale Spec bei Iteration-Start):**

1. Service-Skelett `apps/cms-media-upload/`
   - Go + Chi + slog wie cms-auth
   - Endpoint POST `/presign` mit Auth-Check
   - Endpoint POST `/process` für post-upload Transformations
   - Bind `127.0.0.1:8091`, Caddy proxied unter
     `cms-media.worldweathernews.com`

2. Pre-Signed-URL-Generierung
   - SOPS-Secret `media-storage.sops.env` aus 1.1b lesen
   - aws-sdk-go-v2 für PresignPostObject
   - URL-Ablauf: 5 Min

3. Image-Pipeline nach Upload
   - WebP-Konvertierung (libvips via govips oder bimg)
   - Responsive Sizes: 320 / 640 / 1280 / 1920 px Breite
   - EXIF-Strip (Privacy)
   - Original im Bucket archivieren

4. Sveltia-Config-Switch
   - `apps/frontend/static/admin/config.yml`
   - `media_folder` auf Bucket-Pfad
   - `media_library: { name: 'custom', config: { uploader: 'cms-media-upload' } }`

5. Doku
   - `docs/cms.md` Bild-Upload-Sektion
   - `docs/media-storage.md` Bucket-Layout für variants

**Akzeptanzkriterien (TBD bei Iteration-Start):**

- [ ] Co-Autor:innen können in Sveltia Bilder hochladen
- [ ] Upload landet direkt im Bucket (nicht über App-Server)
- [ ] WebP-Variant in 4 Größen vorhanden
- [ ] EXIF gestrippt
- [ ] Markdown referenziert die generierten Bildpfade

---

### Iteration 1.3 — Original-Spec (historisch, vor Split in 1.3a/1.3b)

**Status: ⏭ Skipped, ersetzt durch 1.3a + 1.3b**

(Behalten als Archiv-Eintrag — die ursprüngliche 1-Tag-Spec für „Sveltia
einbinden" war zu klein für die Realität: OAuth-Proxy musste neu gebaut
werden, Image-Upload braucht eigenen Service. Split in 1.3a/1.3b spiegelt
das wider.)

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

**Status: ✅ Done 2026-05-12 (PR #69, v0.4.0 → v0.4.1 → v0.4.2 live)**

**Was geliefert wurde:**

- DB-Schema: `locations`, `observations` (Hypertable), `forecasts`
  (Hypertable) mit 3 seeded Locations (Potsdam, Berlin, Hamburg)
- Python-Worker: APScheduler-basiert, current alle 10 Min,
  hourly Forecast alle 60 Min
- Backend-Endpoints: `/api/v1/locations` (List) und
  `/api/v1/locations/{slug}` (Detail mit current + 24h Forecast)
- Frontend: `/wetter`-Route mit drei WeatherCards
- Attribution: Open-Meteo-Block auf `/quellen-attribution`,
  Footer-Snippet auf `/wetter`

**Realisierte Entscheidungen (siehe Tranche 8 in
`feature-decisions.md`):**

- B.5 (Scheduling): W1 — APScheduler in-Memory im Worker-Container
- B.6 (Frontend-Position): eigene Route `/wetter`, aktuell CSR-only
- A.20 (OpenAPI): keine `nullable`-Marker, Pointer via `required: false`
- A.21 (sqlc-Schema): Pre-Processing aus goose-Migrations
- A.22 (Deploy): Ansible-staged goose-Migration als Pflicht-Step

**Lessons aus Post-Merge-Verlauf:**

- v0.4.0: erst-deploy failed weil Migration nicht im Deploy-Step
  → Hotfix in PR #70 (v0.4.1)
- v0.4.1: Cleanup failed wegen `/tmp` sticky-bit + postgres-User
  → Hotfix in PR #71 (v0.4.2) mit `docker exec -u 0`
- v0.4.2: stabil, drei Locations antworten, Berlin liefert frisches
  current (9°C @ 09:30Z am 12. Mai 2026)

**Backlog-Themen aus 2.1 (siehe `docs/backlog.md`):**

- W3 Persistent-Job-Store für APScheduler
- SSR-Upgrade für `/wetter` via Internal-API-Hostname
- Daily-Aggregate-Tabelle + Era5-Historie (Klima-Iteration)
- testcontainers-Postgres für Backend-Handler-Tests
- Lighthouse-CI für `/wetter`
- mdsvex-Konvertierung der hardcoded Compliance-Pages
- EN-Übersetzung von `/quellen-attribution`

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

(Detail-Status pro Iteration in `STATUS.md`)

### Track 1

- [x] Iteration 1.1 — Hardcoded-Skelett mit Compliance ✅ 2026-05-07 (PR #45)
- [x] Iteration 1.1b — Object Storage einrichten ✅ 2026-05-08 (PR #44)
- [x] Iteration 1.2 — Markdown-Pipeline + Paraglide ✅ 2026-05-08 (PR #46)
- [x] Iteration 1.3a — Sveltia + OAuth (self-hosted seit 2026-05-11) ✅ (PRs #47/#48/#58/#59)
- [ ] Iteration 1.3b — Image-Pipeline (geplant, wartet auf Blog-Bedarf)
- [ ] Iteration 1.4 — Blog-Collection
- [ ] Iteration 1.5 — Editor-Onboarding

### Track 2

- [x] Iteration 2.1 — Open-Meteo ✅ 2026-05-12 (v0.4.2, PR #69+#70+#71)
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

- 2026-05-11 (Tranche 6) — Implementation-Realität in Roadmap zurück-
  gespiegelt. Vier Iterationen als ✅ Done markiert (1.1, 1.1b, 1.2, 1.3a),
  Status-Tracking aktualisiert. Drei substantielle Spec-Änderungen:

  **Iteration 1.1b**: Caddy-Reverse-Proxy mit Host-Rewrite statt
  direktem CNAME (Hetzner S3 routet host-basiert, direkter CNAME → 400).

  **Iteration 1.3 → 1.3a + 1.3b Split**: Original-1-Tag-Spec war zu klein.
  1.3a: Sveltia-Text-Edit + OAuth-Proxy (self-hosted seit 2026-05-11
  als `apps/cms-auth/` Go-Service, supersedes A.4-CF-Worker-Plan).
  1.3b: Image-Pipeline als eigene Iteration, geplant als
  `apps/cms-media-upload/` Go-Service nach A.19 Self-hosting-Prinzip,
  wartet auf Blog-Bedarf in 1.4.

  **Iteration 1.2 Lesson**: `apps/frontend/pnpm-lock.yaml` ist standalone,
  Workspace-Lockfile überschattet ihn. Bei Dependency-Updates: temp-hide
  `pnpm-workspace.yaml`, dann `pnpm install`. Pattern dokumentiert, taucht
  bei jedem Frontend-Update wieder auf.
