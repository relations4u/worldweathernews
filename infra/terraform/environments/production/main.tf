// Production-Environment, aktuell auf eigenem Proxmox-Host.
//
// Achtung: wwn-prod und wwn-mon existieren bereits manuell. `terraform apply`
// würde sie ohne vorheriges `terraform import` neu anlegen. Siehe
// infra/terraform/README.md für den Import-Workflow.

provider "proxmox" {
  endpoint  = var.proxmox_endpoint
  api_token = var.proxmox_api_token
  insecure  = var.proxmox_insecure

  ssh {
    agent    = true
    username = var.proxmox_ssh_user
  }
}

variable "proxmox_endpoint" {
  type        = string
  description = "Proxmox-API-URL, z. B. https://proxmox.lan:8006/"
}

variable "proxmox_api_token" {
  type        = string
  sensitive   = true
  description = "API-Token im Format \"USER@REALM!TOKENID=SECRET\"."
}

variable "proxmox_insecure" {
  type        = bool
  default     = false
  description = "TLS-Verifikation überspringen (nur für selbstsigniertes Cluster-Cert)."
}

variable "proxmox_ssh_user" {
  type        = string
  default     = "root"
  description = "SSH-User auf dem Proxmox-Node (für File-Uploads)."
}

variable "proxmox_node_name" {
  type        = string
  description = "Name des Proxmox-Nodes."
}

variable "proxmox_template_id" {
  type        = number
  description = "VMID der vorbereiteten Ubuntu-24.04-cloud-init-Template-VM."
}

variable "proxmox_datastore_id" {
  type        = string
  default     = "local-lvm"
  description = "Storage-Pool für VM-Disks und cloud-init-Snippets."
}

variable "ipv4_gateway" {
  type        = string
  description = "LAN-Default-Gateway, z. B. 10.100.100.1"
}

locals {
  authorized_keys = [
    "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIC8I0s/wZ7BLj+m+T3anSKZoHGwzf7DkrxIMj4y3WB1S hwr@relations4u.de"
  ]
  cloud_init_user_data = templatefile("${path.module}/cloud-init.yaml.tftpl", {
    deploy_user     = "deploy"
    authorized_keys = local.authorized_keys
  })
}

module "wwn_prod" {
  source       = "../../modules/server-proxmox"
  name         = "wwn-prod"
  node_name    = var.proxmox_node_name
  template_id  = var.proxmox_template_id
  vm_id        = 121
  cores        = 4
  memory_mb    = 8192
  disk_size_gb = 64
  datastore_id = var.proxmox_datastore_id
  ipv4_address = "10.100.100.70/24"
  ipv4_gateway = var.ipv4_gateway
  user_data    = local.cloud_init_user_data
  tags         = ["wwn", "wwn-prod", "app"]
}

module "wwn_mon" {
  source       = "../../modules/server-proxmox"
  name         = "wwn-mon"
  node_name    = var.proxmox_node_name
  template_id  = var.proxmox_template_id
  vm_id        = 122
  cores        = 2
  memory_mb    = 4096
  disk_size_gb = 64
  datastore_id = var.proxmox_datastore_id
  ipv4_address = "10.100.100.69/24"
  ipv4_gateway = var.ipv4_gateway
  user_data    = local.cloud_init_user_data
  tags         = ["wwn", "wwn-mon", "monitoring"]
}

output "wwn_prod_ipv4" {
  value = module.wwn_prod.ipv4
}

output "wwn_mon_ipv4" {
  value = module.wwn_mon.ipv4
}
