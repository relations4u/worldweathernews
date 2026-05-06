#!/usr/bin/env bash
# App-Deploy-Wrapper. Ruft Ansible-Playbook deploy.yml gegen das gewählte
# Environment mit der gewünschten Image-Version.
#
# Usage:
#   bash scripts/deploy.sh <environment> <version>
#   bash scripts/deploy.sh production 0.1.0
#
# Aktuell ist nur `production` konfiguriert (Forschungs-Phase). `staging`
# bleibt im Wrapper als Hinweis enthalten, sobald ein zweites Inventory
# existiert.

set -euo pipefail

ENV="${1:-}"
VERSION="${2:-}"

usage() {
	cat >&2 <<EOF
Usage: $0 <environment> <version>
  environment: production
  version:     git-tag oder explizite Version (z. B. 0.1.0)

Examples:
  $0 production 0.1.0
EOF
	exit 1
}

[ -z "$ENV" ] && usage
[ -z "$VERSION" ] && usage

case "$ENV" in
production)
	# OK
	;;
staging)
	echo "ERROR: staging-Inventory existiert noch nicht. Erst anlegen unter" >&2
	echo "       infra/ansible/inventories/staging/ — siehe README." >&2
	exit 1
	;;
*)
	usage
	;;
esac

if [ "$ENV" = "production" ]; then
	echo "⚠  Deploying to PRODUCTION (version $VERSION)"
	read -rp "Type 'production' to confirm: " confirm
	if [ "$confirm" != "production" ]; then
		echo "Aborted."
		exit 1
	fi
fi

REPO_ROOT="$(git rev-parse --show-toplevel)"
cd "${REPO_ROOT}/infra/ansible"

# Der App-Deploy braucht root für Datei/Template-Tasks unter /opt/wwn — der
# `deploy`-Default-User hat aus Sicherheitsgründen NUR docker-NOPASSWD-sudo,
# kein generelles. Deshalb läuft der Wrapper zuverlässig nur über `hwr`.
ansible-playbook \
	-i "inventories/${ENV}/hosts.yml" \
	playbooks/deploy.yml \
	-e "ansible_user=hwr" \
	-e "target_version=${VERSION}"
