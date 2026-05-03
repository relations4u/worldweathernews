# Lokaler Compose-Stack

Definiert alle Infrastruktur-Services für die Entwicklung. App-Services
(Backend, Frontend, Workers) kommen ab Session 4 dazu.

Start vom Repo-Root: `make dev`. Stoppen: `make dev-down`.
Komplett zurücksetzen (löscht Volumes): `make dev-reset`.

## Services

| Service  | Image                              | Zweck                                | Host-Ports          |
|----------|------------------------------------|--------------------------------------|---------------------|
| postgres | `timescale/timescaledb-ha:pg16`    | Hauptdatenbank (mit PostGIS + TS)    | `127.0.0.1:5432`    |
| redis    | `redis:7-alpine`                   | Cache, Queue                         | `127.0.0.1:6379`    |
| caddy    | `caddy:2-alpine`                   | Reverse-Proxy auf `*.localhost`      | `80`, `443`         |
| mailhog  | `mailhog/mailhog`                  | SMTP-Catcher + Web-UI für Mail-Tests | `127.0.0.1:1025/8025` |

## Sicherheits-Konventionen

- **Datenbank- und Cache-Ports binden ausschließlich an `127.0.0.1`**, damit
  sie nicht versehentlich aus dem LAN erreichbar sind. Nur Caddy hat
  öffentlich gebundene Ports (80/443).
- `redis` läuft ohne Persistenz (`--save "" --appendonly no`) — lokales
  State ist mit dem Container weg, das ist Absicht.
- **Keine Production-Secrets in `.env`.** Wird in Sessions 11/12 für
  staging/production über SOPS gelöst.

## `*.localhost`-Routing

Caddy bedient `app.localhost`, `api.localhost` und `mail.localhost`.

- macOS und Linux resolven `*.localhost` per Default auf `127.0.0.1`
  (RFC 6761) — keine Konfiguration nötig.
- Falls dein System das nicht tut (z.B. ältere Windows-Versionen), in die
  Hosts-Datei aufnehmen:
  ```
  127.0.0.1  app.localhost api.localhost mail.localhost
  ```

## Datenbank zurücksetzen

```bash
make dev-reset    # stoppt Stack + löscht Volumes (postgres_data, caddy_*)
make dev          # bringt frisch hoch, Init-Scripts laufen erneut
```

## Init-Scripts (`postgres-init/`)

Postgres führt SQL-Files aus `/docker-entrypoint-initdb.d/` **nur beim
allerersten Start** aus, wenn `/var/lib/postgresql/data/` leer ist. Bei
späteren Starts werden Änderungen an den Scripts ignoriert.

Wenn du eine neue Extension oder ein neues Init-SQL brauchst:
1. Datei in `postgres-init/` ergänzen (Lexikographische Reihenfolge: `02-…`).
2. `make dev-reset && make dev`.

Die regulären Schema-Migrations laufen ab Session 4 über `goose` (in
`infra/migrations/`), nicht über Init-Scripts.

## Convenience-Targets

```bash
make dev-logs SERVICE=postgres   # Logs eines Services tailen
make dev-psql                    # psql-Shell auf der DB
make dev-redis                   # redis-cli auf den Cache
```
