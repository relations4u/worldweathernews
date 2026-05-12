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

### 12) "Open-Meteo-Worker liefert keine Daten"

**Symptome**: `/api/v1/locations/{slug}` antwortet mit `current: null`
oder einem `observedAt`, das deutlich älter als 15 min ist. Frontend
zeigt „Noch keine aktuelle Beobachtung vorliegend" auf den Cards.
`forecast` ist leer oder veraltet (forecastFor in der Vergangenheit).

**Diagnose-Reihenfolge** (von DB nach Quelle):

```bash
# 1) DB-Stand prüfen — wann wurde zuletzt eine Beobachtung
# persistiert? Sollte alle 10-15 min Fortschritt zeigen.
docker exec wwn-postgres psql -U wwn -d wwn -c \
  "SELECT l.slug, o.observed_at, o.fetched_at, NOW() - o.fetched_at AS age
   FROM observations o JOIN locations l ON l.id = o.location_id
   WHERE o.observed_at = (SELECT MAX(observed_at) FROM observations
                          WHERE location_id = l.id)
   ORDER BY l.slug;"
# age > 30 min für alle Locations: Worker-Pipeline steht.
# age > 30 min für einzelne: locations.active prüfen.

# 2) Pyworkers-Container gesund + Scheduler läuft?
docker ps --filter name=wwn-pyworkers --format '{{.Status}}'
docker logs --since 5m wwn-pyworkers 2>&1 | grep -E \
  "scheduler_started|open_meteo|error|Exception" | tail -20
# Erwartet: 'scheduler_started' mit open_meteo_enabled=true und
# regelmäßige 'open_meteo_current_persisted'-Logs alle 10 min.

# 3) Metrik-Counter prüfen (wenn Prometheus läuft)
# wwn_open_meteo_fetches_total{kind="current",status="error"}
# steigt → Open-Meteo-API oder DB-Schreibfehler.
# wwn_open_meteo_fetches_total{kind="current",status="ok"}
# stagniert → Scheduler firet nicht.

# 4) Open-Meteo-API direkt prüfen
curl -s "https://api.open-meteo.com/v1/forecast?latitude=52.5&longitude=13.4&current=temperature_2m&timezone=Europe/Berlin"
# Erwartet: JSON mit current.temperature_2m. 5xx oder Timeout →
# Open-Meteo-Down (Statusseite: https://open-meteo.com).

# 5) Worker manuell triggern (umgeht Scheduler-Frage)
docker exec wwn-pyworkers python -c '
import asyncio
from pyworkers.config import load_settings
from pyworkers.storage.postgres import create_pool
from pyworkers.jobs import open_meteo

async def main():
    settings = load_settings()
    pool = await create_pool(str(settings.database_url))
    await open_meteo.run_current(pool)
    await pool.close()

asyncio.run(main())
'
# Wenn das eine neue observation schreibt → Scheduler-Problem
# (APScheduler hängt? Restart hilft).
# Wenn das fehlschlägt → siehe Stack-Trace.
```

**Häufigste Ursachen**:

- **Worker-Container läuft alte Code-Version**: nach Code-Update
  ohne Container-Restart greift der bind-mount-Reload (watchfiles)
  manchmal nicht zuverlässig. `docker compose restart pyworkers`
  forciert einen frischen Start mit neuem Code.
- **DB-Connection-Problem**: asyncpg-Pool-Acquire-Fehler in den
  Logs (z. B. nach Postgres-Restart ohne Worker-Reconnect).
- **Open-Meteo-Rate-Limit**: extrem unwahrscheinlich (10 000/Tag
  Free-Tier, wir liegen bei ~580/Tag), aber bei Spike-Tests
  möglich.
- **Falsche Locations-Konfiguration**: `locations.active = FALSE`
  oder `source != 'open-meteo'` → Worker iteriert über leere Liste.
  Quick-Check: `SELECT slug, active, source FROM locations`.
- **Schema-Drift**: nach Migration ohne Worker-Restart können die
  asyncpg-Prepared-Statements ungültig sein. Worker-Restart.

**Rollback / Workaround**: Open-Meteo ist seit Iteration 2.2 nicht
mehr die einzige Quelle. Für die drei Stadt-Slugs (berlin, hamburg,
potsdam) liefert auch DWD aktuelle Werte; die Backend-API wählt
DWD als Default. Wenn Open-Meteo ausfällt, betrifft das primär die
24-h-Stundenvorhersage (DWD hat 2.2 noch keinen Forecast-Pfad).
Brocken / Helgoland / Zugspitze sind DWD-only — OM-Ausfall berührt
sie nicht. Automatisches Multi-Source-Failover für `current` ist
weiterhin Backlog-Thema.

### 13) "Nach Deploy: API gibt 500 / relation does not exist"

**Symptome**: Direkt nach einem `scripts/deploy.sh production X.Y.Z`
liefert das Backend 500er auf bisher funktionsfähige Endpoints.
`/api/v1/ping` ist noch ok (nutzt keine Tabelle), aber
`/api/v1/locations` oder `/api/v1/locations/{slug}` antworten mit
`{"title":"internal_error","status":500}`. Backend-Logs zeigen
SQLSTATE `42P01` und Meldungen wie
`ERROR: relation "locations" does not exist`.

**Ursache**: Eine neue goose-Migration unter `infra/migrations/` ist
mit dem Release gebaut/getaggt worden, aber auf wwn-prod noch nicht
angewendet. Tritt klassisch beim **ersten** Release auf, der eine
Schema-Änderung mitbringt — passierte konkret bei v0.4.0 am 12. Mai 2026 (Open-Meteo Hello World, erste echte Migration nach
Setup-Phase).

**Diagnose**:

```bash
# 1) Aus Public-View direkt verifizieren (production-User reicht):
curl -s https://api.research.worldweathernews.com/api/v1/locations \
  | jq .
# Erwartet bei Schema-Drift:
# {"title":"internal_error","status":500,"traceId":"..."}

# 2) Auf wwn-prod: Backend-Logs zeigen die genaue Relation:
ssh deploy@wwn-prod 'docker logs --tail 50 wwn-backend 2>&1 | grep -i "relation"'
# z. B.: ERROR: relation "locations" does not exist (SQLSTATE 42P01)

# 3) Postgres-State prüfen — fehlt die Tabelle wirklich?
ssh deploy@wwn-prod 'docker exec wwn-postgres psql -U wwn -d wwn -c "\dt"'
# Wenn nur goose_db_version (oder gar nichts) auftaucht → Migration fehlt.
```

**Sofort-Fix (manuell, wie am 12. Mai durchgeführt)**: postgres
exposed keinen Host-Port — also goose innerhalb des Containers
laufen lassen.

```bash
# Auf dem Maintainer-Host (wwn-dev):
scp $(which goose) deploy@wwn-prod:/tmp/goose
rsync -av infra/migrations/ deploy@wwn-prod:/tmp/wwn-migrations/

# Auf wwn-prod:
ssh deploy@wwn-prod
docker cp /tmp/wwn-migrations wwn-postgres:/tmp/wwn-migrations
docker cp /tmp/goose wwn-postgres:/tmp/goose
docker exec wwn-postgres bash -c '
  /tmp/goose -dir /tmp/wwn-migrations postgres \
    "postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@127.0.0.1:5432/${POSTGRES_DB}?sslmode=disable" \
    up
'
# Cleanup
docker exec wwn-postgres rm -rf /tmp/wwn-migrations /tmp/goose
rm /tmp/goose && rm -rf /tmp/wwn-migrations
```

Backend muss nicht neu gestartet werden — pgx baut neue Statements
beim nächsten Query auf. Bei Workern, die asyncpg-prepared-Statements
cachen (pyworkers), `docker compose restart pyworkers` nach der
Migration.

**Langfristiger Fix**: Migrations sind seit diesem PR Bestandteil
des Ansible-Deploys
(`infra/ansible/roles/app/tasks/main.yml`, Task „Apply database
migrations via goose-in-postgres-container"). Sie laufen
automatisch nach `scripts/deploy.sh production X.Y.Z`, zwischen
„Postgres healthy" und „Backend/Pyworkers starten". Damit ist
dieses Szenario im Normalbetrieb nicht mehr erreichbar — der
Sofort-Fix oben bleibt als Notfall-Rezept für den Fall, dass die
Ansible-Pipeline selbst kaputt ist.

**Smoke-Test nach der Migration**:

```bash
curl -s https://api.research.worldweathernews.com/api/v1/locations \
  | jq '.results | length'
# Erwartet: 6 (Berlin, Brocken, Hamburg, Helgoland, Potsdam, Zugspitze
# ab v0.5.0); 3 vor 2.2.
```

### 14) "DWD-POI-Worker liefert keine Daten"

**Symptome**: `/api/v1/locations/{slug}` antwortet mit `current: null`
für eine DWD-Location, oder das `observedAt` ist deutlich älter als
60 min (DWD-POI publiziert halbstündlich, eine halbe Stunde
Verzögerung ist normal — über 1 h ist verdächtig). `?source=dwd` für
eine OM-+-DWD-Stadt liefert auch keine Werte.

**Diagnose-Reihenfolge** (von DB nach Quelle):

```bash
# 1) DB-Stand pro DWD-Station prüfen
docker exec wwn-postgres psql -U wwn -d wwn -c \
  "SELECT l.slug, l.dwd_station_id,
          MAX(o.observed_at) AS latest,
          NOW() - MAX(o.fetched_at) AS age
   FROM locations l
   LEFT JOIN observations o
     ON o.location_id = l.id AND o.source = 'dwd'
   WHERE l.dwd_station_id IS NOT NULL
   GROUP BY l.id, l.slug, l.dwd_station_id
   ORDER BY l.slug;"
# age > 90 min: Worker-Pipeline steht. age NULL: noch nie gepersistet
# (frische Location? Worker disabled?). Einzelne Stations betroffen:
# DWD-POI-File für diese station_id prüfen (Schritt 4).

# 2) Pyworkers-Container gesund + DWD-Job läuft?
docker logs --since 5m wwn-pyworkers 2>&1 | grep -E \
  "scheduler_started|dwd_poi|dwd_fetches" | tail -20
# Erwartet: 'scheduler_started' mit dwd_enabled=true und alle 30 min
# 'dwd_poi_persisted'-Logs für jede Station.

# 3) Metrik-Counter prüfen (wenn Prometheus läuft)
# wwn_dwd_fetches_total{status="error"} steigt → siehe Worker-Logs
# wwn_dwd_fetches_total{status="empty"} steigt → CSV ist da, aber
#   Parser findet keine Daten-Rows (Header-Format geändert? Station
#   hat keine recent observations?)
# wwn_dwd_fetches_total{status="ok"} stagniert → Scheduler firet nicht.

# 4) DWD-POI-File für eine konkrete Station prüfen
curl -sI "https://opendata.dwd.de/weather/weather_reports/poi/10384-BEOB.csv"
# 200 + content-length ~7000: alles ok. 404: Station nicht (mehr) in
# der DWD-POI-Liste — gegen
# https://opendata.dwd.de/weather/weather_reports/poi/ abgleichen.
# 5xx oder Timeout: DWD-File-Server-Outage (selten, kein Status-Page).

# 5) Worker manuell triggern (umgeht Scheduler-Frage)
docker exec wwn-pyworkers /app/.venv/bin/python -c '
import asyncio, os, asyncpg
from pyworkers.jobs.dwd import run_poi

async def main():
    pool = await asyncpg.create_pool(dsn=os.environ["WWN_PY_DATABASE_URL"])
    await run_poi(pool)
    await pool.close()

asyncio.run(main())'
# Schreibt für alle aktiven DWD-Locations je ~25 Rows; bei Fehlern
# erscheint der Stack-Trace im Container-Log.
```

**Häufigste Ursachen**:

- **Station ist aus dem POI-Verzeichnis verschwunden**: DWD rotiert
  seine POI-Liste gelegentlich. `dwd_station_id` per Hotfix-Migration
  auf eine sinnvolle Ersatz-Station ändern, oder Location auf
  `active = FALSE` setzen.
- **Header-Format geändert**: DWD ändert sehr selten die englischen
  Variablen-Namen in Zeile 1. Sichtbar als `status="empty"`-Metrik
  oder als `temperature=None`-Felder. Parser-Mapping in
  `apps/pyworkers/pyworkers/jobs/dwd.py` (`DWD_COL_*` Konstanten)
  abgleichen mit aktuellem CSV-Header.
- **Worker-Container läuft alte Code-Version**: nach Code-Update
  ohne Container-Restart greift watchfiles nicht zuverlässig.
  `docker compose restart pyworkers` forciert frischen Start.
- **DB-Connection-Problem**: asyncpg-Pool-Acquire-Fehler in den
  Logs. Worker-Restart fixt es meist; sonst Postgres prüfen.
- **PK-Mismatch nach Migration**: wenn jemand `observations` per
  Hand-Migration anfasst und den PK auf `(location_id, observed_at)`
  zurücksetzt, scheitert der DWD-Upsert mit
  `InvalidColumnReferenceError: there is no unique or exclusion
constraint matching the ON CONFLICT specification`. Die korrekte
  PK ist `(location_id, source, observed_at)` seit Migration 0002.

**Rollback / Workaround**: für die drei Stadt-Slugs gibt es
Open-Meteo als Fallback (`?source=open-meteo`). Frontend zeigt für
diese drei dann OM-Daten mit OM-Source-Badge. Für Brocken /
Helgoland / Zugspitze gibt es bei DWD-Ausfall keinen Fallback —
Frontend zeigt „Noch keine aktuelle Beobachtung". Wartebenachrichtigung
auf Statusseite reicht in der Forschungs-Phase.

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
