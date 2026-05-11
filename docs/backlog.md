# Backlog

Liste von technischen Punkten, die bewusst auf später verschoben sind.
Pflegen wie ein leichtes Issue-Tracking — Stichpunkt + warum-jetzt-nicht

- wann/woran erkennbar, dass es jetzt Zeit ist. Wenn ein Eintrag in
  GitHub-Issues konvertiert wird, hier mit Issue-Nr ergänzen statt
  löschen, damit die Spur erhalten bleibt.

Stand: 2026-05-06.

## Operations / Infrastruktur

- **Backend- und Pyworkers-Metrics von wwn-mon scrapen** — Container
  binden ihre Metrics-Ports aktuell auf `127.0.0.1`, also nicht von
  wwn-mon (10.100.100.22) erreichbar. Optionen: zweites Port-Binding
  auf `10.100.100.21` mit ufw-Restrict zu mon, oder Push-Sidecar
  (vmagent o. ä.) auf wwn-prod, der remote-write zu Prometheus macht.
  Code-Stelle: `infra/ansible/roles/monitoring-stack/files/prometheus/prometheus.yml`.
- **node-exporter für wwn-mon im Stack ergänzen** — wwn-mon hat
  aktuell keine Host-Metriken (der Stack-Promtail ist nur für Logs).
  Folge-PR.
- **Caddy-Admin-Metrics in Prometheus hängen** — Caddy-Admin ist im
  Caddyfile nicht freigegeben; sobald aktiviert, scrape-Job in
  `infra/monitoring/prometheus/prometheus.yml` einkommentieren.
- **Tracing-Sampler für Production** — aktuell `AlwaysSample`. Vor
  signifikantem Wachstum auf
  `ParentBased(TraceIDRatioBased(0.1))` umstellen. Code-Stelle:
  `apps/backend/internal/observability/tracing.go`.
- **Tempo-Storage auf S3/MinIO statt lokal-Filesystem** — sobald
  Trace-Volumes wachsen.
- **Postgres-Backups automatisieren** — aktuell manuelle `pg_dump`-
  Kommandos im Runbook. Pflicht, bevor wir echte User-Daten halten.
  Optionen: pgBackRest, Borg, externer Backup-Worker.
- **Proxmox-Snapshot-Cadence formalisieren** — aktuell manuelle
  Snapshots durch den Maintainer (z. B. `caddy-online` 6. Mai 2026,
  `setup-complete` 6. Mai 2026 nach v0.0.2 + Metrics-Wiring auf
  wwn-prod und wwn-mon). Vor jedem Production-Deploy ohne klares
  Rollback-Pfad: snapshot. Mittelfristig: Proxmox Backup Server als
  separate VM auf demselben Host plus externer Off-Site-Sync.
  Snapshot- vs. Backup-Trennung dokumentieren.
- **Container-Resource-Metrics** (cAdvisor / Docker-Stats-Exporter)
  und **postgres_exporter / redis_exporter** — der _Infra Overview_-
  Dashboard hat dafür Stubs.
- **CDN/Edge-Cache vor `media.worldweathernews.com`** — aktuell
  proxied Caddy auf wwn-prod direkt zum Hetzner-Bucket; jeder
  Request läuft über den Heim-Anschluss, kein Edge-Cache. Optionen,
  wenn Bandwidth oder Latenz drücken: (a) Cloudflare-Worker mit
  Cache-API vor dem Hetzner-Endpoint (setzt Workers-Aktivierung
  voraus, siehe nächster Punkt), (b) Migration zu Cloudflare R2 mit
  nativer Worker-Binding-API (würde A.13 in `feature-decisions.md`
  revidieren — DSGVO-Frage neu prüfen). Code-Stelle:
  `infra/caddy/prod/Caddyfile` (Block `media.worldweathernews.com`).
- **Cloudflare-Workers-Subscription-Status klären** — bei der
  Iteration-1.1b-Recherche waren „Custom Domains" und „Routes" für
  Workers im Cloudflare-Dashboard ausgegraut. Unklar ob Free-Tier
  gar nicht aktiviert oder spezifische Sub-Funktion Paid. Voraussetzung
  für den vorigen Punkt (Cloudflare-Worker als CDN-Edge).

## CMS

- **Iteration 1.3b — Image-Pipeline** — Pre-Signed-S3-URL-Worker,
  WASM-libvips Konvertierung zu WebP + responsive Sizes
  (320/640/1280/1920) + EXIF-Strip, Sveltia `media_library`-Switch.
  Voraussetzung: 1.3a stabil, mindestens eine Bild-bedürftige Page
  in Sicht (Blog 1.4). Code-Stellen: zweiter Cloudflare-Worker
  `wwn-media-upload` parallel zu `wwn-cms-auth`, plus
  `apps/frontend/static/admin/config.yml` (`media_folder`).
- **Decap-Fallback dokumentiert halten** — falls Sveltia-Wartung
  jemals stagniert: drop-in-kompatibel über Loader-Script-Tausch
  in `apps/frontend/static/admin/index.html`. Auth-Worker und
  `config.yml` bleiben identisch.
- **HSTS double-header auf media.worldweathernews.com** — Caddy
  und Hetzner Object Storage senden beide `Strict-Transport-
Security`. Browser akzeptieren nur das erste, Doppelung ist
  technisch egal, kosmetisch unsauber. Optionen: Hetzner-Header
  in Caddy strippen (`header_down -Strict-Transport-Security`)
  oder Caddy-Header weglassen für diesen Vhost.

## Tooling / Build

- **Mailhog → Mailpit** — Mailhog ist amd64-only, läuft auf Apple
  Silicon nur unter Rosetta. `axllent/mailpit` ist drop-in,
  multi-arch, aktiv gewartet, kompatibles SMTP/UI. Migration
  nicht kritisch, aber low-effort.
- **i18n-Library wählen** — `svelte-i18n` vs. Paraglide vs. Inlang.
  Entscheidung in der ersten Feature-Session, die User-facing Text
  einführt. Code-Stelle: `apps/frontend/src/lib/i18n/index.ts`.
- **Default-Versionen in `infra/ansible/inventories/production/group_vars/all.yml`**
  — aktuell `0.0.0` als bewusster „darf nicht starten"-Marker. Sobald
  ein Tagged Release als „Default" gelten soll, hier setzen statt
  per `-e target_version=…` zu übergeben.

## Sicherheit

- **DMARC** auf `p=reject` umstellen — aktuell `p=quarantine`. Nach
  1–2 Wochen Beobachtung ohne legitime Mails in Quarantäne.
- **HSTS `includeSubDomains`** — aktuell bewusst weg, weil zukünftige
  interne Subdomains evtl. lange ohne TLS sein. Sobald alle
  geplanten Subdomains zuverlässig TLS haben: aktivieren, später
  `preload` evaluieren.
- **Cosign-Signaturen in der Deploy-Pipeline verifizieren** — die
  Pipeline signiert beim Push, der Pull macht's nicht automatisch.
  Ansible-Task `cosign verify` vor dem `docker compose pull`
  ergänzen.
- **Security-Scan-Workflow: behoben 2026-05-11** — die vier Befunde
  aus dem ursprünglichen Backlog-Eintrag (Trivy-SARIF-403, pnpm audit,
  pip-audit, govulncheck) sind durch die Triage-Runde im Branch
  `chore/security-triage-post-v0-0-4` weggegangen. Die zugehörigen
  Commits:
  1. `af63796` (Go-Toolchain 1.25.10) — schließt govulncheck-CVEs
     GO-2026-4982 / -4980 / -4971 / -4918 (Stdlib) und GO-2025-3770
     (chi v5.2.5 in cms-auth).
  2. `ce8da5c` (urllib3 >= 2.7.0) — schließt CVE-2026-44431 / -44432
     im pyworkers-Dep-Graph; pip-audit damit grün.
  3. `0e53867` (`security-scan.yml`-Fixes) — Trivy lädt SARIF jetzt
     als Artifact (gleicher Pattern wie `release.yml`, weil
     codeql-action/upload-sarif ohne GHAS auf privaten Repos 403
     liefert); pnpm audit läuft mit `--prod`, weil die hochstufigen
     Findings transitiv durch `@redocly/cli` kamen — Build-only-Tooling,
     kein Runtime-Risiko.

  Bewusst aus der Runde rausgelassen (Folge-Iterationen):
  - **devDep-Updates `@redocly/cli`** — pulls vulnerable
    `fast-xml-parser` und `fast-uri` via `redoc → @redocly/openapi-core
→ @redocly/ajv`. Mit `--prod` aus dem CI-Gate raus, aber im lokalen
    Build trotzdem im node_modules-Tree. Als eigene devDep-Pflege-
    Iteration auf die aktuelle `@redocly/cli`-Major (oder Migration
    zu `vacuum` / `spectral`) anstoßen, sobald die nächste
    OpenAPI-Schema-Pflege ansteht.
  - **CodeQL-Action v3 → v4 Migration** —
    `github/codeql-action/*` ist bei v4. Wir nutzen `@v3` nach den
    Workflow-Fixes nirgends mehr (upload-sarif ist raus). Falls wir
    CodeQL irgendwo neu hinzufügen, direkt mit v4 anfangen, nicht
    aus alten Beispielen kopieren.
  - **SARIF-Konsumenten-Frage** — `release.yml` und `security-scan.yml`
    legen Trivy-SARIF jetzt als Artifact ab, ohne dass etwas die liest.
    Wenn GHAS aktiviert wird oder das Repo public geht: zurück auf
    `github/codeql-action/upload-sarif@v4` (und die `security-events:
write`-Permission wieder rein). Bis dahin SARIF manuell ziehen,
    wenn Trivy rot wird — Hinweis in beiden Workflow-Files vorhanden.

## Produkt / Features (gehören eigentlich nicht hier rein, aber als Reminder)

- **Erste Datenquelle integrieren** — Open-Meteo ist die Phase-1-Wahl.
- **Locations-Suche** real machen (Geocoding, DB-Schema, Endpoint, UI).
- **Authentifizierung** (Sessions oder OAuth/OIDC).
- **Karten-Komponente** mit MapLibre.
- **Transactional-Mail-Provider** wählen (Postmark / Brevo / SES /
  eigener SMTP).

---

Wenn ein Eintrag „dran ist": GitHub-Issue aufmachen, Eintrag mit
Issue-Nr ergänzen oder umziehen lassen. Dieser Backlog ist die low-
ceremony Vorstufe.
