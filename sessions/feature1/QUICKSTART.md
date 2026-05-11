# Quick-Start für den Maintainer

Dieses Dokument ist der **5-Minuten-Überblick** zum Loslegen mit
Iteration 1.1 + 1.1b. Wenn du wenig Zeit hast: hier reinschauen,
nicht in `README.md` (die ist ausführlicher).

---

## Was passiert in welcher Reihenfolge?

```
1. Du bereitest die Pre-Implementation-Tasks vor (15-30 Min)
2. Du startest zwei Claude-Code-Sessions auf wwn-dev (5 Min)
3. Claude Code arbeitet, fragt vor jedem Commit (Stunden bis Tage)
4. Du gibst Freigaben oder forderst Änderungen (jeweils 1-5 Min)
5. Pull Requests werden erstellt, du reviewst, merged
6. v0.1.0 als erstes Feature-Release
```

---

## Pre-Implementation-Tasks (vor Session-Start)

### Vor Iteration 1.1 (Compliance-Skelett)

- [ ] **eRecht24-Account anlegen** (oder Anwalt-Termin)
  - Für Impressum-Generator und Datenschutz-Generator
  - Iteration 1.1 läuft auch ohne, aber juristische Abnahme ist
    Akzeptanzkriterium vor Live-Schaltung

### Vor Iteration 1.1b (Object Storage)

- [ ] **Hetzner Cloud Console öffnen**
  - https://console.hetzner.cloud/
  - Projekt vorbereiten (Vorschlag: "worldweathernews" anlegen)
  - Object Storage in der Region Falkenstein aktivieren
  - **Bucket noch nicht erstellen** — das machen wir mit Claude Code zusammen

- [ ] **aws-cli verfügbar machen**
  - Auf deinem Mac oder auf wwn-dev
  - Auf Mac: `brew install awscli`
  - Auf wwn-dev: `apt install awscli` oder via uv
  - Test: `aws --version`

- [ ] **Bitwarden/1Password bereit**
  - Für sicheres Backup der Hetzner S3-Credentials
  - Notiz vorbereiten: "wwn Hetzner Object Storage"

### Optional aber empfohlen

- [ ] **Cloudflare-Dashboard offen** in einem Tab
  - Für DNS-Eintrag `media.worldweathernews.com` in Iteration 1.1b
  - Für spätere Worker-Aktivierung in Iteration 1.3

---

## Sessions starten

### Variante A — Eine Session nach der anderen (sicherer)

```bash
ssh hwr@10.100.100.113
cd ~/repos/worldweathernews
claude code
```

Dann den **kompletten Inhalt** aus `prompt-iteration-1-1.md` (Abschnitt
"Prompt für Claude Code (Copy-Paste ab hier)") als ersten Prompt
eingeben.

Wenn 1.1 fertig: neue Session, Inhalt aus `prompt-iteration-1-1b.md`.

### Variante B — Zwei Terminals parallel (schneller)

Terminal 1 (für 1.1 Frontend-Skelett):

```bash
ssh hwr@10.100.100.113
cd ~/repos/worldweathernews
git checkout main && git pull
claude code
# → Prompt aus prompt-iteration-1-1.md
```

Terminal 2 (für 1.1b Object Storage):

```bash
ssh hwr@10.100.100.113
cd ~/repos/worldweathernews
git checkout main && git pull
claude code
# → Prompt aus prompt-iteration-1-1b.md
```

**Wichtig**: beide Sessions arbeiten auf eigenen Branches, kollidieren
also nicht. Du wirst trotzdem aufmerksam zwischen ihnen wechseln müssen
für Freigaben.

---

## Tracking-Dokumente bereitstellen

Claude Code braucht Zugriff auf die Konzept-Dokumente. Du musst sie
auf wwn-dev verfügbar machen.

```bash
# Auf deinem Mac:
scp /pfad/zu/feature-decisions.md hwr@10.100.100.113:/home/hwr/wwn-handover/
scp /pfad/zu/feature-roadmap.md hwr@10.100.100.113:/home/hwr/wwn-handover/

# Oder direkt über Browser-Download und scp aus deinem Browser-DL-Folder
```

Auf wwn-dev sollte am Ende existieren:

```
/home/hwr/wwn-handover/
├── feature-decisions.md
├── feature-roadmap.md
├── status-snapshot.md     (optional)
└── handover/
    ├── README.md
    ├── QUICKSTART.md
    ├── prompt-iteration-1-1.md
    └── prompt-iteration-1-1b.md
```

In den Prompts wird der Pfad `/home/hwr/wwn-handover/` als Default
referenziert. Wenn du einen anderen Pfad nutzt, im Prompt anpassen
oder Claude Code den Pfad mitteilen.

---

## Forschungs-Modus — was du beachten musst

Claude Code wird **vor jedem Commit** explizit fragen. Beispiel:

> Claude Code: "Ich habe die ResearchBanner-Komponente fertig.
> Soll ich das jetzt committen?"

Deine Antworten:

| Antwort                   | Wirkung                      |
| ------------------------- | ---------------------------- |
| "OK" / "commit" / "merge" | Wird committed/gemerged      |
| "warte" / "nochmal"       | Wartet, bessert nach         |
| "anders: ..."             | Macht Änderungen vor Commit  |
| "push" / "PR aufmachen"   | Pusht zu GitHub, erstellt PR |
| "show diff"               | Zeigt git diff vor Commit    |

**Wichtig**: stiller Schweigen ist **kein** "OK". Claude Code soll im
Zweifel warten und nachfragen, nicht selbst entscheiden.

---

## Wenn etwas schief läuft

### Claude Code committed ohne dich zu fragen

Selbst-Fix von Claude Code:

```bash
git reset --soft HEAD~1     # Commit rückgängig, Änderungen bleiben
```

Dann: Claude Code an Forschungs-Modus erinnern. Steht im Prompt, sollte
nicht passieren, aber falls doch.

### Akzeptanzkriterium nicht erreicht

Claude Code soll **ehrlich melden**, nicht überspielen. Wenn du den
Eindruck hast, dass etwas geschönt wird:

- "Zeig mir den Lighthouse-Report im Detail"
- "Was sind die Akzeptanzkriterien, die noch offen sind?"
- "Welche TODOs hast du im Code gelassen?"

### Architektur-Frage taucht auf, die im Konzept nicht steht

Claude Code soll **pausieren**, nicht raten. Du klärst die Frage in
einer eigenen Konzept-Session (separater Chat) und pflegst die
Entscheidung in `feature-decisions.md`. Dann zurück zur Implementation.

### Implementation-Bug, blockierend

Claude Code dokumentiert in `docs/backlog.md` als Folge-Issue, baut
einen Workaround. Du entscheidest, ob der Workaround OK ist oder ob
wir das Konzept anpassen.

---

## Nach Abschluss von 1.1 + 1.1b

```
1. Beide PRs gemerged auf main
2. CI grün auf main
3. Tag v0.1.0 setzen (erstes Feature-Release)
4. release.yml deployt auf wwn-prod
5. Smoke-Tests grün:
   - https://worldweathernews.com (Apex)
   - https://research.worldweathernews.com (mit allen 8 Routes)
   - https://media.worldweathernews.com/site/test.txt (Bucket)
6. Status-Snapshot aktualisieren
7. CLAUDE.md Changelog ergänzen
8. Mir Bescheid sagen → wir starten Track 2 oder Iteration 1.2
```

---

## Hilfe-Trigger für mich (Konzept-Claude)

Wenn du in Claude-Code-Session feststeckst und Konzept-Klärung brauchst:

```
"Ich brauche Konzept-Hilfe für: <kurze Beschreibung>"
```

Dann öffnest du eine eigene Chat-Session (nicht Claude Code, normales
Claude.ai), gibst diese Frage rein zusammen mit dem aktuellen
`status-snapshot.md`. Ich helfe dir, die Frage zu klären, und du
bringst die Antwort zurück in die Claude-Code-Session.
