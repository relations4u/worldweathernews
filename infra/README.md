# infra/

Alles, was nicht im Anwendungscode lebt: Compose-Stacks, Reverse-Proxy,
Migrations, Monitoring, Deployment-Skripte, Secrets, Server-Konfiguration.

## Layout

| Pfad          | Zweck                                                                  |
| ------------- | ---------------------------------------------------------------------- |
| `caddy/`      | Caddyfile-Varianten (Dev = Root, Prod-Stub = `prod/`)                  |
| `compose/`    | Compose-Stacks für Dev und Production (App-Services, Monitoring)       |
| `deploy/`     | Stand-alone Deploy-Skripte (Phase vor Ansible/Session 11)              |
| `ansible/`    | Server-Konfiguration (Playbooks, Roles) — wird in Session 11 ausgebaut |
| `terraform/`  | Provisionierung — Proxmox jetzt, Hetzner-Migration später              |
| `migrations/` | goose-DB-Migrationen, sprachunabhängig                                 |
| `monitoring/` | Prometheus/Grafana/Loki/Tempo-Configs                                  |
| `secrets/`    | SOPS-verschlüsselte ENV-Files                                          |

## Stack-Topologie auf wwn-prod (Phase 1, Mai 2026)

```
[Internet] → [Hardware-Firewall] → wwn-prod (10.100.100.70)
                                       │
                                       ├─ /srv/wwn/caddy   (caddy/prod/, network_mode: host)
                                       └─ /srv/wwn/app     (compose/compose.prod.yml, später)
```

Caddy läuft bewusst getrennt vom App-Stack: in Phase 1 reicht ein TLS-
terminierender Stub auf den 4 öffentlichen Hostnames. Der App-Stack zieht
nach, sobald die Service-Images in `ghcr.io` sind und Secrets via SOPS
deployed werden.

## Deploys

| Stack | Wie                                 | Wer      |
| ----- | ----------------------------------- | -------- |
| Caddy | `bash infra/deploy/deploy-caddy.sh` | manuell  |
| App   | _wird Session 11 via Ansible_       | _später_ |

## Caddy: Dev vs. Prod

| Umgebung | Datei                  | Modus                                 |
| -------- | ---------------------- | ------------------------------------- |
| Dev      | `caddy/Caddyfile`      | `auto_https off`, `*.localhost`       |
| Prod     | `caddy/prod/Caddyfile` | Let's-Encrypt-ACME, 4 echte Hostnames |

Der Prod-Caddy läuft als eigener Compose-Stack auf wwn-prod
(`/srv/wwn/caddy`, `network_mode: host`) und wird via
`infra/deploy/deploy-caddy.sh` gepflegt. Er ist bewusst getrennt vom
App-Stack in `compose/compose.prod.yml`.
