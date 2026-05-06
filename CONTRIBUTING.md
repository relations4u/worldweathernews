# Contributing

Vielen Dank fürs Mitwirken. Diese Datei beschreibt, wie der Beitragsworkflow
in diesem Repo funktioniert. Tiefere Spielregeln für die generelle
Entwicklungsarbeit (inkl. Versions-Pinning, Secrets, Hosting) stehen in
[`CLAUDE.md`](CLAUDE.md) — die ist für Mensch wie KI gleichermaßen
verbindlich.

## Setup

Siehe [README → Quickstart](README.md#quickstart).

Vor dem ersten Commit:

```bash
mise install
make bootstrap
pre-commit install --install-hooks
```

## Branching

- `main` ist immer deploybar — Direct-Pushes sind durch Branch-Protection
  blockiert.
- Feature-Branches mit Prefix:
  - `feat/<thema>` für neue Features
  - `fix/<thema>` für Bugfixes
  - `chore/<thema>` für Tooling, Build, Refactoring
  - `ci/<thema>` für CI/CD-Änderungen
  - `docs/<thema>` für reine Doku
- Eine PR = ein logisches Thema. Lieber zwei kleine PRs als eine große.

## Commits

[Conventional Commits](https://www.conventionalcommits.org/) sind Pflicht
und werden von commitlint im CI erzwungen:

- `feat(scope): add X`
- `fix(scope): handle null in Y`
- `chore(deps): bump golangci-lint`
- `docs(adr): add 0006-i18n-strategy`

Übliche Scopes: `backend`, `frontend`, `pyworkers`, `infra`, `api`, `ci`,
`deps`, `docs`, `caddy`, `ansible`.

**Body-Zeilen ≤ 100 Zeichen** (commitlint hat `body-max-line-length: 100`).
Bei Mehrabsatz-Bodies: Heredoc oder `git commit -F /tmp/msg.txt` statt
Multi-`-m`-Zeilen.

**Signierte Commits sind Pflicht**: SSH-Signing mit `~/.ssh/id_ed25519.pub`,
Konfiguration siehe CLAUDE.md → Maintainer-Identität.

## Pull Requests

1. Branch ziehen, lokale Tests vor Push grün machen:

   ```bash
   make lint
   make test
   ```

2. Bei Schema-Änderungen: `make gen && make gen-check` müssen lokal grün
   sein, sonst meckert die CI.
3. PR aufmachen mit aussagekräftiger Beschreibung:
   - **Was** — kurze Zusammenfassung
   - **Warum** — Motivation, Kontext
   - **Test plan** — wie hast du verifiziert?
4. Self-Review vor Reviewer-Anfrage. Nur grüne CI wird gemerged.
5. Squash-Merge ist Default — Linear History auf `main`.

## Code-Style

Wird durch Linter erzwungen (`make lint`):

- Go: `gofmt`, `goimports`, `golangci-lint v2.12.1`
- Python: `ruff` (Lint + Format), `mypy` strict
- TypeScript / Svelte: `eslint`, `prettier`, `svelte-check`
- YAML: `yamllint`, `prettier`
- Dockerfiles: `hadolint`

Auto-Fix: `make fmt`.

## Secrets

`infra/secrets/` ist SOPS-verschlüsselt. **Niemals Plaintext-Secrets
committen** — der Pre-commit-Hook `forbid-unencrypted-secrets` blockt
es ohnehin. Workflow: `docs/secrets.md`.

## Tests

- Go: Tests neben dem Code (`foo.go` + `foo_test.go`), CI läuft mit
  `-race`.
- Python: `tests/`-Verzeichnis, pytest mit `asyncio_mode = "auto"`.
- Frontend: Unit-Tests via Vitest, E2E (geplant) mit Playwright.
- Mindest-Coverage: kein hartes Gate, aber sichtbar in CI.

Vor dem PR: `make test` muss grün sein.

## Reviews

- Reviewer prüft: Logik, Tests, Doku, Side-Effects, Versions-Pins.
- Reviewee fragt nach, wenn Feedback unklar ist.
- Eigene PRs werden in der Solo-Maintainer-Phase ohne Review gemerged,
  sobald CI grün ist; bei Code-Reviews aufgemacht, sobald ein zweiter
  Mensch dazukommt.

## Issues und Bug-Reports

Bug-Reports und Feature-Requests bitte als
[GitHub Issue](https://github.com/relations4u/worldweathernews/issues).

Bei Bugs hilft uns:

- Genaue Reproduktion (Schritte, Erwartung vs. Realität)
- Logs (anonymisiert wenn nötig)
- Versions-Info (`make version` zeigt App-Versionen + Toolchain)
- Browser/OS bei Frontend-Bugs

## Security

Sicherheits-Probleme bitte **nicht** als öffentliches Issue,
sondern privat per E-Mail an `security@worldweathernews.com`
(folgt sobald die Mailbox aktiv ist; bis dahin
`hwr@relations4u.de`).

## Lizenz

Mit dem Einreichen eines Beitrags erklärst du dich damit einverstanden,
dass dein Code unter der [AGPL-3.0](LICENSE) lizenziert wird.
