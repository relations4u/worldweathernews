#!/usr/bin/env bash
# Deploy the stand-alone Caddy stack to wwn-prod.
#
# Source:  infra/caddy/prod/
# Target:  hwr@10.100.100.70:/srv/wwn/caddy
# Action:  rsync, then `docker compose pull && up -d` over SSH.
#
# Idempotent. Safe to run repeatedly. Uses `rsync --delete`, so the remote
# directory mirrors the local one exactly — keep that in mind before adding
# anything manually on the remote side.

set -euo pipefail

REMOTE_USER="hwr"
REMOTE_HOST="10.100.100.70"
REMOTE_PORT="22"
REMOTE_PATH="/srv/wwn/caddy"
REMOTE_TARGET="${REMOTE_USER}@${REMOTE_HOST}"

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
LOCAL_PATH="$(cd "${SCRIPT_DIR}/../caddy/prod" && pwd)"

ssh_run() {
	ssh -p "${REMOTE_PORT}" -o BatchMode=yes "${REMOTE_TARGET}" "$@"
}

echo "==> Source:  ${LOCAL_PATH}"
echo "==> Target:  ${REMOTE_TARGET}:${REMOTE_PATH} (port ${REMOTE_PORT})"

echo "==> Verifying SSH reachability and Docker installation"
ssh_run 'command -v docker >/dev/null && docker compose version >/dev/null' || {
	echo "ERROR: SSH failed, or docker / docker compose plugin missing on remote." >&2
	exit 1
}

echo "==> Checking that ${REMOTE_PATH} exists and is writable by ${REMOTE_USER}"
if ! ssh_run "test -d ${REMOTE_PATH} && test -w ${REMOTE_PATH}"; then
	cat <<EOF >&2
ERROR: ${REMOTE_PATH} either doesn't exist or isn't writable by ${REMOTE_USER}.
Run once on the remote (will prompt for sudo password):

    ssh -t ${REMOTE_TARGET} sudo install -d -o ${REMOTE_USER} -g ${REMOTE_USER} -m 0755 ${REMOTE_PATH}

Then re-run this script.
EOF
	exit 1
fi

# Guard: das compose.yml im Repo nutzt Bind-Mounts (./data, ./config). Wenn die
# Bind-Mount-Verzeichnisse auf der Remote fehlen oder leer sind, würde dieser
# Deploy Caddy mit einem leeren Cert-Store starten — gleichbedeutend mit einer
# erneuten ACME-Issuance, was am Let's-Encrypt-Rate-Limit kratzt.
if grep -q '\./data:/data' "${LOCAL_PATH}/compose.yml"; then
	if ! ssh_run "test -d ${REMOTE_PATH}/data && [ -n \"\$(ls -A ${REMOTE_PATH}/data 2>/dev/null)\" ]"; then
		cat <<EOF >&2
ERROR: compose.yml erwartet Bind-Mount unter ${REMOTE_PATH}/data, aber das
Verzeichnis fehlt oder ist leer auf der Remote. Ein Deploy würde den
laufenden Caddy mit leerem Cert-Store starten und ACME erneut auslösen
(Rate-Limit-Risiko).

Wenn du gerade von Named-Volumes auf Bind-Mount migrierst, nutze:

    bash infra/deploy/migrate-caddy-bindmount.sh

Wenn das ein Fresh-Install ohne bestehende Certs ist, lege die Verzeichnisse
einmalig leer an und akzeptiere die initiale ACME-Issuance:

    ssh -t ${REMOTE_TARGET} 'sudo install -d -o ${REMOTE_USER} -g ${REMOTE_USER} -m 0750 ${REMOTE_PATH}/data ${REMOTE_PATH}/config'
EOF
		exit 1
	fi
fi

echo "==> Syncing Caddy stack"
# data/ und config/ liegen als Bind-Mount auf der Remote, sind dort
# root-owned und niemals lokal versioniert. Tar-Backups (caddy-data-backup-*)
# werden bewusst NICHT mitgesynct, sonst zerstört --delete sie beim nächsten
# Lauf. .terraform-Artefakte gleicher Logik.
rsync -avz --delete \
	-e "ssh -p ${REMOTE_PORT}" \
	--exclude='.DS_Store' \
	--exclude='data/' \
	--exclude='config/' \
	--exclude='caddy-data-backup-*.tar.gz' \
	"${LOCAL_PATH}/" "${REMOTE_TARGET}:${REMOTE_PATH}/"

echo "==> Pulling image and (re)starting stack"
# `docker compose up -d` allein reicht NICHT, wenn nur der Caddyfile-Inhalt
# geändert wurde: Bind-Mount-Files hängen am Inode des Originals beim
# Container-Start, und rsync macht atomic-rename → neuer Inode auf dem Host,
# Container sieht weiterhin den alten. `restart` re-resolved den Bind-Mount.
# Das `data`-Volume mit den TLS-Certs bleibt davon unangetastet.
ssh_run "cd ${REMOTE_PATH} && docker compose pull && docker compose up -d && docker compose restart caddy"

echo "==> Recent Caddy logs (last 40 lines)"
ssh_run "cd ${REMOTE_PATH} && docker compose logs --tail=40 caddy"

cat <<'EOF'

==> Deploy complete. Verify TLS from anywhere with public DNS:

    for h in worldweathernews.com www.worldweathernews.com \
             research.worldweathernews.com api.research.worldweathernews.com; do
        echo "--- $h ---"
        curl -sSI "https://$h" | head -n 5
    done

    echo | openssl s_client -connect worldweathernews.com:443 \
        -servername worldweathernews.com 2>/dev/null \
        | openssl x509 -noout -issuer -subject -dates

EOF
