# Runbook

<!-- TODO: Wird in einer späteren Session strukturiert befüllt (siehe sessions/step12.md). -->

## Bekannte Auffälligkeiten / Migrations-TODOs

- **Mailhog läuft als linux/amd64-Image** (kein arm64-Build verfügbar). Auf
  Apple Silicon läuft das via Rosetta — funktional ok, nicht ideal. Drop-in-
  Alternative: [`axllent/mailpit`](https://github.com/axllent/mailpit), multi-arch
  und aktiver gewartet, bietet kompatibles SMTP (Port 1025) und Web-UI (Port
  8025). Ablösung steht offen — nicht kritisch genug für sofortige Migration.
