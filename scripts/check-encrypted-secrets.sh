#!/usr/bin/env bash
# Pre-commit guard: verweigert das Commit-Setup, wenn ein File unter
# infra/secrets/ unverschlüsselt eincheckt würde.
#
# SOPS-encrypted Files erkennen wir am Marker pro Format:
#   - .env       → enthält Zeile `sops_version=...`
#   - .yaml/.yml → enthält Top-Level-Key `sops:` (Block am Ende)
#   - .json      → enthält Top-Level-Key `"sops":`
#
# Aufruf via .pre-commit-config.yaml mit pass_filenames: true.

set -euo pipefail

failed=0

for file in "$@"; do
	if [ ! -f "$file" ]; then
		continue
	fi

	encrypted=0
	case "$file" in
	*.env)
		grep -q '^sops_version=' "$file" 2>/dev/null && encrypted=1
		;;
	*.yaml | *.yml)
		grep -q '^sops:' "$file" 2>/dev/null && encrypted=1
		;;
	*.json)
		head -c 4096 "$file" | grep -q '"sops":' 2>/dev/null && encrypted=1
		;;
	*)
		# Unbekanntes Format unter infra/secrets/ → konservativ blocken.
		echo "✗ Unknown secret-file extension: $file" >&2
		failed=1
		continue
		;;
	esac

	if [ "$encrypted" = "0" ]; then
		echo "✗ Unencrypted secret detected: $file" >&2
		echo "  Encrypt with: sops --encrypt --in-place $file" >&2
		failed=1
	fi
done

exit "$failed"
