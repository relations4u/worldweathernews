# Deployment

Anleitung für Production-Deployments auf wwn-prod (10.100.100.21) und
wwn-mon (10.100.100.22). Stand: 2026-05-06, nach Abschluss von
Session 11a.

> **Forschungs-Phase, nicht echte Production.** Keine SLA, kein
> automatisches Failover, kein dedizierter DDoS-Schutz. Nutzer werden
> via Banner auf `research.worldweathernews.com` informiert.

## Architektur kurz

- **wwn-prod** — App-Stack (Postgres, Redis, Backend, Frontend,
  Pyworkers) plus eigenständiger Caddy-Stack unter `/srv/wwn/caddy`
- **wwn-mon** — Observability (Prometheus, Loki, Tempo, Grafana)
- Beide sind Proxmox-VMs, hinter einer Hardware-Firewall, NAT 80/443
  zeigt auf wwn-prod

Detail: [`docs/architecture.md`](architecture.md).

## Voraussetzungen

Auf der Maintainer-Maschine:

- mise mit den Tools aus `.mise.toml` installiert
- `~/.config/sops/age/keys.txt` mit dem age-Private-Key (Mode 0600)
- SSH-Zugang zu wwn-prod und wwn-mon mit dem Key, der in
  `infra/ansible/inventories/production/group_vars/all.yml` als
  `maintainer_authorized_keys` eingetragen ist
- Git-Identität konfiguriert (siehe CLAUDE.md → Maintainer-Identität)

## Container-Registry: ghcr.io

Alle drei Service-Images werden via
[Release-Pipeline](../.github/workflows/release.yml) nach
`ghcr.io/relations4u/wwn-{backend,frontend,pyworkers}` gepusht.
Trigger: Tag `v*` auf `main`.

### Auth für den Server-Pull

Auf wwn-prod meldet der `deploy`-User sich gegen ghcr.io mit einem
Token an, das per SOPS in `infra/secrets/production/ghcr.env` liegt.
Die `docker`-Rolle (`infra/ansible/roles/docker/`) liest die Secrets
und macht `docker login`.

**PAT-Scope**: Classic-PAT mit `read:packages` ist der einfache Weg.
Fine-grained PATs müssen `Organization permissions → Packages: Read`
für `relations4u` haben — `Repository permissions` haben **keine**
packages-Variante. (Das hat während Session 11a einen Tag gekostet.)

Token aktualisieren:

```bash
sops infra/secrets/production/ghcr.env
# Editor öffnet entschlüsselte Plaintext, Werte ändern, speichern.
# sops re-encryptet beim Schließen.

# Server-Login refreshen:
ANSIBLE_HOST_KEY_CHECKING=False ansible-playbook \
  -i infra/ansible/inventories/production/hosts.yml \
  infra/ansible/playbooks/site.yml \
  -e ansible_user=hwr --tags docker --limit app
```

### Cosign-Signaturen verifizieren

Alle Images sind keyless via Sigstore signiert:

```bash
cosign verify ghcr.io/relations4u/wwn-backend:0.1.0 \
  --certificate-identity-regexp "^https://github.com/relations4u/worldweathernews/" \
  --certificate-oidc-issuer "https://token.actions.githubusercontent.com"
```

## Erstmaliges Server-Deployment (Bootstrap)

Wenn eine VM frisch ist und nur ein Maintainer-User (`hwr`) existiert:

```bash
# Vom Maintainer-Host aus
cd infra/ansible

# Voller Bootstrap inkl. deploy-User-Anlage, SSH-Hardening, Docker,
# Monitoring-Agent, App-Stack:
ANSIBLE_HOST_KEY_CHECKING=False ansible-playbook \
  -i inventories/production/hosts.yml playbooks/site.yml \
  -e ansible_user=hwr -e target_version=0.0.1-rc4
```

Hintergrund: das Inventory hat per Default `ansible_user=deploy`,
aber der `deploy`-User existiert beim allerersten Lauf noch nicht.
`-e ansible_user=hwr` overridet das nur für diesen Lauf. Sobald
`common`-Role den `deploy`-User angelegt hat, übernimmt der Default.

Hausaufgabe vor dem ersten `terraform apply` (nicht in dieser Session
abgedeckt, aber dokumentiert): bestehende VMs in den Terraform-State
importieren. Siehe `infra/terraform/README.md`.

## Folge-Deployments

Nach einem neuen Release-Tag (z. B. `v0.1.0`):

```bash
bash scripts/deploy.sh production 0.1.0
# Interaktive Bestätigung: "production" tippen.
```

Was der Wrapper macht:

- ruft `playbooks/deploy.yml` mit `-e ansible_user=hwr -e target_version=...`
  auf — der `deploy`-User hat aus Sicherheitsgründen nur
  docker-NOPASSWD und kann keine `/opt/wwn`-Dateitasks
- rendert `compose.prod.yml.j2` mit den Versions-Pins
- pullt die Images, restartet die Container, wartet auf
  `/api/v1/ping` 200

`playbooks/deploy.yml` führt nur die `app`-Rolle aus — common/docker/
monitoring werden nicht angefasst. Wer das Bootstrap-Profil komplett
laufen lassen will, nimmt `playbooks/site.yml`.

## Caddy ist NICHT Teil des App-Stacks

Caddy lebt unter `/srv/wwn/caddy/` mit eigenem Compose-Stack und
eigenem Deploy-Pfad:

```bash
bash infra/deploy/deploy-caddy.sh
```

Das Skript rsynct `infra/caddy/prod/` nach wwn-prod, pullt das Image
und macht `docker compose up -d && docker compose restart caddy`. Der
explizite `restart` ist nötig — Single-File-Bind-Mounts hängen am
Inode beim Container-Start, und rsync's atomic-rename produziert
einen neuen Inode, den der Container ohne Restart nicht sieht. Der
Cert-Volume (`./data` als Verzeichnis-Bind) bleibt davon unberührt.

**Wichtig: Caddy-Cert-Volume schützen.** Vier Let's-Encrypt-
Zertifikate liegen unter `/srv/wwn/caddy/data/`. Verlust = ACME
Re-Issuance, und Let's Encrypt rate-limit'et 50 Issuances/Apex/Woche.

Don'ts:

- ❌ `docker compose down -v` für `wwn-caddy`
- ❌ Snapshot-Rollback auf einen Zeitpunkt VOR Cert-Issuance ohne
  Volume-Restore
- ❌ Caddyfile direkt auf wwn-prod editieren — `--delete` im rsync
  überschreibt das beim nächsten Deploy

## Rollback

### Code-/Image-Rollback

Vorigen Tag finden, redeployen:

```bash
gh release list --limit 5
bash scripts/deploy.sh production 0.0.1-rc3   # vorige Version
```

`compose.prod.yml.j2` wird mit dem alten Pin neu gerendert,
`docker compose pull` zieht das alte Image, der Container startet neu.
Volumes (Postgres, Redis, Caddy-Certs) bleiben unverändert.

### DB-Rollback

```bash
goose -dir infra/migrations down
```

Achtung: Down-Migrations können destruktiv sein. Vor jedem Production-
Down: Snapshot der Postgres-Daten ziehen. Schritt-für-Schritt im
Runbook → "Postgres-Restore aus Backup".

### Caddyfile-Rollback

```bash
git revert <bad-commit>
bash infra/deploy/deploy-caddy.sh
```

## Branch-Protection auf `main`

In der GitHub-UI gepflegt (nicht im Repo dokumentiert, weil API-
Authoring der Settings ein eigener Schmerz ist):

- Direct-Push verboten
- Pull Request mit grüner CI Pflicht
- Signed commits Pflicht
- Linear history (Squash-Merges)

Settings → Branches → `main` → Branch protection rules.

## Smoke-Tests nach Deploy

```bash
# Frontend rendert
curl -fsSI https://research.worldweathernews.com | head -1

# API antwortet mit Trace-ID
curl -fsS https://api.research.worldweathernews.com/api/v1/ping

# CORS-Preflight von Apex aus geht durch zur chi-cors
curl -sI -X OPTIONS \
  -H "Origin: https://worldweathernews.com" \
  -H "Access-Control-Request-Method: GET" \
  https://api.research.worldweathernews.com/api/v1/ping \
  | grep -i access-control

# Cert-Dates haben sich nicht verschoben (gegen ACME-Surprise)
echo | openssl s_client -connect worldweathernews.com:443 \
  -servername worldweathernews.com 2>/dev/null \
  | openssl x509 -noout -dates
```

## Bekannte Stolpersteine

(Vollständigere Liste in CLAUDE.md → Häufige Fallen.)

- **Frontend `PUBLIC_API_BASE_URL`** — build-time. Falscher Wert im
  Bundle = `Failed to fetch` im Browser. Release-Pipeline setzt es
  via `--build-arg`; lokale Builds für andere Umgebungen müssen das
  auch tun.
- **Single-File-Bind-Mounts** — `docker compose restart` nach Config-
  Update bei Caddy und monitoring-stack. Handler/Skript machen das
  schon, aber bei manueller Edit auf dem Server: dran denken.
- **`sudo` über plain SSH (BatchMode)** braucht TTY. Skripte prüfen
  Vorbedingungen mit `test -d` statt `sudo install`. Erstmalige
  Verzeichnis-Anlagen einmalig manuell mit `ssh -t`.
- **`terraform apply` ohne vorigen `import`** — Terraform versucht
  die manuell erstellten VMs neu anzulegen → Konflikt. Workflow im
  `infra/terraform/README.md`.
