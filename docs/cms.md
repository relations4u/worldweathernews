# CMS und Content-Authoring

Stand: 8. Mai 2026 (Iteration 1.2 — vor Sveltia)

worldweathernews.com nutzt eine Markdown-basierte Content-Pipeline mit
mdsvex-Preprocessor und paraglide-js v2 für Internationalisierung.
Sveltia als visueller Editor folgt in Iteration 1.3.

Aktuell werden Inhalte direkt im Repo gepflegt (Pull Request → Merge), bis
Sveltia eingerichtet ist.

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

Diese Pages bleiben in 1.2 manuell gepflegt; Sveltia ab 1.3 fokussiert auf
Markdown-Pages.

---

## Verweise

- [docs/architecture.md](architecture.md) — Plattform-Gesamtarchitektur
- [docs/development.md](development.md) — Lokale Dev-Umgebung
- [feature-decisions A.16](https://...) — Warum mdsvex statt Decap
- [feature-decisions A.17](https://...) — Warum Paraglide.js v2
- mdsvex-Doku: <https://mdsvex.pngwn.io/>
- Inlang Paraglide-Doku:
  <https://inlang.com/m/gerre34r/library-inlang-paraglideJs/sveltekit>
