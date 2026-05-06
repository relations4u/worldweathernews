# Secret-Management mit SOPS

Secrets werden mit [SOPS](https://getsops.io/) und age-Keys verschlüsselt im
Git-Repo gehalten. Auf den Servern werden sie zur Deploy-Zeit von Ansible
entschlüsselt und als `.env`-Files unter `/opt/wwn/` mit Mode `0600` abgelegt.

Die Forschungs-Phase nutzt **nur** das `production`-Environment
(`infra/secrets/production/`). Eine `staging`-Schiene ist in `.sops.yaml` schon
vorgesehen, aber leer — solange sie keine Files enthält, fällt sie durchs Raster.

## Tools

| Tool         | Quelle            | Pin (`.mise.toml`)            |
| ------------ | ----------------- | ----------------------------- |
| `sops`       | mise (`ubi:`)     | `ubi:getsops/sops = "3.12"`   |
| `age`        | System (apt/brew) | nicht via mise — System-Paket |
| `age-keygen` | wie `age`         | wie oben                      |

`make bootstrap` ruft `mise install` und installiert SOPS automatisch. `age`
bringt jede halbwegs aktuelle Linux-Distribution direkt mit
(`apt install age`).

## Erst-Setup für einen Maintainer

1. age-Keypair generieren:

   ```bash
   mkdir -p ~/.config/sops/age
   chmod 700 ~/.config/sops/age
   age-keygen -o ~/.config/sops/age/keys.txt
   chmod 600 ~/.config/sops/age/keys.txt
   ```

2. Public-Key kopieren — das ist die Zeile, die `age-keygen` als
   `# public key: age1...` ausgibt (auch im File ganz oben kommentiert).

3. **Privaten Key sofort sichern.** Ohne ihn kommt niemand mehr an die Secrets.
   Empfehlung: in den persönlichen Passwort-Manager kopieren und in einer
   verschlüsselten Off-Site-Kopie ablegen. Es gibt **keinen zweiten Weg**.

4. Den Public-Key zur `.sops.yaml` im Repo-Root hinzufügen lassen
   (PR an bestehenden Maintainer):

   ```yaml
   keys:
     - &hwr age1cee9fufhrs0yt85py2gfchytxnm6ze4qxmvds9almqym7y2qg55quelefm
     - &alice age1xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
   creation_rules:
     - path_regex: ^infra/secrets/(staging|production)/.+\.env$
       age:
         - *hwr
         - *alice
   ```

5. Bestehender Maintainer rotiert die Schlüssel-Liste in jedem Secret-File:

   ```bash
   for f in infra/secrets/production/*.env; do sops updatekeys "$f"; done
   ```

6. Neuer Maintainer prüft Lese-Zugriff:

   ```bash
   sops --decrypt infra/secrets/production/backend.env
   ```

## Secret editieren

```bash
sops infra/secrets/production/backend.env
```

Öffnet das File entschlüsselt im `$EDITOR`, verschlüsselt beim Speichern wieder.

## Neues Secret anlegen

```bash
# Plaintext temporär schreiben, encrypten, Original löschen.
printf 'MY_KEY=value\n' > infra/secrets/production/feature-x.env
sops --encrypt --in-place infra/secrets/production/feature-x.env
git add infra/secrets/production/feature-x.env
```

Der Pre-commit-Hook `forbid-unencrypted-secrets` blockt den Commit, falls
das `--in-place`-Encrypten vergessen wurde.

## Ad-hoc Decrypt für Skript-Nutzung

```bash
# Dotenv-Variable ins aktuelle Shell sourcen (nicht ideal — nur für Debug):
set -a && source <(sops --decrypt infra/secrets/production/backend.env) && set +a
```

Auf den Servern macht das Ansible's `community.sops.load_vars` bzw. das
`sops`-Modul. Manuelle Sourcen sollten die Ausnahme sein.

## Recovery

Wenn der Private-Key verloren geht:

1. Bestehende Maintainer haben weiter Zugriff
2. Aus Backup wiederherstellen, falls vorhanden
3. Wenn kein Backup: alle Secrets rotieren — bestehende Maintainer
   entschlüsseln, neue Werte erzeugen (Passwörter, Tokens), neu encrypten,
   externe Systeme (DB, ghcr.io, Proxmox-API) auf neue Werte aktualisieren

## Was NICHT in SOPS gehört

- Niemals Klartext-Secrets in `.env.example`
- Keine Secrets in CI-Logs ausgeben (Workflow-Step `set +x` bei Bedarf)
- Keine Secrets in Container-Image-Layern
- Keine Secrets als Build-Args (sind in der Image-History sichtbar)
- Keine Secrets in `infra/compose/compose.*.yml` direkt — die kommen via
  `env_file:` aus den vom Ansible deployten `/opt/wwn/.env.*`-Files

## Aktueller Bestand

| Pfad                                     | Inhalt                                       |
| ---------------------------------------- | -------------------------------------------- |
| `infra/secrets/production/backend.env`   | Backend-DB-/Redis-URLs, Log-Level            |
| `infra/secrets/production/frontend.env`  | Frontend-ENV (PUBLIC_API_BASE_URL, NODE_ENV) |
| `infra/secrets/production/pyworkers.env` | Worker-DB-/Redis-URLs, Log-Level             |
| `infra/secrets/production/postgres.env`  | Postgres User/Password/DB                    |
| `infra/secrets/production/ghcr.env`      | GitHub-PAT für ghcr.io-Pulls (read:packages) |
| `infra/secrets/production/proxmox.env`   | Proxmox-API-Endpoint + Token (für Terraform) |

Alle Werte mit `CHANGE_ME_*`-Platzhaltern müssen vor dem ersten echten Deploy
durch reale Werte ersetzt werden — `sops <file>` öffnet das jeweilige File im
Editor.
