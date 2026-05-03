# Session 3 — Compose-Stack mit DB und Redis

**Phase**: B (Services)
**Geschätzte Dauer**: 1-2 Stunden
**Vorbedingung**: Session 2 abgeschlossen, Docker und Docker Compose installiert.

## Ziel

`make dev` startet einen lokalen Stack mit PostgreSQL (mit PostGIS und TimescaleDB),
Redis, Caddy als Reverse-Proxy und Mailhog. Alle Services werden healthy.

App-Services (Backend, Frontend, Workers) sind noch nicht dabei — sie kommen in
Sessions 4-6 dazu.

## Aufgaben

### 1. `infra/compose/compose.dev.yml`

#### Services:

**postgres**:
- Image: `timescale/timescaledb-ha:pg16` (enthält PostGIS und TimescaleDB).
  **Verifiziere das.** Falls dieses Image PostGIS nicht enthält, nutze stattdessen
  `postgis/postgis:16-3.4` und installiere TimescaleDB als Extension via init-Script.
- Container-Name: `wwn-postgres`
- Volume: `postgres_data:/var/lib/postgresql/data`
- Init-Script-Mount: `./postgres-init:/docker-entrypoint-initdb.d:ro`
- Environment aus `.env`:
  - `POSTGRES_USER=wwn`
  - `POSTGRES_PASSWORD=${POSTGRES_PASSWORD}`
  - `POSTGRES_DB=wwn`
- Port-Binding: `127.0.0.1:5432:5432` (nur loopback)
- Healthcheck: `pg_isready -U wwn -d wwn`, interval 5s, retries 10, start_period 10s

**redis**:
- Image: `redis:7-alpine`
- Container-Name: `wwn-redis`
- Command: `redis-server --save "" --appendonly no` (keine Persistenz lokal)
- Port-Binding: `127.0.0.1:6379:6379`
- Healthcheck: `redis-cli ping | grep PONG`, interval 5s

**caddy**:
- Image: `caddy:2-alpine`
- Container-Name: `wwn-caddy`
- Volume:
  - `../caddy/Caddyfile:/etc/caddy/Caddyfile:ro`
  - `caddy_data:/data`
  - `caddy_config:/config`
- Ports:
  - `80:80`
  - `443:443`
- Logs als JSON (in Caddyfile konfiguriert)
- Restart: `unless-stopped`

**mailhog**:
- Image: `mailhog/mailhog`
- Container-Name: `wwn-mailhog`
- Ports:
  - `127.0.0.1:1025:1025` (SMTP)
  - `127.0.0.1:8025:8025` (Web-UI)

#### Networks:
- Ein Default-Network `wwn-net`, alle Services darauf

#### Volumes:
- `postgres_data`
- `caddy_data`
- `caddy_config`

### 2. `infra/compose/postgres-init/01-extensions.sql`

```sql
-- Wird beim allerersten Start ausgeführt, wenn DB-Verzeichnis leer ist
CREATE EXTENSION IF NOT EXISTS postgis;
CREATE EXTENSION IF NOT EXISTS timescaledb;
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS pg_trgm;
```

Falls TimescaleDB-HA-Image: `timescaledb` ist evtl. schon angelegt, dann
`CREATE EXTENSION IF NOT EXISTS` als no-op.

### 3. `infra/caddy/Caddyfile`

```caddy
{
	# Globale Optionen
	auto_https off
	admin off
	log {
		output stdout
		format json
	}
}

# Frontend
app.localhost, localhost:80 {
	respond "503 — frontend not yet running (kommt in Session 5)" 503
}

# Backend API
api.localhost {
	respond "503 — backend not yet running (kommt in Session 4)" 503
}

# Mailhog UI (Convenience für lokale Entwicklung)
mail.localhost {
	reverse_proxy mailhog:8025
}
```

In Sessions 4 und 5 werden die `respond`-Direktiven durch echte `reverse_proxy`-Targets ersetzt.

### 4. `.env.example` im Repo-Root

```bash
# PostgreSQL
POSTGRES_PASSWORD=changeme_in_dev_only

# Backend (kommt in Session 4)
WWN_DATABASE_URL=postgres://wwn:changeme_in_dev_only@postgres:5432/wwn?sslmode=disable
WWN_REDIS_URL=redis://redis:6379/0
WWN_LOG_LEVEL=debug
WWN_LOG_FORMAT=text
WWN_ENVIRONMENT=dev
WWN_HTTP_PORT=8080

# Python Workers (kommt in Session 6)
WWN_PY_DATABASE_URL=postgres://wwn:changeme_in_dev_only@postgres:5432/wwn
WWN_PY_REDIS_URL=redis://redis:6379/0
WWN_PY_LOG_LEVEL=debug
WWN_PY_LOG_FORMAT=text

# Frontend (kommt in Session 5)
PUBLIC_API_BASE_URL=http://api.localhost
```

`.env` muss in `.gitignore` stehen (sollte schon).

### 5. Top-Level `compose.yml`

Entscheidung: **dünne Datei mit `include`** statt Symlink (portabler).

```yaml
# Dieser File ist der Default-Entry für `docker compose`.
# Er includiert das eigentliche Dev-Compose.
include:
  - infra/compose/compose.dev.yml
```

Falls `include` nicht stabil läuft (Compose-Version): stattdessen Hinweis in README,
direkt mit `-f infra/compose/compose.dev.yml` zu arbeiten, oder Symlink + Hinweis
für Windows-User.

### 6. `Makefile`-Updates

Die Targets `dev`, `dev-down`, `dev-reset` aus Session 2 funktionieren jetzt erstmals
echt. Verifiziere dass:

- `make dev` → bringt Stack hoch, tailt Logs
- `make dev-down` → stoppt
- `make dev-reset` → stoppt + löscht Volumes

Optional ergänzen:
- `make dev-logs SERVICE=postgres` → tailt nur einen Service
- `make dev-shell SERVICE=postgres` → öffnet psql/redis-cli/sh

```makefile
dev-logs: ## Logs eines Services tailen (SERVICE=name)
	docker compose logs -f $(SERVICE)

dev-psql: ## psql-Shell auf der DB
	docker compose exec postgres psql -U wwn -d wwn

dev-redis: ## redis-cli auf dem Cache
	docker compose exec redis redis-cli
```

### 7. `infra/compose/README.md`

Erkläre:
- Welche Services laufen, mit welchem Zweck
- Welche Ports gemappt sind
- Warum `127.0.0.1`-Binding (Sicherheit, kein versehentlicher LAN-Zugriff)
- Wie `*.localhost` funktioniert (auf modernen Betriebssystemen automatisch
  auf 127.0.0.1; bei Problemen Hinweis auf `/etc/hosts` oder `dnsmasq`)
- Wie man die Datenbank zurücksetzt
- Wie man init-Scripts ändert (nur bei leerem Volume wirksam)

### 8. Erste Smoke-Tests in der Session

Nach Hochfahren:

```bash
docker compose ps                      # alle healthy
docker compose exec postgres pg_isready -U wwn
docker compose exec redis redis-cli ping
curl -I http://localhost:80            # Caddy antwortet (503 ist OK)
curl http://api.localhost              # 503 mit unserer Message
curl http://mail.localhost             # mailhog UI
```

Outputs zeigen.

## Vorgehen (verbindlich)

1. Plan zeigen, ich review
2. Freigabe abwarten
3. Implementieren
4. **Selbst `make dev` starten und verifizieren**, dass alle Services healthy werden
5. `docker compose ps` zeigen
6. Smoke-Tests laufen lassen, Ergebnisse zeigen
7. Wenn etwas nicht healthy: **fixen, nicht ignorieren**, dann erneut zeigen
8. Nicht committen

## Erfolgs-Kriterien

- [ ] `make dev` startet alle vier Services
- [ ] Alle Healthchecks werden grün innerhalb von 30s
- [ ] PostgreSQL hat PostGIS und TimescaleDB als Extensions geladen
  (verifiziere mit `SELECT extname FROM pg_extension;`)
- [ ] Redis antwortet auf PING
- [ ] Caddy routet `mail.localhost` korrekt zu Mailhog
- [ ] Caddy gibt für `api.localhost` und `app.localhost` die Platzhalter-503 zurück
- [ ] `make dev-down` und `make dev-reset` funktionieren

## Mögliche Stolpersteine

- **`*.localhost`-Routing**: funktioniert nicht überall ohne Konfiguration.
  Auf macOS und Linux meist ja; auf Windows oft `/etc/hosts`-Einträge nötig.
  In README dokumentieren.
- **Caddy ohne TLS lokal**: `auto_https off` ist wichtig, sonst versucht Caddy
  ACME für `*.localhost`, was scheitert.
- **TimescaleDB-Image-Auswahl**: Das HA-Image ist groß. Wenn das Probleme macht
  oder nicht open-source-frei ist, alternativ `timescale/timescaledb:latest-pg16`
  und PostGIS via Init-Script. Verifiziere die Lizenz-Situation kurz und melde
  zurück, wenn unsicher.
- **Volume-Namespace**: bei mehreren Compose-Projekten am selben Host können
  Volume-Namen kollidieren. `name:` im Volume-Block setzen oder explizit
  Compose-Project-Name via `COMPOSE_PROJECT_NAME=wwn` in `.env`.

## Was diese Session NICHT tut

- Keine App-Services (Backend, Frontend, Workers)
- Kein Hot-Reload-Setup für App-Services (kommt mit den Services in 4-6)
- Kein Monitoring (Session 10)
- Keine Production-Compose-Datei (Session 11)

## Suggested Commit-Message

```
feat(infra): add local compose stack with postgres, redis, caddy, mailhog
```
