# wwn-cms-auth

Self-hosted GitHub OAuth proxy for Sveltia CMS.

Hat den ursprünglichen Cloudflare-Worker am 11. Mai 2026 abgelöst, um
die Cloudflare-Abhängigkeit zu reduzieren. Funktional 1:1 — Sveltia
spricht den gleichen postMessage-Handshake.

## Endpoints

| Pfad            | Zweck                                       |
| --------------- | ------------------------------------------- |
| `GET /auth`     | 302 nach `github.com/login/oauth/authorize` |
| `GET /callback` | Tauscht `code` gegen Token, post-Message    |
| `GET /healthz`  | 200 ok (Docker HEALTHCHECK)                 |
| `GET /`         | Plain-Banner für manuelle Smoketests        |

## Konfiguration (ENV)

| Variable                       | Pflicht | Default | Bedeutung                                                   |
| ------------------------------ | ------- | ------- | ----------------------------------------------------------- |
| `WWN_CMS_AUTH_LISTEN`          | nein    | `:8090` | TCP-Listen-Adresse                                          |
| `WWN_CMS_AUTH_PUBLIC_BASE_URL` | ja      | —       | Externe URL (z. B. `https://cms-auth.worldweathernews.com`) |
| `WWN_CMS_AUTH_CLIENT_ID`       | ja      | —       | GitHub OAuth App Client ID                                  |
| `WWN_CMS_AUTH_CLIENT_SECRET`   | ja      | —       | GitHub OAuth App Client Secret                              |
| `WWN_CMS_AUTH_ALLOWED_DOMAINS` | ja      | —       | CSV erlaubter Hostnames für `site_id`                       |
| `WWN_CMS_AUTH_LOG_FORMAT`      | nein    | `json`  | `json` oder `text`                                          |

`ALLOWED_DOMAINS` matched Subdomains automatisch — `worldweathernews.com`
matched auch `blog.worldweathernews.com`.

## Lokal entwickeln

```bash
cd apps/cms-auth
go test ./...
go run ./cmd/cms-auth   # benötigt die Pflicht-ENVs
```

## GitHub OAuth App

In **Settings → Developer settings → OAuth Apps**:

- Homepage URL: `https://worldweathernews.com`
- Authorization callback URL: `https://cms-auth.worldweathernews.com/callback`

`Client ID` und das Generated `Client Secret` landen SOPS-verschlüsselt
in `infra/secrets/production/cms-auth.env`.

## Deployment

- Image: `ghcr.io/relations4u/wwn-cms-auth:<version>` (via Release-Pipeline)
- Container in `infra/compose/compose.prod.yml` und der Ansible-Template-
  Variante unter `infra/ansible/roles/app/templates/compose.prod.yml.j2`
- Bind auf `127.0.0.1:8090` — Caddy proxied von außen
- Caddy-Block: `cms-auth.worldweathernews.com` in `infra/caddy/prod/Caddyfile`
