# Übergabe an Claude Code — Track 1 Iteration 1.1 + 1.1b

Stand: 7. Mai 2026
Maintainer: Heinz W. Richter <hwr@relations4u.de>

Dieses Verzeichnis enthält die Übergabe-Prompts für Claude Code, um
Track 1 der Feature-Phase zu starten.

---

## Wie nutze ich diese Prompts?

Die Prompts sind so geschrieben, dass du sie **als ersten Prompt einer
neuen Claude-Code-Session auf wwn-dev** verwenden kannst.

```
ssh hwr@10.100.100.113     # auf wwn-dev
cd ~/repos/worldweathernews
claude code                # neue Session starten
# → ersten Prompt aus prompt-iteration-1-1.md einfügen
```

Jeder Prompt:

- Setzt den Kontext (warum, was, wofür)
- Definiert das konkrete Ziel der Iteration
- Listet alle Schritte mit Akzeptanzkriterien
- Enthält Verweise auf die maßgeblichen Konzept-Dokumente
- Aktiviert den **Forschungs-Modus** mit expliziter Freigabe
- Erklärt das Branch- und Commit-Verhalten

---

## Iterations-Reihenfolge

```
1. prompt-iteration-1-1.md     Hardcoded-Skelett mit Compliance
   ⏱  geschätzt 2-3 Tage Arbeit
   📦  Liefert: 8 Routes, Cookie-Banner, Forschungs-Banner, Layouts

2. prompt-iteration-1-1b.md    Hetzner Object Storage einrichten
   ⏱  geschätzt 1 Tag Arbeit
   📦  Liefert: media.worldweathernews.com Bucket, SOPS-Secrets
   ℹ️  Kann parallel zu 1.1 laufen
```

Iterationen 1.2 bis 1.5 (mdsvex, Sveltia, Blog, Onboarding) folgen
in eigenen Prompts, sobald 1.1 fertig ist.

---

## Forschungs-Modus für Claude Code

**Wichtige Abweichung von der Setup-Phase-Regel:**

In der Setup-Phase galt: „Maintainer committet selbst."

In der Feature-Phase gilt: **„Claude Code committet nach expliziter
Freigabe durch den Maintainer."**

Workflow für Claude Code:

1. Branch anlegen für die Iteration (`feat/iteration-1-1-skeleton`)
2. Implementation in mehreren Commits auf dem Branch
3. Vor jedem Commit: Maintainer um Freigabe fragen
4. Bei „OK" oder „commit" oder „merge": committen oder mergen
5. Bei „warte" oder „nochmal" oder „anders": warten und nachbessern
6. Push zu GitHub: erst nach explizitem „push" oder „PR aufmachen"

**Kein eigenständiger Commit ohne Freigabe.** Auch wenn der Code
fertig aussieht — immer fragen.

Die Setup-Phase-Disziplin („nie eigenständig committen") wird durch
„nach expliziter Freigabe committen" ersetzt, nicht durch „committen
wann immer es sinnvoll erscheint".

---

## Konzept-Quellen für Claude Code

Die Konzept-Dokumente, die Claude Code lesen sollte:

```
/home/wwn-handover/feature-decisions.md
  → Alle Architektur-Entscheidungen mit Begründung
  → Bei Unklarheit: hier nachschauen, dann fragen

/home/wwn-handover/feature-roadmap.md
  → Konkrete Schritte je Iteration mit Akzeptanzkriterien
  → Bei „wie geht das?": hier nachschauen

CLAUDE.md (im Repo-Root)
  → Setup-Phase-Doku, Conventions, Tech-Stack
  → Bei „warum so?": hier nachschauen

apps/frontend/  (im Repo)
  → Bestehender SvelteKit-Code, der erweitert wird
  → Bei „was ist da schon?": hier nachschauen
```

---

## Pre-Implementation Tasks (Maintainer vor Claude-Code-Sessions)

Diese Tasks sollten **vor** der ersten Claude-Code-Session erledigt sein:

### Vor Iteration 1.1 (Compliance-Skelett)

- [ ] **Anwalt-Vorbereitung**: Termin oder eRecht24-Account aufsetzen
      für Impressum + Datenschutz-Review. Kann parallel zur Implementation
      laufen, aber juristische Abnahme ist Akzeptanzkriterium.

### Vor Iteration 1.1b (Hetzner Object Storage)

- [ ] **Hetzner Cloud Account**: bereits vorhanden ✅
- [ ] Cloud Console öffnen, Object Storage aktivieren
- [ ] Bucket-Naming-Convention bestätigen: `media-worldweathernews-prod`

### Vor Iteration 1.3 (später, Sveltia einbinden)

- [ ] **GitHub OAuth-App** erstellen:
  - GitHub → Settings → Developer settings → OAuth Apps → New
  - Name: `worldweathernews-cms`
  - Homepage: `https://research.worldweathernews.com`
  - Callback: `https://wwn-cms-auth.<workers-domain>/callback`
    (genaue Worker-Domain wird in 1.3 bestimmt)
  - Client ID + Secret notieren in Bitwarden/1Password

- [ ] **Cloudflare Worker Account** vorbereiten:
  - Cloudflare Dashboard → Workers & Pages → Aktivieren
  - Free Tier reicht für CMS-Login-Volumen
  - Wrangler-CLI lokal installieren (in 1.3-Prompt enthalten)

---

## Maintainer-Kontrollpunkte je Iteration

Hier kann der Maintainer eingreifen oder Feedback geben:

```
🔵 Branch wird angelegt           → optional: anderer Branch-Name?
🔵 Erste Commits werden gemacht   → Code-Style passt? Tests OK?
🔵 Akzeptanzkriterium erreicht    → vor PR-Erstellung
🔵 PR-Erstellung                   → Maintainer reviewed im PR
🔵 Merge nach grünem CI            → final OK
🔵 Anschließendes Deployment       → next release-Tag
```

Claude Code wird vor jedem dieser Punkte explizit fragen.

---

## Was tun bei Problemen während Implementation?

### Wenn Claude Code unsicher ist:

→ fragen statt raten (CLAUDE.md-Regel bleibt aktiv)

### Wenn ein Akzeptanzkriterium nicht erreicht wird:

→ ehrlich melden, nicht überspielen
→ analysieren, ob das Konzept-Decision in `feature-decisions.md`
vielleicht angepasst werden muss
→ Maintainer entscheidet über Anpassung oder Workaround

### Wenn unerwartete Architektur-Frage auftaucht:

→ Pause der Implementation
→ neue Frage in `feature-decisions.md` als `[OPEN]` eintragen
→ Maintainer und (parallel laufende) Konzept-Claude-Session
klären, dann zurück zur Implementation

### Wenn ein Tool-/Library-Bug blockiert:

→ Workaround dokumentieren in „Häufige Fallen" der CLAUDE.md
→ Issue im jeweiligen Library-Repo (Sveltia, Paraglide, etc.)
→ Backlog-Eintrag in `docs/backlog.md` für späteren Folge-Fix

---

## Nach Iteration 1.1 + 1.1b — was kommt?

Nach erfolgreichem Abschluss beider Iterationen:

1. **Tag v0.1.0** als erstes Feature-Release setzen
2. **Maintainer-Review** der Live-Site
3. **Status-Update** in `STATUS.md` und `CLAUDE.md` Changelog
4. **Folge-Iteration** 1.2 (mdsvex + Paraglide) starten

Die Übergabe-Prompts für 1.2 bis 1.5 werden zu gegebener Zeit ergänzt.
