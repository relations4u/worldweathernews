# wwn-cms-auth — GitHub OAuth Proxy für Sveltia CMS

Minimaler Cloudflare-Worker, der den OAuth-Flow zwischen Sveltia CMS und
GitHub vermittelt. Sveltia kann das Token nicht direkt holen, weil GitHub
keinen `Origin: https://worldweathernews.com` als Redirect-Target zulässt
ohne eine OAuth-App und Client-Secret-Schutz.

Der Worker liegt zwischen beiden, bewahrt das Client-Secret, und stellt
sicher, dass nur erlaubte Origins (`ALLOWED_DOMAINS`) den Flow starten
können.

---

## Deploy (Maintainer-Aufgabe)

Vorbereitung — einmalig pro Cloudflare-Account und GitHub-Org:

### 1. GitHub OAuth-App anlegen

Unter <https://github.com/organizations/relations4u/settings/applications>
(oder im persönlichen Profil unter Settings → Developer settings →
OAuth Apps) eine neue App registrieren:

| Feld                       | Wert                                                          |
| -------------------------- | ------------------------------------------------------------- |
| Application name           | `worldweathernews-cms`                                        |
| Homepage URL               | `https://worldweathernews.com`                                |
| Authorization callback URL | `https://wwn-cms-auth.<dein-cf-account>.workers.dev/callback` |

Nach „Register application": Client ID notieren und ein Client Secret
generieren. Beide werden gleich als Worker-Secrets eingespielt.

### 2. Worker deployen

```bash
cd infra/cloudflare-worker-cms-auth
pnpm install --ignore-workspace   # `--ignore-workspace`, weil das Repo ein
                                  # pnpm-Workspace ist und der Worker bewusst
                                  # davon getrennt deployt
pnpm wrangler login               # einmal pro Maschine, öffnet Browser
pnpm wrangler secret put GITHUB_CLIENT_ID      # Wert aus Schritt 1 einfügen
pnpm wrangler secret put GITHUB_CLIENT_SECRET  # Wert aus Schritt 1 einfügen
pnpm wrangler deploy
```

Nach `wrangler deploy` zeigt die CLI die Worker-URL, z. B.
`https://wwn-cms-auth.<dein-cf-account>.workers.dev`. Diese URL wird
in Schritt 3 in `apps/frontend/static/admin/config.yml` eingetragen.

### 3. Sveltia-Config aktualisieren

In `apps/frontend/static/admin/config.yml` den Platzhalter ersetzen:

```yaml
backend:
  name: github
  repo: relations4u/worldweathernews
  branch: main
  base_url: https://wwn-cms-auth.<dein-cf-account>.workers.dev
  auth_endpoint: auth
```

Commit, mergen, deployen wie üblich (`scripts/deploy.sh production X.Y.Z`).

### 4. Login-Test

1. <https://worldweathernews.com/admin/> öffnen.
2. Auf „Login with GitHub" klicken — Popup öffnet sich.
3. GitHub bittet um Zustimmung für die OAuth-App (einmalig pro User).
4. Popup schließt sich, Sveltia zeigt die `pages_de` und `pages_en`
   Collections.

Wenn der Popup mit „Origin not allowed" abbricht: in `wrangler.toml`
unter `[vars] ALLOWED_DOMAINS` die nutzende Domain ergänzen und
`pnpm wrangler deploy` erneut.

---

## Lokal testen (optional)

```bash
cp .dev.vars.example .dev.vars   # falls Beispiel vorhanden
# .dev.vars: GITHUB_CLIENT_ID=… GITHUB_CLIENT_SECRET=…
pnpm wrangler dev
```

`wrangler dev` läuft auf `http://localhost:8787`. Sveltia kann darauf
zeigen via `base_url: http://localhost:8787` in einer lokalen Variante
der `config.yml`. `localhost` ist als Origin in `ALLOWED_DOMAINS`
freigegeben.

---

## Sicherheits-Hinweise

- **Client Secret rotiert** wenn jemand den Worker-Source bekommt: einfach
  in der GitHub-OAuth-App regenerieren und `wrangler secret put GITHUB_CLIENT_SECRET`
  neu setzen.
- **`ALLOWED_DOMAINS` strict halten** — sonst wird der Worker zum
  Open-OAuth-Relay für jede Site, die `site_id` setzt. Subdomains der
  gelisteten Hosts werden automatisch akzeptiert.
- **CSRF-Schutz** über Secure-HttpOnly-Cookie mit `state`. Der Worker
  validiert beim `/callback` dass der Cookie-State zum URL-State passt.
- **Token-Scope** `repo,user` (Sveltia-Default). Bei privaten Repos
  zwingend; bei rein öffentlichen Repos könnte `public_repo,user` reichen,
  aber Sveltia fragt aktuell `repo` an — also belassen.

---

## Verwandte Dateien

- `apps/frontend/static/admin/index.html` — Sveltia-Loader
- `apps/frontend/static/admin/config.yml` — Sveltia-Konfiguration mit
  `base_url` auf diese Worker-URL
- `docs/cms.md` — Authoring-Workflow inkl. Login-Test
- Upstream-Vorlage: <https://github.com/sveltia/sveltia-cms-auth> (MIT)
