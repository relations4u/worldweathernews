# infra/ansible

Server-Konfiguration und Deploy-Workflow für die Forschungs-Phase. Ziele:

1. Erst-Setup eines neuen Ubuntu-24.04-Hosts (`site.yml`)
2. Wiederholtes App-Update mit neuer Image-Version (`deploy.yml`)
3. Rollback auf eine frühere Version (`rollback.yml`)

**Diese Session installiert nur das Skelett.** Tatsächliches Provisioning ist
Hausaufgabe des Maintainers.

## Voraussetzungen

| Tool                                              | Quelle                     |
| ------------------------------------------------- | -------------------------- |
| `ansible-core` 2.20                               | `mise install`             |
| `ansible-lint` 26                                 | `mise install`             |
| `sops` 3.12                                       | `mise install`             |
| `age` / `age-keygen`                              | System (`apt install age`) |
| Privater age-Key in `~/.config/sops/age/keys.txt` | siehe `docs/secrets.md`    |
| SSH-Zugriff auf die Zielhosts                     | manuell                    |

Collections nach Erst-Klon installieren:

```bash
cd infra/ansible
ansible-galaxy collection install -r requirements.yml
```

## Inventory

`inventories/production/hosts.yml` definiert zwei Hosts:

- `wwn-prod` (10.100.100.70) — Application-Stack
- `wwn-mon` (10.100.100.69) — Monitoring-Stack (zentral, Stack-Deploy ist Follow-up)

Beide gehören zur Gruppe `wwn`. Untergruppen `app` und `mon` steuern, welche
Rolle wo greift.

## Erst-Setup eines Hosts (Bootstrap)

Auf einem frisch provisionierten Host (z. B. via Terraform-cloud-init mit
`deploy`-User vorbereitet) reicht:

```bash
ansible-playbook -i inventories/production/hosts.yml playbooks/site.yml
```

Auf wwn-prod, der **manuell** angelegt wurde und nur den `hwr`-User kennt,
einmalig mit Override:

```bash
ansible-playbook -i inventories/production/hosts.yml \
    playbooks/site.yml -e ansible_user=hwr
```

Die `common`-Rolle legt dann den `deploy`-User an, autorisiert den
Maintainer-Public-Key, härtet SSH und richtet UFW/fail2ban ein. Anschließend
nutzt jedes weitere Playbook den Default `ansible_user=deploy`.

## Deploy einer neuen Version

```bash
bash scripts/deploy.sh production 0.1.0
```

Der Wrapper ruft `ansible-playbook playbooks/deploy.yml` und prompt
`target_version`. Versionen sind die Image-Tags aus
`ghcr.io/relations4u/wwn-{backend,frontend,pyworkers}`.

## Rollback

```bash
ansible-playbook -i inventories/production/hosts.yml playbooks/rollback.yml
```

Promptet die Ziel-Version, zeigt den aktuellen State, fordert Bestätigung,
re-running die `app`-Rolle mit der gewählten Version.

## Was diese Skelett-Session NICHT tut

- Kein tatsächliches Provisioning (siehe `docs/development.md`-Roadmap)
- Keine zentrale Monitoring-Stack-Rolle für `wwn-mon` — nur ein
  `monitoring-agent`-Stub. Der zentrale Stack auf wwn-mon ist Follow-up.
- Kein DB-Backup-Workflow (separater Schritt, Session 12 oder Follow-up)
- Kein `terraform import` der bestehenden VMs — siehe `infra/terraform/README.md`

## Lokale Validierung

```bash
cd infra/ansible
ansible-playbook --syntax-check playbooks/site.yml
ansible-playbook --syntax-check playbooks/deploy.yml
ansible-playbook --syntax-check playbooks/rollback.yml
ansible-lint .
```

`ansible-lint` warnt evtl. zu Best-Practices, die wir im Skelett bewusst nicht
adressieren (z. B. fehlende `name:`-Felder bei trivialen Tasks). Errors müssen
grün sein.

## Caddy

**Caddy wird von dieser Rolle nicht angefasst.** Reverse-Proxy läuft als
eigener Stack unter `/srv/wwn/caddy` auf wwn-prod und wird über
`infra/deploy/deploy-caddy.sh` gepflegt. Der Caddy-Block in
`infra/compose/compose.prod.yml` wird in einer separaten PR entfernt.
