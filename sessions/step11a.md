# Session 11a — Komplettes Deployment auf wwn-prod und wwn-mon

**Phase**: D (Ops)
**Geschätzte Dauer**: 3-5 Stunden, davon ~2h aktive Arbeit, Rest Wartezeiten
**Vorbedingung**: Sessions 10a, 10b, 11 abgeschlossen.

## Ziel

Das gesamte System läuft auf den eigenen Proxmox-VMs:

- **wwn-prod (10.100.100.21):** Application-Stack (Postgres, Redis, Backend,
  Frontend, Pyworkers) hinter dem schon laufenden Caddy. Caddy schaltet von
  Stub-`respond` auf `reverse_proxy` zu den App-Containern um.
- **wwn-mon (10.100.100.22):** Zentraler Observability-Stack (Prometheus,
  Loki, Tempo, Grafana) — intern erreichbar via SSH-Tunnel oder VPN, nicht
  öffentlich.
- **End-to-End-Datenfluss:** Client → Caddy → App-Container; Container-Logs
  und -Traces via Promtail/OTLP → wwn-mon; Metriken via node_exporter +
  Backend-/Pyworkers-Prometheus-Endpoints → wwn-mon.

**Nicht-Ziel:** echte Production-SLA, Backup-Automation, K8s-Migration.
Alles davon kommt später (siehe „Was diese Session NICHT tut").

## Kritisch: SSL-Zertifikate dürfen NICHT verloren gehen

Vier Let's-Encrypt-Zertifikate liegen aktuell in Docker named volume
`wwn_caddy_data` auf wwn-prod (`/var/lib/docker/volumes/wwn_caddy_data/_data/`),
gemounted als `/data` in `wwn-caddy`. Verlust bedeutet Re-Issuance — und
Let's Encrypt rate-limit'et 50 Issuances pro Apex pro Woche, plus 5
Duplicate-Cert pro Woche. Bei wiederholtem Re-Test schnell ausgeschöpft.

### Was die Certs zerstört

- `docker compose down -v` auf den Caddy-Stack (entfernt named volumes)
- `docker volume rm wwn_caddy_data`
- `rm -rf /var/lib/docker/volumes/wwn_caddy_data/`
- Snapshot-Rollback auf einen Zeitpunkt VOR der Cert-Issuance
- Disk-Wipe / VM-Rebuild ohne Volume-Backup

### Schutzmaßnahmen für 11a

1. **Proxmox-Snapshot von wwn-prod** vor jeder Phase, plus zusätzlich vor
   der Bind-Mount-Migration (Phase 2) ein Tar-Backup des Volume-Inhalts.
2. **Bind-Mount-Migration früh** (Phase 2), damit ab dann Backups einfach
   per `tar czf … /srv/wwn/caddy/data/` gehen — nicht mehr quer durch
   `/var/lib/docker/`.
3. **Caddy-Stack wird in Phase 3-4 nicht angefasst** — App-Stack-Deploy ist
   ein separater Compose-Stack, kennt den Caddy-Stack gar nicht.
4. **Caddyfile-Cutover (Phase 5) ist Hot-Reload** — `docker compose up -d`
   nach rsync-Update lässt den Container leben, das Volume bleibt
   unverändert. Kein `docker compose down`.
5. **Verifikation nach Cutover:** `notBefore` der Zertifikate muss identisch
   zum Pre-Cutover-Stand sein. Wenn das Datum sich verschoben hat → Caddy
   hat neue Certs gezogen → entweder waren die alten weg oder Caddyfile
   hatte falsche Hostnames.

### Don't-Liste für die ganze Session

- ❌ `docker compose down -v` für `wwn-caddy` (oder einen anderen Stack mit
  geteilten Volumes)
- ❌ `docker volume prune` ohne vorherige Volume-Liste-Inspektion
- ❌ Manuelle Edits in `/var/lib/docker/volumes/wwn_caddy_data/_data/` —
  nur via Volume-Bind-Mount oder `docker exec`
- ❌ Snapshot-Rollback auf einen Stand VOR der Cert-Issuance ohne
  Volume-Backup-Restore
- ❌ Caddyfile-Edits direkt auf wwn-prod (gehen beim nächsten rsync verloren) —
  immer im Repo editieren, dann `deploy-caddy.sh`

## Vor-Klärungen (geklärt)

1. **Monitoring-Stack-Rolle für wwn-mon:** wird in dieser Session als neue
   Ansible-Rolle `roles/monitoring-stack/` geschrieben. Adaptiert vom
   `monitoring`-Profile in `infra/compose/compose.dev.yml`. Konfigurations-
   Files unter `infra/monitoring/` (Prometheus, Loki, Tempo, Grafana
   Provisioning) werden vom Repo nach wwn-mon synchronisiert.
2. **Cert-Volume-Strategie:** Migration von Docker named volume zu
   Bind-Mount `/srv/wwn/caddy/data:/data`. Einfacheres Backup, klarere
   Eigentümerschaft, weniger Überraschung bei `compose down -v`.
3. **App-Backend Health-Endpoint:** `/api/v1/ping` (per existierender
   app-Rolle). Wenn der noch nicht existiert: in der Backend-Codebasis
   ergänzen, separater PR vor 11a-Implementierung.

## Phasen

### Phase 0 — Pre-flight (Maintainer, manuell)

**Aktionen:**

1. Real-Werte in SOPS-Secrets eintragen:

   ```bash
   cd ~/projects/worldweathernews
   sops infra/secrets/production/postgres.env       # POSTGRES_PASSWORD
   sops infra/secrets/production/backend.env        # WWN_DATABASE_URL mit echtem PW
   sops infra/secrets/production/pyworkers.env      # WWN_PY_DATABASE_URL mit echtem PW
   sops infra/secrets/production/ghcr.env           # GHCR_TOKEN (PAT mit read:packages)
   ```

   `frontend.env` und `proxmox.env` brauchen keine echten Werte; `proxmox.env`
   wird in 11a nicht genutzt (kein `terraform apply`).

2. Verifikation:

   ```bash
   bash scripts/check-encrypted-secrets.sh infra/secrets/production/*.env
   sops --decrypt infra/secrets/production/postgres.env  # echte Werte sichtbar?
   ```

3. Real-Image-Versionen in `ghcr.io/relations4u/`:

   ```bash
   gh release list --repo relations4u/worldweathernews --limit 5
   # Aktuelle Tag wählen, z. B. v0.1.0 → für deploy.sh als Argument
   ```

4. Proxmox-Snapshots vor Beginn:
   - wwn-prod: `pre-deploy-11a`
   - wwn-mon: `pre-deploy-11a`

5. Caddy-Volume-Tar-Backup auf wwn-prod (zusätzlich zum Snapshot):
   ```bash
   ssh hwr@10.100.100.21 'sudo tar czf /srv/wwn/caddy-data-backup-$(date +%F).tar.gz \
       -C /var/lib/docker/volumes/wwn_caddy_data/_data .'
   ssh hwr@10.100.100.21 'ls -la /srv/wwn/caddy-data-backup-*.tar.gz'
   ```
   Backup zusätzlich nach lokal kopieren (für Off-Host-Sicherheit):
   ```bash
   scp hwr@10.100.100.21:/srv/wwn/caddy-data-backup-*.tar.gz ~/backups/
   ```

**SSL-Risiko:** keiner.

**Erfolgs-Check:** Snapshots in der Proxmox-UI sichtbar, Tar-Backup
≥ 100 KB groß, decrypt der SOPS-Files zeigt echte Passwörter.

---

### Phase 1 — Ansible-Bootstrap beider VMs

**Ziel:** `deploy`-User anlegen, SSH-Hardening, ufw, fail2ban,
unattended-upgrades, chrony — auf beiden Hosts. Docker ist auf beiden
schon installiert (Session 10a) — die `docker`-Rolle muss idempotent
sein, was sie via apt-Module ist.

**Aktionen:**

```bash
cd infra/ansible
ansible-galaxy collection install -r requirements.yml

# Dry-run zuerst, mit ansible_user=hwr (deploy gibt's noch nicht):
ansible-playbook -i inventories/production/hosts.yml playbooks/site.yml \
    -e ansible_user=hwr --tags common,docker --check --diff

# Wenn das Diff plausibel ist, real anwenden:
ansible-playbook -i inventories/production/hosts.yml playbooks/site.yml \
    -e ansible_user=hwr --tags common,docker
```

Nach erfolgreichem Lauf existiert auf beiden VMs der `deploy`-User mit
dem Maintainer-SSH-Key. Future-Runs brauchen das `ansible_user=hwr`-
Override nicht mehr.

```bash
# Login als deploy testen:
ssh deploy@10.100.100.21 'whoami && groups'   # erwartet: deploy, docker
ssh deploy@10.100.100.22 'whoami && groups'
```

**SSL-Risiko:** keiner — Caddy-Stack bleibt unangetastet, der eigene
Compose-Verbund unter `/srv/wwn/caddy` läuft weiter.

**Rollback:**

- Bei Fehler in der `common`-Rolle (z. B. ufw sperrt SSH aus):
  Snapshot-Rollback wwn-prod / wwn-mon auf `pre-deploy-11a`.
- Bei kleineren Fehlern (z. B. einzelner Task failt): Task fixen,
  Playbook idempotent erneut laufen lassen.

**Verifikation:**

- `ssh deploy@10.100.100.21` und `ssh deploy@10.100.100.22` funktioniert.
- `ufw status` zeigt nur 22/80/443 inbound erlaubt.
- `systemctl is-active fail2ban chrony` = active.

---

### Phase 2 — Caddy-Volume zu Bind-Mount migrieren

**Ziel:** Cert- und Caddy-Config-Daten von Docker named volume nach
Bind-Mount unter `/srv/wwn/caddy/data` und `/srv/wwn/caddy/config`
verschieben — einfacheres Backup, klarere Datei-Eigentümerschaft.

**Aktionen (manuell, ein-malig):**

1. Caddy-Stack stoppen (Container weg, Volumes bleiben):

   ```bash
   ssh hwr@10.100.100.21 'cd /srv/wwn/caddy && docker compose stop caddy'
   ```

2. Volume-Inhalt nach Bind-Mount-Pfad kopieren (`-a` erhält Permissions):

   ```bash
   ssh hwr@10.100.100.21 '
       sudo install -d -o hwr -g hwr -m 0750 /srv/wwn/caddy/data /srv/wwn/caddy/config
       sudo cp -a /var/lib/docker/volumes/wwn_caddy_data/_data/.   /srv/wwn/caddy/data/
       sudo cp -a /var/lib/docker/volumes/wwn_caddy_config/_data/. /srv/wwn/caddy/config/
       sudo chown -R hwr:hwr /srv/wwn/caddy/data /srv/wwn/caddy/config
   '
   ```

3. Verifizieren, dass die vier Cert-Verzeichnisse drin sind:

   ```bash
   ssh hwr@10.100.100.21 'ls /srv/wwn/caddy/data/caddy/certificates/acme-v02.api.letsencrypt.org-directory/'
   # erwartet: 4 Verzeichnisse für die 4 Hostnames
   ```

4. `infra/caddy/prod/compose.yml` lokal editieren — Volumes-Block ersetzen:

   **Vorher:**

   ```yaml
   volumes:
     - ./Caddyfile:/etc/caddy/Caddyfile:ro
     - caddy_data:/data
     - caddy_config:/config
   ...
   volumes:
     caddy_data:
       name: wwn_caddy_data
     caddy_config:
       name: wwn_caddy_config
   ```

   **Nachher:**

   ```yaml
   volumes:
     - ./Caddyfile:/etc/caddy/Caddyfile:ro
     - ./data:/data
     - ./config:/config

   # Top-level volumes-Block entfällt komplett.
   ```

5. Deploy:

   ```bash
   bash infra/deploy/deploy-caddy.sh
   # Skript syncen die compose.yml + Caddyfile, dann docker compose up -d
   # Caddy startet mit den bind-mounted Daten — gleiche Certs wie vorher.
   ```

6. Verifikation: `notBefore` darf sich nicht geändert haben.

   ```bash
   for h in worldweathernews.com research.worldweathernews.com api.research.worldweathernews.com; do
       echo "=== $h ==="
       echo | openssl s_client -connect "$h:443" -servername "$h" 2>/dev/null \
         | openssl x509 -noout -dates
   done
   # notBefore = ursprüngliches Datum (6. Mai), KEIN frisches Datum.
   ```

7. Wenn alles passt: alte Volumes nach 1-2 Tagen Beobachtungszeit entfernen
   (NICHT sofort — Sicherheitsnetz):
   ```bash
   ssh hwr@10.100.100.21 'docker volume rm wwn_caddy_data wwn_caddy_config'
   ```

**SSL-Risiko: HOCH.** Hier liegt der riskanteste Schritt der Session.
Mitigation: Tar-Backup aus Phase 0, Proxmox-Snapshot, sorgfältige
notBefore-Verifikation.

**Rollback:**

- Wenn `cp -a` fehlschlägt oder Caddy nach Restart die Certs nicht findet:
  ```bash
  cd /srv/wwn/caddy
  # alte compose.yml zurück (git pull oder manuell), dann:
  docker compose up -d
  # Caddy nutzt wieder die alten named volumes — Certs sind dort noch.
  ```
- Wenn die Volumes versehentlich entfernt wurden: Tar-Backup aus Phase 0
  zurück nach `/srv/wwn/caddy/data/`.

**Verifikation:**

- `curl -sSI https://worldweathernews.com` antwortet 200, Server: Caddy.
- `openssl s_client | openssl x509 -dates` zeigt unverändertes
  `notBefore` (6. Mai 2026).
- `docker exec wwn-caddy ls /data/caddy/certificates/.../` zeigt die
  vier Hostname-Verzeichnisse.

---

### Phase 3 — Monitoring-Stack auf wwn-mon (neue Ansible-Rolle)

**Ziel:** Prometheus, Loki, Tempo, Grafana laufen auf wwn-mon. Konfigs
adaptiert vom existierenden dev-`monitoring`-Profile.

**Code-Aufgabe (eigener PR vor der Deploy-Aktion):**

Neue Rolle `infra/ansible/roles/monitoring-stack/`:

```
infra/ansible/roles/monitoring-stack/
├── defaults/main.yml             # Ports, Retention-Settings
├── tasks/main.yml                # Compose deployen, configs syncen
├── templates/
│   └── compose.yml.j2            # adaptiert von compose.dev.yml
└── files/
    ├── prometheus/prometheus.yml
    ├── loki/local-config.yaml
    ├── tempo/tempo.yaml
    ├── promtail/config.yaml      # für Self-Monitoring von wwn-mon
    └── grafana/
        ├── provisioning/         # Datasources Loki/Prom/Tempo
        └── dashboards/           # aus infra/monitoring/grafana/dashboards/ kopiert
```

`compose.yml.j2` läuft ohne `network_mode: host`, alle Services in einem
Bridge-Netz. Ports an 127.0.0.1 binden (interner Zugriff per
SSH-Tunnel), KEIN Public-Forward.

`playbooks/site.yml` erweitern:

```yaml
- name: Monitoring stack on wwn-mon
  hosts: mon
  become: true
  roles:
    - monitoring-stack
```

`group_vars/all.yml` ergänzen:

```yaml
monitoring_app_dir: /opt/wwn/monitoring-stack
monitoring_grafana_admin_password: "{{ vault_grafana_admin_password }}" # aus SOPS
```

Neues SOPS-File `infra/secrets/production/grafana.env` mit
`GRAFANA_ADMIN_PASSWORD`.

**Deploy-Aktion (nachdem die Rolle gemerged ist):**

```bash
cd infra/ansible
ansible-playbook -i inventories/production/hosts.yml playbooks/site.yml \
    --limit mon --tags monitoring-stack --check --diff
ansible-playbook -i inventories/production/hosts.yml playbooks/site.yml \
    --limit mon --tags monitoring-stack
```

Verifikation per SSH-Tunnel:

```bash
ssh -L 3000:127.0.0.1:3000 -L 9090:127.0.0.1:9090 -L 3100:127.0.0.1:3100 \
    deploy@10.100.100.22
# Im zweiten Terminal:
curl -sSI http://127.0.0.1:9090/-/ready    # Prometheus
curl -sSI http://127.0.0.1:3100/ready      # Loki
curl -sSI http://127.0.0.1:3000/api/health # Grafana
```

**SSL-Risiko:** keiner — wwn-mon hat keine Certs.

**Rollback:** wwn-mon Snapshot zurück auf `pre-deploy-11a`. Stack stoppen
mit `docker compose down`.

**Verifikation:** Grafana erreichbar via Tunnel, alle drei Datasources
(Prometheus/Loki/Tempo) zeigen „Connected" im Setup.

---

### Phase 4 — App-Stack auf wwn-prod

**Ziel:** Postgres, Redis, Backend, Frontend, Pyworkers laufen via
existierender `app`-Rolle. Container exposen Backend auf
127.0.0.1:8080 und Frontend auf 127.0.0.1:3000 — bereit für Caddy-
Cutover in Phase 5.

**Aktionen:**

1. SOPS-Files mit Real-Werten sind schon befüllt (Phase 0).

2. Image-Versionen prüfen — die Images müssen in ghcr.io
   `relations4u/wwn-{backend,frontend,pyworkers}` mit dem gewählten Tag
   existieren:

   ```bash
   for s in backend frontend pyworkers; do
       gh api -H "Accept: application/vnd.github+json" \
           "/orgs/relations4u/packages/container/wwn-$s/versions" \
           --jq '.[0:3] | .[] | .metadata.container.tags'
   done
   ```

3. Deploy via Wrapper:

   ```bash
   bash scripts/deploy.sh production 0.1.0
   # interaktive Bestätigung: "production" tippen
   ```

4. Container-Status:

   ```bash
   ssh deploy@10.100.100.21 'cd /opt/wwn && docker compose -f compose.prod.yml ps'
   ```

5. Backend-Health:
   ```bash
   ssh deploy@10.100.100.21 'curl -sS http://127.0.0.1:8080/api/v1/ping'
   # erwartet 200 OK mit JSON-Body
   ```

**SSL-Risiko:** keiner — Caddy-Stack ist getrennt.

**Rollback:** `ansible-playbook playbooks/rollback.yml` mit voriger
Version. Im Notfall: wwn-prod Snapshot zurück (kostet die Caddy-
Volumes-Migration aus Phase 2 — Tar-Backup einspielen).

**Verifikation:**

- 5 Container laufen (postgres, redis, backend, frontend, pyworkers).
- Postgres-Healthcheck grün.
- Backend antwortet auf `/api/v1/ping`.
- Frontend antwortet auf `/` mit SvelteKit-HTML.
- Pyworkers-Heartbeat in den Logs.

---

### Phase 5 — Caddy-Cutover (Stub → Reverse-Proxy)

**Ziel:** Caddyfile von `respond "wwn caddy stub: ..."` auf echtes
`reverse_proxy` zu den App-Containern auf 127.0.0.1 umschreiben.
Hot-Reload via `docker compose up -d` — Container bleibt am Leben,
Volume bleibt unverändert, Certs bleiben.

**Aktionen:**

1. `infra/caddy/prod/Caddyfile` editieren — pro Hostname `respond` durch
   `reverse_proxy` ersetzen:

   ```caddy
   worldweathernews.com {
       import hsts
       reverse_proxy 127.0.0.1:3000
   }

   www.worldweathernews.com {
       import hsts
       redir https://worldweathernews.com{uri} permanent
   }

   research.worldweathernews.com {
       import hsts
       reverse_proxy 127.0.0.1:3000
   }

   api.research.worldweathernews.com {
       import hsts
       encode zstd gzip
       reverse_proxy 127.0.0.1:8080

       @options method OPTIONS
       respond @options 204
   }
   ```

2. Vor dem Deploy: `notBefore` der vier Certs notieren:

   ```bash
   for h in worldweathernews.com www.worldweathernews.com \
            research.worldweathernews.com api.research.worldweathernews.com; do
       echo -n "$h: "
       echo | openssl s_client -connect "$h:443" -servername "$h" 2>/dev/null \
         | openssl x509 -noout -dates | grep notBefore
   done | tee /tmp/pre-cutover-cert-dates.txt
   ```

3. Deploy via Caddy-Skript (rsync + `docker compose up -d`):

   ```bash
   bash infra/deploy/deploy-caddy.sh
   ```

   Caddy reloaded den Caddyfile — der Container bleibt am Leben, das
   Volume `/srv/wwn/caddy/data` (jetzt Bind-Mount) wird nicht angefasst.

4. Verifikation, dass Cert-Daten unverändert sind:

   ```bash
   for h in worldweathernews.com www.worldweathernews.com \
            research.worldweathernews.com api.research.worldweathernews.com; do
       echo -n "$h: "
       echo | openssl s_client -connect "$h:443" -servername "$h" 2>/dev/null \
         | openssl x509 -noout -dates | grep notBefore
   done | tee /tmp/post-cutover-cert-dates.txt

   diff /tmp/pre-cutover-cert-dates.txt /tmp/post-cutover-cert-dates.txt && \
       echo "✅ Certs unverändert" || echo "❌ Cert-Daten haben sich geändert!"
   ```

5. End-to-End-Test:
   ```bash
   curl -sSI https://research.worldweathernews.com    # SvelteKit-Headers
   curl -sS  https://api.research.worldweathernews.com/api/v1/ping  # JSON
   curl -sSI https://worldweathernews.com             # 200, Frontend
   curl -sSI https://www.worldweathernews.com         # 301 → Apex
   ```

**SSL-Risiko: niedrig** dank Hot-Reload, aber durch die diff-Verifikation
wasserdicht messbar.

**Rollback:**

- Caddyfile lokal auf vorigen Stand zurücksetzen (git revert), erneut
  `bash infra/deploy/deploy-caddy.sh`. Caddy reloaded auf den Stub-Stand;
  Stub-Antworten kommen wieder.
- Wenn Caddy nach Reload nicht hochkommt: `docker compose logs caddy`,
  Caddyfile-Syntax prüfen, fixen, erneut deployen. Container läuft im
  Zweifel weiter mit alter Config (Caddy-Reload-Verhalten bei
  Konfig-Fehlern).

**Verifikation:**

- Alle 4 Hostnames antworten mit App-Content statt Stub-Strings.
- `diff` aus Schritt 4 zeigt keine Änderung der `notBefore`-Daten.
- Caddy-Logs zeigen reverse_proxy-Requests, keine ACME-Re-Issuance.

---

### Phase 6 — End-to-End-Verifikation und Post-Deploy-Snapshot

**Aktionen:**

1. Funktionale Smoke-Tests:

   ```bash
   # Frontend lädt
   curl -fsSI https://research.worldweathernews.com | head -1
   curl -fsSI https://worldweathernews.com | head -1

   # API antwortet
   curl -fsS  https://api.research.worldweathernews.com/api/v1/ping
   curl -fsSI https://api.research.worldweathernews.com/api/v1/locations?q=berlin

   # www-Redirect
   curl -fsSI https://www.worldweathernews.com | head -3
   ```

2. Monitoring-Sicht prüfen (via SSH-Tunnel auf wwn-mon → Grafana):
   - Backend-Dashboard: Request-Rate > 0 (von den Smoke-Tests)
   - Pyworkers-Dashboard: Heartbeat-Counter steigt
   - Loki: Logs der drei Service-Container indexiert
   - Tempo: Traces der API-Requests sichtbar
   - Cross-Lookup Loki → Tempo via Trace-ID funktioniert

3. Cert-Renewal-Sanity-Check (nicht zwingend für 11a, aber nice):

   ```bash
   ssh deploy@10.100.100.21 'cd /srv/wwn/caddy && \
       docker compose logs caddy 2>&1 | grep -i "cert.*renew\|tls.cache" | tail -5'
   ```

4. Proxmox-Snapshots `post-deploy-11a` für beide VMs.

5. STATUS.md und CLAUDE.md updaten:
   - Session 10c → ✅ (in 11a abgedeckt)
   - Session 11a → ✅
   - Snapshot-Namen, Deploy-Datum, eingesetzte Image-Versionen.

**SSL-Risiko:** keiner.

**Erfolgs-Check:** alle vier Hostnames servieren App-Content,
`notBefore`-Diff aus Phase 5 zeigt keine Änderung, Monitoring-Stack
zeigt Live-Daten, Snapshots gesetzt.

---

## Snapshot-Strategie kompakt

| Zeitpunkt                            | wwn-prod                                           | wwn-mon           |
| ------------------------------------ | -------------------------------------------------- | ----------------- |
| vor Phase 0                          | `pre-deploy-11a`                                   | `pre-deploy-11a`  |
| zusätzlich vor Phase 2               | (Tar-Backup `caddy-data-backup-YYYY-MM-DD.tar.gz`) | —                 |
| nach erfolgreicher Phase 6           | `post-deploy-11a`                                  | `post-deploy-11a` |
| Snapshot `caddy-online` (vom 6. Mai) | bleibt — als Pre-Migration-Anker erhalten          | n/a               |

## Erfolgs-Kriterien (Checkliste am Ende)

- [ ] Alle 4 öffentlichen Hostnames servieren App-Content (nicht Stub)
- [ ] `notBefore` der LE-Certs unverändert seit dem 6. Mai 2026
- [ ] Caddy-Volume liegt als Bind-Mount unter `/srv/wwn/caddy/data`
      und `/srv/wwn/caddy/config`
- [ ] App-Stack auf wwn-prod: 5 Container `Up (healthy)`, Backend-Ping
      antwortet
- [ ] Monitoring-Stack auf wwn-mon: Prometheus/Loki/Tempo/Grafana
      erreichbar via SSH-Tunnel; alle drei Datasources in Grafana
      „Connected"
- [ ] node_exporter auf wwn-prod scraped, Promtail pusht Logs zu Loki
      auf wwn-mon
- [ ] Snapshots `post-deploy-11a` auf beiden VMs gesetzt
- [ ] Tar-Backup `caddy-data-backup-YYYY-MM-DD.tar.gz` lokal
      auf Maintainer-Maschine (Off-Host-Sicherheit)
- [ ] CLAUDE.md und sessions/STATUS.md sind aktualisiert
- [ ] Alte Docker named volumes `wwn_caddy_data`, `wwn_caddy_config`
      nach Beobachtungszeit (~1-2 Tage) entfernt

## Common Pitfalls (aus Lessons-Learned-Antizipation)

- **`docker compose down -v` aus Reflex** beim Caddy-Stack — niemals tun.
  Wenn ein Container neu gestartet werden muss: `docker compose restart`
  oder `docker compose up -d --force-recreate <service>`, beides ohne
  Volume-Verlust.
- **`cp -a` ohne Punkt am Ende des Quellpfads** — das `.` ist wichtig,
  sonst wandert der Inhalt eine Hierarchie-Ebene tiefer
  (`/srv/wwn/caddy/data/_data/...` statt `/srv/wwn/caddy/data/...`).
- **Caddyfile-Syntax-Fehler im Reverse-Proxy** — vor dem Deploy lokal
  validieren:
  ```bash
  docker run --rm -v $PWD/infra/caddy/prod/Caddyfile:/etc/caddy/Caddyfile:ro \
      caddy:2-alpine caddy validate --config /etc/caddy/Caddyfile --adapter caddyfile
  ```
- **`network_mode: host` Caddy proxied auf Container-IP statt Loopback** —
  beim Wechsel zu reverse_proxy unbedingt 127.0.0.1:PORT, nicht den
  Container-Namen. Caddy auf dem Host kann den Service-DNS-Namen aus dem
  app-Bridge-Netz NICHT auflösen, deshalb müssen die App-Container die
  Ports auf 127.0.0.1 binden (was die `compose.prod.yml.j2` schon tut).
- **CORS-Konfiguration im Backend** — wenn das Frontend auf
  `research.` läuft und die API auf `api.research.` antwortet, muss
  das Backend `Access-Control-Allow-Origin: https://research.worldweathernews.com`
  setzen. Vor dem Cutover prüfen, dann der echte Reverse-Proxy bringt
  das schnell zu einem Browser-Fehler, falls vergessen.
- **Image-Tag-Mismatch zwischen `gh release list` und ghcr.io** — Tags
  in der GitHub-Release-UI sind nicht automatisch in ghcr.io. Für jeden
  Tag braucht es einen erfolgreichen Release-Workflow-Run. Vor Phase 4
  via `gh api /orgs/.../packages/...` verifizieren.
- **Grafana-Admin-Passwort als Default `admin`** — niemals so in
  Production. Über SOPS-File rotieren, vor erstem Login.

## Was diese Session NICHT tut

- Kein `terraform apply` — die VMs existieren manuell (siehe
  `infra/terraform/README.md` für den `terraform import`-Workflow als
  Maintainer-Hausaufgabe)
- Keine Backup-Automation für Postgres / Caddy-Daten / Monitoring-Daten
  — kommt in Session 12 als Runbook + Cron-Konfiguration
- Keine externe Uptime-Prüfung (Uptime Kuma als Container ist im
  monitoring-stack drin, aber Konfiguration der Checks bleibt manuell)
- Keine Hetzner-Migration (separate ADR + zukünftige Session)
- Keine GitHub-Actions-basierte Deploy-Automation — alles Hand-Triggered
  via `scripts/deploy.sh`

## Vorgehen — verbindliche Reihenfolge

1. **Code-PRs zuerst, dann Deploy:**
   - PR A: neue `monitoring-stack`-Ansible-Rolle + Files unter
     `infra/monitoring/` adaptieren
   - PR B: Bind-Mount-Migration in `infra/caddy/prod/compose.yml`
     (Volumes-Block ändern; Skelett-PR)
   - PR C: ggf. Caddyfile-Cutover-Vorlage (auskommentiert oder als
     `Caddyfile.cutover`-Variante zum manuellen Aktivieren)
2. **Beide PRs gemerged**, Tags der Backend/Frontend/Pyworkers-Releases
   stehen in ghcr.io.
3. **Phase 0-6 wie oben**, jeweils mit Snapshot vorher.
4. **STATUS.md und CLAUDE.md aktualisieren**, zwei Sub-Sessions als ✅
   markieren (10c über 11a abgedeckt, 11a selbst).

## Suggested Commit-Messages

- Code-PR A: `feat(infra): add monitoring-stack ansible role for wwn-mon`
- Code-PR B: `refactor(infra): migrate caddy data volume to bind-mount`
- Code-PR C (falls separat): `feat(infra): add caddy reverse_proxy cutover config`
- Final-Doku-PR nach Deploy: `docs(sessions): mark 10c and 11a done after live deploy`
