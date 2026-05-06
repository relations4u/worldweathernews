// TODO Stub für die Hetzner-Migration (siehe ADR — Hetzner ist Migrations-
// Pfad nach der Forschungs-Phase). Aktuell noch nicht aktiv genutzt;
// `terraform validate` muss das Modul trotzdem akzeptieren.
//
// Wenn Hetzner aktiviert wird:
//   - environments/production/main.tf: zusätzliches `module "server_hetzner"`
//   - hcloud_token via Secret aus infra/secrets/production/hetzner.env
//   - cloud-init wird aus `templatefile()` gerendert wie beim Proxmox-Modul

variable "name" {
  type        = string
  description = "Hostname und Hetzner-Server-Name."
}

variable "server_type" {
  type        = string
  default     = "cpx21"
  description = "Hetzner-Server-Typ. cpx21 = 3 vCPU, 4 GB RAM (AMD)."
}

variable "image" {
  type    = string
  default = "ubuntu-24.04"
}

variable "location" {
  type        = string
  default     = "fsn1"
  description = "Hetzner-Standort (fsn1 = Falkenstein, nbg1 = Nürnberg)."
}

variable "ssh_keys" {
  type        = list(string)
  description = "Liste der Hetzner-SSH-Key-IDs (numerisch oder Name)."
}

variable "user_data" {
  type    = string
  default = ""
}

resource "hcloud_server" "this" {
  name        = var.name
  server_type = var.server_type
  image       = var.image
  location    = var.location
  ssh_keys    = var.ssh_keys
  user_data   = var.user_data

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
