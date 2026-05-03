# Setup-Sessions für worldweathernews.com

Diese Verzeichnis enthält die strukturierten Sessions zum Aufbau der DevOps-Pipeline.
Jede Datei ist eine in sich abgeschlossene Aufgabenbeschreibung für eine Claude-Code-Session.

## Wie diese Sessions zu nutzen sind

1. **Eine Session = eine Claude-Code-Session.** Starte Claude Code im Repo-Root.
2. Beginne die Session mit:
   ```
   Lies CLAUDE.md und sessions/stepNN.md. Bestätige dass du verstanden hast, was zu tun ist, dann zeig mir den Plan.
   ```
3. Lass Claude den Plan zeigen, review ihn, gib Freigabe.
4. Nach Abschluss: du committest selbst, dann Eintrag in `STATUS.md`.
5. Nächste Session: neuer Claude-Code-Kontext (`/clear` oder neue Instanz).

## Reihenfolge

| #   | Datei     | Phase     | Inhalt                                 |
| --- | --------- | --------- | -------------------------------------- |
| 1   | step01.md | Fundament | Repo-Skelett, Tooling, mise            |
| 2   | step02.md | Fundament | Pre-commit, Makefile, lokale Workflows |
| 3   | step03.md | Services  | Compose-Stack mit DB/Redis/Caddy       |
| 4   | step04.md | Services  | Go-Backend-Skelett                     |
| 5   | step05.md | Services  | SvelteKit-Frontend-Skelett             |
| 6   | step06.md | Services  | Python-Workers-Skelett                 |
| 7   | step07.md | CI/CD     | OpenAPI-Schema + Type-Generation       |
| 8   | step08.md | CI/CD     | GitHub Actions CI-Workflows            |
| 9   | step09.md | CI/CD     | Release-Workflow + Container-Registry  |
| 10  | step10.md | Ops       | Observability-Stack lokal              |
| 11  | step11.md | Ops       | Ansible + SOPS + Terraform-Skelett     |
| 12  | step12.md | Ops       | Dokumentation finalisieren             |

## Vor Session 1

Manuell zu erledigen:

```bash
# Repo erstellen
mkdir worldweathernews && cd worldweathernews
git init
gh repo create worldweathernews --private --source=. --remote=origin

# CLAUDE.md aus diesem Handover ins Repo legen
# sessions/ aus diesem Handover ins Repo legen

# .gitignore (Minimal-Version)
cat > .gitignore <<'EOF'
.env
.env.local
*.sops.decrypted
node_modules/
.svelte-kit/
build/
dist/
__pycache__/
*.pyc
.venv/
bin/
tmp/
.DS_Store
*.swp
.idea/
.vscode/
!.vscode/settings.json
EOF

# Initial-Commit
git add CLAUDE.md sessions/ .gitignore
git commit -m "chore: initial repo setup with session plan"
git push -u origin main

# Claude Code installieren falls noch nicht
npm install -g @anthropic-ai/claude-code

# Erste Session starten
claude
```

## Allgemeine Taktik-Tipps

- **Plan-Mode immer nutzen** für Aufgaben über trivial. Shift+Tab.
- **Auto-Accept-Edits aus.** Du willst jeden Edit sehen.
- **`/clear` zwischen Sessions** für sauberen Context.
- **Bei Tunnelblick: neue Session.** Lieber zwei kurze als eine vermurkste lange.
- **Commit nur du.** Nie Claude committen lassen ohne explizite Freigabe.
- **Bei Halluzinationen: stop und korrigieren.** "Du hast X erfunden, das steht nicht in CLAUDE.md."

## STATUS.md

Eine separate Datei `STATUS.md` in diesem Verzeichnis trackt den Fortschritt.
Format:

```
## Session 1 — Repo-Skelett
Status: ✅ Done
Datum: 2026-05-04
Commit: abc1234
Notizen: Lizenz-Frage offen, in step12 zu klären.
```
