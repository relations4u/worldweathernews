# Feature-Phase — Decisions Log

Stand: 7. Mai 2026
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

[DECIDED 2026-05-07, SUPERSEDED 2026-05-11] ~~**Cloudflare Worker** mit
Sveltias offiziellem `sveltia-cms-auth` Worker.~~

**Update 2026-05-11 — self-hosted Container statt CF-Worker.**
Cloudflare-Worker-Abhängigkeit wurde im laufenden Betrieb als unerwünschte
zusätzliche Drittanbieter-Verflechtung eingestuft (DNS bei Cloudflare ist
ohnehin schon; ein weiterer kritischer Pfad dort hebt das Migrations-Risiko
unnötig). Die Auth läuft jetzt als Go-Service `apps/cms-auth/` im App-
Compose-Stack auf wwn-prod hinter Caddy unter `cms-auth.worldweathernews.com`.
Logik ist 1:1 vom CF-Worker portiert (Chi statt itty-router, sonst gleicher
postMessage-Handshake). Migration-Doku in `docs/cms.md` →
„Maintainer-Aufgaben für Erst-Aktivierung". Zugehöriger CLAUDE.md-Eintrag
unter „Beantwortete Entscheidungen ab 2026-05-11".

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

[DECIDED 2026-05-07] **Hetzner Object Storage** in Falkenstein (FSN1).

Konfiguration:

- Bucket: `media-worldweathernews-prod`
- Region: Falkenstein (DSGVO-konform, deutsche Server)
- Auslieferung: `media.worldweathernews.com` (Cloudflare CNAME)
- Pricing: **€6.49/Monat** Basispreis ab 1. April 2026 (vorher €4.99),
  inkludiert 1 TB Storage und 1 TB Egress.
- Mehrkosten: €6.49/Monat pro weitere TB Storage,
  €1.30/TB Egress über Free-Frame.

Begründung:

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

Verworfene Alternativen:

- **MinIO-VM auf eigenem Proxmox**: macht Backup-Last bei uns,
  ohne Pricing-Vorteil bei realistischen Volumen. Bleibt als Plan B
  wenn Hetzner-Pricing weiter eskaliert.
- **Backblaze B2**: günstiger pro GB, aber US-Anbieter
  (DSGVO-Komplexität)
- **Hetzner Storage Box**: kein S3-API, Sveltia-Integration komplexer
- **GCP Cloud Storage**: 4-6x teurer für vergleichbare Leistung

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

---

## B — Wetterdaten-Import

### B.1 — Erste Datenquelle (Architektur-Härtetest)

[OPEN 2026-05-07] Open-Meteo als Hello World, oder direkt DWD?
Empfehlung: Open-Meteo zuerst (REST/JSON, kein Auth, simpel),
trainiert das Worker → DB → API → Frontend Pattern. Dann DWD als
zweite Quelle, schon mit erprobter Pipeline.

### B.2 — Wetterkarten-Strategie

[OPEN 2026-05-07] Selbst rendern (ICON-Daten + Cartopy → PNG, hoher
Aufwand, hohe Kontrolle) oder externe Services einbinden (windy.com,
Ventusky, DWD NinJo via iframe, niedriger Aufwand, Lizenz-Themen).
Frage an Maintainer.

### B.3 — Storage für große Datasets

[OPEN 2026-05-07] GRIB-Modelldaten und Radar passen nicht in Postgres
(zu groß). S3-kompatibel nötig: Hetzner Storage Box jetzt einrichten,
oder MinIO als VM auf demselben Proxmox-Host?

### B.4 — Daten-Lizenzen

[OPEN 2026-05-07] DWD: GeoNutzV CC-BY mit Quellenangabe — bestätigen.
EUMETSAT: free für nicht-kommerzielle Nutzung — Forschungs-Phase OK?
Open-Meteo: CC-BY-4.0 — bestätigen. Konkrete Attribution-Anforderungen
werden in `docs/attribution.md` dokumentiert.

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
