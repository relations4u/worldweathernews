# infra/deploy

Deployment-Skripte für den Self-Hosting-Stack auf den Proxmox-VMs. Kein
Ansible, kein Terraform — bewusst klein, bis Session 11 das übernimmt.

## deploy-caddy.sh

Synct `infra/caddy/prod/` nach `hwr@10.100.100.21:/srv/wwn/caddy` und startet
den Caddy-Stack via `docker compose pull && up -d`.

### Voraussetzungen

- SSH-Login `hwr@10.100.100.21` ohne Passwort (Public-Key)
- `hwr` auf wwn-prod hat passwordless `sudo` (für initiales `mkdir` in `/srv`)
- Docker Engine + Compose-Plugin auf wwn-prod
- DNS-Records für die 4 Hostnames zeigen via `gate.hw7.eu` auf den Host
- Hardware-Firewall leitet Ports 80 und 443 auf wwn-prod weiter

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

`caddy_data` bleibt erhalten (Volume `wwn_caddy_data`) — wichtig, damit beim
nächsten Start die Zertifikate nicht erneut gezogen werden müssen (Let's-
Encrypt-Rate-Limit beachten).
