# infra/terraform

Provisioning der Infrastruktur. Aktuell aktiv: **Proxmox** für die
Forschungs-Phase. Stub: **Hetzner** für eine spätere Migration nach
docs/adr/0002-hetzner-migration.md (TODO).

**Diese Session installiert nur das Skelett.** Kein `terraform apply` durch
Claude Code — siehe „Bestehende VMs einbinden" unten.

## Layout

```
infra/terraform/
├── versions.tf                 # Provider-Pins (Proxmox + Hetzner)
├── modules/
│   ├── server-proxmox/         # Aktives Modul: cloud-init-Template klonen
│   └── server-hetzner/         # Stub für Migration
└── environments/
    └── production/
        ├── main.tf
        ├── cloud-init.yaml.tftpl
        └── terraform.tfvars.example   # Vorlage für lokale terraform.tfvars
```

## Voraussetzungen

| Tool                   | Quelle                                            |
| ---------------------- | ------------------------------------------------- |
| `terraform` 1.15       | `mise install`                                    |
| Proxmox-Cluster läuft  | manuell                                           |
| Cloud-init-Template-VM | manuell (siehe `terraform.tfvars.example`)        |
| API-Token              | Proxmox-UI: Datacenter → Permissions → API Tokens |

## Initialer Apply (für neuen Host, nicht für die bestehenden!)

```bash
cd infra/terraform/environments/production

cp terraform.tfvars.example terraform.tfvars
$EDITOR terraform.tfvars                # Werte einsetzen

terraform init
terraform plan
terraform apply
```

Output: `wwn_prod_ipv4`, `wwn_mon_ipv4`. In Ansible-Inventory eintragen
(falls vom Default `10.100.100.{21,22}` abweichend).

## Bestehende VMs einbinden (Maintainer-Hausaufgabe)

`wwn-prod` und `wwn-mon` wurden manuell in der Proxmox-UI erstellt. Bevor
`terraform apply` jemals gegen Production läuft, müssen die VMs in den
Terraform-State importiert werden — sonst legt Terraform sie als „neu" an
und kollidiert mit den bestehenden VMIDs.

Schritte:

1. `terraform.tfvars` ausfüllen (`proxmox_endpoint`, `proxmox_api_token`, …)
2. `terraform init`
3. Pro VM:
   ```bash
   # Format: <node>/qemu/<vmid>
   terraform import 'module.wwn_prod.proxmox_virtual_environment_vm.this' <node>/qemu/121
   terraform import 'module.wwn_mon.proxmox_virtual_environment_vm.this'  <node>/qemu/122
   ```
4. `terraform plan` muss anschließend „No changes" zeigen. Wenn nicht: die
   VM-Konfiguration in `main.tf` an den realen Stand anpassen, NICHT
   umgekehrt — sonst würde Terraform die VMs neu konfigurieren und
   möglicherweise Daten verlieren.

Wenn die VMs noch nicht zum Importieren bereit sind (z. B. weil Disk-Größe
oder Netzwerk-Bridge in `main.tf` nicht zur Realität passen): erst Werte
in `main.tf` und/oder Variablen in `terraform.tfvars` an den Ist-Zustand
anpassen, dann importieren.

## State

- Initial: lokal im `terraform/`-Workingdir, in `.gitignore`.
- Backup: `terraform.tfstate` regelmäßig in Off-Site-Backup mitnehmen
  (gleicher Sicherheits-Bereich wie SOPS-Keys; State enthält Token-Werte).
- Team-Setup später: Hetzner Object Storage als S3-kompatibles Backend,
  Beispiel-Snippet steht auskommentiert in `versions.tf`-Kommentar.

## Hetzner-Stub aktivieren (zukünftiger Schritt)

```bash
cd infra/terraform/environments/production
terraform import …                # noch nichts; Hetzner ist Stub
```

Erst wenn die ADR genehmigt ist:

1. Modul in `main.tf` instantiieren (`module "wwn_prod_hetzner"`)
2. `hcloud_token` als Variable, Secret aus SOPS
3. DNS umstellen — siehe Cloudflare-Doku in `CLAUDE.md`

## Lokale Validierung

```bash
cd infra/terraform
terraform fmt -recursive -check
terraform -chdir=environments/production init -backend=false
terraform -chdir=environments/production validate
```

`-backend=false` springt am State-Backend vorbei und braucht keinen Proxmox-
Zugriff zum Validieren.
