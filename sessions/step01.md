# Session 1 — Repo-Skelett und Tooling

**Phase**: A (Fundament)
**Geschätzte Dauer**: 1-2 Stunden
**Vorbedingung**: Repo initialisiert, `CLAUDE.md` und `sessions/` committed.

## Ziel

Am Ende dieser Session existiert die volle Verzeichnisstruktur des Monorepos,
alle Sprach-Toolchains sind über `mise` reproduzierbar definiert und installiert,
Editor-Konfiguration steht.

Es läuft noch **nichts** außer der Tool-Installation. Das ist Absicht.

## Aufgaben

### 1. Verzeichnisstruktur anlegen

Lege folgende Struktur an. Leere Verzeichnisse mit `.gitkeep` markieren:

```
worldweathernews/
├── apps/
│   ├── backend/
│   ├── frontend/
│   └── pyworkers/
├── packages/
│   ├── api-schema/
│   └── shared-types/
├── infra/
│   ├── compose/
│   ├── caddy/
│   ├── monitoring/
│   │   ├── prometheus/
│   │   ├── grafana/
│   │   └── loki/
│   ├── ansible/
│   │   ├── inventories/
│   │   ├── playbooks/
│   │   └── roles/
│   ├── terraform/
│   │   ├── modules/
│   │   └── environments/
│   ├── secrets/
│   │   ├── staging/
│   │   └── production/
│   └── migrations/
├── scripts/
├── docs/
│   └── adr/
└── .github/
    └── workflows/
```

### 2. `.mise.toml` erstellen

Im Repo-Root mit folgenden Tools:

- `go = "1.23"` (oder die aktuelle stabile Version, falls neuer)
- `node = "22"` (LTS)
- `python = "3.12"`
- `pnpm = "latest"`
- `uv = "latest"`
- `golangci-lint = "latest"`
- `sqlc = "latest"`
- `goose = "latest"`
- `task = "latest"` (Task-Runner als Alternative zu Make für Service-interne Tasks)

Wenn `mise` einzelne Tools nicht direkt unterstützt: dokumentieren wie sie sonst
zu installieren sind (z.B. `go install ...` Hinweis in der README).

### 3. `.editorconfig` anlegen

Sinnvolle Defaults für:

- Go: Tabs für Einrückung
- Python: 4 Leerzeichen
- JavaScript/TypeScript/Svelte: 2 Leerzeichen
- YAML/JSON/Markdown: 2 Leerzeichen
- Trim trailing whitespace, final newline, UTF-8, LF line-endings

### 4. `.gitattributes` anlegen

- LF line-endings für alle Text-Files
- Linguist-Hints: generierte Files (`*.gen.go`, `**/api/types.ts`) als generated markieren
- Binary-Files korrekt markieren

### 5. Initiale `README.md` im Root

Inhalt minimal, aber strukturiert:

- Eine Zeile Beschreibung
- Status-Hinweis "WIP — initial setup phase"
- Quickstart-Block (Platzhalter, Befehle existieren noch nicht):
  ```bash
  mise install
  make bootstrap
  make dev
  ```
- Verweise auf `docs/architecture.md` (Datei mit TODO-Kommentar anlegen)
- Verweis auf `CLAUDE.md` für Mitarbeitende
- Lizenz-Sektion: **frag den Maintainer**, welche Lizenz gewünscht ist.
  Default bis zur Antwort: keine Lizenz, Hinweis "All rights reserved" mit
  Kommentar dass das in Session 12 entschieden wird.

### 6. Platzhalter-Files

- `docs/architecture.md`: nur eine Überschrift + TODO-Kommentar
- `docs/runbook.md`: dito
- `docs/deployment.md`: dito
- `docs/development.md`: dito

## Vorgehen (verbindlich)

1. **Plan zeigen**, ich review.
2. Auf Freigabe warten.
3. Schrittweise umsetzen, Status-Updates an markanten Punkten.
4. Am Ende:
   - `git status` zeigen
   - `mise install` als Befehl mir zur Ausführung geben (du läufst es nicht selbst,
     ich führe es aus damit ich sehe, was passiert)
   - Verzeichnisbaum mit `tree -L 3` oder `find . -type d` zeigen
5. **NICHT committen.** Ich committe selbst, sobald `mise install` durchgelaufen ist.

## Erfolgs-Kriterien

- [ ] Verzeichnisstruktur entspricht der Spec
- [ ] `.mise.toml` listet alle definierten Tools
- [ ] Editor-Config greift in mindestens VS Code und Vim sichtbar
- [ ] `git status` zeigt nur die geplanten neuen Files
- [ ] Keine Platzhalter-Files mit "Lorem ipsum" oder erfundenen Inhalten
- [ ] README enthält keine Anleitungen für Befehle, die noch nicht existieren
      (außer als Quickstart-Vorschau mit Hinweis)

## Mögliche Stolpersteine

- **mise und go-task**: `go-task` heißt im mise-Plugin oft `task`, prüfen.
- **mise auf macOS vs. Linux**: Kommandos können leicht abweichen, README sollte beides erwähnen.
- **`.gitkeep` vs. echte Files**: `.gitkeep` ist eine Konvention, kein Git-Feature.
  Reine leere Datei reicht.

## Was diese Session NICHT tut

- Keine Make-Targets (kommt Session 2)
- Keine Pre-commit-Hooks (kommt Session 2)
- Keine Compose-Files (kommt Session 3)
- Keine echten Service-Dateien (kommen ab Session 4)

## Nach der Session

Trage in `sessions/STATUS.md` ein:

```
## Session 1 — Repo-Skelett und Tooling
Status: ✅ Done
Datum: <heute>
Commit: <SHA>
Notizen: <Auffälligkeiten, offene Lizenz-Frage etc.>
```

Suggested Commit-Message:

```
chore: initial repo skeleton with mise tooling
```
