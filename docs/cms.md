# CMS und Content-Authoring

Stand: 8. Mai 2026 (Iteration 1.3a — Sveltia text-only)

worldweathernews.com nutzt eine Markdown-basierte Content-Pipeline mit
mdsvex-Preprocessor und paraglide-js v2 für Internationalisierung.
Sveltia CMS unter `/admin/` ist seit Iteration 1.3a aktiv für Text-Edit;
Bilder folgen in 1.3b mit Pre-Signed-S3-Pipeline und WASM-Optimierung.

Inhalte können auf zwei Wegen gepflegt werden:

1. **Sveltia CMS** unter <https://worldweathernews.com/admin/> — visueller
   Editor mit Editorial-Workflow (Edits landen als Pull Request, nicht
   direkt auf main).
2. **Direkt im Repo** — pull request, merge wie üblich. Für komplexere
   Änderungen (neue Routes, neue Components) weiter der Standard-Weg.

---

## Verzeichnisstruktur

```
apps/frontend/
├── messages/
│   ├── de-de.json        # UI-Strings für Deutsch
│   └── en.json           # UI-Strings für Englisch
├── project.inlang/
│   └── settings.json     # Inlang-Projekt-Konfiguration
└── src/
    ├── content/
    │   └── pages/
    │       ├── de/<slug>.md   # Deutsche Version
    │       └── en/<slug>.md   # Englische Version (sofern bilingual)
    ├── lib/
    │   ├── content-components/   # mdsvex-Components (DataSourceCard, Callout)
    │   └── paraglide/            # generated, in .gitignore
    └── routes/
        ├── [slug]/+page.{svelte,ts}   # Dynamic Markdown-Renderer
        └── sitemap.xml/+server.ts     # Auto-generated Sitemap
```

---

## Eine neue Page schreiben

### 1. Markdown-Datei anlegen

Lege die DE-Version unter `src/content/pages/de/<slug>.md` an. Wenn die Page
auch auf Englisch erscheinen soll, dazu eine parallele `src/content/pages/en/<slug>.md`.

Der Slug ist der URL-Pfad (ohne führenden Slash). Beispiel: `methodik`,
`klimadaten`, `historische-rekorde`.

### 2. Frontmatter

Jede Markdown-Datei beginnt mit einem YAML-Frontmatter-Block:

```yaml
---
title: "Methodik"
slug: methodik
lang: de-de
lead: "Wie worldweathernews.com Wetterdaten zusammenführt und einordnet."
updated_at: 2026-05-08
status: published
---
```

| Feld         | Pflicht | Bedeutung                                      |
| ------------ | ------- | ---------------------------------------------- |
| `title`      | ja      | H1 der Page, wird auch im Browser-Tab gezeigt  |
| `slug`       | ja      | URL-Pfad-Segment, muss zum Dateinamen passen   |
| `lang`       | ja      | `de-de` oder `en`                              |
| `lead`       | nein    | Untertitel + meta-description + og:description |
| `updated_at` | nein    | ISO-Datum, wird im Footer der Page angezeigt   |
| `status`     | ja      | `draft` (nicht ausgespielt) oder `published`   |

Pages mit `status: draft` liefern HTTP 404 bis sie auf `published` gesetzt werden.

### 3. Inhalt schreiben

Direkt unter dem Frontmatter folgt der Markdown-Body. Standard-Markdown ist
unterstützt: Überschriften, Listen, Links, Hervorhebungen, Code-Blöcke.

Für interaktive oder gestylte Elemente (Live-Daten, Karten, Hinweis-Boxen)
nutzt man Svelte-Components — siehe **Components** weiter unten.

---

## Components in Markdown

mdsvex erlaubt Svelte-Components direkt im Markdown. Importiere sie über
einen `<script>`-Block am Anfang der Datei (nach dem Frontmatter):

```svelte
<script lang="ts">
	import DataSourceCard from '$lib/content-components/DataSourceCard.svelte';
	import Callout from '$lib/content-components/Callout.svelte';
</script>
```

**Wichtig:** Schreibe Component-Tags auf eine Zeile (Open- und Close-Tag mit
Inhalt dazwischen). Prettier zerlegt mehrzeilige Tags in .md-Dateien sonst
in unbrauchbare Fragmente. Deshalb ist `src/content/` in `.prettierignore`.

### `<DataSourceCard />`

Quellen-Box mit Status-Badge. Für die Methodik- und Quellen-Attribution-Pages.

```svelte
<DataSourceCard
	name="Open-Meteo"
	url="https://open-meteo.com"
	license="CC-BY-4.0 (historisch)"
	region="EU-basiert, global"
	status="planned">Beschreibung der Quelle.</DataSourceCard>
```

| Prop      | Pflicht | Werte                                             |
| --------- | ------- | ------------------------------------------------- |
| `name`    | ja      | Anzeigename der Quelle                            |
| `url`     | nein    | Externer Link (öffnet in neuem Tab, rel=noopener) |
| `license` | nein    | Lizenz-Kurztext                                   |
| `region`  | nein    | Geografische Abdeckung                            |
| `status`  | nein    | `active` (grün) oder `planned` (amber, default)   |
| Slot      | nein    | Frei formulierte Beschreibung                     |

Status-Labels werden i18n-aware gerendert (`aktiv`/`geplant` vs. `active`/`planned`).

### `<Callout />`

Hervorgehobene Hinweis-Box.

```svelte
<Callout variant="info" title="Forschungs-Phase">Inhalt der Box.</Callout>
```

| Prop      | Pflicht | Werte                                                            |
| --------- | ------- | ---------------------------------------------------------------- |
| `variant` | nein    | `info` (sky), `warning` (amber), `note` (slate). Default `info`. |
| `title`   | nein    | Eigener Titel; ohne wird Variant-Label genutzt                   |
| Slot      | ja      | Inhalt                                                           |

Variant-Labels sind i18n-aware (`Hinweis`/`Achtung`/`Notiz` vs. `Info`/`Warning`/`Note`).

### Eigene Components

Neue Content-Components landen unter `src/lib/content-components/`. Achte
darauf:

- **i18n** — UI-Strings über `$lib/paraglide/messages` ziehen, nicht hartkodieren
- **Tailwind** — Klassen auf einer Zeile schreiben, keine dynamischen Class-Strings
  bauen (Tailwind kann sonst nicht tree-shaken)
- **External links** — `rel="noopener noreferrer"` und `target="_blank"`
- **Snippets** — `children?: import('svelte').Snippet` für freien Slot-Inhalt
- **Accessibility** — semantische HTML-Elemente (`<aside>`, `<figure>`, …)

Component-Imports im Markdown sind explizit, kein Auto-Import — so erkennt der
Sveltia-Editor (ab 1.3) die genutzten Components am `<script>`-Block.

---

## Mehrsprachigkeit

worldweathernews.com läuft mit zwei Locales: `de-de` (Default, ohne
URL-Prefix) und `en` (mit `/en/`-Prefix).

### Eine neue Übersetzung anlegen

Lege parallel zur DE-Datei eine `src/content/pages/en/<slug>.md` an. Slug
identisch, Frontmatter-Felder gleich, aber:

- `lang: en`
- Übersetzte `title` und `lead`

### hreflang und Sitemap

Beides wird automatisch generiert:

- **hreflang-Tags** — `[slug]/+page.svelte` emittiert `<link rel="alternate" hreflang>`
  für jede Locale, in der die Page mit `status: published` existiert, plus
  `x-default = de-de`.
- **sitemap.xml** unter `/sitemap.xml` — listet alle bekannten Routes.
  Markdown-Pages bekommen pro Locale einen Entry mit Cross-Alternates.

Keine manuelle Sitemap-Pflege. Bei Änderungen an deLocalisierungs-Patterns
(`vite.config.ts → paraglideVitePlugin → strategy`) automatisch übernehmen.

### UI-Strings übersetzen

Statische UI-Strings (Header, Footer, Banner-Text, Component-Labels) leben
in `messages/de-de.json` und `messages/en.json`. Für jede neue Message:

1. Key in beiden Dateien hinzufügen
2. In Components nutzen: `import * as m from '$lib/paraglide/messages'` →
   `{m.dein_key()}`
3. Vite-Plugin kompiliert automatisch beim nächsten `pnpm build` oder
   `pnpm dev`. Manuell: `pnpm exec paraglide-js compile --project ./project.inlang --outdir ./src/lib/paraglide`

Inlang Sherlock (VSCode) hilft beim Pflegen — ist aber optional.

---

## Locale-Switcher

Ist im Layout-Header bereits eingebaut (`src/lib/components/LocaleSwitcher.svelte`),
ein einfaches `<select>`. Beim Switch wird `setLocale()` aus dem
paraglide-runtime aufgerufen, was die URL und das Cookie umschreibt und einen
Reload triggert.

Die Locale-Strategie ist `['url', 'cookie', 'baseLocale']` — Default-URL-Pattern
prefixt non-base-Locales (`/en/...`), de-de bleibt unprefix.

---

## Page mit Server-Logik

Wenn eine Page über reines Markdown hinausgeht (Datenladen, Form-Handling),
ist sie keine `[slug]`-Page mehr, sondern bekommt eine eigene Route unter
`src/routes/<pfad>/+page.svelte` (+ optional `+page.server.ts` oder `+page.ts`).

Diese Pages bleiben manuell gepflegt; Sveltia fokussiert auf
Markdown-Pages.

---

## Sveltia CMS (Iteration 1.3a)

Sveltia ist eine drop-in-kompatible Alternative zu Decap CMS. Die Konfiguration
liegt unter `apps/frontend/static/admin/config.yml`, der Loader unter
`apps/frontend/static/admin/index.html`. Beim Build kopiert SvelteKit den
gesamten `static/`-Baum unverändert nach `/admin/index.html` und
`/admin/config.yml`.

### Login-Flow

1. <https://worldweathernews.com/admin/> öffnen.
2. „Login with GitHub" klicken — Popup öffnet sich.
3. GitHub OAuth-Consent (einmalig pro Account).
4. Popup schließt, Sveltia zeigt zwei Collections: **Pages — Deutsch**
   und **Pages — English**.

Hinter den Kulissen:

- Popup ruft `https://wwn-cms-auth.<account>.workers.dev/auth?provider=github&site_id=worldweathernews.com&scope=repo,user` auf.
- Worker leitet auf GitHub um, GitHub kommt mit `code` zurück, Worker
  tauscht gegen Token, schickt per `postMessage` an Sveltia.
- Sveltia speichert das Token in IndexedDB und nutzt es direkt gegen die
  GitHub-API. Kein Backend-State.

Wenn der Worker noch nicht deployed ist (Platzhalter `PLACEHOLDER` in
`base_url`), bricht der Login mit „Backend not reachable" ab — das ist
beabsichtigt und kein Bug.

### Edit-Workflow

1. Page in einer der Collections auswählen oder „New Page" klicken.
2. Felder bearbeiten (Title, Slug, Lead, Status, Inhalt). Slug muss
   weiterhin zum Dateinamen passen — Sveltia nutzt `slug: '{{fields.slug}}'`
   im Pattern, übernimmt es also automatisch in den Pfad.
3. „Save" → Sveltia schreibt einen Branch `cms/<title-slug>` mit dem Edit
   und öffnet einen Pull-Request gegen main.
4. CI läuft (lint + test + build), Maintainer reviewt + merged.
5. Release-Pipeline + Deploy nach `scripts/deploy.sh` wie üblich.

Direct push to main ist nicht möglich — branch protection auf main blockt
es ohnehin, und `publish_mode: editorial_workflow` zwingt Sveltia in den
PR-Modus.

### Markdown-Body in Sveltia

Sveltia rendert den Body als Markdown mit Live-Preview. Der `<script>`-Block
am Anfang einer Page (für mdsvex-Imports) und Component-Tags wie
`<DataSourceCard>` oder `<Callout>` werden als raw HTML behandelt und
unverändert durchgereicht.

**Nicht aus dem Markdown-Body entfernen:**

- Den führenden `<script lang="ts">…</script>`-Block
- Component-Tags und ihre schließenden Tags
- Component-Props (auch wenn sie wie unbekannte HTML-Attribute aussehen)

**Sicher zu bearbeiten:**

- Prosa zwischen Component-Tags
- Überschriften, Listen, Links, Code-Blöcke
- Frontmatter-Felder über die Sveltia-Form

Wenn ein Component-Tag versehentlich entfernt wird: Edit verwerfen
(„Discard Changes") oder den entsprechenden Tag aus dem letzten Repo-Commit
zurückkopieren.

### Image-Upload — bewusst deaktiviert

In 1.3a ist `media_folder: ""` gesetzt. Versuche, Bilder per Drag-and-Drop
in den Markdown-Editor zu ziehen, schlagen fehl. Hintergrund: Bilder
landen sonst direkt im Git-Repo, was bei mehreren MB pro Bild den Clone
über Monate aufbläht und zudem keine WebP-Konvertierung oder responsive
Sizes erzeugt.

Iteration 1.3b führt eine Pre-Signed-URL-S3-Upload-Pipeline ein:
Sveltia → Cloudflare-Worker → Hetzner Object Storage → CDN-Edge. Bilder
werden dabei vom Worker per WASM-libvips zu WebP konvertiert, in vier
Größen ausgeliefert (320/640/1280/1920) und EXIF-stripped.

Bis dahin: Bild-bedürftige Pages weiter direkt im Repo pflegen, oder
Bilder als externe URL einbetten (`![alt](https://…)`).

### Pages anlegen

Eine neue Page über Sveltia:

1. „Pages — Deutsch" → „New Page" klicken.
2. Slug eingeben (`klimadaten` oder ähnlich).
3. Title, Lead, Updated-At, Status setzen — `status: draft` schaltet die
   Page erst nach „published" live.
4. Save → PR wird geöffnet → CI grün → mergen → next deploy live.

Bilingual: nach dem DE-Save dieselbe Slug-ID in „Pages — English" anlegen
mit übersetzten Texten. hreflang-Tags und sitemap.xml werden automatisch
ausgeliefert (siehe „Mehrsprachigkeit" oben).

### Maintainer-Aufgaben für Erst-Aktivierung

Sveltia ist erst voll funktional, wenn folgende Punkte erledigt sind. Sie
sind bewusst nicht im Code automatisiert, weil sie OAuth-Secrets und
Cloudflare-Account-Zugang erfordern:

1. **GitHub OAuth-App registrieren** (relations4u-Org oder persönlich).
   Callback-URL: `https://wwn-cms-auth.<cf-account>.workers.dev/callback`.
2. **Cloudflare-Worker deployen** — siehe
   [`infra/cloudflare-worker-cms-auth/README.md`](../infra/cloudflare-worker-cms-auth/README.md).
3. **Worker-URL eintragen** in `apps/frontend/static/admin/config.yml`
   unter `backend.base_url`. Commit, mergen, deployen.
4. **Login testen** auf <https://worldweathernews.com/admin/>.

### Decap-Fallback

Sveltia ist seit 2024 sehr aktiv gepflegt, aber falls das Projekt jemals
stagniert: Decap CMS ist API-kompatibel. Migration besteht aus genau
drei Schritten: Loader-Script-URL in `static/admin/index.html` von
`@sveltia/cms` auf `decap-cms` ändern, dieselbe `config.yml` weiter
nutzen, denselben OAuth-Worker weiter nutzen. Im `docs/backlog.md` als
Eventual-Backup vermerkt.

---

## Verweise

- [docs/architecture.md](architecture.md) — Plattform-Gesamtarchitektur
- [docs/development.md](development.md) — Lokale Dev-Umgebung
- [feature-decisions A.16](https://...) — Warum mdsvex statt Decap
- [feature-decisions A.17](https://...) — Warum Paraglide.js v2
- mdsvex-Doku: <https://mdsvex.pngwn.io/>
- Inlang Paraglide-Doku:
  <https://inlang.com/m/gerre34r/library-inlang-paraglideJs/sveltekit>
