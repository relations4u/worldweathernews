# Iteration 1.1 — Hardcoded-Skelett mit Compliance

**Übergabe-Prompt für Claude Code auf wwn-dev**

---

## Verwendung

Diesen Prompt **als ersten Prompt einer neuen Claude-Code-Session**
auf wwn-dev (10.100.100.113) verwenden, im Repo-Root von
`worldweathernews`.

```
ssh hwr@10.100.100.113
cd ~/repos/worldweathernews
claude code
# → kompletten Inhalt unten als ersten Prompt einfügen
```

---

## Prompt für Claude Code (Copy-Paste ab hier)

---

Hallo Claude Code. Wir starten die Feature-Phase auf
worldweathernews.com. Die Setup-Phase (Sessions 1-12) ist abgeschlossen,
v0.0.2 läuft live auf wwn-prod. Lies bitte zuerst diese Dokumente in
dieser Reihenfolge:

1. `CLAUDE.md` im Repo-Root — die zentralen Spielregeln, Tech-Stack,
   Conventions
2. `STATUS.md` und `sessions/STATUS.md` — letzter Stand der Sessions
3. Externe Tracking-Dokumente, die ich dir gleich mitgebe

Sobald du diese gelesen hast, melde dich kurz mit einer Zusammenfassung
des Stack-Stands, damit ich sicher bin dass du den Kontext hast.

## Feature-Phase-Modus

**Wichtige Abweichung von der Setup-Phase:** In der Setup-Phase galt
"Maintainer committet selbst." In der Feature-Phase gilt **"Claude Code
committet nach expliziter Freigabe."**

Workflow:

1. Branch anlegen (`feat/iteration-1-1-skeleton`)
2. Implementation in mehreren Commits auf dem Branch
3. **Vor jedem Commit: mich um Freigabe fragen**
4. Bei "OK" oder "commit" oder "merge": committen oder mergen
5. Bei "warte" oder "nochmal" oder "anders": warten und nachbessern
6. Push zu GitHub: erst nach explizitem "push" oder "PR aufmachen"

Kein eigenständiger Commit ohne Freigabe. Alle anderen Regeln aus
`CLAUDE.md` (Plan vor Ausführung, Fragen statt annehmen, Identität
prüfen, Linter und Tests grün) bleiben aktiv.

## Was diese Iteration liefert

Wir bauen das **Hardcoded-Skelett mit Compliance-Pages**. Acht Routes,
ein TTDSG-konformer Cookie-Banner, ein Forschungs-Banner, alle
rechtlich verbindlichen Pflicht-Seiten gemäß DDG/DSGVO/TTDSG/BFSG
für deutsche Forschungs-Plattformen. Plus die Layout-Komponenten
(Header, Footer mit allen Pflicht-Links).

Diese Iteration nutzt **noch kein CMS** — alle Inhalte sind hardcoded
in Svelte-Files, weil rechtlich verbindliche Pages PR-Review brauchen,
keinen Browser-Edit.

i18n-Library (Paraglide.js) und Markdown-Pipeline (mdsvex) kommen
in Iteration 1.2 — diese Iteration ist reines Skelett-Bauen.

## Konzept-Dokumente

Du brauchst Zugriff auf zwei Dokumente außerhalb des Repos. Maintainer
stellt dir die Pfade zur Verfügung (vermutlich `/home/hwr/wwn-handover/`
oder ähnlich):

- `feature-decisions.md` — alle Architektur-Entscheidungen
  - Insbesondere relevant für diese Iteration:
    A.7 (Pflicht-Pages-Liste), A.10 (Forschungs-Banner),
    A.14 (Cookie-Strategie), A.15 (Quellen-Attribution)

- `feature-roadmap.md` — konkrete Iterations-Schritte mit
  Akzeptanzkriterien
  - Insbesondere: Abschnitt "Iteration 1.1 — Hardcoded-Skelett mit
    Compliance" mit den 8 Schritten

Bitte lies beide einmal komplett, bevor wir mit der Implementation
loslegen. Die Konzept-Dokumente überschreiben nichts in `CLAUDE.md`,
sind aber spezifischer für die aktuelle Phase.

## Iterations-Plan

### Schritt 1 — Branch + Plan

1. Branch anlegen: `feat/iteration-1-1-skeleton`
2. Verifikation: bist du auf `wwn-dev` (uname -n)? Bist du im richtigen
   Repo-Root (git rev-parse --show-toplevel)?
3. Maintainer-Identität prüfen (`git config --get user.email` muss
   `hwr@relations4u.de` zeigen)
4. Sobald alles OK: kurzen Plan zeigen wie du Schritt 2-8 angehen
   willst, dann Freigabe vom Maintainer abwarten

### Schritt 2 — SvelteKit-Routes anlegen (Stub-Level)

Acht Routes, alle erst als Stub mit Platzhalter-Inhalt:

```
apps/frontend/src/routes/+page.svelte                        (Startseite)
apps/frontend/src/routes/impressum/+page.svelte
apps/frontend/src/routes/datenschutz/+page.svelte
apps/frontend/src/routes/barrierefreiheit/+page.svelte
apps/frontend/src/routes/about/+page.svelte
apps/frontend/src/routes/kontakt/+page.svelte
apps/frontend/src/routes/quellen-attribution/+page.svelte
apps/frontend/src/routes/cookie-einstellungen/+page.svelte
```

Jede Stub-Page bekommt: H1, kurzer Lead-Text, TODO-Kommentar mit
Verweis auf Iteration-Schritt-Nummer. Reine Textinhalte erst in
Schritt 7.

### Schritt 3 — Forschungs-Banner-Komponente

`apps/frontend/src/lib/components/ResearchBanner.svelte`

Anforderungen:

- Sticky oben (CSS `position: sticky` oder fixed mit Spacer)
- Schließbar via Klick auf X-Button
- State persistent in `localStorage` (Key: `wwn-research-banner-closed`)
- Inhalt: kurzer Hinweistext + Link zu /methodik
- Wording: konsultiere mich für den finalen Text, ich liefere dir die
  exakte Formulierung
- Style: dezent aber sichtbar (z. B. amber/yellow Hintergrund)

### Schritt 4 — Cookie-Banner-Komponente (TTDSG-konform!)

`apps/frontend/src/lib/components/CookieBanner.svelte`

**Strikte TTDSG-§-25-Anforderungen** (BGH I ZR 7/16):

- Granulare Wahl: Essenziell / Funktional / Analytics / Marketing
- "Ablehnen" muss **genauso prominent** wie "Akzeptieren" sein
  (gleiche Größe, gleiche Farbe, gleiche Position-Hierarchie)
- **Keine Vorab-Häkchen** außer für "Essenziell" (das ist immer aktiv,
  nicht abwählbar)
- Schließen-X **darf nicht** als implizite Zustimmung gelten
- Speichert Consent in localStorage mit Versions-Stamp:
  `wwn-cookie-consent-v1` mit JSON `{essential, functional, analytics, marketing, timestamp, version}`

Phase-1-Reality:

- Wir setzen aktuell **gar keine** nicht-essenziellen Cookies
- Banner trotzdem TTDSG-konform einbauen, weil:
  - Bei späterer Aktivierung von z. B. Plausible Infrastruktur da
  - Bei externen Embeds (Wetterkarten Drittanbieter) wird Consent nötig

Frage falls unklar: ob der Banner per Default sofort beim ersten Besuch
erscheint, oder ob er nur bei Setzen-Versuch eines nicht-essenziellen
Cookies triggern soll. Empfehlung: per Default beim ersten Besuch, weil
einfacher und juristisch sauberer.

### Schritt 5 — Cookie-Settings-Page

`apps/frontend/src/routes/cookie-einstellungen/+page.svelte`

Erlaubt nachträgliche Anpassung der Cookie-Einstellungen. Im Footer
permanent erreichbar als Link "Cookie-Einstellungen".

UI-Elemente:

- Liste aller Cookie-Kategorien mit Beschreibung
- Toggle pro Kategorie (Essenziell ist disabled, weil immer aktiv)
- "Speichern"-Button mit Erfolgs-Feedback
- Status-Anzeige: "Aktuelle Einstellung: ..."

### Schritt 6 — Layout-Komponenten

Update bestehender Layout-Files:

`apps/frontend/src/routes/+layout.svelte` erweitern um:

- Header mit Logo (Platzhalter), Navigation
- ResearchBanner oben
- CookieBanner als Overlay (nur wenn Consent fehlt)
- Footer mit allen Pflicht-Links:
  - Impressum
  - Datenschutz
  - Barrierefreiheit
  - Cookie-Einstellungen
  - Quellen-Attribution
  - Kontakt
- Konsistente Tailwind-Styles, mobile-first

### Schritt 7 — Inhalts-Drafts erstellen

Für jede der 8 Routes konkreten Inhalt schreiben. Hier brauchen wir
Maintainer-Input für die rechtlich verbindlichen Texte:

- **Impressum** (§ 5 DDG): Heinz W. Richter als Diensteanbieter mit
  Adresse, Mail, ggf. Steuer-/Berufsangaben — Maintainer liefert exakte
  Daten, eRecht24 Generator als Vorlage
- **Datenschutz**: für Forschungs-Phase ohne Tracking. Server-Logs
  (IP, User-Agent — Speicherdauer dokumentieren), Cookie-Liste (nur
  essenziell), Drittland-Transfers (Cloudflare USA-Hosting → SCC),
  Betroffenen-Rechte, Verantwortliche Person
- **Barrierefreiheit**: Erklärung gemäß BFSG, Stand der Konformität,
  Kontakt für Feedback
- **Quellen-Attribution**: Liste der Datenquellen mit Lizenz-Hinweisen
  (Phase 1: noch leer, vorbereitend für Track 2 strukturieren)
- **About**: Mission der Plattform, Maintainer-Vorstellung
- **Kontakt**: Mailto-Link an `ops@worldweathernews.com`, kein Formular
- **Startseite**: Hero mit Wertversprechen, kurze Erklärung, Link zu
  Methodik (kommt in Iteration 1.2)

Für jede dieser Pages: erst Stub schreiben mit klaren TODO-Markern für
juristisch zu klärende Stellen, dann Maintainer-Review.

### Schritt 8 — Tests und Smoke-Checks

- Alle 8 Routes lokal erreichbar (`pnpm --filter frontend dev`)
- Mobile-Responsive (Browser-DevTools, mobile Profile)
- Lighthouse-Check:
  - Performance > 90
  - Accessibility > 95
  - Best Practices > 90
- Cookie-Banner Verhalten testen:
  - Akzeptieren: localStorage gesetzt, Banner verschwindet
  - Ablehnen: localStorage gesetzt, Banner verschwindet
  - Cookie-Settings-Page öffnen: Re-Configure möglich
- Forschungs-Banner Verhalten:
  - Schließen: localStorage gesetzt, Banner verschwindet
  - Reload: Banner bleibt zu

## Akzeptanzkriterien (komplette Liste)

- [ ] Alle 8 Routes live im Dev-Server
- [ ] Forschungs-Banner sichtbar, schließbar, schließt persistent
- [ ] Cookie-Banner TTDSG-konform (Ablehnen ≥ Akzeptieren)
- [ ] Cookie-Settings-Page funktioniert
- [ ] Lighthouse Performance > 90, Accessibility > 95
- [ ] Mobile- und Desktop-Layout sauber
- [ ] Footer enthält alle 6 Pflicht-Links
- [ ] Impressum und Datenschutz haben TODO-Marker für juristisch zu
      klärende Stellen, gehen in eRecht24-Generator-Review
- [ ] Quellen-Attribution-Page strukturiert leer für Track-2-Befüllung
- [ ] Linter und Tests grün (`make lint && make test`)
- [ ] PR-Erstellung erst nach finalem OK des Maintainers

## Was du **noch nicht** baust

Diese Dinge sind explizit für spätere Iterationen:

- **Markdown-Pipeline** (mdsvex) → Iteration 1.2
- **Paraglide-i18n** (DE/EN parallel) → Iteration 1.2
- **Sveltia CMS** → Iteration 1.3
- **Blog-Routes** → Iteration 1.4
- **Methodik-Page** als Markdown → Iteration 1.2

Wenn du verlockt bist, diese Dinge schon einzubauen — widerstehen,
Phase-Disziplin halten.

## Notes für Sprache

Diese Iteration ist **erstmal nur deutsch**. Englische Übersetzungen
kommen in Iteration 1.2 mit Paraglide. Das heißt:

- Routes haben noch kein `/de/` Prefix
- Texte sind direkt deutsch
- Kein i18n-Setup
- Layout muss aber so aufgebaut sein, dass i18n in 1.2 ohne große
  Refactorings nachgerüstet werden kann (z. B. Locale-Switcher-Slot
  im Header schon einplanen)

## Wenn etwas unklar ist

Frag mich. Lieber drei Fragen vor einem Commit als ein falscher Commit.

Insbesondere:

- Bei juristischen Texten: ich (Maintainer) liefere die finale Version
- Bei Wording (Banner-Text, Hero-Text): mit mir abstimmen
- Bei Architektur-Frage, die in `feature-decisions.md` nicht steht:
  pause die Implementation, dokumentier die Frage, ich kläre und
  pflege die Entscheidung in das Tracking-File

Lass uns loslegen. Bestätige mir kurz, dass du die Dokumente
gelesen hast, und schlag den ersten Schritt vor.
