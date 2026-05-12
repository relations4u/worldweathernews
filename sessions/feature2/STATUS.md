# Track 2 — Wetterdaten-Import — Status

Pflege diese Datei am Ende jeder Iteration. Format analog zu
`sessions/feature1/STATUS.md`.

Status-Legende: ✅ Done · 🟡 In Progress · ⏳ Geplant · ❌ Blocked · ⏭ Skipped

Stand: 2026-05-12

---

## Konzept-Phase

Status: ✅ Done (für Iteration 2.1)
Datum: 2026-05-11

Vier B-Punkte aus `sessions/feature1/feature-decisions.md` Abschnitt B:

- **B.1 (Erste Datenquelle)** ✅ DECIDED 2026-05-11
  Open-Meteo zuerst. 3 Städte (Potsdam, Berlin, Hamburg),
  4 Variablen (Temperatur, Niederschlag, Windgeschwindigkeit,
  Windrichtung), 2 Frequenzen (current + hourly 24h). Keine Historie
  in 2.1. Storage in Postgres + TimescaleDB-Hypertables.
- **B.2 (Wetterkarten)** ⏳ POSTPONED 2026-05-11
  Eigene Konzept-Diskussion nach Iteration 2.1. Drei Optionen
  (K1 selbst rendern / K2 extern / K3 hybrid) dokumentiert.
- **B.3 (Storage für Datasets)** ⏳ POSTPONED 2026-05-11
  Für 2.1 nicht relevant. Wiederaufnahme bei Iterationen 2.4/2.5
  (Satellitenbilder, Radar).
- **B.4 (Daten-Lizenzen)** ✅ DECIDED 2026-05-11
  Attribution-Pattern: Footer-Link auf jeder Page + Detail-Page
  `/quellen-attribution`. Strings in Paraglide-Messages.

---

## Geplante Iterationen

### Iteration 2.1 — Open-Meteo Hello World

Status: ⏳ Geplant (Voraussetzung erfüllt seit v0.0.5)
Übergabe-Prompt: `prompt-iteration-2-1.md` (542 Zeilen, 9 Schritte)
Geschätzte Dauer: 3-4 Tage
Geplanter Tag: **v0.1.0** (erstes Feature-Release)

**Voraussetzungen:**

- [x] Security-Triage-PR gemerged (PR #67, v0.0.5 live seit 12. Mai)
- [x] Konzept (B.1, B.4) entschieden
- [x] Übergabe-Prompt geschrieben

**Offene Klärungs-Fragen für Implementations-Start:**

- Scheduling-Wahl: W1 (APScheduler im Worker-Container) bevorzugt,
  finale Entscheidung im Implementations-Schritt 4
- Frontend-Position: Startseite erweitern vs eigene `/wetter`-Route
  — Entscheidung mit Maintainer

**Akzeptanzkriterien:** siehe `prompt-iteration-2-1.md`

### Iteration 2.2 — DWD-Adapter

Status: ⏳ Geplant (Plan-Skizze fertig, Übergabe-Prompt nach 2.1)
Plan-Skizze: `plan-iteration-2-2.md`
Geschätzte Dauer: 4-6 Tage (DWD-Format-Komplexität)
Geplanter Tag: **v0.2.0**

**Voraussetzungen:**

- [ ] Iteration 2.1 gemerged und v0.1.0 live
- [ ] Worker-Pattern aus 2.1 erprobt (lessons learned eingearbeitet)
- [ ] DWD-OpenData-Recherche durchgeführt (siehe Plan-Skizze)
- [ ] Konkrete Stations-Auswahl mit Maintainer abgestimmt
- [ ] Übergabe-Prompt ausgearbeitet (`prompt-iteration-2-2.md`)

### Iteration 2.3 — Stations-Map mit MapLibre

Status: ⏳ Geplant (Plan-Skizze fertig, Übergabe-Prompt nach 2.2)
Plan-Skizze: `plan-iteration-2-3.md`
Geschätzte Dauer: 3-4 Tage
Geplanter Tag: **v0.3.0**

**Voraussetzungen:**

- [ ] Iteration 2.2 gemerged und v0.2.0 live
- [ ] Tile-Quelle entschieden (siehe Plan-Skizze: OSM / Stadiamaps /
      MapTiler)
- [ ] Cookie-Banner-Implikationen für externe Tile-Quelle geprüft
- [ ] Übergabe-Prompt ausgearbeitet (`prompt-iteration-2-3.md`)

---

## Folge-Iterationen (Konzept-Diskussion ausstehend)

### Iteration 2.4 — Satellitenbilder

Status: ⏳ Konzept offen — braucht B.2-Wiederaufnahme + EUMETSAT-
Lizenz-Bestätigung

### Iteration 2.5 — Radar

Status: ⏳ Konzept offen — braucht B.2-Wiederaufnahme + DWD-Radolan-
Recherche

### Iteration 2.6 — ICON-Modelle (komplette Modellläufe)

Status: ⏳ Konzept offen — braucht B.3-Wiederaufnahme (Storage für
GRIB-Dateien, mehrere GB pro Modelllauf)

---

## Tag-Roadmap

```
v0.0.5      Security-Triage post-v0.0.4               ✅ 2026-05-12
                ↓
v0.1.0      Iteration 2.1 (Open-Meteo Hello World)    ⏳ ~16. Mai 2026
                ↓
v0.2.0      Iteration 2.2 (DWD-Adapter)               ⏳ ~22. Mai 2026
                ↓
v0.3.0      Iteration 2.3 (Stations-Map)              ⏳ ~28. Mai 2026
                ↓
Konzept-Session vor Track-2-Fortsetzung:
  - B.2 Wetterkarten-Strategie
  - B.3 Storage für große Datasets
  - EUMETSAT-Lizenz-Status für Phase 1
                ↓
v0.4.0+     2.4 / 2.5 / 2.6 nach Konzept-Session       ⏳ Juni 2026+
```

Daten sind Schätzungen, kein Commitment. Iteration startet wenn
Voraussetzungen erfüllt sind, nicht nach Kalender.

---

## Querschnitt-Themen

### 1.3b — Image-Pipeline (Track 1, ausgesetzt)

Status: ⏭ Skipped bis Blog-Bedarf entsteht (Iteration 1.4)
Begründung: keine bildbedürftige Page in Sicht, Pipeline ohne
Use-Case wäre theoretisch. Wird mit 1.4 (Blog) zusammen
gebündelt oder unmittelbar davor implementiert.

---

## Refs

- Übergeordnete Decisions: `../feature1/feature-decisions.md` Abschnitt B
- Übergeordnete Roadmap: `../feature1/feature-roadmap.md`
- Track-1-Status: `../feature1/STATUS.md`
- Setup-Phase-Status: `../STATUS.md`
