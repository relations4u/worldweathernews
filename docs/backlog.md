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
- **Container-Resource-Metrics** (cAdvisor / Docker-Stats-Exporter)
  und **postgres_exporter / redis_exporter** — der _Infra Overview_-
  Dashboard hat dafür Stubs.

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
