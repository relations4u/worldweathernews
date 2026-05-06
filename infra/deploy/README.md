# infra/deploy

Deployment-Skripte für den Self-Hosting-Stack auf den Proxmox-VMs. Kein
Ansible, kein Terraform — bewusst klein, bis Session 11a Ansible übernimmt.

## Skript-Übersicht

| Skript                       | Zweck                                               |
| ---------------------------- | --------------------------------------------------- |
| `deploy-caddy.sh`            | Caddy-Stack auf wwn-prod (re)deployen               |
| `migrate-caddy-bindmount.sh` | Einmalige Migration des Cert-Volumes auf Bind-Mount |

## deploy-caddy.sh

Synct `infra/caddy/prod/` nach `hwr@10.100.100.21:/srv/wwn/caddy` und startet
den Caddy-Stack via `docker compose pull && up -d`.

### Voraussetzungen

- SSH-Login `hwr@10.100.100.21` ohne Passwort (Public-Key)
- Docker Engine + Compose-Plugin auf wwn-prod
- DNS-Records für die 4 Hostnames zeigen via `gate.hw7.eu` auf den Host
- Hardware-Firewall leitet Ports 80 und 443 auf wwn-prod weiter
- Zielverzeichnis `/srv/wwn/caddy` existiert und gehört `hwr:hwr`. Einmalig
  manuell anlegen (sudo-Passwort wird abgefragt):

  ```bash
  ssh -t hwr@10.100.100.21 sudo install -d -o hwr -g hwr -m 0755 /srv/wwn/caddy
  ```

### Ausführung

```bash
bash infra/deploy/deploy-caddy.sh
```

Das Skript ist idempotent. `rsync --delete` spiegelt das Remote-Verzeichnis
exakt auf das lokale — keine manuellen Änderungen direkt auf wwn-prod, sonst
gehen sie beim nächsten Deploy verloren.

### Verifikation nach Deploy

```bash
for h in worldweathernews.com www.worldweathernews.com \
         research.worldweathernews.com api.research.worldweathernews.com; do
    echo "--- $h ---"
    curl -sSI "https://$h" | head -n 5
done

echo | openssl s_client -connect worldweathernews.com:443 \
    -servername worldweathernews.com 2>/dev/null \
    | openssl x509 -noout -issuer -subject -dates
```

Erwartet: `HTTP/2 200` (Apex, research, api.research) bzw. `HTTP/2 301`
(www → Apex), Issuer `Let's Encrypt`, gültiges Zertifikat (`notBefore` heute,
`notAfter` in ~90 Tagen).

### Log-Tail bei Bedarf

```bash
ssh hwr@10.100.100.21 'cd /srv/wwn/caddy && docker compose logs -f caddy'
```

### Stack stoppen

```bash
ssh hwr@10.100.100.21 'cd /srv/wwn/caddy && docker compose down'
```

**NIEMALS `docker compose down -v`** — das `-v` würde Bind-Mount-Pfade nicht
betreffen, aber alte Named-Volumes (vor der Bind-Mount-Migration) und ist
generell ein Footgun. Nach dem regulären `down` bleiben die Cert-Daten unter
`/srv/wwn/caddy/data/` erhalten.

## migrate-caddy-bindmount.sh

Einmalige Migration: kopiert die Cert-Daten aus den Docker-Named-Volumes
(`wwn_caddy_data`, `wwn_caddy_config`) in Bind-Mount-Pfade unter
`/srv/wwn/caddy/data` und `/srv/wwn/caddy/config`. Das `compose.yml` im Repo
wurde gleichzeitig auf Bind-Mount umgestellt, der `deploy-caddy.sh`-Guard
blockt einen Deploy ohne vorherige Migration.

**Wann anwenden:** genau einmal nach dem Merge der Bind-Mount-Migrations-PR
auf den Stand mit den Live-Certs (Caddy auf wwn-prod läuft seit 6. Mai 2026).

**Aufruf:**

```bash
bash infra/deploy/migrate-caddy-bindmount.sh
```

Das Skript ist idempotent — bei erneutem Aufruf nach erfolgreicher Migration
ist es ein No-op.

**Was es tut, in Reihenfolge:**

1. Idempotenz-Check: ist `/srv/wwn/caddy/data/caddy/certificates/` schon befüllt?
2. Vorab-Snapshot der `notBefore`-Daten der vier öffentlichen Hostnames
   (Apex, www, research, api.research).
3. Tar-Backup des Named-Volumes nach `/srv/wwn/caddy/caddy-data-backup-<TS>.tar.gz`.
4. Caddy stoppen, Bind-Mount-Verzeichnisse anlegen (Owner `hwr:hwr`,
   Mode 0750), Volume-Inhalt mit `cp -a` rüberkopieren (Permissions
   bleiben erhalten).
5. Aufruf von `deploy-caddy.sh` — der Guard passt nun, neues `compose.yml`
   wird gesynct, Caddy startet mit Bind-Mount-Daten.
6. Vergleich der `notBefore`-Daten vorher/nachher. Wenn sie sich unterscheiden,
   bricht das Skript mit Fehler ab — Caddy hätte dann frisch ge-ACME't, was
   nicht erwartet ist.

**Voraussetzung:** Die Named-Volumes `wwn_caddy_data` und `wwn_caddy_config`
müssen auf wwn-prod existieren (Stand: ja, vom Setup am 6. Mai). `hwr` braucht
`sudo`-Rechte für `tar`, `install -d` und `cp` — die `sudo`-Aufrufe nutzen
`ssh -t`, also wird das Passwort lokal abgefragt (kein NOPASSWD nötig).

**Aufräumen nach 1-2 Tagen Beobachtungszeit:**

```bash
ssh hwr@10.100.100.21 docker volume rm wwn_caddy_data wwn_caddy_config
```

Den Tar-Backup von der Maintainer-Maschine ziehen und off-host sichern:

```bash
scp hwr@10.100.100.21:/srv/wwn/caddy/caddy-data-backup-*.tar.gz ~/backups/
```
