# Session 11 — Ansible, SOPS, Terraform-Skelett

**Phase**: D (Ops)
**Geschätzte Dauer**: 2-3 Stunden
**Vorbedingung**: Sessions 9 und 10 abgeschlossen, Container-Images werden via Tag in ghcr.io gebaut.

## Ziel

Der Deployment-Pfad ist klar definiert und schreib-dokumentiert, **ohne dass
tatsächlich auf einem Server deployed wird**. Diese Session liefert:

- SOPS-basiertes Secret-Management mit age-Keys, in Git verschlüsselt
- Pre-commit-Hook der unverschlüsselte Secrets blockt
- Ansible-Rollen und -Playbooks für Erst-Setup und Deployment eines Hosts
- Terraform-Skelett für den gewählten Hosting-Provider
- Ein `scripts/deploy.sh`-Wrapper für manuelles Deployment

Tatsächliches Provisioning eines Servers ist Hausaufgabe des Maintainers nach
dieser Session.

## Vor-Klärung (kritisch)

Ohne diese Antworten bleibt das Skelett zu generisch. **Frag den Maintainer:**

1. **Hosting-Provider**: Hetzner Cloud (Default-Empfehlung, Standort Falkenstein
   oder Nürnberg), Strato, eigene Hardware, andere?
2. **age-Key**: Hat der Maintainer schon einen age-Public-Key für SOPS, oder
   neu generieren? Empfehlung neu generieren mit Backup-Anleitung:
   `age-keygen -o ~/.config/sops/age/keys.txt`. Public-Key (beginnt mit `age1...`)
   wird gebraucht.
3. **Domain-Setup**: DNS bei welchem Provider? (Cloudflare, Hetzner, Inwx, ...)
   Der Provider beeinflusst, ob wir dafür auch ein Terraform-Modul brauchen.
4. **SSH-Public-Key** des Maintainers für initial-User-Setup auf dem Server
5. **GitHub-Org** für ghcr.io-Pulls (für Server-PAT)

Mit diesen Antworten weitermachen.

## Aufgaben

### 1. SOPS-Setup

#### 1.1 `.sops.yaml`

```yaml
keys:
  - &maintainer <AGE_PUBLIC_KEY_HIER>
  # Weitere Maintainer hier ergänzen, mit YAML-Anker

creation_rules:
  - path_regex: infra/secrets/staging/.*\.env$
    age:
      - *maintainer
  - path_regex: infra/secrets/staging/.*\.yaml$
    age:
      - *maintainer
  - path_regex: infra/secrets/production/.*\.env$
    age:
      - *maintainer
  - path_regex: infra/secrets/production/.*\.yaml$
    age:
      - *maintainer
```

#### 1.2 Verzeichnisstruktur

```
infra/secrets/
├── staging/
│   ├── backend.env        # SOPS-encrypted
│   ├── frontend.env       # SOPS-encrypted
│   ├── pyworkers.env      # SOPS-encrypted
│   └── postgres.env       # SOPS-encrypted
├── production/
│   ├── backend.env
│   ├── frontend.env
│   ├── pyworkers.env
│   └── postgres.env
└── README.md
```

#### 1.3 Beispiel-Secret-File anlegen (verschlüsselt)

Beispiel `infra/secrets/staging/backend.env`:

```bash
# Plaintext vor encryption (zur Demonstration):
WWN_DATABASE_URL=postgres://wwn:STAGING_PG_PASS@postgres:5432/wwn?sslmode=disable
WWN_REDIS_URL=redis://redis:6379/0
WWN_LOG_LEVEL=info
WWN_LOG_FORMAT=json
WWN_ENVIRONMENT=staging
```

Verschlüsseln:

```bash
sops --encrypt --in-place infra/secrets/staging/backend.env
```

Das resultierende File mit `sops:`-Header wird committed.

#### 1.4 Pre-commit-Hook gegen unverschlüsselte Secrets

`.pre-commit-config.yaml` erweitern:

```yaml
- repo: local
  hooks:
    - id: forbid-unencrypted-secrets
      name: Forbid unencrypted secrets
      entry: scripts/check-encrypted-secrets.sh
      language: script
      files: ^infra/secrets/.*\.(env|yaml)$
      pass_filenames: true
```

`scripts/check-encrypted-secrets.sh`:

```bash
#!/usr/bin/env bash
set -euo pipefail

failed=0
for file in "$@"; do
    # SOPS-encrypted Files haben "sops" als Top-Level-Key
    if ! grep -q '^sops:' "$file" 2>/dev/null && \
       ! head -c 4096 "$file" | grep -q '"sops":'; then
        echo "✗ Unencrypted secret detected: $file"
        echo "  Encrypt with: sops --encrypt --in-place $file"
        failed=1
    fi
done
exit $failed
```

Ausführbar machen.

#### 1.5 `docs/secrets.md`

````markdown
# Secret-Management mit SOPS

Secrets werden mit [SOPS](https://getsops.io/) und age-Keys verschlüsselt im
Git-Repo gehalten. Auf den Servern werden sie zur Deploy-Zeit von Ansible
entschlüsselt und als `.env`-Files abgelegt.

## Erst-Setup für einen neuen Maintainer

1. age-Keypair generieren:
   ```bash
   age-keygen -o ~/.config/sops/age/keys.txt
   chmod 600 ~/.config/sops/age/keys.txt
   ```
````

2. Public-Key kopieren (beginnt mit `age1...`)
3. Public-Key zu `.sops.yaml` hinzufügen lassen (PR an bestehenden Maintainer)
4. **Backup**: Private-Key in einen Passwort-Manager kopieren. Ohne Private-Key
   keine Entschlüsselung. Kein zweiter Weg.
5. Bestehender Maintainer rotiert Secrets:
   ```bash
   sops updatekeys infra/secrets/staging/backend.env
   ```
   für jedes Secret-File.

## Secret editieren

```bash
sops infra/secrets/staging/backend.env
```

Öffnet das File entschlüsselt im `$EDITOR`, verschlüsselt beim Speichern wieder.

## Neues Secret anlegen

```bash
echo "MY_KEY=value" > /tmp/new-secret.env
sops --encrypt /tmp/new-secret.env > infra/secrets/staging/new-secret.env
rm /tmp/new-secret.env
git add infra/secrets/staging/new-secret.env
```

## Recovery

Wenn der Private-Key verloren ist:

1. Bestehende Maintainer haben weiter Zugriff
2. Aus Backup wiederherstellen, falls vorhanden
3. Wenn kein Backup: alle Secrets rotieren und neu verschlüsseln

## Was NICHT in SOPS gehört

- Niemals Klartext-Secrets in `.env.example`
- Keine Secrets in CI-Logs ausgeben
- Keine Secrets in Container-Image-Layers
- Keine Secrets als Build-Args (sind in History sichtbar)

````

### 2. Ansible-Struktur

#### 2.1 `infra/ansible/ansible.cfg`

```ini
[defaults]
inventory = inventories
roles_path = roles
host_key_checking = True
retry_files_enabled = False
forks = 5
stdout_callback = yaml
bin_ansible_callbacks = True
gather_timeout = 30

[ssh_connection]
pipelining = True
control_master = auto
control_persist = 10m
````

#### 2.2 `inventories/staging/hosts.yml`

```yaml
all:
  children:
    wwn:
      hosts:
        wwn-staging-01:
          ansible_host: <STAGING_IP_HIER>
          ansible_user: deploy
          ansible_port: 22
          environment: staging
```

#### 2.3 `inventories/production/hosts.yml` (analog)

#### 2.4 `group_vars/all.yml`

```yaml
# Globale Defaults
deploy_user: deploy
app_dir: /opt/wwn
docker_compose_file: compose.prod.yml
image_namespace: <ORG_HIER>
default_versions:
  backend: latest
  frontend: latest
  pyworkers: latest

# SSH-Hardening
sshd_port: 22
sshd_permit_root_login: "no"
sshd_password_authentication: "no"
sshd_pubkey_authentication: "yes"

# Firewall
ufw_allowed_ports:
  - { port: 22, proto: tcp, comment: "SSH" }
  - { port: 80, proto: tcp, comment: "HTTP" }
  - { port: 443, proto: tcp, comment: "HTTPS" }
```

#### 2.5 Rollen

**`roles/common/`**:

- Tasks:
  - User `deploy` anlegen mit SSH-Public-Key
  - sudo-Rechte (NOPASSWD für `docker compose`-Wrapper-Script, nicht generell)
  - SSH-Hardening: `/etc/ssh/sshd_config`-Template
  - `unattended-upgrades` installieren und konfigurieren
  - `fail2ban` installieren mit sshd-Jail
  - `ufw` aufsetzen mit Default-Deny-Inbound, allowed_ports
  - Time-Sync (`systemd-timesyncd` oder `chrony`)
  - Hostname setzen
- Handler: SSH-Reload
- Templates: `sshd_config.j2`, `unattended-upgrades.j2`

**`roles/docker/`**:

- Docker-Engine via offiziellem Repo
- docker-compose-Plugin
- User `deploy` zur `docker`-Group hinzufügen
- Daemon-Config: log-driver, log-opts, default-address-pools für Konflikt-
  Vermeidung mit anderen Diensten
- ghcr.io-Login: PAT aus SOPS, in `~deploy/.docker/config.json` ablegen
  (mit korrekten Permissions)

**`roles/app/`**:

- `/opt/wwn/`-Struktur erzeugen mit korrekten Owner/Permissions
- SOPS-Decrypt der relevanten Secret-Files nach `/opt/wwn/.env.*`
  (root-only readable, mode 0600)
- `compose.prod.yml`, `Caddyfile.prod`, `postgres-init/` rüberkopieren
- ENV-Substitution in `compose.prod.yml` für VERSION/IMAGE_NAMESPACE
- `docker compose pull`
- `docker compose up -d`
- Health-Check: warten bis `/health` von `localhost:80/api/v1/ping` antwortet
- Optional: ältere ungenutzte Images prunen

**`roles/caddy-host/`** (optional, falls Caddy nicht im Compose, sondern auf
dem Host läuft. Empfehlung: Caddy IM Compose lassen. Diese Rolle dann nicht
nötig — entscheide nach Hosting-Setup.)

**`roles/monitoring-agent/`**:

- node_exporter Container starten
- promtail Container starten, sendet an Loki (zentraler Loki im Compose oder
  externe Instanz — frag bei Setup)

#### 2.6 Playbooks

**`playbooks/site.yml`** — voller Setup eines Hosts:

```yaml
---
- name: Initial host setup
  hosts: wwn
  become: true
  roles:
    - common
    - docker
    - monitoring-agent

- name: Application deployment
  hosts: wwn
  become: true
  roles:
    - app
```

**`playbooks/deploy.yml`** — schnelles App-Update:

```yaml
---
- name: Deploy application
  hosts: wwn
  become: true
  vars_prompt:
    - name: target_version
      prompt: "Version to deploy (z.B. 0.1.0)"
      private: false
  vars:
    versions:
      backend: "{{ target_version }}"
      frontend: "{{ target_version }}"
      pyworkers: "{{ target_version }}"
  roles:
    - role: app
      vars:
        skip_secret_decrypt: false
```

**`playbooks/rollback.yml`** — auf vorherige Version zurück:

```yaml
---
- name: Rollback application
  hosts: wwn
  become: true
  vars_prompt:
    - name: rollback_to
      prompt: "Version to roll back to"
      private: false
  tasks:
    - name: Show current versions
      ansible.builtin.command: docker compose -f /opt/wwn/compose.prod.yml ps --format json
      changed_when: false
      register: current

    - name: Confirm rollback
      ansible.builtin.pause:
        prompt: "Confirm rollback to {{ rollback_to }}?"

    - include_role: { name: app }
      vars:
        versions:
          backend: "{{ rollback_to }}"
          frontend: "{{ rollback_to }}"
          pyworkers: "{{ rollback_to }}"
```

#### 2.7 `requirements.yml` für Ansible-Collections

```yaml
collections:
  - name: community.general
  - name: community.docker
  - name: ansible.posix
```

### 3. Terraform-Skelett

Anhand des gewählten Providers, Beispiel **Hetzner Cloud**:

#### 3.1 `infra/terraform/versions.tf`

```hcl
terraform {
  required_version = ">= 1.7"
  required_providers {
    hcloud = {
      source  = "hetznercloud/hcloud"
      version = "~> 1.45"
    }
  }
}
```

#### 3.2 `infra/terraform/modules/server/`

`main.tf`:

```hcl
variable "name" {
  type = string
}

variable "server_type" {
  type    = string
  default = "cpx21"   # 3 vCPU, 4GB RAM, AMD
}

variable "image" {
  type    = string
  default = "ubuntu-24.04"
}

variable "location" {
  type    = string
  default = "fsn1"    # Falkenstein
}

variable "ssh_keys" {
  type = list(string)
}

variable "user_data" {
  type    = string
  default = ""
}

resource "hcloud_server" "this" {
  name         = var.name
  server_type  = var.server_type
  image        = var.image
  location     = var.location
  ssh_keys     = var.ssh_keys
  user_data    = var.user_data
  public_net {
    ipv4_enabled = true
    ipv6_enabled = true
  }
  labels = {
    project = "wwn"
    env     = var.name
  }
}

output "ipv4" {
  value = hcloud_server.this.ipv4_address
}

output "ipv6" {
  value = hcloud_server.this.ipv6_address
}
```

#### 3.3 `infra/terraform/environments/staging/main.tf`

```hcl
provider "hcloud" {
  token = var.hcloud_token
}

variable "hcloud_token" {
  type      = string
  sensitive = true
}

variable "ssh_key_ids" {
  type = list(string)
}

module "server" {
  source      = "../../modules/server"
  name        = "wwn-staging-01"
  server_type = "cpx21"
  ssh_keys    = var.ssh_key_ids
  user_data   = file("${path.module}/cloud-init.yaml")
}

output "staging_ip" {
  value = module.server.ipv4
}
```

`cloud-init.yaml`:

```yaml
#cloud-config
users:
  - name: deploy
    sudo: ALL=(ALL) NOPASSWD:ALL
    shell: /bin/bash
    ssh_authorized_keys:
      - <SSH_PUBLIC_KEY_HIER>

package_update: true
package_upgrade: true
packages:
  - python3
  - sudo
```

#### 3.4 `environments/production/main.tf` (analog mit größerem Server)

#### 3.5 `terraform.tfvars.example`

```hcl
hcloud_token = "GET_FROM_HETZNER_CONSOLE"
ssh_key_ids  = ["123456"]   # IDs aus `hcloud ssh-key list`
```

`terraform.tfvars` selbst nicht committen (in `.gitignore`).

#### 3.6 State-Backend

Initial: lokaler State, in `.gitignore`. README erklärt Migration zu
S3-kompatiblem Backend (Hetzner Object Storage) für Team-Setup:

```hcl
# Später in versions.tf:
# terraform {
#   backend "s3" {
#     bucket   = "wwn-tfstate"
#     key      = "staging/terraform.tfstate"
#     endpoint = "https://fsn1.your-objectstorage.com"
#     region   = "eu-central"
#     skip_credentials_validation = true
#     skip_region_validation      = true
#     skip_metadata_api_check     = true
#   }
# }
```

#### 3.7 Terraform-README

`infra/terraform/README.md`:

- Voraussetzungen (Token, SSH-Key)
- Erstmaliger Apply für staging
- Output → Ansible-Inventory mappen
- destroy-Hinweis mit Warnung

### 4. `scripts/deploy.sh`

```bash
#!/usr/bin/env bash
set -euo pipefail

ENV="${1:-}"
VERSION="${2:-}"

usage() {
    echo "Usage: $0 <staging|production> <version>"
    echo "Example: $0 staging 0.1.0"
    exit 1
}

[ -z "$ENV" ] && usage
[ -z "$VERSION" ] && usage

case "$ENV" in
    staging|production) ;;
    *) usage ;;
esac

if [ "$ENV" = "production" ]; then
    echo "⚠  Deploying to PRODUCTION (version $VERSION)"
    read -rp "Type 'production' to confirm: " confirm
    [ "$confirm" = "production" ] || { echo "Aborted."; exit 1; }
fi

cd "$(git rev-parse --show-toplevel)/infra/ansible"

ansible-playbook \
    -i "inventories/${ENV}/hosts.yml" \
    playbooks/deploy.yml \
    -e "target_version=${VERSION}"
```

Ausführbar machen.

### 5. `infra/ansible/README.md`

Kurz-Anleitung:

- Voraussetzungen (Ansible installieren, age-Key, Access)
- Erstmaliges Setup eines Hosts: Terraform → Inventory ergänzen → site.yml
- Deploy: `bash scripts/deploy.sh staging 0.1.0`
- Rollback: `ansible-playbook -i inventories/staging/hosts.yml playbooks/rollback.yml`

### 6. ENV-Substitution für `compose.prod.yml`

`compose.prod.yml` braucht `${VERSION}` und `${IMAGE_NAMESPACE}`. Variante:

- Ansible-Template-Modul mit Vars: schreibt fertige Datei nach Server
- ODER: das File wird as-is rüberkopiert, ENV-Vars kommen aus
  `/opt/wwn/.env.compose`

Empfehlung: Template via Ansible. Sauberer.

### 7. CI-Integration (optional, frag)

`.github/workflows/deploy.yml` (manueller Trigger):

```yaml
name: Deploy
on:
  workflow_dispatch:
    inputs:
      environment:
        type: choice
        options: [staging, production]
      version:
        type: string
        required: true

jobs:
  deploy:
    runs-on: ubuntu-latest
    environment: ${{ inputs.environment }}
    steps:
      - uses: actions/checkout@v4
      - name: Set up SSH
        run: |
          mkdir -p ~/.ssh
          echo "${{ secrets.DEPLOY_SSH_KEY }}" > ~/.ssh/id_ed25519
          chmod 600 ~/.ssh/id_ed25519
          ssh-keyscan -H ${{ secrets.DEPLOY_HOST }} >> ~/.ssh/known_hosts
      - name: Set up age key
        run: |
          mkdir -p ~/.config/sops/age
          echo "${{ secrets.SOPS_AGE_KEY }}" > ~/.config/sops/age/keys.txt
      - name: Install Ansible
        run: pip install ansible
      - name: Deploy
        run: bash scripts/deploy.sh ${{ inputs.environment }} ${{ inputs.version }}
```

In separater PR aktivieren, wenn manuelles Deployment einmal erfolgreich gelaufen ist.

## Vorgehen (verbindlich)

1. Plan zeigen, **vorher die Klärungs-Fragen oben mit dem Maintainer durchgehen**
2. Freigabe abwarten
3. Implementierung in Schritten:
   a) SOPS-Setup + `.sops.yaml` + Pre-commit + Beispiel-Secret
   b) Ansible-Struktur (Rollen, Playbooks, Inventories)
   c) Terraform-Skelett für gewählten Provider
   d) Deploy-Script und CI-Workflow-Stub
4. **Nicht** auf einem Server ausführen — alle Tests sind syntaktisch:
   - `ansible-lint`
   - `ansible-playbook --syntax-check`
   - `terraform fmt -check`
   - `terraform validate` (wenn Provider erreichbar; ggf. mit `-backend=false`)
5. Pre-commit-Hook lokal testen mit einem Test-Secret
6. Nicht committen

## Erfolgs-Kriterien

- [ ] `.sops.yaml` syntaktisch korrekt, age-Key vorhanden
- [ ] Beispiel-Secret-File ist verschlüsselt (sieht aus wie YAML mit `sops:`-Header)
- [ ] Pre-commit-Hook blockt unverschlüsseltes Test-File
- [ ] `ansible-lint infra/ansible/` grün (Warnings OK, Errors nicht)
- [ ] `ansible-playbook --syntax-check playbooks/site.yml` grün
- [ ] `ansible-playbook --syntax-check playbooks/deploy.yml` grün
- [ ] `terraform fmt -check` grün
- [ ] `terraform validate` grün (falls Provider verfügbar)
- [ ] `scripts/deploy.sh` ausführbar, hat Confirmation-Prompt für production
- [ ] `docs/secrets.md` enthält klare Anleitung für Erst-Setup und Recovery
- [ ] `infra/ansible/README.md` und `infra/terraform/README.md` vorhanden

## Mögliche Stolpersteine

- **age-Key vs. GPG**: SOPS unterstützt beides. age ist neuer und einfacher.
  Falls der Maintainer GPG bevorzugt: anpassen, aber Default ist age.
- **Ansible und Docker-Compose**: `community.docker.docker_compose_v2`-Modul
  ist die richtige Wahl. Alternativ Shell-Calls — pragmatisch aber weniger
  idempotent.
- **ghcr.io-Pull-Authentifizierung**: Server braucht PAT (read:packages-Scope).
  Token in SOPS, via Ansible in `~/.docker/config.json`.
- **Terraform-State**: lokaler State ist für Solo-Maintainer OK, Backup nicht
  vergessen. Bei Team-Setup muss der State in einen Remote-Backend.
- **First Apply ohne State**: `terraform plan` zeigt vollständige Erstellung,
  bei zweiter Maschine das gleiche Modul wiederverwenden.
- **Hetzner Cloud-Limits**: Standard-Account hat Server-Limit. Ggf. erhöhen
  lassen vor Production.
- **TLS und Caddy auf Hetzner**: braucht Domain-DNS, die auf die IP zeigt,
  damit ACME-HTTP-Challenge funktioniert. Reihenfolge: Terraform apply → IP
  notieren → DNS setzen → Ansible playbook.

## Was diese Session NICHT tut

- Kein tatsächliches Provisioning eines Servers
- Kein tatsächliches Deployment
- Keine DB-Backup-Konfiguration (separater Schritt)
- Keine Disaster-Recovery-Übung
- Keine Multi-Host- oder Cluster-Architektur

## Nach der Session — Hausaufgabe für den Maintainer

1. Hetzner-Account, API-Token erzeugen
2. SSH-Public-Key hochladen
3. `terraform apply` für staging
4. DNS-Eintrag setzen (api-staging.worldweathernews.com → IP)
5. Server-IP in Ansible-Inventory eintragen
6. age-Key auf Maintainer-Maschine, Secrets befüllen, encrypten
7. `ansible-playbook -i inventories/staging/hosts.yml playbooks/site.yml`
8. Health-Check: https://api-staging.worldweathernews.com/health
9. Wenn alles läuft: production analog, mit größerem Server

## Suggested Commit-Message

```
feat(infra): add ansible deployment, sops secrets, terraform skeleton
```
