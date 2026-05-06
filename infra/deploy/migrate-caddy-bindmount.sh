#!/usr/bin/env bash
# Einmalige Migration des Caddy-Daten-Volumes von Docker named volume auf
# Bind-Mount.
#
# Quelle:  /var/lib/docker/volumes/{wwn_caddy_data,wwn_caddy_config}/_data/
# Ziel:    /srv/wwn/caddy/{data,config}/
#
# Das Skript ist idempotent. Wenn die Bind-Mount-Pfade schon Daten enthalten,
# läuft die Migration nicht erneut.
#
# Sicherheits-Garantien:
#   - Vor jeder destruktiven Aktion wird ein lokaler Tar-Backup angelegt.
#   - Caddy wird nur kurz gestoppt; Container und Image bleiben.
#   - Cert-Dates (notBefore) werden vor und nach der Migration verglichen.
#     Wenn sie sich unterscheiden, hat Caddy frisch ge-ACME't — Skript
#     bricht mit Fehler ab und der Maintainer muss prüfen.

set -euo pipefail

REMOTE_USER="hwr"
REMOTE_HOST="10.100.100.21"
REMOTE_PORT="22"
REMOTE_PATH="/srv/wwn/caddy"
REMOTE_TARGET="${REMOTE_USER}@${REMOTE_HOST}"

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
HOSTNAMES=(
	worldweathernews.com
	www.worldweathernews.com
	research.worldweathernews.com
	api.research.worldweathernews.com
)

ssh_run() {
	ssh -p "${REMOTE_PORT}" -o BatchMode=yes "${REMOTE_TARGET}" "$@"
}

ssh_run_tty() {
	ssh -t -p "${REMOTE_PORT}" "${REMOTE_TARGET}" "$@"
}

cert_dates() {
	# Gibt für jeden Hostname "host: notBefore=..." aus.
	for h in "${HOSTNAMES[@]}"; do
		local out
		out=$(echo | openssl s_client -connect "${h}:443" -servername "${h}" 2>/dev/null \
			| openssl x509 -noout -dates 2>/dev/null \
			| grep notBefore || echo "notBefore=UNKNOWN")
		printf '%-40s %s\n' "${h}:" "${out}"
	done
}

echo "==> Migration: caddy named volume → bind-mount"
echo "==> Remote: ${REMOTE_TARGET}:${REMOTE_PATH}"
echo

echo "==> Step 1/6: Idempotenz-Check — sind die Bind-Mounts schon befüllt?"
if ssh_run "test -d ${REMOTE_PATH}/data/caddy/certificates && [ -n \"\$(ls -A ${REMOTE_PATH}/data/caddy/certificates 2>/dev/null)\" ]"; then
	echo "  Bind-Mount ${REMOTE_PATH}/data/caddy/certificates ist bereits befüllt."
	echo "  Migration wurde offenbar schon durchgeführt — nichts zu tun."
	exit 0
fi
echo "  Bind-Mount noch leer/fehlt — Migration läuft jetzt."
echo

echo "==> Step 2/6: Vorab-Snapshot der Cert-Dates (öffentliche DNS-Auflösung)"
PRE_DATES_FILE="$(mktemp -t caddy-cert-dates-pre.XXXXXX)"
cert_dates | tee "${PRE_DATES_FILE}"
echo

echo "==> Step 3/6: Tar-Backup des Named-Volumes (sicher gegen alle Folge-Fehler)"
# /var/lib/docker/volumes ist root-only (mode 0700), deshalb braucht der
# Existenz-Check sudo. Auf wwn-prod ist NOPASSWD aktiv, läuft non-interactive.
ssh_run "sudo -n test -d /var/lib/docker/volumes/wwn_caddy_data/_data" || {
	echo "ERROR: /var/lib/docker/volumes/wwn_caddy_data/_data existiert nicht auf ${REMOTE_HOST}." >&2
	echo "       Entweder ist das Volume schon entfernt oder die Caddy-Stack lief nie." >&2
	echo "       Wenn das ein Fresh-Install ist, lege /srv/wwn/caddy/data und /config einfach manuell an:" >&2
	echo "         ssh -t ${REMOTE_TARGET} 'sudo install -d -o ${REMOTE_USER} -g ${REMOTE_USER} -m 0750 ${REMOTE_PATH}/data ${REMOTE_PATH}/config'" >&2
	exit 1
}

BACKUP_NAME="caddy-data-backup-$(date +%F-%H%M).tar.gz"
echo "  Schreibe ${REMOTE_PATH}/${BACKUP_NAME} (sudo)"
ssh_run_tty "sudo tar czf ${REMOTE_PATH}/${BACKUP_NAME} -C /var/lib/docker/volumes/wwn_caddy_data/_data . && sudo chown ${REMOTE_USER}:${REMOTE_USER} ${REMOTE_PATH}/${BACKUP_NAME}"
ssh_run "ls -lh ${REMOTE_PATH}/${BACKUP_NAME}"
echo

echo "==> Step 4/6: Caddy stoppen, Bind-Mount-Verzeichnisse anlegen, Daten kopieren"
ssh_run "cd ${REMOTE_PATH} && docker compose stop caddy"
ssh_run_tty "
	set -e
	sudo install -d -o ${REMOTE_USER} -g ${REMOTE_USER} -m 0750 ${REMOTE_PATH}/data ${REMOTE_PATH}/config
	sudo cp -a /var/lib/docker/volumes/wwn_caddy_data/_data/.   ${REMOTE_PATH}/data/
	sudo cp -a /var/lib/docker/volumes/wwn_caddy_config/_data/. ${REMOTE_PATH}/config/
"
ssh_run "ls -la ${REMOTE_PATH}/data/caddy/certificates/acme-v02.api.letsencrypt.org-directory/ 2>&1 | head -20"
echo

echo "==> Step 5/6: Neues compose.yml (Bind-Mount-Variante) ausrollen + Caddy starten"
bash "${SCRIPT_DIR}/deploy-caddy.sh"
echo

echo "==> Step 6/6: Cert-Dates vergleichen"
sleep 5
POST_DATES_FILE="$(mktemp -t caddy-cert-dates-post.XXXXXX)"
cert_dates | tee "${POST_DATES_FILE}"
echo

if diff -u "${PRE_DATES_FILE}" "${POST_DATES_FILE}" >/dev/null; then
	echo "✅ Cert-Dates unverändert. Migration erfolgreich."
	echo
	echo "Tar-Backup auf wwn-prod:    ${REMOTE_PATH}/${BACKUP_NAME}"
	echo "Empfehlung: Backup zusätzlich auf Maintainer-Maschine ziehen:"
	echo "  scp ${REMOTE_TARGET}:${REMOTE_PATH}/${BACKUP_NAME} ~/backups/"
	echo
	echo "Alte Docker-Volumes können nach 1-2 Tagen Beobachtungszeit entfernt werden:"
	echo "  ssh ${REMOTE_TARGET} docker volume rm wwn_caddy_data wwn_caddy_config"
else
	echo "❌ Cert-Dates haben sich geändert — Caddy hat ACME erneut gemacht!" >&2
	echo "   Das ist KEIN normaler Migrations-Verlauf." >&2
	echo "   Vergleiche:" >&2
	diff -u "${PRE_DATES_FILE}" "${POST_DATES_FILE}" >&2 || true
	echo >&2
	echo "Backup zum Restore: ${REMOTE_TARGET}:${REMOTE_PATH}/${BACKUP_NAME}" >&2
	exit 1
fi
