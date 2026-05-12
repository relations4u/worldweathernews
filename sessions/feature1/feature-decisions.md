# Feature-Phase — Decisions Log

Stand: 12. Mai 2026 (post-Iteration-2.1, v0.4.2 live)
Maintainer: Heinz W. Richter <hwr@relations4u.de>

Dieses Dokument hält atomar fest, **was entschieden** ist. Begründungen kurz.
Verwerfungs-Ausnahmen mit „SUPERSEDED am Datum" markieren — nichts löschen.
Roadmap und Vorgehensschritte stehen separat in `feature-roadmap.md`.

---

## A — Frontend, Inhalte, CMS

### A.1 — Wer pflegt Inhalte

[DECIDED 2026-05-07] Heinz plus 3 oder mehr weitere Autor:innen (Variante A2).
Erst-Bestand 1 Person, wächst über Zeit. Editorische Workflows ergeben sich.

### A.2 — CMS-Wahl

[DECIDED 2026-05-07] **Sveltia CMS** für Phase 1.
Begründung: gleicher Setup-Aufwand wie Decap, aber bessere UX für
nicht-technische Co-Autor:innen, Mobile-First-Editor, native i18n,
aktive Entwicklung. Migration zu Decap als Notausgang ist 1-Zeilen-Change
(kompatible config.yml), wird im Backlog als Fallback-Plan dokumentiert.

### A.3 — Repo-Topologie für Inhalte

[DECIDED 2026-05-07] **Hybrid: Markdown im Frontend-Repo, Bilder im
Object Storage.**

Begründung: pure Inline-Speicherung würde das Monorepo bei realistischem
Wachstum überfordern (~5000 Bilder × 200 KB = 1 GB Binary in Git über
12 Monate). Markdown allein bleibt sehr klein (1664 MD-Files ≈ 5 MB).

Architektur:

- Markdown in `apps/frontend/src/content/` (im Repo, normaler Git-Flow)
- Bilder in S3-kompatiblem Bucket `media.worldweathernews.com`
- Sveltia-Bild-Upload geht via Worker direkt in den Bucket
- Markdown referenziert Bilder über absolute URLs

Storage-Provider: TBD (siehe A.13)

Verworfene Alternativen:

- Eigenes Content-Repo (Submodule): doppelte Verwaltung ohne Gegenwert
  bei 4 Autor:innen. Migrationspfad bleibt offen für später.
- Git LFS: Sveltia-Support nicht primary, LFS-Kosten ungünstig
  skalierend, CDN-Auslieferung sowieso besser.

### A.4 — OAuth-Proxy-Hosting

[SUPERSEDED 2026-05-11] Ursprünglich [DECIDED 2026-05-07]:
~~**Cloudflare Worker** mit Sveltias offiziellem `sveltia-cms-auth`
Worker.~~

**Neuer Stand seit 2026-05-11 — self-hosted Container statt CF-Worker.**

Cloudflare-Worker-Abhängigkeit wurde im laufenden Betrieb als unerwünschte
zusätzliche Drittanbieter-Verflechtung eingestuft (DNS bei Cloudflare ist
ohnehin schon; ein weiterer kritischer Pfad dort hebt das Migrations-Risiko
unnötig). Die Auth läuft jetzt als Go-Service `apps/cms-auth/` im App-
Compose-Stack auf wwn-prod hinter Caddy unter `cms-auth.worldweathernews.com`.
Logik ist 1:1 vom CF-Worker portiert (Chi statt itty-router, sonst gleicher
postMessage-Handshake, ~170 LOC). Distroless-Image, HEALTHCHECK über
Binary-Subcommand. Bind auf `127.0.0.1:8090`, Caddy proxied unter
`cms-auth.worldweathernews.com`.

Live seit v0.0.4 auf wwn-prod (PR #58 für Service, PR #59 für Cleanup
des CF-Worker-Codes).

Migration-Doku in `docs/cms.md` → „Maintainer-Aufgaben für Erst-Aktivierung".
Zugehöriger CLAUDE.md-Eintrag unter „Beantwortete Entscheidungen ab 2026-05-11".

**Leitlinie als Konsequenz (siehe neuer Punkt A.19):** Self-hosting-
Prinzip — neue Services landen erst-mal als Container im App-Stack, nicht
als CF-Worker. Konkret betrifft das Iteration 1.3b (Image-Pipeline): wird
als `apps/cms-media-upload/` Go-Service geplant, nicht als CF-Worker.

### A.5 — Render-Strategie

[DECIDED 2026-05-07] **Build-time (statisch)** — Markdown wird beim
Frontend-Build zu HTML. Inhalts-Änderungen lösen via GitHub-Actions
neuen Container-Build aus, Deploy nach 3-5 Minuten live.
Begründung: höchste Performance, einfachste Architektur, keine
Sync-Logik zwischen Container und Repo nötig. Latenz von 3-5 Min für
Inhalts-Änderungen in Forschungs-Phase akzeptabel.

### A.6 — Markdown-Pipeline

[DECIDED 2026-05-07] **mdsvex** für alle Markdown-Inhalte.

Begründung: Live-Diagramme und interaktive Karten in Artikeln sind
gewünscht (Maintainer-Statement: "Live-Diagramme, interaktive Karten
in Artikeln - ja"). mdsvex erlaubt Inline-Components im Markdown, was
für Wetter-Daten-Snippets, Live-Charts und Map-Embeds essentiell ist.

Trade-off akzeptiert:

- Sveltia kann mdsvex-Components nicht im Live-Preview rendern
  (zeigt sie als Platzhalter-Tags)
- Co-Autor:innen müssen lernen, dass `<TemperatureChart slug="berlin" />`
  ein Component ist und keine HTML-Tag
- Workaround: Component-Referenz-Page in `docs/cms.md` mit allen
  verfügbaren Components dokumentieren

Implementation:

- `apps/frontend` mit `@sveltejs/adapter-node` plus `mdsvex`-Preprocessor
- Components-Library unter `apps/frontend/src/lib/content-components/`
- Phase 1: einige Basis-Components (TemperatureChart, MapEmbed, DataSourceCard)
- Erweiterung kontinuierlich

### A.7 — Erste Seiten-Liste (Phase 1.1, hardcoded)

[DECIDED 2026-05-07] Hardcoded-Pflicht-Seiten ohne CMS, gepflegt nur
über PR-Review.

**Pflicht-Seiten (deutsches Recht 2026):**

- `/impressum` — § 5 DDG (Diensteanbieter-Pflicht)
- `/datenschutz` — DSGVO Art. 13/14 + BDSG
- `/barrierefreiheit` — BFSG (verpflichtend ab 28.06.2025 für
  kommerzielle Online-Angebote; Forschungs-Phase technisch nicht
  zwingend, aber als „state of the art" sauber)

**Pflicht-Komponente (in jedem Layout):**

- Cookie-Banner — TTDSG § 25, granulare Consent-Wahl, Ablehnen so
  einfach wie Annehmen, keine Vorab-Häkchen (BGH I ZR 7/16)
- Cookie-Settings-Link permanent im Footer

**Wichtige Pages (Phase 1.1 hardcoded oder MD via 1.2):**

- `/` Startseite
- `/about` Über uns
- `/kontakt` (Mailto-Link, kein Formular)
- `/quellen-attribution` Datenquellen-Lizenzen ehrlich angeben

**Optional, später:**

- `/agb` falls API-Nutzungsbedingungen
- `/api-terms` bei öffentlicher API

Begründung: rechtlich verbindliche Seiten brauchen PR-Review, kein
Browser-Edit. Editor-Risiko vermeidbar.

### A.8 — Erste editierbare Seite (Phase 1.2, MD via Sveltia)

[DECIDED 2026-05-07] `/methodik` als erste Markdown-Seite, gepflegt
über Sveltia. Beweisstück, dass die Pipeline funktioniert.
Spätere Seiten: `/datenquellen`, `/team`, weitere Methodik-Unterseiten.

### A.9 — Sprachen

[DECIDED 2026-05-07] **Deutsch und Englisch parallel** ab Phase 1.

Begründung: Maintainer ist Englisch-fähig, internationaler Reichweite
für Forschungs-Plattform. Sveltia hat erstklassigen i18n-Support, also
guter Zeitpunkt um es richtig zu setupen.

Implementation:

- i18n-Library wird in eigenem Diskussions-Punkt entschieden (siehe A.17)
- URL-Struktur: `/de/...` und `/en/...` mit Default `/de/`
- Sveltia-Collections mit i18n-Multifile-Modus pro Locale
- Cloudflare CNAME für `en.worldweathernews.com` (optional, später)

Pflicht-Pages bilingual:

- /impressum + /en/legal-notice (oder ähnlich)
- /datenschutz + /en/privacy
- /barrierefreiheit + /en/accessibility
- /quellen-attribution + /en/source-attribution
- /about + /en/about

ToDo: deutsche und englische Pflicht-Texte für rechtliche Pages
parallel erstellen lassen.

### A.17 — i18n-Library für SvelteKit

[DECIDED 2026-05-07] **Paraglide.js** (von Inlang) mit
`@inlang/paraglide-sveltekit` Adapter.

Begründung:

- **Bundle-Size**: ~250 Bytes plus nur tatsächlich genutzte Messages,
  vs. ~12 KB svelte-i18n / ~20 KB sveltekit-i18n. Bei mobilen Nutzer:innen
  einer Wetter-Plattform messbarer Performance-Vorteil.
- **Type-Safety**: Compile-time-Generierung typisierter TS-Module verhindert
  „Tippfehler in Translation-Keys"-Klasse von Bugs. Kritisch bei 3-4
  Co-Autor:innen.
- **SvelteKit-native**: Offizieller Adapter handhabt Locale-Routing
  (`/de/methodik` ↔ `/en/methodology`), Sprach-Detection, Locale-Switcher.
- **Ökosystem**: Inlang Studio als Web-IDE für Co-Autor:innen optional.
- **Lizenz**: Apache 2.0, kompatibel mit AGPL-3.0.

Strategie für Messages-Verwaltung: **Strategie A** (zentrale
`messages/de.json` und `messages/en.json`). Co-Autor:innen finden alle
Texte an einem Ort, Standard-Pattern.

Trade-off akzeptiert:

- Jüngere Library (~3 Jahre), weniger Stack-Overflow-Fundus als
  svelte-i18n. In eurem DE/EN-LTR-Setup unkritisch.
- Compile-Step im Build-Prozess (automatisiert via Paraglide-Watcher
  im Dev-Server).
- Migration zu svelte-i18n falls nötig: Helper-Funktion plus Such-
  Ersetzen, ~1 Tag Arbeit.

**Verworfene Alternativen:**

- **svelte-i18n**: keine Type-Safety, kein Tree-Shaking, größerer Bundle
- **sveltekit-i18n**: zwischen den Welten — nicht so etabliert wie
  svelte-i18n, nicht so modern wie Paraglide

**Architektur-Klarstellung Sveltia vs. Paraglide:**

- **Paraglide** verwaltet UI-Strings (Buttons, Menüs, Cookie-Banner)
  in `messages/de.json` und `messages/en.json`
- **Sveltia** verwaltet Inhalts-Markdown (Methodik-Page, Blog-Artikel)
  in `src/content/pages/de/methodik.md` und `src/content/pages/en/...`
- Beide laufen parallel ohne Konflikt. Sveltia nutzt eigenen i18n-
  Multi-File-Modus pro Locale, Paraglide kümmert sich um den Rest.

### A.10 — Forschungs-Phase-Banner

[DECIDED 2026-05-07] Permanent oben (sticky), schließbar mit
localStorage. Inhalt: kurzer Text plus Link zu `/methodik` für Details.
Begründung: rechtlich-ethische Pflicht, Nutzer:innen über Status zu
informieren. Aber nicht aufdringlich (schließbar).

### A.11 — Sveltia Editorial Workflow

[DECIDED 2026-05-07] **Kein formaler Approval-Workflow** für Co-Autor:innen
in Phase 1.

Co-Autor:innen mit Schreibrechten committen direkt in `main` über Sveltia.
Maintainer-Review erfolgt nachträglich (post-publish). Bei groben Fehlern
Revert via Git oder Edit über Sveltia.

Begründung: Maintainer-Statement: "keine komplizierten Freigabe-
mechanismen jetzt für co Autoren". Schnellerer Workflow, weniger
Reibung beim Onboarding. Risiko ist begrenzt durch:

- Git-Versionskontrolle (Revert immer möglich)
- 3-5 Min Build-Latenz gibt kurzes Zeitfenster für Reaktion
- Bei rechtlich heiklen Pages: hardcoded mit PR-Pflicht (siehe A.7)

**KI-Agenten arbeiten anders** (Maintainer-Statement). Agenten haben
ein eigenes Workflow-Modell:

- Agenten-Output landet erst in Draft-Status
- Maintainer-Review vor Publishing
- Detail-Definition in Track 3 (siehe C.2)

Diese Trennung — Mensch direkt, Agent via Review — ist bewusst gewählt.

### A.12 — CMS-Collections für Phase 1

[DECIDED 2026-05-07] Zwei Collections:

- `pages` (statische Seiten, MD)
- `blog` (Blog-Artikel mit Slug-Schema YYYY-MM-DD-titel)
  Spätere Collections: `data-sources`, `team`, ggf. `weather-stations`.

### A.13 — Storage-Provider für Bilder und Media-Assets

[DECIDED 2026-05-07, UPDATED 2026-05-11] **Hetzner Object Storage** in
Falkenstein (FSN1), Auslieferung über **Caddy-Reverse-Proxy** auf wwn-prod.

Konfiguration:

- Bucket: `media-worldweathernews-prod`
- Region: Falkenstein (DSGVO-konform, deutsche Server)
- Bucket-Endpoint: `https://fsn1.your-objectstorage.com`
- Auslieferung: `https://media.worldweathernews.com` — **Reverse-Proxy
  über Caddy auf wwn-prod**, der den Host-Header auf den Hetzner-Bucket-
  Endpoint rewritet
- DNS-Verlauf: `media → home → gate.hw7.eu → Public-IP → Firewall-NAT → Caddy`
- Cloudflare-Proxy: **AUS** (graue Wolke, DNS-only)
- Pricing: **€6.49/Monat** Basispreis ab 1. April 2026 (vorher €4.99),
  inkludiert 1 TB Storage und 1 TB Egress.
- Mehrkosten: €6.49/Monat pro weitere TB Storage,
  €1.30/TB Egress über Free-Frame.

**Warum Caddy-Proxy statt direkter CNAME** (Lesson aus Iteration 1.1b,
2026-05-08):
Hetzner S3 routet **host-basiert**. Ein direkter CNAME mit Cloudflare-Proxy
liefert dem Bucket den Client-Host `media.worldweathernews.com`, was
Hetzner mit `400 BadRequest` quittiert. Caddy auf wwn-prod terminiert TLS,
forwardet GET/HEAD an den Bucket und setzt dabei den Host-Header auf
`media-worldweathernews-prod.fsn1.your-objectstorage.com` um. Read-Only-
Proxy — Schreibzugriffe (Sveltia-Upload) gehen direkt zum Bucket-Endpoint
über Pre-Signed-URLs, nicht durch Caddy.

Begründung Hetzner OS:

- S3-kompatibel — Sveltia kann direkt drauf zugreifen
- DSGVO-trivial (deutscher Anbieter, deutsche Server)
- Pay-as-you-go für Wachstum
- Passt zum Hetzner-Migrations-Pfad in Track 1
- Plus: löst auch B.3 (Storage für GRIB-Daten in Track 2) elegant
- Auch nach Preiserhöhung 1. April 2026 günstiger als US-Alternativen
  (CDN Interconnect Egress GCP wird zur gleichen Zeit verdoppelt)

Pricing-Kontext:
Hetzner hat ab 1. April 2026 alle Preise erhöht — Cloud +30-43%,
Object Storage +30%, Memory-Add-ons +575% wegen DRAM-Preisanstieg
(+171% YoY durch AI-Compute-Boom). Object Storage von €4.99 auf €6.49
ist relativ moderat im Vergleich zu Cloud-Servern (+30-43%).

**Backlog-Punkte:**

- CDN/Edge-Cache vor `media.worldweathernews.com` (aktuell: Caddy
  liefert direkt, kein Edge-Cache) — siehe `docs/backlog.md`
- Cloudflare-Workers-Subscription-Status im Account `hwr-06e` klären
  (Workers-Menüs ausgegraut, Idee einer CF-Worker-basierten Edge-Cache-
  Lösung pausiert)

Verworfene Alternativen:

- **MinIO-VM auf eigenem Proxmox**: macht Backup-Last bei uns,
  ohne Pricing-Vorteil bei realistischen Volumen. Bleibt als Plan B
  wenn Hetzner-Pricing weiter eskaliert.
- **Backblaze B2**: günstiger pro GB, aber US-Anbieter
  (DSGVO-Komplexität)
- **Hetzner Storage Box**: kein S3-API, Sveltia-Integration komplexer
- **GCP Cloud Storage**: 4-6x teurer für vergleichbare Leistung
- **Direkter CNAME ohne Proxy**: scheitert am Hetzner-Host-Routing (400)

### A.16 — Compute-Hosting-Strategie

[DECIDED 2026-05-07] **Self-hosted Proxmox bleibt für Phase 1 und 2.**

Hetzner Cloud bleibt als Migration-Plan im Repo dokumentiert
(Terraform-Stub `hetznercloud/hcloud ~> 1.48` aktiv halten).

**GCP wurde geprüft und ausdrücklich verworfen.** Begründung:

- **TimescaleDB-Inkompatibilität (Killer-Faktor)**: Cloud SQL und
  AlloyDB unterstützen TimescaleDB-Extension nicht. Eure Wetterdaten-
  Architektur hängt an TimescaleDB-Hypertables. Self-managed Postgres
  auf GCE oder TimescaleDB Cloud kosten €145-190/Monat extra.
- **Pricing 4-6x höher**: Realistische Phase-1-Last kommt auf
  €320-400/Monat (vs. €60-75/Monat self-hosted). Das sind +€3000-4000
  pro Jahr ohne klaren Mehrwert in Forschungs-Phase.
- **Egress-Pricing-Trend**: Google verdoppelt CDN Interconnect Rates
  ab Mai 2026 — Pricing-Risiko für skalierbare Plattform.
- **Vendor-Lock-in**: VPC Connectors, IAM-Komplexität, Cloud-spezifisches
  Logging und Monitoring würde die Architektur-Migration umkehren.
- **DSGVO**: GCP-Frankfurt ist möglich, aber Free-Tier-Compute nur in
  US-Regionen — kein Vorteil für unsere Daten.

**Mögliche Ausnahme später**: GCP **könnte** für KI-Agent-Workloads
in Track 3 attraktiv werden (Vertex AI / Gemini-Integration, Cloud-Run-
Pricing-Modell für sporadische Agent-Calls passt). Diese Entscheidung
wird beim LLM-Provider-Topic (C.3) neu bewertet — nicht jetzt.

### A.18 — Wann Migration auf dedicated Server (Trigger-Bedingungen)

[DECIDED 2026-05-07] **Wechsel auf dedicated Server (Hetzner AX-Serie
oder IONOS AE-Serie) wird ausgelöst durch einen der folgenden Trigger:**

1. **Hardware-Limit erreicht**: 32 GB RAM oder 500 GB SSD vom Proxmox-
   Host wird zu 80%+ ausgelastet (3 VMs zusammen)
2. **Hardware-Probleme**: physische Hardware-Defekte oder häufige
   Strom-/Internet-Ausfälle bei Maintainer
3. **Skalierungs-Need**: Track-2-Storage-Bedarf (GRIB, Radar) sprengt
   500 GB
4. **Service-Level-Need**: Plattform geht in echte Production
   (über Forschungs-Phase hinaus), SLA-Versprechen werden gemacht
5. **Maintainer-Verfügbarkeits-Konflikt**: weniger Zeit für Hardware-Ops

**Wenn Trigger eintritt — aktuell evaluierte Optionen:**

```
Hetzner AX42 — Entry-Level dedicated
  AMD Ryzen 5 7600, 64 GB RAM, 2x 512 GB NVMe
  €57.30/Monat (ab April 2026), Setup-Fee €71.40 einmalig
  Vergleichbar mit aktuellem Proxmox-Host, mehr RAM

Hetzner AX52 — Mittlere Klasse (Empfehlung)
  AMD Ryzen 7 7700, 64 GB RAM, 2x 1 TB NVMe
  ~€70-80/Monat
  Schneller als aktueller Proxmox + doppelter Storage

Hetzner SX64 — Storage-Server (für Track 2 wenn massiv)
  64 GB RAM, 4x 22 TB HDD + 2x 480 GB NVMe
  ~€140-180/Monat — overkill für Phase 1, lohnt nur bei TB-skaligen
  Wetter-Datasets

IONOS AE6-32 — IONOS Entry
  AMD EPYC 8024P, 32 GB RAM, 1 TB NVMe
  €47-55/Monat (Aktionspreise 24-Monats-Verträge)
  Etwas schwächer als AX42, aber etablierter Provider mit C5-Zertifikat

IONOS AE16-128 — IONOS Mittelklasse
  AMD EPYC, 16 cores, 128 GB RAM
  €128/Monat (24 Monate), danach €150/Monat
  Stark, aber teurer als Hetzner AX52
```

**Empfehlung wenn Trigger eintritt:**

- Standard-Wechsel auf **Hetzner AX52** für Compute (Tracks 1+3)
- Plus **SX64-Server** wenn Track 2 große Datenmengen erfordert
- IONOS als Alternative wenn Compliance-Zertifikate (C5/BSI)
  business-relevant werden — aktuell nicht der Fall

**Migration-Vorbereitung (heute schon):**

- Ansible-Playbooks aus Session 11 sind hardware-agnostisch
- Terraform-Provider `hetznercloud/hcloud` aktuell halten
- Snapshot-Strategie unabhängig von Proxmox dokumentieren
  (wenn Proxmox wegfällt, was ersetzt Snapshot-Workflow?)

**Re-Evaluation:** Trigger-Bedingungen alle 3 Monate prüfen.

### A.19 — Self-hosting-Prinzip für neue Services

[DECIDED 2026-05-11] **Neue Services landen erst-mal als Container im
App-Compose-Stack auf wwn-prod, nicht als Cloudflare-Worker oder anderer
Drittanbieter-Compute.**

Hintergrund:
DNS läuft bereits über Cloudflare, das ist ein unvermeidlicher kritischer
Pfad. Jeder zusätzliche kritische Pfad bei Cloudflare (Workers, Pages,
KV-Storage, R2) hebt das Migrations-Risiko unnötig und macht uns von einer
weiteren Komponente des gleichen Anbieters abhängig.

Konkreter Auslöser:
Iteration 1.3a wurde am 2026-05-08 mit Cloudflare Worker als OAuth-Proxy
für Sveltia released. Drei Tage später (2026-05-11) als self-hosted Go-
Service neu gebaut (A.4 SUPERSEDED). Lessons:

- 1:1-Port von Worker zu Go-Container ist machbar in ~170 LOC
- Self-hosted Variante fügt sich in bestehendes Compose-/Caddy-/Monitoring-
  Setup natürlich ein (keine separate Wrangler-Toolchain, kein extra
  Dashboard, keine eigene CI-Pipeline)
- Deploy-Workflow ist identisch zu Backend/Frontend/Pyworkers — eine
  Versionspflege weniger

Anwendung auf kommende Iterationen:

- **Iteration 1.3b (Image-Pipeline)**: als `apps/cms-media-upload/`
  Go-Service, nicht als CF-Worker. Pre-Signed-URL-Generierung +
  WebP-Konvertierung + EXIF-Strip im selben Service oder zweiter Container.
- **Track 3 (KI-Agenten)**: LLM-Provider-Wahl (C.3) ist davon NICHT
  betroffen — Cloud-LLM-APIs sind etwas anderes als Compute-Hosting.
  Der Agent-Container, der die APIs aufruft, läuft self-hosted.

Verworfene Ausnahmen:

- **CF-Worker-Edge-Cache für `media.worldweathernews.com`**: wurde
  als Backlog-Item dokumentiert, weil der Account-Subscription-Status
  unklar ist (siehe A.13). Wenn das geklärt wird, ist es eine reine
  Cache-Schicht — kein kritischer Pfad. Dafür wäre A.19 keine
  Ausschluss-Regel.
- **Cloudflare R2 statt Hetzner OS**: wurde explizit verworfen, weil
  Hetzner OS DSGVO-trivialer und billiger ist (siehe A.13). A.19
  bestätigt diese Wahl zusätzlich.

Die Regel gilt für **Compute**, nicht für reine **Edge-/Cache-Schichten**
oder **DNS**. Cloudflare bleibt DNS-Anbieter. Cloudflare-Cache vor einer
self-hosted Origin ist okay. CF-Worker, die unabhängig App-Logik halten,
sind nicht okay.

### A.14 — Cookie-Strategie und Banner

[DECIDED 2026-05-07] Cookie-Banner einbauen, auch wenn Phase 1 selbst
keine nicht-essenziellen Cookies setzt.

**Phase 1: Tracking-frei**

- Keine Analytics (kein Plausible, kein Matomo, kein Google Analytics)
- Keine Marketing-Cookies
- Keine externen Embeds, die ohne Consent laden
- Nur localStorage für UI-State (Banner-Schließen, Theme, Sprache)

**Banner trotzdem einbauen, weil:**

- Bei späterer Aktivierung (z. B. Plausible) ist Infrastruktur da
- Bei externen Embeds (Wetterkarten Drittanbieter) wird Consent nötig
- Anwaltlich vorsorgliche Empfehlung

**Anforderungen TTDSG § 25 (BGH I ZR 7/16):**

- Granulare Einwilligungs-Wahl
- Ablehnen genauso einfach wie Annehmen
- Keine Vorab-Häkchen
- Vorhandene Einstellung permanent über Footer-Link änderbar
- Cookie-Settings-Page als eigene Route

**Library-Optionen:**

- Eigenständige Implementierung (volle Kontrolle, ~150 Zeilen Svelte)
- `cookieconsent` von Orestbida (open-source, gut)
- Klaro (DSGVO-fokussiert, deutsch)

Empfehlung: eigenständige Implementierung — weniger JS, volle Kontrolle,
keine zusätzliche Dependency. Wenn später Tracking aktiviert wird,
Library-Wechsel möglich.

### A.15 — Quellen-Attribution-Page

[DECIDED 2026-05-07] Eigene `/quellen-attribution`-Seite mit allen
Daten-Lizenzen und Quellenangaben. Pflicht-Inhalt:

- DWD-Daten: „Datenbasis: Deutscher Wetterdienst, eigene Bearbeitung"
- Open-Meteo: CC-BY-4.0-Attribution
- EUMETSAT: Lizenz-Status pro genutztem Service
- Weitere je nach späteren Quellen

Plus: ähnlicher Hinweis im Footer mit Link zur Page.

### A.20 — OpenAPI 3.1 ohne `nullable`-Marker

[DECIDED 2026-05-12, aus Iteration 2.1] Optionale/nullable Felder in
OpenAPI-Schemas werden über **`required: false` plus Pointer-Typen
in Go** ausgedrückt — **nicht** über das OpenAPI-3.0-`nullable`-Keyword
und auch nicht über OpenAPI-3.1-`type: [string, "null"]`-Arrays.

Hintergrund (aus 2.1-Implementation):

- oapi-codegen v2.4.1 kennt OpenAPI-3.1 `type: [string, "null"]`-Arrays
  nicht und erzeugt fehlerhaften Go-Code
- redocly verbietet OpenAPI-3.0 `nullable: true` in 3.1-Specs
- Ausweg: optionale Felder bleiben `required: false`; oapi-codegen
  generiert dann automatisch `*float32`-/`*string`-/`*time.Time`-
  Pointer-Felder in den Go-Structs

Konsequenz für alle künftigen OpenAPI-Schema-Definitionen:

- **Nullable-Marker NICHT verwenden** — weder 3.0- noch 3.1-Syntax
- Optionalität über `required`-Listen ausdrücken
- Bei wirklich-nullable-Bedeutung im JSON: Custom-`MarshalJSON`/
  `UnmarshalJSON` in den Generated-Types oder Wrapper-Layer
- Re-Evaluation, sobald oapi-codegen OpenAPI-3.1 vollständig supportet
  (Tracking: github.com/oapi-codegen/oapi-codegen issues)

### A.21 — sqlc-Schema-Input via Pre-Processing aus goose-Migrations

[DECIDED 2026-05-12, aus Iteration 2.1] sqlc liest sein DB-Schema
nicht direkt aus den goose-Up-Migrations, sondern aus einer
generierten Datei `apps/backend/internal/storage/schema.sql`. Diese
wird von `scripts/build-sqlc-schema.py` aus den `+goose Up`-Sections
in `infra/migrations/` extrahiert und konkateniert.

Workflow:

1. Migration in `infra/migrations/NNN_*.sql` schreiben
2. `make sqlc-schema` (Pre-Processing-Skript)
3. `make sqlc-generate` (sqlc) erzeugt typsichere Go-Funktionen
4. `make gen-check` verifiziert, dass Generated-Files committed sind

Begründung gegen Alternativen:

- **sqlc direkt mit Migrations-Liste**: sqlc kann das technisch
  (config.yaml: `schema: ["migrations/*.sql"]`), aber goose-Down-Sections
  würden Schema-Inhalt verwirren — `+goose Up`/`Down`-Marker sind
  goose-spezifisch
- **Schema duplizieren** (manuell pflegen): wartungsanfällig
- **Test-DB-Spinner mit `pg_dump`**: zu langsam für jeden Codegen-Run

Generated-Files-Policy: `schema.sql` ist **committed** im Repo, nicht
gitignored. Drift wird per `make gen-check` in CI gefangen.

### A.22 — DB-Migrations als Pflicht-Deploy-Step

[DECIDED 2026-05-12, aus Iteration 2.1 Hotfix-Pattern] DB-Schema-
Migrations sind **integraler Teil des Ansible-Deploys**, nicht
optional/manuell. Konkret:

- Ansible-App-Rolle staged das `goose`-Binary in den postgres-
  Container vor `docker compose up`
- `docker exec` der Migration läuft mit `-u 0` (root), weil sticky-bit
  in `/tmp` sonst Cleanup-Failures bei docker-cp-staged Files
  verursacht (siehe „Häufige Fallen" in CLAUDE.md)
- Cleanup der staged Files nach erfolgreicher Migration

Pattern hat sich in 2.1 nicht-trivial entwickelt:

- v0.4.0-Deploy schlug fehl mit `relation "locations" does not exist`
  (Migration nicht ausgeführt) → Hotfix per `docker cp` + `docker exec`
- v0.4.1-Deploy schlug fehl beim Cleanup (sticky-bit) → erneuter Hotfix
  mit `-u 0` Flag
- v0.4.2-Deploy lief sauber durch — Pattern jetzt stabil

Konsequenz für 2.2 und alle Folge-Iterationen mit DB-Schema-Änderung:
**Migration automatisch beim Deploy**, kein manueller Schritt nötig.
Akzeptanzkriterium für alle Schema-ändernden Iterationen: Deploy auf
wwn-prod muss ohne manuelle Migrations-Schritte funktionieren.

---

## B — Wetterdaten-Import

### B.1 — Erste Datenquelle (Architektur-Härtetest)

[DECIDED 2026-05-11] **Open-Meteo zuerst, als Hello World für die
Worker → DB → API → Frontend-Pipeline.**

Begründung: REST/JSON-API, kein Auth, kein Aggregations-Vorprozessieren
nötig, einfachste Datenquelle zum Pattern-Aufbau. DWD als zweite Quelle
in Iteration 2.2 mit der erprobten Pipeline.

**Scope für Iteration 2.1:**

- **Locations**: drei Städte initial — Potsdam, Berlin, Hamburg.
  Potsdam als Maintainer-Lokal, Berlin als Hauptstadt-Referenz,
  Hamburg als zweite Klimazone (Küste). Erweiterung zu „Top-N
  deutsche Städte" kommt mit Locations-Suche (eigene spätere
  Iteration).
- **Variablen** (4): Temperatur, Niederschlag, Windgeschwindigkeit,
  Windrichtung. Genug für „typischer Wetter-Schnappschuss", erweiterbar
  ohne Schema-Migration (Open-Meteo liefert alle weiteren Variablen
  im selben Request).
- **Frequenzen** (2): `current` (Single-Datapoint „jetzt") plus
  `hourly` (24h Forecast). `daily` und Tagesaggregate kommen mit
  Klima-Features.
- **Historie**: KEINE in Iteration 2.1. Open-Meteo Archive-API
  (Era5-basiert) ist frei verfügbar und kommt mit der ersten
  Klima-Iteration. Forecast-Pattern erstmal sauber aufbauen.
- **Storage**: Postgres + TimescaleDB-Hypertables. Keine S3-/Blob-
  Storage-Komplexität. Open-Meteo-Daten sind nicht groß genug.
  Schema-Skizze:
  - `locations` (kleine reguläre Tabelle: id, name, lat, lon, slug)
  - `observations` (Hypertable, time-partitioned, station_id +
    variable + value + timestamp)
  - `forecasts` (Hypertable, time-partitioned, plus `run_at` column
    für Forecast-Generations)
- **Worker-Frequenz**: TBD im Implementations-Prompt (Vorschlag:
  current alle 10 Min, hourly-Forecast alle 60 Min)
- **Attribution**: „Daten von Open-Meteo.com, CC BY 4.0" auf
  jeder Page als Footer-Snippet, plus zentrale Detail-Page auf
  `/quellen-attribution` (siehe A.15 + B.4)

### B.2 — Wetterkarten-Strategie

[POSTPONED 2026-05-11] Verschoben in eine spätere 2.x-Konzept-Diskussion,
nachdem Iteration 2.1 (Open-Meteo) live ist und die Daten-Pipeline-
Architektur erprobt ist.

Ursprünglich [OPEN 2026-05-07] Selbst rendern (ICON-Daten + Cartopy →
PNG, hoher Aufwand, hohe Kontrolle) oder externe Services einbinden
(windy.com, Ventusky, DWD NinJo via iframe, niedriger Aufwand,
Lizenz-Themen). Drei Optionen werden bei B.2-Wiederaufnahme bewertet:

- K1: Selbst rendern (volle Kontrolle, hoher Initial-Aufwand)
- K2: Externe einbinden (sofort verfügbar, Lizenz-/Cookie-Themen)
- K3: Hybrid — eigene Stations-Visualisierungen, externe Modell-Karten
  als Outbound-Link (kein Embed)

### B.3 — Storage für große Datasets

[POSTPONED 2026-05-11] Für Iteration 2.1 nicht relevant — Open-Meteo-
Daten passen in Postgres+TimescaleDB. Wird mit Iteration 2.4
(Satellitenbilder) und 2.5 (Radar) wieder aufgenommen, wo echte
GB-Mengen an Binärdaten anfallen.

Ursprüngliche Optionen bleiben dokumentiert: Hetzner Storage Box,
MinIO-VM auf eigenem Proxmox, oder direkt-Nutzung des bereits
vorhandenen Hetzner-Object-Storage-Buckets (A.13).

### B.4 — Daten-Lizenzen

[DECIDED 2026-05-11] Lizenz-Status pro Quelle bestätigt, Attribution-
Pattern für Iteration 2.1 festgelegt:

- **Open-Meteo**: CC-BY-4.0. Attribution: „Daten von Open-Meteo.com,
  CC BY 4.0" auf jeder Page mit Open-Meteo-Daten als Footer-Snippet,
  plus Detail-Eintrag auf `/quellen-attribution`.
- **DWD**: GeoNutzV (funktional CC-BY-äquivalent für offene Geodaten).
  Wording: „Datenbasis: Deutscher Wetterdienst, eigene Bearbeitung".
  Wird in Iteration 2.2 (DWD) relevant.
- **EUMETSAT**: free für non-commercial, kommerziell lizenzpflichtig.
  Für Forschungs-Phase okay, aber Status wird bei Iteration 2.4
  (Satellitenbilder) erneut geprüft — nicht jetzt.

Attribution-Strategie:

- **Anzeige-Pflicht**: Footer-Link „Datenquellen" auf jeder Page
- **Inhalt-Pflicht**: Detail-Page `/quellen-attribution` mit allen
  Lizenz-Texten
- Konkrete Strings landen in `messages/de.json` und `messages/en.json`
  (Paraglide), nicht hardcoded — damit Updates an die Quellen-Liste
  ohne Code-Changes möglich sind.

### B.5 — Worker-Scheduling-Pattern

[DECIDED 2026-05-12, aus Iteration 2.1] **W1 — APScheduler im
Worker-Container, in-Memory-Job-State** ist Default-Pattern für
Phase 1 aller Datenquellen-Worker.

Konkret in Iteration 2.1 erprobt: APScheduler läuft im `pyworkers`-
Container, scheduled `fetch_current` alle 10 Min und `fetch_hourly`
alle 60 Min. Container-Restart vergisst Job-Run-History, was für
idempotente Fetcher (current/hourly mit overwrite-on-conflict) okay
ist — beim Restart läuft der nächste scheduled Run, kein Backfill-
Bedarf.

**Migration zu W3** (PostgresJobStore mit Persistent-State) ist
Backlog-Punkt und kommt, wenn:

- Jobs nicht-idempotent werden (z.B. Klima-Backfills mit historischer
  Reihenfolge-Abhängigkeit)
- Job-Run-History zu Observability/Audit-Zwecken nötig wird
- Multi-Replica-Setup mit Coordination-Bedarf entsteht

Aktuell keine dieser Bedingungen aktiv. W1 bleibt für 2.2 (DWD),
2.3, etc.

**Verworfene Alternativen** (siehe Vergleich in 2.1-Übergabe-Prompt):

- **W2 — Separater Cron-Container** (z.B. mcuadros/ofelia oder
  systemd-Timer): zweites Tool, zwei Stellen für Job-Definition.
- **W3 — APScheduler + PostgresJobStore**: jetzt overkill, später
  drop-in-Migration ohne Schema-Bruch.

### B.6 — Frontend-Position für Daten-Features

[DECIDED 2026-05-12, aus Iteration 2.1] **Eigene Route pro Feature**,
nicht Hero-Erweiterungen der Startseite.

Konkret in Iteration 2.1: Wetter-Cards landeten auf `/wetter` als
eigene SvelteKit-Route, nicht als Block auf `/`. Die Startseite bleibt
Compliance-/Mission-/Editorial-Fokussiert.

Begründung:

- Klare URL pro Feature (`/wetter`, später `/karte`, `/klima`, `/blog`)
- Separates Caching/SSR-Verhalten je Route ohne Side-Effects auf
  andere Pages
- Bessere Code-Organization: feature-spezifische Components,
  Loader, Tests in eigenem Route-Verzeichnis
- Vortrag-/Navigations-fähig: klare Sitebar-Einträge

Implementations-Detail aus 2.1: `/wetter` ist aktuell `ssr = false`,
weil das Frontend-Container das Backend nur über die Public-URL
kennt (`PUBLIC_API_BASE_URL`). SSR-Upgrade per Internal-API-Hostname
ist Backlog-Punkt — Architektur-Entscheidung als solche bleibt B.6
unberührt.

Verworfene Alternativen:

- **Sammel-Page mit allen Features als Sections**: schwer zu
  cachen, schwer zu skalieren, schwer zu navigieren.
- **Modal/Drawer-Overlays auf Startseite**: gleicher Nachteil plus
  schlechtes SEO.

---

## C — KI-Agenten-Netzwerk

### C.1 — Welche Agent-Rollen für Phase 1

[OPEN 2026-05-07] Aus 6 vorgeschlagenen Rollen maximal 3 priorisieren:

- Wetterlagen-Einordnung
- News-Aggregator
- Klima-Kontext-Annotator
- Citizen-Science-Moderation
- Quality Assurance (Cross-Source-Anomalien)
- Inhalts-Generator für statische Seiten
  Frage an Maintainer.

### C.2 — Mensch-in-the-Loop-Niveau für KI-Agenten

[DECIDED 2026-05-07] **Phase 1: Agent-Output immer als Draft**, manuelles
Review/Publish durch Maintainer oder Co-Autor:innen.

Begründung: Maintainer-Statement: "KI Agenten arbeiten ja anders in das
system" — bewusste Trennung von Mensch-Workflow (direkt-publish) und
Agent-Workflow (review-required). Schützt Plattform-Qualität in der
Lernphase.

Implementation-Skizze:

- Agent-Output landet als MD-Datei in Draft-Branch oder draft-Status-
  Frontmatter (`status: draft`)
- Build-Pipeline rendert nur `status: published`
- Sveltia-CMS zeigt Draft-Pages mit Hinweis
- Reviewer setzt Status auf `published` oder löscht/editiert
- Phase 2 ggf. selektiv automatisch je Agent-Confidence-Score

### C.3 — LLM-Anbieter-Strategie

[OPEN 2026-05-07] Cloud (Anthropic/OpenAI/Google), EU-Cloud (Mistral,
Aleph Alpha), Self-hosted (Ollama mit Llama/Mistral 7B), oder Mix?
Trade-offs: Qualität vs. DSGVO vs. Kosten vs. Latenz.

### C.4 — DSGVO-Strategie für Agent-Inputs

[OPEN 2026-05-07] Wenn User-Daten (z. B. eingesendete Beobachtungen) in
LLM-Prompts gehen, ist das datenschutzrechtlich heikel. Phase-1-Linie:
keine personenbezogenen User-Daten in Cloud-LLM-Prompts, ggf.
Anonymisierungs-Pipeline.

### C.5 — Budget-Rahmen für LLM-Calls

[OPEN 2026-05-07] €50/€200/€1000 pro Monat? Bestimmt LLM-Wahl und
Frequenz der Agent-Runs.

### C.6 — Lizenz für Agent-Code

[PROPOSED 2026-05-07] Bleibt AGPL-3.0 wie restliches Repo. Kompatible
Frameworks: LangChain (MIT), LlamaIndex (MIT), CrewAI (MIT), AutoGen (MIT).
ENTSCHEIDUNG NOCH AUSSTEHEND.

---

## Cross-Cutting

### X.1 — Feature-Session-Struktur

[OPEN 2026-05-07] Setup-Sessions waren 12 nummerierte Schritte mit
Plan-Files. Feature-Phase ist anders: User Stories, Backlog-getrieben.
Vorschlag: GitHub-Issues für einzelne Features, Milestones für
Iterationen. `sessions/` bleibt als Setup-Archiv.

### X.2 — Fortschritts-Dokumentation

[DECIDED 2026-05-07] Dieses File (`feature-decisions.md`) wird der
zentrale Decisions-Log für die Feature-Phase. Roadmap mit Vorgehens-
Schritten in `feature-roadmap.md`. Beide Files initial außerhalb des
Repos auf wwn-handover, später Move-Entscheidung pro File.

---

## Status-Legende

- `[OPEN]` — noch nicht diskutiert oder entschieden
- `[PROPOSED]` — Empfehlung von Claude, Maintainer-Bestätigung erbeten
- `[DECIDED]` — entschieden, Datum dokumentiert
- `[SUPERSEDED]` — durch neue Entscheidung ersetzt, alte bleibt sichtbar

## Changelog

- 2026-05-07 — Initiale Version, Sveltia-CMS-Wahl bestätigt (A.2),
  alle weiteren Punkte als `[PROPOSED]` oder `[OPEN]` zur weiteren
  Diskussion eingetragen.
- 2026-05-07 (Tranche 2) — Sechs Punkte auf `[DECIDED]` gesetzt:
  A.4 (Cloudflare Worker OAuth), A.5 (Build-time-Render),
  A.7 (Pflicht-Pages-Liste vollständig nach DDG/DSGVO/TTDSG/BFSG),
  A.8 (Methodik als erste MD-Page), A.10 (Forschungs-Banner),
  A.12 (Collections pages + blog).
  A.3 mit Hybrid-Lösung entschieden: Markdown im Repo, Bilder im
  Object Storage (löst Monorepo-Größen-Sorge).
  Drei neue Punkte hinzugefügt: A.13 (Storage-Provider, OPEN —
  Hetzner-Object-Storage-Pricing zu recherchieren), A.14 (Cookie-
  Banner-Strategie, DECIDED — tracking-frei in Phase 1, Banner
  trotzdem einbauen für TTDSG-Compliance), A.15 (Quellen-
  Attribution-Page, DECIDED).
- 2026-05-07 (Tranche 3) — Hosting-Strategie und Restpunkte Track 1
  entschieden (vor Tranche-4-Update):
  A.6 (Markdown-Pipeline) DECIDED auf **mdsvex** — Inline-Components
  nötig für Live-Diagramme und interaktive Karten in Artikeln.
  A.9 (Sprachen) DECIDED auf **DE+EN parallel** ab Phase 1.
  A.17 NEU (i18n-Library, OPEN): svelte-i18n vs. Paraglide vs.
  sveltekit-i18n.
  A.11 (Editorial Workflow) DECIDED: kein formaler Approval-Mechanismus
  für Co-Autor:innen, direkt-Commit, Git-Revert als Sicherheit.
  C.2 (Mensch-in-Loop für Agenten) DECIDED: Agent-Output immer als
  Draft, manuelles Review/Publish — bewusste Trennung von Mensch-
  Workflow (direkt) und Agent-Workflow (review).
  A.16 NEU: Self-hosted Proxmox bleibt für Phase 1+2. GCP **explizit
  verworfen** wegen TimescaleDB-Inkompatibilität, 4-6x höhere Kosten,
  Vendor-Lock-in. GCP-Ausnahme für Track 3 (Vertex AI / KI-Agenten)
  wird beim LLM-Topic neu bewertet.

- 2026-05-07 (Tranche 4) — Hetzner-/IONOS-Vergleich für dedicated
  Server vor finaler Storage-Entscheidung durchgeführt. Erkenntnis:
  Hetzner-Preiserhöhung 1. April 2026 macht Object Storage von €4.99
  auf €6.49 — bleibt aber günstigste sinnvolle Option. Dedicated
  Server bei Hetzner (AX42 €57.30, AX52 ~€75) sind nach Preiserhöhung
  ähnlich teuer wie eigener Proxmox-Strom-Aufwand bei doppelter
  Hardware. IONOS AE6-32 (€47-55) etwas günstiger, aber 24-Monats-
  Vertragspflicht und schwächere CPU.

  A.13 (Storage) bleibt **Hetzner Object Storage**, Pricing aktualisiert
  auf €6.49/Monat (post April 2026). Self-Hosting bleibt Phase-1.

  A.18 NEU (DECIDED): Trigger-Bedingungen für Compute-Migration auf
  dedicated Server dokumentiert. Fünf Trigger (Hardware-Limit,
  Hardware-Probleme, Storage-Bedarf, SLA-Need, Maintainer-Zeitkonflikt).
  Konkrete Optionen recherchiert: Hetzner AX42/AX52/SX64, IONOS
  AE6-32/AE16-128. Empfehlung wenn Trigger eintritt: AX52 standard,
  SX64 wenn Track-2-Storage-Bedarf groß. Re-Evaluation alle 3 Monate.

- 2026-05-07 (Tranche 5) — Track 1 abgeschlossen.
  A.17 (i18n-Library) DECIDED auf **Paraglide.js** (von Inlang) mit
  `@inlang/paraglide-sveltekit` Adapter. Begründung: Bundle-Size
  dramatisch besser, Type-Safety verhindert Co-Autor:innen-Bugs,
  SvelteKit-native Routing-Integration. Verworfen: svelte-i18n
  (kein Tree-Shaking) und sveltekit-i18n (zwischen den Welten).
  Strategy A für Message-Files (zentrale messages/de.json + en.json).
  Architektur-Klarstellung dokumentiert: Paraglide für UI-Strings,
  Sveltia für Inhalts-Markdown — beide parallel ohne Konflikt.

  **Track 1 Status: 18 von 18 A-Punkten entschieden.**

- 2026-05-08 bis 2026-05-11 — Implementation läuft, Lessons fließen
  in Decisions zurück:

  **A.13 (Storage) UPDATED**: Caddy-Reverse-Proxy mit Host-Rewrite
  statt direktem CNAME mit Cloudflare-Proxy. Hetzner S3 routet
  host-basiert, direkter CNAME bekam 400 BadRequest. Lesson aus
  Iteration 1.1b (PR #44, 2026-05-08).

  **A.4 (OAuth-Proxy) SUPERSEDED 2026-05-11**: Self-hosted Go-Service
  `apps/cms-auth/` statt Cloudflare Worker. 1:1-Port der Worker-Logik
  in ~170 LOC. Cloudflare-Worker-Quellcode aus Repo entfernt (PR #58
  für Service, #59 für Cleanup). Live in v0.0.4.

  **A.19 NEU (DECIDED 2026-05-11)**: Self-hosting-Prinzip als
  Architektur-Leitlinie. Neue Services landen als Container im
  App-Compose-Stack, nicht als CF-Worker. Konsequenz für Iteration
  1.3b: Image-Pipeline wird `apps/cms-media-upload/` Go-Service.

  **Iteration 1.2 fertig (PR #46)**: Paraglide-i18n live, DE/EN
  parallel funktional, mdsvex-Components-Library mit DataSourceCard.
  Folge-Bug: Standalone `apps/frontend/pnpm-lock.yaml` musste manuell
  resync werden (Docker-Build-Context-Falle, Fix in #32f571d).

  **Iteration 1.3a fertig (PR #47/#48)**: Sveltia-CMS unter /admin,
  Editorial-Workflow per `publish_mode: editorial_workflow`. Image-
  Upload bewusst aus (`media_folder: ""`) bis 1.3b.

  **Tailwind v3 → v4 Migration (PR #63)**: 1-Line-Dependabot-Bump
  wurde durch saubere Migration ersetzt. `@tailwindcss/vite` als
  Vite-Plugin, `autoprefixer` raus, Config in `@theme`-Block in
  `src/app.css`. shadcn-svelte HSL-Variablen via `--color-*` gebrückt.

  **Dependabot-Triage 2026-05-11 (PR #60, #62)**: 21 PRs auf 1
  reduziert. Toolchain-Major-Bumps (Go 1.25→1.26, Python 3.12→3.14,
  Node 22→26) geschlossen — verstoßen gegen Pin-Regel. Dependabot-
  Config gehärtet, dass sie nicht zurückkommen. cms-auth als vierter
  Docker-Watcher hinzugefügt. Frontend-Patches konsolidiert in #62.

  **security-scan.yml ❌ failing**: Trivy SARIF upload 403, pnpm
  audit + pip-audit Findings, govulncheck Exit 3 (False-Positives
  vermutet). Backlog-PR #64 dokumentiert die Triage. Workflow läuft,
  aber Alarme aktuell rauschend.

- 2026-05-11 (Tranche 7) — Track 2 Einstieg, drei B-Punkte aus
  Konzept-Session entschieden:

  **B.1 (Erste Datenquelle) DECIDED**: Open-Meteo zuerst (REST/JSON,
  kein Auth). Iteration-2.1-Scope festgelegt: drei Städte (Potsdam,
  Berlin, Hamburg), vier Variablen (Temperatur, Niederschlag,
  Windgeschwindigkeit, Windrichtung), zwei Frequenzen (current +
  hourly 24h), keine Historie in 2.1. Storage: Postgres+TimescaleDB-
  Hypertables, kein S3. Schema-Skizze für `locations`,
  `observations`, `forecasts` dokumentiert.

  **B.2 (Wetterkarten) POSTPONED**: bewusst verschoben in spätere
  2.x-Konzept-Diskussion, nach Iteration 2.1. Drei Optionen
  (K1 selbst rendern / K2 extern / K3 hybrid) bleiben als
  Diskussions-Material dokumentiert.

  **B.3 (Storage für Datasets) POSTPONED**: für 2.1 nicht relevant
  (TimescaleDB reicht). Wiederaufnahme mit 2.4 (Satellitenbilder)
  und 2.5 (Radar). Hetzner-Object-Storage aus A.13 ist verfügbar
  und wäre der Default, wenn dann relevant.

  **B.4 (Daten-Lizenzen) DECIDED**: Status pro Quelle bestätigt
  (Open-Meteo CC-BY-4.0, DWD GeoNutzV, EUMETSAT non-commercial).
  Attribution-Pattern: Footer-Link auf jeder Page erfüllt Anzeige-
  Pflicht, Detail-Page `/quellen-attribution` erfüllt Inhalt-Pflicht.
  Attribution-Strings in Paraglide-Messages, nicht hardcoded.
  EUMETSAT-Status wird erst bei Iteration 2.4 erneut geprüft.

  **Track 2 Status: 4 von 4 B-Punkten in Iteration-2.1-Scope geklärt
  (2 DECIDED, 2 POSTPONED). Implementation-Übergabe-Prompt
  `prompt-iteration-2-1.md` bereit für Claude Code.**

- 2026-05-12 (Tranche 8) — Iteration 2.1 (Open-Meteo Hello World)
  durch (v0.4.0/v0.4.1/v0.4.2 live). Fünf neue Architektur-/Pattern-
  Entscheidungen aus der Implementation:

  **B.5 (Worker-Scheduling) DECIDED**: W1 — APScheduler im
  Worker-Container, in-Memory-Job-State — als Default-Pattern für
  alle Datenquellen-Worker in Phase 1. W3 (PostgresJobStore) als
  Migration-Pfad, wenn Jobs nicht-idempotent werden oder
  Job-Run-History für Audit nötig.

  **B.6 (Frontend-Position) DECIDED**: Eigene Route pro Feature
  (`/wetter`, später `/karte`, `/klima`), nicht Hero-Erweiterungen
  der Startseite. Aktuell `ssr = false` weil
  `PUBLIC_API_BASE_URL` browser-orientiert ist — SSR-Upgrade per
  Internal-API-Hostname im Backlog, betrifft B.6 nicht.

  **A.20 (OpenAPI 3.1 ohne nullable) DECIDED**: oapi-codegen kennt
  3.1 `type: [..., "null"]`-Arrays nicht, redocly verbietet
  3.0 `nullable`. Lösung: optionale Felder bleiben `required: false`
  → oapi-codegen erzeugt Pointer-Felder in den Go-Structs.

  **A.21 (sqlc-Schema-Input) DECIDED**: `scripts/build-sqlc-schema.py`
  baut `apps/backend/internal/storage/schema.sql` aus den
  `+goose Up`-Sections in `infra/migrations/`. Generated-File ist
  committed, `make gen-check` verifiziert Drift.

  **A.22 (DB-Migrations als Deploy-Step) DECIDED**: Ansible-App-Rolle
  staged `goose`-Binary in postgres-Container, führt Migration vor
  `docker compose up` aus, mit `-u 0` für sticky-bit-Cleanup.
  Pattern stabil seit v0.4.2 (post-Hotfix-Tranche).

  **Tag-Schema-Korrektur**: Track 2 Iterations-Tags starten bei
  **v0.4.0**, nicht v0.1.0 — Track-1-Iterationen hatten v0.1.0,
  v0.2.0, v0.3.0 bereits belegt. Tag-Schema ist fortlaufend
  v0.X.Y über alle Tracks, keine Track-spezifische Re-
  Initialisierung. (Spätere Iterationen: 2.2→v0.5.0, 2.3→v0.6.0,
  etc.)

  **Track 1 Backlog-Punkt eröffnet**: SSR-Upgrade für `/wetter`
  via separater Internal-API-Hostname (in `apps/frontend/`
  Compose-Network-Auflösung). Code-Stelle:
  `apps/frontend/src/lib/api/client.ts` plus
  `apps/frontend/src/routes/wetter/+page.ts`.

  Track 2 Status: **4 von 4 B-Punkten für 2.1-Scope geklärt
  (B.1, B.4 DECIDED, B.2, B.3 POSTPONED), plus zwei post-hoc
  B-Punkte aus Implementation (B.5 Scheduling, B.6 Frontend-Position)**.
