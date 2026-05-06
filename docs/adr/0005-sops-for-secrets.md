# 5. SOPS+age für Secrets statt Vault o. ä.

Date: 2026-05-06
Status: Accepted

## Context

Production-Secrets (DB-Passwords, ghcr-Token, Grafana-Admin-Pass,
Backend-CORS-Konfig, …) müssen versioniert, auditierbar und
kontrolliert verteilt sein. Plaintext im Git ist ausgeschlossen. Ein
externer Dienst wie HashiCorp Vault wäre möglich, aber das wäre **noch
ein** kritischer Self-Hosted-Service mit eigener Hochverfügbarkeit-
Anforderung.

Die Plattform hat aktuell eine geringe Secret-Komplexität: zwei VMs,
ein Maintainer, alle Werte sind langlebig (kein dynamisches Issuance
nötig).

## Decision

Wir nutzen **[SOPS](https://github.com/getsops/sops)** mit
**[age](https://github.com/FiloSottile/age)** als Master-Key-Backend.

- Verschlüsselte Secret-Files leben unter `infra/secrets/<env>/<service>.env`
  (Dotenv) bzw. `<service>.yml` (YAML).
- `.sops.yaml` im Repo-Root definiert creation_rules und Public-Keys.
- Maintainer-Private-Key liegt unter `~/.config/sops/age/keys.txt`
  mit `chmod 0600`. Der Public-Key ist im Repo (`.sops.yaml` →
  `keys: - &hwr age1…`).
- Pre-commit-Hook `forbid-unencrypted-secrets` blockt Plaintext-Files
  unter `infra/secrets/`.
- Ansible-Rollen lesen Secrets via
  `lookup('community.sops.sops', secrets_dir + '/<file>')`, parsen
  sie ggf. als Dotenv und geben sie als ENV-Files an die Container.

## Consequences

**Positiv**:

- **Kein zusätzlicher Service** — SOPS ist ein CLI-Tool, age ist
  ein CLI-Tool. Beide unter 5 MB.
- **Audit über git-history** — jeder Wert hat einen Commit, ein
  Diff zeigt _was_ geändert wurde (verschlüsselt, aber Struktur ist
  sichtbar).
- Maintainer-Workflow ist einfach: `sops <file>` öffnet Editor mit
  entschlüsseltem Inhalt, Speichern re-encryptet.
- Mehrere Maintainer trivial hinzufügbar — neuer age-Pubkey in
  `.sops.yaml`, dann einmal `sops updatekeys` über alle Files.
- Funktioniert offline (kein Netzwerk-Roundtrip wie bei Vault).
- Unter dem CIS-Benchmark für Secrets-Management praktikabel:
  Verschlüsselung at rest, granulares Access-Control via Pubkey-
  Listen pro Path.

**Negativ**:

- **Kein dynamisches Issuance** — DB-Credentials sind statisch
  vergeben. Für die aktuelle Phase OK.
- **Key-Rotation ist manuell** — alter Pubkey raus, neuer rein,
  `sops updatekeys` über alle Files, Master-Private-Key ersetzen.
  Workflow muss dokumentiert sein (siehe `docs/secrets.md`).
- **Privater Key ist Single Point of Trust** — Verlust bedeutet
  Verlust aller verschlüsselten Werte. Mitigation: Off-Site-Backup
  des Private-Keys auf separatem Medium.
- **Teilweise SOPS-Edge-Cases**: `community.sops.load_vars` will
  YAML, nicht Dotenv → wir parsen Dotenv-Files manuell mit
  `regex_findall` in den Ansible-Rollen.

## Alternatives Considered

- **HashiCorp Vault** — mächtig, aber Operations-Overhead (Auto-Unseal,
  HA, Audit-Logging, Backup) ist für ein 1-Service-Self-Hosting-Setup
  unverhältnismäßig. License-Drama (BSL) ist zusätzlich gegen ihn.
- **AWS KMS / Secrets Manager** — wir wollen self-hostable bleiben,
  Cloud-Lock-in widerspricht der Hosting-Strategie (siehe CLAUDE.md).
- **Doppler / Infisical** — externe SaaS, Vendor-Risiko, kostet ab
  bestimmter Größe Geld.
- **gitcrypt** — funktioniert, aber das Re-Key-Modell ist klobiger
  und CI-Integration ist schwächer als bei SOPS.
- **GPG statt age** — GPG ist powerful, aber sein Keyring-Modell
  und Trust-Web ist erheblich komplexer als age (die explizite
  Pubkey-Liste). age erschien dem Maintainer nach Erfahrung als der
  richtige Trade-off.
