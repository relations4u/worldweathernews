#!/usr/bin/env bash
# Deploy the stand-alone Caddy stack to wwn-prod.
#
# Source:  infra/caddy/prod/
# Target:  hwr@10.100.100.21:/srv/wwn/caddy
# Action:  rsync, then `docker compose pull && up -d` over SSH.
#
# Idempotent. Safe to run repeatedly. Uses `rsync --delete`, so the remote
# directory mirrors the local one exactly — keep that in mind before adding
# anything manually on the remote side.

set -euo pipefail

REMOTE_USER="hwr"
REMOTE_HOST="10.100.100.21"
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

echo "==> Ensuring ${REMOTE_PATH} exists and is writable by ${REMOTE_USER}"
ssh_run "sudo install -d -o ${REMOTE_USER} -g ${REMOTE_USER} -m 0755 ${REMOTE_PATH}"

echo "==> Syncing Caddy stack"
rsync -avz --delete \
	-e "ssh -p ${REMOTE_PORT}" \
	--exclude='.DS_Store' \
	"${LOCAL_PATH}/" "${REMOTE_TARGET}:${REMOTE_PATH}/"

echo "==> Pulling image and (re)starting stack"
ssh_run "cd ${REMOTE_PATH} && docker compose pull && docker compose up -d"

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
