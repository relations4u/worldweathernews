# Architektur

Stand: 2026-06-21. Dieses Dokument beschreibt das Gesamtsystem
worldweathernews.com so, wie es nach Session 13 tatsächlich läuft —
nicht den theoretischen Endzustand. Wenn sich Architektur ändert, wird
diese Datei mitgeführt.

Seit Session 13 (Strategie R1, siehe `sysadmin`-Repo, ADR-0002) ist der
öffentliche Internet-Ingress der eigenständige Host **gate**
(10.100.100.151). gate terminiert öffentliches TLS + HSTS und reicht alle
`worldweathernews.com`-Namen mit erhaltenem Host-Header an den _internen_
wwn-Caddy auf wwn-prod durch. Der wwn-Caddy ist damit kein TLS-Terminator
mehr, sondern interner HTTP-Router (`auto_https off`, `default_bind
10.100.100.70:80`); die WWN-Routing-Logik bleibt unverändert in diesem Repo.

## System-Überblick

```mermaid
graph TB
    subgraph external["External"]
        User[User Browser]
        WeatherAPIs[National Weather Services<br/>DWD, NOAA, Open-Meteo, ...]
    end

    subgraph proxmox["Proxmox-Host (Ryzen 7, 32 GB)"]
        subgraph gatevm["gate (10.100.100.151)"]
            Gate[Caddy 2 — Ingress<br/>public TLS + HSTS<br/>Let's Encrypt]
        end
        subgraph wwnprod["wwn-prod (10.100.100.70)"]
            Caddy[Caddy 2 — interner Router<br/>network_mode: host<br/>auto_https off, :80]
            Frontend[SvelteKit<br/>Node-Adapter<br/>:3000]
            Backend[Go API<br/>Chi + sqlc + pgx<br/>:8080]
            Workers[Python Workers<br/>asyncio + structlog<br/>:9100 metrics]
            Postgres[(PostgreSQL 16<br/>+PostGIS<br/>+TimescaleDB)]
            Redis[(Redis 7)]
            NodeExp[node-exporter<br/>:9101]
            PromtailA[promtail<br/>agent]
        end
        subgraph wwnmon["wwn-mon (10.100.100.69)"]
            Prometheus[Prometheus]
            Loki[Loki]
            Tempo[Tempo]
            Grafana[Grafana<br/>:3000 LAN-only]
            PromtailM[promtail<br/>self]
        end
    end

    User -->|HTTPS| Gate
    Gate -->|HTTP, Host erhalten<br/>10.100.100.70:80| Caddy
    Caddy -->|127.0.0.1:3000| Frontend
    Caddy -->|127.0.0.1:8080| Backend
    Frontend -.->|fetch via api.research| Gate
    Backend --> Postgres
    Backend --> Redis
    Workers --> Postgres
    Workers --> Redis
    Workers -->|HTTP/GRIB| WeatherAPIs

    NodeExp -.->|10.100.100.70:9101| Prometheus
    PromtailA -.->|push| Loki
    PromtailM -.->|push| Loki
    Backend -.->|OTLP| Tempo
    Workers -.->|OTLP| Tempo
    Grafana --> Prometheus
    Grafana --> Loki
    Grafana --> Tempo
```

## Hosts

Die Plattform läuft in der Forschungs-Phase auf eigener Hardware
(Proxmox VE auf Ryzen 7, 32 GB RAM, Hardware-Firewall davor). VMs:

| VM       | IP             | Rolle                                                            | Größe    |
| -------- | -------------- | ---------------------------------------------------------------- | -------- |
| wwn-dev  | 10.100.100.113 | Entwicklung (Editor, mise, Compose-Stack)                        | 8 GB RAM |
| gate     | 10.100.100.151 | Öffentlicher Ingress (Caddy, public TLS + HSTS), `sysadmin`-Repo | s. dort  |
| wwn-prod | 10.100.100.70  | App-Stack + interner Caddy-Router (HTTP :80, kein public TLS)    | 8 GB RAM |
| wwn-mon  | 10.100.100.69  | Observability-Stack (Prometheus/Loki/Tempo/Grafana), LAN only    | 4 GB RAM |

`gate` wird im separaten `sysadmin`-Repo verwaltet (Ansible + Caddy-Rolle,
Entscheidung dort als ADR-0002, Strategie R1). Dieses Repo besitzt nur den
_internen_ wwn-Caddy — die WWN-Routing-Logik wird auf gate nicht dupliziert.

Die Aufgabentrennung wwn-prod/wwn-mon ist bewusst: bei einem Crash auf
wwn-prod bleibt Telemetrie auf wwn-mon abrufbar, plus die hohe I/O-Last
von Prometheus/Loki konkurriert nicht mit Application-Code.

Migration auf Hetzner Cloud ist als Option im Repo skizziert
(`infra/terraform/modules/server-hetzner/`), aber nicht aktiv.

## Service-Verantwortlichkeiten

### Backend (Go 1.25, Chi)

- HTTP-API für Frontend und potentielle Drittsysteme
- AuthN/AuthZ (Sessions, später API-Keys / OAuth)
- Geschäftslogik (Locations, User-Profile, Alerts, …)
- Caching von hot Read-Pfaden (Redis)
- OpenAPI 3.1 ist Single Source of Truth — Server-Stubs werden via
  oapi-codegen generiert (siehe [ADR-0001](adr/0001-openapi-as-source-of-truth.md))
- **Kein** ETL/Batch — das machen die pyworkers

Health: `/api/v1/ping` (öffentlich, gibt JSON mit Trace-ID).

### Pyworkers (Python 3.12, asyncio)

- Pull externer Wetterdaten (DWD, NOAA, Open-Meteo, später EUMETSAT)
- GRIB-/NetCDF-Parsing (xarray, cfgrib)
- Normalisierung in TimescaleDB-Hypertables
- Periodische Aggregationen (Tages-/Monats-Klima-Stats)
- Sentinel/Heartbeat-Job alle 30 s

Metrics: `:9100/metrics`. Scheduler ist APScheduler 3.x AsyncIOScheduler.

### Frontend (SvelteKit, Node-Adapter)

- SSR für SEO und schnellen First-Paint
- Hydration für Interaktivität (Svelte 5 Runes)
- Karten-Komponente (MapLibre, geplant)
- Personalisierung (Locations, Einheiten, Sprache)
- API-Client unter `apps/frontend/src/lib/api/client.ts`,
  `PUBLIC_API_BASE_URL` ist build-time pinned (siehe Caveat unten)

### gate (öffentlicher Ingress, `sysadmin`-Repo)

- Einziger öffentlicher Internet-Ingress seit Session 13 (Strategie R1)
- Terminiert öffentliches TLS via Let's Encrypt und setzt HSTS
- Reicht alle `worldweathernews.com`-Namen mit erhaltenem Host-Header an
  den internen wwn-Caddy (`10.100.100.70:80`) durch
- Alleinige Cert-Instanz — die Let's-Encrypt-Zertifikate liegen auf gate,
  nicht mehr auf wwn-prod
- Verwaltung, Deploy und Cutover-Runbook im `sysadmin`-Repo
  (`docs/operations/reverse-proxy-caddy.md`)

### wwn-Caddy (2-alpine, eigener Compose-Stack — interner Router)

- Interner HTTP-Router für Apex/www/research/api.research/cms-auth/media —
  **kein** öffentliches TLS mehr (`auto_https off`)
- Lauscht nur intern auf `10.100.100.70:80` (`default_bind`); HSTS setzt
  jetzt gate
- Routing/Upstreams unverändert: `127.0.0.1:{3000,8080,8090}`,
  S3-`media`-Host-Rewrite, CMS-OAuth, OPTIONS-Pass-through (chi-cors)
- Läuft weiter mit `network_mode: host` — nötig für die Loopback-Upstreams
- **Nicht** Teil des App-Compose-Stacks — getrennter Lifecycle, eigener
  Deploy-Pfad via `infra/deploy/deploy-caddy.sh`

## Datenflüsse

### Read-Pfad (User schaut Wetter)

```
Browser → gate → wwn-Caddy → Frontend (SSR)
                       ↓
                    Hydration
                       ↓
Browser → gate → wwn-Caddy → Backend → Redis (cache hit)  oder
                       ↓
                  Postgres (cache miss → set Redis → return)
                       ↓
                  Browser
```

### Write-Pfad (User postet Beobachtung)

```
Browser → gate → wwn-Caddy → Frontend → Backend (Auth) → Postgres
```

### Ingest-Pfad (Wetterdaten holen)

```
APScheduler → pyworkers job → External API (HTTP/GRIB)
                                  ↓
                              normalize
                                  ↓
                              Postgres (TimescaleDB Hypertable)
```

## Datenmodell (Skizze)

Konkrete Tabellen kommen mit den ersten Features. Das initiale Modell
sieht so aus:

- `locations` — geographische Orte (UUID, Name, Country, Geo via PostGIS)
- `weather_observations` — TimescaleDB-Hypertable, time-series der
  Beobachtungen pro Location
- `users` — Mitglieder
- `posts` — User-generated Content
- `weather_stations` — citizen-science Stations

Migrations leben in `infra/migrations/` (sprachunabhängig, goose).

## Observability

- **Logs** — strukturiert in JSON: `slog` (Go), `structlog` (Python),
  `pino` (Node, geplant). Promtail liest Container-Logs auf wwn-prod
  und pusht zu Loki auf wwn-mon.
- **Metrics** — Prometheus pull-basiert. node-exporter auf wwn-prod
  ist live; Backend/Pyworkers binden ihre Metrics-Ports aktuell nur
  auf 127.0.0.1 (siehe Caveat).
- **Traces** — OpenTelemetry → OTLP → Tempo. Backend nutzt
  otelchi+otelhttp; Pyworkers das Auto-Instrumentation-Paket. Trace-IDs
  landen in den Logs (Cross-Lookup Loki → Tempo).
- **Dashboards** — drei provisionierte Dashboards in Grafana
  (`Backend Overview`, `Pyworkers Overview`, `Infra Overview`).

Grafana ist auf 0.0.0.0:3000 gebunden und nur aus dem Manager-LAN
(10.100.100.0/24) erreichbar — UFW-Regeln in der `common`-Rolle.

## Caveats und offene Punkte

- **Backend-/Pyworkers-Metrics von wwn-mon aus** sind aktuell nicht
  scrape-bar. Die Metrics-Ports binden 127.0.0.1 only. Entscheidung
  zwischen LAN-Bind+UFW vs. Push-Sidecar (vmagent o. ä.) steht aus.
  Notiert als TODO in `infra/ansible/roles/monitoring-stack/files/prometheus/prometheus.yml`.
- **node-exporter für wwn-mon selbst** läuft nicht im Stack; Host-
  Metriken für wwn-mon sind aktuell blind. Folge-PR.
- **Frontend `PUBLIC_API_BASE_URL` ist build-time** (SvelteKit
  `$env/static/public`). Runtime-ENV in Compose hat **keinen** Effekt.
  Der Wert wird in der Release-Pipeline als `--build-arg` gesetzt
  (`.github/workflows/release.yml`). Wer das Frontend für eine andere
  Umgebung baut, muss den build-arg setzen — sonst landet
  `http://api.localhost` im JS-Bundle.

## Skalierungs-Annahmen

- Single-Host-Deployment ist initial ausreichend
- Horizontale Skalierung (mehrere Backend-Replicas + externer LB) ist
  möglich, aber für die Forschungs-Phase nicht nötig
- TimescaleDB skaliert vertikal weit; Sharding kommt erst bei vielen TB
- Compose → K3s ist der Wachstumspfad, aber bewusst noch nicht
  beschritten ([ADR-0004](adr/0004-compose-before-k3s.md))

## Externe Abhängigkeiten

| Service                     | Zweck                            | Kritikalität     |
| --------------------------- | -------------------------------- | ---------------- |
| Joker.com (Domain + DynDNS) | Domain-Registrar, gate.hw7.eu    | Kritisch         |
| Cloudflare DNS              | DNS Free-Plan, DNS-only Modus    | Kritisch         |
| ProtonMail                  | Mail für `@worldweathernews.com` | Hoch             |
| GitHub                      | Source-Hosting, CI, ghcr.io      | Kritisch         |
| Let's Encrypt               | TLS-Zertifikate via Caddy        | Hoch (ersetzbar) |
| Sigstore                    | Cosign keyless signing           | Mittel           |
| Open-Meteo                  | Wetterdaten Phase 1              | Hoch (ersetzbar) |
| DWD OpenData                | Deutschland-Daten                | Hoch (ersetzbar) |

## Sicherheits-Annahmen

- Hardware-Firewall vor dem Proxmox-Host (NAT 80/443 auf gate, SSH/22 wie bisher)
- UFW auf jedem Host: Default-deny inbound, nur whitelisted Ports
- SSH key-only, fail2ban, Hardening via `common`-Rolle
- Container laufen non-root (eigene UIDs pro Service)
- Secrets ausschließlich SOPS-encrypted in Git
- Signierte Commits Pflicht auf `main`

Details: [`docs/secrets.md`](secrets.md), [`docs/runbook.md`](runbook.md).
