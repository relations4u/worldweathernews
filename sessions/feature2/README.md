# Track 2 — Wetterdaten-Import

Stand: 11. Mai 2026
Maintainer: Heinz W. Richter <hwr@relations4u.de>

Dieser Folder enthält die Konzept- und Übergabe-Dokumente für **Track 2**
der Feature-Phase — Wetterdaten-Import in die worldweathernews.com-
Plattform.

---

## Hintergrund

Track 1 (Frontend, Inhalte, CMS) ist bei Iteration 1.3a abgeschlossen,
Plattform läuft auf v0.0.4. Track 2 bringt **echte Wetterdaten** ins
System. Die Architektur-Entscheidungen sind in
`sessions/feature1/feature-decisions.md` (Abschnitt B) dokumentiert:

- **B.1 DECIDED**: Open-Meteo zuerst, drei Städte, vier Variablen
- **B.2 POSTPONED**: Wetterkarten — eigene Konzept-Diskussion nach 2.1
- **B.3 POSTPONED**: Storage für große Datasets — relevant ab 2.4
- **B.4 DECIDED**: Daten-Lizenzen + Attribution-Pattern

## Iterations-Reihenfolge

```
1. prompt-iteration-2-1.md    Open-Meteo Hello World
   ⏱  geschätzt 3-4 Tage Arbeit
   📦  Liefert: 3 Städte, 4 Variablen, Worker→DB→API→Frontend-Pattern
   📌  Voraussetzung: Security-Triage-PR gemerged (✅ v0.0.5 live)
   🚀  Status: BEREIT FÜR CLAUDE CODE
   🏷️  Tag: v0.1.0

2. plan-iteration-2-2.md      DWD-Adapter (Plan-Skizze)
   ⏱  geschätzt 4-6 Tage Arbeit
   📦  Liefert: erste deutsche-Wetterdienst-Daten + DWD-Station-IDs
   📌  Voraussetzung: 2.1 gemerged, Konzept-Fragen Q1-Q6 geklärt
   🚀  Status: PLAN-SKIZZE, prompt-iteration-2-2.md nach 2.1
   🏷️  Tag: v0.2.0

3. plan-iteration-2-3.md      Stations-Map (Plan-Skizze)
   ⏱  geschätzt 3-4 Tage Arbeit
   📦  Liefert: Interaktive Karte mit Wetter-Marker pro Station
   📌  Voraussetzung: 2.2 gemerged, Tile-Quelle entschieden
   🚀  Status: PLAN-SKIZZE, prompt-iteration-2-3.md nach 2.2
   🏷️  Tag: v0.3.0

4. Track-2-Folge-Konzept: B.2 Wetterkarten + B.3 Storage
   ⏳ Konzept-Diskussion nach 2.3 live ist
   📦  Vorbereitung für 2.4 (Satellitenbilder), 2.5 (Radar),
       2.6 (ICON-Modelle)
```

## Workflow: Plan-Skizze → Übergabe-Prompt

Plan-Skizzen sind **früh** entstanden und enthalten offene Konzept-
Fragen. Sie werden zu submission-ready Übergabe-Prompts ausgearbeitet,
**nachdem die vorhergehende Iteration gemerged ist**. Grund:
Erkenntnisse aus der laufenden Implementation fließen in den nächsten
Prompt ein.

Lebenszyklus:

```
PLAN-SKIZZE → KONZEPT-DISKUSSION → ÜBERGABE-PROMPT → CLAUDE CODE
  (jetzt)    (zwischen Iterationen)  (vor Submission)   (Implementation)
```

## Forschungs-Modus

Wie in Track 1: Claude Code committet **nach expliziter Freigabe**.
Detail im Übergabe-Prompt selbst.

## Pre-Implementation-Tasks für 2.1

- [ ] **Security-Triage-PR gemerged**: `chore/security-triage-post-v0-0-4`
      muss durch sein bevor 2.1 startet. Begründung: 2.1 baut neue
      Python-Module mit httpx-Calls — wenn urllib3 noch auf 2.6.3 ist,
      kollidiert das mit der pyproject.toml-Pin (Commit 2 der Security-
      Triage)
- [ ] **Open-Meteo Rate-Limits geprüft** (optional, präventiv):
      Open-Meteo free tier ist großzügig (10000 Calls/Tag), unsere
      Frequenz (3 Städte × 6 Calls/h current + 1 Call/h hourly = ~150/Tag)
      ist locker drin. Kein Account nötig.
- [ ] **Sessions-Workflow vorbereiten**: feature1- und feature2-
      Tracking-Docs auf wwn-dev verfügbar (Pfad `/home/hwr/wwn-handover/`
      oder so)

## Konzept-Quellen für Claude Code

```
sessions/feature1/feature-decisions.md  → Architektur-Entscheidungen
  → B.1, B.4 als DECIDED, B.2/B.3 POSTPONED
  → A.16 Self-hosted Compute
  → A.17 Paraglide für UI-Strings
  → A.19 Self-hosting-Prinzip

sessions/feature1/feature-roadmap.md    → Iterations-Schritte
  → Track 2 Iteration 2.1 grobe Skizze (TBD-Vermerk wird ersetzt)

sessions/feature2/prompt-iteration-2-1.md → konkreter Übergabe-Prompt
  → Detail-Plan für 9 Schritte
```

## Maintainer-Kontrollpunkte je Iteration

Wie bei Track 1:

```
🔵 Branch wird angelegt           → optional anderer Name
🔵 Erste Commits werden gemacht   → Code-Style passt? Tests OK?
🔵 Akzeptanzkriterium erreicht    → vor PR-Erstellung
🔵 PR-Erstellung                   → Maintainer reviewed im PR
🔵 Merge nach grünem CI            → final OK
🔵 Anschließendes Deployment       → next release-Tag (v0.1.0
   als erstes Feature-Release wenn 2.1 live geht)
```

## Files in dieser Folder

```
README.md                  ← du liest sie gerade
STATUS.md                  ← Track-2-Status, ähnlich feature1/STATUS.md
prompt-iteration-2-1.md    ← Übergabe-Prompt Iteration 2.1
                              (Open-Meteo Hello World, 9 Schritte)
plan-iteration-2-2.md      ← Plan-Skizze Iteration 2.2
                              (DWD-Adapter, 6 offene Konzept-Fragen)
plan-iteration-2-3.md      ← Plan-Skizze Iteration 2.3
                              (Stations-Map, 6 offene Konzept-Fragen)
```

Wenn 2.1 gemerged ist:

- Plan-Skizze 2.2 wird zu `prompt-iteration-2-2.md` ausgearbeitet
  (mit Lessons aus 2.1 und geklärten Konzept-Fragen)
- STATUS.md wird aktualisiert

Wenn 2.2 gemerged ist:

- Plan-Skizze 2.3 wird zu `prompt-iteration-2-3.md` ausgearbeitet
- STATUS.md wird aktualisiert

Spätere Iterationen (2.4-2.6) brauchen eigene Konzept-Diskussion
(B.2 Wetterkarten, B.3 Storage, EUMETSAT-Lizenz) bevor Plan-Skizzen
sinnvoll sind.
