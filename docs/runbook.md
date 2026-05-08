# Runbook

Was tun, wenn etwas kaputt ist. Pro Szenario: Symptome → Sofort-
maßnahmen → Diagnose-Schritte → Fix → Postmortem-Notiz.

Stand: 2026-05-06. Reflektiert das Setup nach Session 11a (wwn-prod
auf 10.100.100.21, wwn-mon auf 10.100.100.22).

## Schnellzugriffe

| Was              | Wo                                                                                         |
| ---------------- | ------------------------------------------------------------------------------------------ |
| Public Frontend  | <https://research.worldweathernews.com>                                                    |
| Public API       | <https://api.research.worldweathernews.com/api/v1/ping>                                    |
| Grafana (LAN)    | <http://10.100.100.22:3000> (admin + Pass aus `grafana.env`)                               |
| Prometheus (LAN) | SSH-Tunnel: `ssh -L 9090:127.0.0.1:9090 hwr@10.100.100.22`                                 |
| App-Logs         | Grafana → Explore → Loki → `{service="wwn-backend"}`                                       |
| Tempo Traces     | Grafana → Explore → Tempo                                                                  |
| Compose-Stacks   | wwn-prod: `/opt/wwn` (App), `/srv/wwn/caddy` (Caddy); wwn-mon: `/opt/wwn/monitoring-stack` |

## Container-Quick-Reference

```bash
ssh hwr@10.100.100.21 'sudo -u deploy docker ps'      # App-Stack
ssh hwr@10.100.100.21 'sudo -u deploy docker logs wwn-backend --tail 200 -f'
ssh hwr@10.100.100.22 'sudo -u deploy docker ps'      # Monitoring-Stack
```

## Wo finde ich was im Monitoring (Dev und Prod)

| Service      | Dev (Compose-Profile `monitoring`) | Prod                                              |
| ------------ | ---------------------------------- | ------------------------------------------------- |
| Grafana      | <http://localhost:3000>            | <http://10.100.100.22:3000> (LAN-only via UFW)    |
| Prometheus   | <http://localhost:9090>            | SSH-Tunnel oder `curl 127.0.0.1:9090` auf wwn-mon |
| Loki direkt  | <http://localhost:3100>            | `10.100.100.22:3100` (UI ist Grafana → Explore)   |
| Tempo direkt | <http://localhost:3200>            | `10.100.100.22:3200` (UI ist Grafana → Explore)   |
| OTLP gRPC    | `tempo:4317` (intern)              | `10.100.100.22:4317` (von wwn-prod aus)           |

Drei vorprovisionierte Dashboards liegen unter `worldweathernews/`:
**WWN Backend Overview**, **WWN Pyworkers Overview**,
**WWN Infra Overview**.

---

## Szenarien

### 1) "Backend antwortet nicht" / 5xx

**Symptome**: Frontend zeigt _Backend offline_ / API-Curl gibt 5xx
oder Timeout.

**Sofort**:

```bash
# Service-Status prüfen
ssh hwr@10.100.100.21 'sudo -u deploy docker ps --filter name=wwn-backend'

# Public erreichbar?
curl -fsS https://api.research.worldweathernews.com/api/v1/ping
```

**Diagnose**:

1. Container `unhealthy`? → `docker inspect wwn-backend --format '{{.State.Health}}'`
2. Container restartet (`RestartCount > 0`)? → letzten Run-Logs
   anschauen: `docker logs wwn-backend --tail 200`
3. Caddy reverse_proxy zum Backend ok? → `docker logs wwn-caddy --tail 50`
4. DB erreichbar? `docker exec wwn-postgres pg_isready`

**Fix**:

- Bei reinem Crash-Loop: Image-Version checken, ggf. Rollback auf
  vorigen Tag (`bash scripts/deploy.sh production <prev-version>`)
- Bei Caddy-Problem: Caddyfile-Edit + `bash infra/deploy/deploy-caddy.sh`
- Bei DB-Outage: Postgres-Container neu starten (Daten in Volume),
  bei Daten-Korruption → Restore aus Backup (Szenario 10)

**Postmortem**: Trace-ID aus dem Fehler-Log in Tempo öffnen, Span-
Tree dokumentieren.

---

### 2) "Backend offline" im Frontend, Backend-Container ist gesund

**Symptom**: API-Curl funktioniert, aber Frontend-Banner zeigt
"Backend offline / Failed to fetch".

Das ist **kein** Backend-Problem, sondern fast immer eines von vier:

1. **CORS-Preflight schlägt fehl** — Browser blockiert echte Anfrage,
   weil OPTIONS keine `Access-Control-Allow-Origin` zurückgibt.
   Test:

   ```bash
   curl -sI -X OPTIONS \
     -H "Origin: https://worldweathernews.com" \
     -H "Access-Control-Request-Method: GET" \
     https://api.research.worldweathernews.com/api/v1/ping \
     | grep -i access-control
   ```

   Wenn keine `access-control-*`-Zeile dabei ist: Backend hat den
   Origin nicht in `WWN_HTTP_CORSORIGINS`. Fix:
   `sops infra/secrets/production/backend.env`, Origin ergänzen,
   `bash scripts/deploy.sh production <version>`.

2. **Caddy fängt OPTIONS ab** — wenn der `api.research`-Vhost ein
   `respond @options 204` hat, kommt chi-cors gar nicht zum Zug. Soll
   nicht da sein, falls doch: aus Caddyfile entfernen, redeployen.

3. **Frontend-Bundle hat falsche API-URL gebacken** — SvelteKit
   `PUBLIC_API_BASE_URL` ist build-time. Wenn der Release ohne
   `--build-arg` gebaut wurde, steht `http://api.localhost` im JS.
   Test: Frontend-Container inspizieren:

   ```bash
   ssh hwr@10.100.100.21 'sudo -u deploy docker exec wwn-frontend \
     sh -c "grep -r api.localhost /app/build/client/_app/immutable/ | head -3"'
   ```

   Wenn `api.localhost` gefunden: Release.yml-build-args prüfen, neuen
   Tag bauen.

4. **Browser-Cache** — Hard-Refresh (Ctrl/Cmd+Shift+R) probieren.
   SvelteKit serviert HTML mit ETag, Browser kann ältere HTML +
   Asset-Pfade cachen.

---

### 3) "DB-Connections am Limit"

**Symptome**: `pgx: too many connections`, Backend-Logs voll mit
`acquire timeout`.

**Sofort**:

```sql
ssh hwr@10.100.100.21 'sudo -u deploy docker exec wwn-postgres \
  psql -U wwn -c "select count(*) from pg_stat_activity;"'
```

**Diagnose**:

- pgx-Pool-Größe in `apps/backend/internal/config/`
  (`database.maxOpenConns`)
- Postgres-Limit in `infra/migrations/postgres-init/` (default 100)
- Idle-Connections vom Backend? Pool-Lifetime zu hoch?
- Sind viele lang laufende Queries aktiv? (`SELECT * FROM pg_stat_activity
WHERE state != 'idle' AND query_start < now() - '30s'::interval;`)

**Fix**:

- Kurzfristig: Backend restarten (`docker restart wwn-backend`) —
  alle Connections werden frei, Pool wird neu aufgebaut.
- Mittelfristig: Pool-Größe vs. Postgres-Limit balancieren (Backend ≤
  ~25, das Connection-Budget gehört auch noch Workers + Migrations).
- Bei lang laufenden Queries: pg_stat_statements einschalten, Top-
  Queries optimieren.

---

### 4) "Externe Wetter-API gibt 5xx oder Timeout"

**Symptome**: Pyworkers-Job-Run-Rate fällt ab, Loki zeigt
`httpx.RemoteProtocolError` oder `503` aus dem Provider.

**Sofort**: Den Job kann man so lassen — er retry'd beim nächsten
Tick. Wenn der Provider länger weg ist, der Job aber kritisch:

```bash
# Auf einen Failover-Provider wechseln (sobald implementiert), oder
# Job temporär pausieren:
ssh hwr@10.100.100.21 'sudo -u deploy docker exec wwn-pyworkers \
  python -m pyworkers.cli pause <job_id>'
```

**Diagnose**:

1. Grafana → _WWN Pyworkers Overview_ → Panel "Job Run Rate (by status)"
2. Loki: `{service="pyworkers"} | json | level="error"` filtern
3. Trace-ID → Tempo, Auto-Instrumentation zeigt HTTP-Span mit Provider-
   Antwort

**Fix**:

- Provider down → warten, Backoff im httpx-Client schon eingebaut
- Provider rate-limit → Job-Frequenz drosseln
- Schema-Change beim Provider → Parser anpassen, Schema-Test schreiben

---

### 5) "Wir wurden DDoS'd"

**Symptome**: Caddy-Logs explodieren, Latency hoch, normale User
können nicht mehr durch.

**Sofort**:

1. Hardware-Firewall (vor dem Proxmox-Host) — wenn der Provider Notbrems-
   Filter hat, einschalten.
2. Cloudflare-Proxy für die Subdomain einschalten (Free-Plan reicht für
   L7) — DNS-Eintrag von DNS-only auf Proxied umstellen, dann wartet
   man auf TTL-Ablauf.

**Diagnose**:

```bash
# Top Source-IPs aus Caddy-Logs
ssh hwr@10.100.100.21 'sudo -u deploy docker logs wwn-caddy --tail 1000 \
  | jq -r ".request.remote_ip" | sort | uniq -c | sort -rn | head -20'
```

**Fix**:

- Bösartige IPs/Subnetze in Caddy-Block oder UFW droppen
- Bei länger anhaltend: Cloudflare-Proxy bleibt aktiv, ggf. Pro-Plan
  evaluieren

**Postmortem**: Logs sichern (`docker logs --tail all`), in einem
GitHub-Issue Pattern festhalten.

---

### 6) "Disk fast voll"

**Symptome**: `no space left on device` in App-Logs, Postgres oder
Loki schreibt nicht mehr.

**Sofort**:

```bash
ssh hwr@10.100.100.21 'df -h /'
ssh hwr@10.100.100.21 'sudo du -sh /var/lib/docker/volumes/* 2>/dev/null | sort -h | tail'
```

**Diagnose** — typische Übeltäter:

- Postgres WAL nicht archiviert/abgeräumt
- Loki retention zu locker konfiguriert
- Docker dangling images (bei vielen Deploys)
- Tar-Backups, die nie wegrotiert wurden

**Fix**:

```bash
# Dangling Images / Build-Cache
ssh hwr@10.100.100.21 'sudo docker image prune -a -f'
ssh hwr@10.100.100.21 'sudo docker builder prune -f'

# Loki TSDB → Retention-Policy in infra/ansible/roles/monitoring-stack/
# files/loki/local-config.yaml; aktuell ungetuned, Default ist 168h (7d)

# Postgres WAL — pg_archivecleanup oder retain via pg_wal_replay_lsn
```

Mittelfristig: Disk-Nutzung pro Volume in Grafana _Infra Overview_-
Dashboard sichtbar machen, Alert bei > 80 %.

---

### 7) "Ein Container ist im Restart-Loop"

**Symptome**: `docker ps` zeigt `Restarting (1) X seconds ago` für
einen Service.

**Diagnose**:

```bash
ssh hwr@10.100.100.21 'sudo -u deploy docker logs wwn-<service> --tail 200'
ssh hwr@10.100.100.21 'sudo -u deploy docker inspect wwn-<service> \
  --format "{{.State.Health}} {{.State.ExitCode}} {{.State.Error}}"'
```

Häufige Ursachen:

- Healthcheck-Bug (siehe Frontend-IPv4-Issue als Beispiel — Healthcheck
  hit `localhost`, busybox-wget resolved IPv6, Server bindet IPv4)
- Migrations-Fehler beim Start (Backend versucht migrate-on-start)
- Falsche oder fehlende ENV-Variable

**Fix**:

- Healthcheck-Bug → Image fixen, neuen Tag bauen, redeployen
- ENV-Bug → `sops infra/secrets/production/<service>.env`
- Migrations-Bug → manuell `goose -dir infra/migrations <up|down>`

---

### 8) "Memory-Leak in Backend"

**Symptome**: Backend-RSS wächst monoton, Container OOM-killed.

**Sofort**: `docker restart wwn-backend` — Memory wird freigegeben.

**Diagnose**:

1. pprof-Endpoint enabled? In Production aktuell nicht standardmäßig.
   Temporär per ENV einschalten und SSH-Tunnel auf `:6060`.
2. `pprof heap` aufzeichnen:

   ```bash
   ssh -L 6060:127.0.0.1:6060 hwr@10.100.100.21
   curl -o heap.pb.gz http://localhost:6060/debug/pprof/heap
   go tool pprof -http=:8080 heap.pb.gz
   ```

3. Vergleich vor/nach Last → Top-Allocation finden.

**Fix**: meistens `defer` vergessen, ein nicht geschlossener Reader,
oder eine Map ohne TTL. PR mit Leak-Test schreiben.

---

### 9) "Letzter Release ist kaputt — wie rolle ich zurück?"

```bash
# Letzten ge­funden Release-Tag finden
gh release list --limit 5

# Auf den vorigen redeployen
bash scripts/deploy.sh production <previous-tag>
```

Was passiert:

- `compose.prod.yml.j2` wird mit dem alten Pin neu gerendert
- Image wird gepullt, Container restartet, Health gewartet
- Volumes (Postgres, Redis, Caddy) bleiben unverändert

**Wichtig**: Wenn der kaputte Release eine Migration enthielt, Down-
Migration läuft **nicht** automatisch. Manuell:

```bash
ssh hwr@10.100.100.21 'sudo -u deploy docker exec wwn-backend \
  goose -dir /app/migrations down'
```

Vorher Postgres-Snapshot ziehen.

---

### 10) "Postgres-Restore aus Backup"

**Voraussetzung**: Es gibt überhaupt ein Backup. Aktuell **nicht**
automatisiert (steht in `docs/backlog.md`). Workaround bis dahin:

```bash
# Manueller Dump, regelmäßig vom Maintainer-Host triggern
ssh hwr@10.100.100.21 'sudo -u deploy docker exec wwn-postgres \
  pg_dump -U wwn -Fc wwn' > "wwn-$(date +%F).dump"
```

**Restore**:

```bash
# Container muss laufen und DB existieren (oder neu anlegen)
scp wwn-2026-05-06.dump hwr@10.100.100.21:/tmp/
ssh hwr@10.100.100.21 'sudo -u deploy docker cp /tmp/wwn-2026-05-06.dump wwn-postgres:/tmp/'
ssh hwr@10.100.100.21 'sudo -u deploy docker exec wwn-postgres \
  pg_restore -U wwn -d wwn --clean --if-exists /tmp/wwn-2026-05-06.dump'
```

**Postmortem**: Pflicht, Backup-Cadence festzulegen (Cron oder
externer Backup-Worker, siehe Backlog).

---

### 11) "media.worldweathernews.com nicht erreichbar"

**Symptome**: Bilder im Frontend laden nicht, `curl
https://media.worldweathernews.com/site/<key>` gibt 4xx/5xx oder
Timeout.

**Diagnose-Reihenfolge** (von außen nach innen):

```bash
# 1) DNS — löst der Name auf gate.hw7.eu / Public-IP?
dig +short media.worldweathernews.com
# Erwartet: CNAME-Kette media → home → gate, am Ende eine A-Record IP.
# Falls nicht: Cloudflare-DNS-Tab prüfen, CNAME media → home setzen,
# Proxy AUS (graue Wolke).

# 2) Caddy auf wwn-prod erreichbar und hat Cert?
ssh hwr@10.100.100.21 'sudo -u deploy docker logs caddy --tail 100 | grep -i media'
curl -vI https://media.worldweathernews.com/site/test.txt 2>&1 | grep -E '^(<|>)' | head
# Cert sichtbar? Wenn TLS handshake fehlschlägt: Caddy holt LE-Cert
# erst nach erstem HTTPS-Hit; ggf. minutenlang warten und nochmal.
# `sudo -u deploy docker exec caddy ls /data/caddy/certificates/...`
# zeigt ob LE-Issuance erfolgt ist.

# 3) Caddy → Hetzner-Upstream — Host-Header korrekt rewritten?
ssh hwr@10.100.100.21 \
  'sudo -u deploy docker exec caddy curl -sI \
   -H "Host: media-worldweathernews-prod.fsn1.your-objectstorage.com" \
   https://media-worldweathernews-prod.fsn1.your-objectstorage.com/site/test.txt'
# Erwartet: HTTP/2 200. Wenn 400 BadRequest: Bucket-Name in
# infra/caddy/prod/Caddyfile (header_up Host …) prüfen.

# 4) Bucket selbst — Object da, Public-Read aktiv?
aws s3api head-object \
  --bucket media-worldweathernews-prod \
  --key site/test.txt \
  --endpoint-url https://fsn1.your-objectstorage.com
# Wenn 403 ohne Credentials, aber Object existiert: Bucket-Policy
# fehlt oder Pfad nicht in der Whitelist (site/blog/pages/team).
# Re-apply: aws s3api put-bucket-policy --bucket … --policy file://infra/object-storage/bucket-policy.json
```

**Häufigste Ursachen**:

- DNS-Eintrag verschwunden oder versehentlich auf orange-Cloud
  gestellt (Cloudflare-Proxy würde HTTP-01-Challenge brechen).
- Caddy-Stack nach Update nicht neugestartet (Bind-Mount-Inode-Falle
  — `infra/deploy/deploy-caddy.sh` macht `restart caddy`, nicht
  `up -d`).
- Hetzner-Credentials rotiert, Bucket-Policy versehentlich gelöscht.

**Rollback / Workaround**: Wenn das Bucket-Routing kaputt ist und
das Frontend Bilder zeigt, die jetzt nicht laden — Sveltia (kommt
in Iteration 1.3) kennt einen Maintenance-Mode, bis dahin
nicht relevant. Static Fallback-Bilder im Frontend-Build sind
keine geplant.

---

## Bekannte Lücken (in andere Sessions / Folge-PRs verschoben)

- **Caddy-Metriken** sind in `prometheus.yml` auskommentiert — Caddy-
  Admin ist im Caddyfile nicht freigegeben. Folge-PR.
- **Container-CPU/Memory** (cAdvisor / Docker-Stats-Exporter) und
  **postgres_exporter / redis_exporter** sind noch nicht gesetzt —
  das _Infra Overview_-Dashboard hat dafür einen Stub.
- **Backend-/Pyworkers-Metriken von wwn-mon** scrapen nicht — Ports
  binden 127.0.0.1, ufw kennt sie noch nicht oder Push-Sidecar
  fehlt. Notiert in `prometheus.yml`.
- **node-exporter auf wwn-mon** läuft nicht (mon's Stack hat keinen);
  Host-Metriken für wwn-mon sind blind.
- **Sampler** ist `AlwaysSample`. Vor Wachstum auf
  `ParentBased(TraceIDRatioBased(0.1))` umstellen (TODO in
  `tracing.go`).
- **Tempo-Storage** ist lokal-filesystem. Für höhere Volumes gehört
  S3/MinIO dahinter.
- **Postgres-Backups** sind nicht automatisiert.
- **Mailhog ist amd64-only** — Migration auf `axllent/mailpit` offen.
