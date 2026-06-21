# Handoff-Prompt — Repo `worldweathernews` (für Claude Code)

Du übernimmst das WWN-Monorepo mitten in **Session 13** (Reverse-Proxy-Ingress
zieht auf den Host `gate`). Lies zuerst `CLAUDE.md` (besonders Maintainer-Identität/
SSH-Signing, Versions-Pinning, Workflow-Regeln), dann den Eintrag **Session 13** in
`sessions/STATUS.md` sowie `infra/caddy/prod/Caddyfile` und `infra/caddy/prod/compose.yml`.

## Kontext

Der öffentliche Internet-Ingress wandert von **wwn-prod** auf den neuen Host
**gate** (10.100.100.151), verwaltet im separaten Repo `sysadmin` (Ansible +
Caddy). Entscheidung dort als ADR-0002, **Strategie R1**: gate terminiert
öffentliches TLS + HSTS und reicht alle `worldweathernews.com`-Namen mit
erhaltenem Host-Header an den _internen_ wwn-Caddy durch. Die WWN-Routing-Logik
bleibt hier im Repo — gate dupliziert sie nicht.

## Bereits vorbereitet (UNcommitted, dieser Stand)

`infra/caddy/prod/Caddyfile` + `compose.yml` wurden vom öffentlichen TLS-Terminator
zum **internen HTTP-Router** umgestellt:

- Global: `auto_https off`, `default_bind 10.100.100.70`.
- Alle Sites auf `http://…` (Port 80). HSTS + `email` entfernt (macht jetzt gate).
- Routing/Upstreams **unverändert**: `127.0.0.1:{3000,8080,8090}`, S3-`media`-
  Host-Rewrite, CMS-OAuth, OPTIONS-Pass-through (chi-cors im Backend).
- `network_mode: host` bleibt — sonst kein Zugriff auf die Loopback-Upstreams.
- `default_bind` braucht Caddy ≥ 2.7 (Image `caddy:2-alpine` erfüllt das).

## 🔒 Regeln (aus CLAUDE.md — strikt)

- **Commits signiert** (SSH, `hwr@relations4u.de`). Vor Commit-Vorschlag Identität
  verifizieren (`git config user.email`, `gpg.format=ssh`). Bei Abweichung: STOP.
- **Nie eigenständig committen / nie auf `main` pushen.** Feature-Branch + PR,
  Squash, Conventional Commits. Maintainer committet selbst.
- **Kein `latest`/`stable`** als Pin. (Hinweis: `caddy:2-alpine` ist ein
  schwebender Major-Tag, Altbestand — ggf. als Folge-Punkt voll pinnen.)
- Plan zeigen bei Änderungen > 3 Dateien. Doku Deutsch. Editor vi. Kein
  zsh-Heredoc für YAML.

## Nächste Schritte

1. **Vorbereiteten Caddy-Change reviewen + committen** (signiert, Feature-Branch
   `chore/caddy-internal-router`, Conventional Commit, PR). Vorher validieren:
   `docker compose -f infra/caddy/prod/compose.yml config` und
   `caddy validate --config infra/caddy/prod/Caddyfile --adapter caddyfile`.
2. **Cutover koordinieren** (gemeinsam mit `sysadmin`/gate, siehe dort
   `docs/operations/reverse-proxy-caddy.md`): WWN-Caddy-Deploy
   (`cd /srv/wwn/caddy && docker compose up -d`) + OPNsense-Forward 80/443
   wwn-prod→gate **im selben Wartungsfenster**. Kurzer ACME-Blip auf gate ist ok
   (Forschungs-Instanz ohne SLA). wwn-Caddy danach NICHT abschalten — er ist jetzt
   der interne Router.
3. **Folge-Doku angleichen (eigener Commit, nach Cutover):**
   `docs/architecture.md` (≈ Zeile 65 „App-Stack + Caddy, public via research…"
   und ≈ Zeile 110 „Reverse-Proxy für Apex/www/research/api.research") und
   `docs/deployment.md` beschreiben wwn-prod noch als öffentlichen Ingress →
   auf „gate = Ingress, wwn-Caddy = interner Router" aktualisieren. Prüfen, ob
   `docs/cms.md` / `docs/media-storage.md` Hostnamen-Annahmen betroffen sind.
4. **STATUS.md** Session 13 auf ✅ setzen + Commit-SHA eintragen, wenn durch.

## Verifikation

- Nach Deploy intern: `curl --resolve worldweathernews.com:80:10.100.100.70 http://worldweathernews.com/`.
- Nach Cutover öffentlich: `curl -I https://worldweathernews.com` (+ research.,
  api.research., cms-auth., media.). CMS-Login (Sveltia) und Media-GET prüfen.
