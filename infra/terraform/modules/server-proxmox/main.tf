// Modul für eine Proxmox-VM. Nutzt das Klonen einer Cloud-init-fähigen
// Template-VM (Maintainer-Hausaufgabe: Template anlegen, Storage und Bridge
// im environments/<env>/terraform.tfvars setzen).

variable "name" {
  type        = string
  description = "Hostname und VM-Name."
}

variable "node_name" {
  type        = string
  description = "Proxmox-Node, auf dem die VM laufen soll."
}

variable "template_id" {
  type        = number
  description = "VMID der Cloud-init-Template-VM (vorab manuell angelegt)."
}

variable "vm_id" {
  type        = number
  description = "Gewünschte VMID. Muss eindeutig sein."
}

variable "cores" {
  type        = number
  default     = 2
  description = "Anzahl vCPUs."
}

variable "memory_mb" {
  type        = number
  default     = 4096
  description = "RAM in MiB."
}

variable "disk_size_gb" {
  type        = number
  default     = 32
  description = "Disk-Größe in GiB."
}

variable "datastore_id" {
  type        = string
  description = "Storage-Pool für Disk und Cloud-init-ISO (z. B. local-lvm)."
}

variable "network_bridge" {
  type        = string
  default     = "vmbr0"
  description = "Proxmox-Bridge fürs Netzwerk-Interface."
}

variable "ipv4_address" {
  type        = string
  description = "Statische IPv4 inkl. Präfix, z. B. \"10.100.100.21/24\"."
}

variable "ipv4_gateway" {
  type        = string
  description = "Default-Gateway."
}

variable "user_data" {
  type        = string
  description = "Cloud-init user-data (YAML, gerendert vom env-Modul)."
}

variable "tags" {
  type        = list(string)
  default     = ["wwn"]
  description = "Tags für die VM."
}

resource "proxmox_virtual_environment_vm" "this" {
  name      = var.name
  node_name = var.node_name
  vm_id     = var.vm_id
  tags      = var.tags

  clone {
    vm_id = var.template_id
    full  = true
  }

  cpu {
    cores = var.cores
    type  = "host"
  }

  memory {
    dedicated = var.memory_mb
  }

  disk {
    datastore_id = var.datastore_id
    interface    = "scsi0"
    size         = var.disk_size_gb
  }

  network_device {
    bridge = var.network_bridge
  }

  initialization {
    datastore_id = var.datastore_id
    ip_config {
      ipv4 {
        address = var.ipv4_address
        gateway = var.ipv4_gateway
      }
    }
    user_data_file_id = proxmox_virtual_environment_file.user_data.id
  }
}

resource "proxmox_virtual_environment_file" "user_data" {
  content_type = "snippets"
  datastore_id = var.datastore_id
  node_name    = var.node_name

  source_raw {
    file_name = "${var.name}-user-data.yaml"
    data      = var.user_data
  }
}

output "ipv4" {
  value = trimsuffix(split("/", var.ipv4_address)[0], "/")
}

output "vm_id" {
  value = proxmox_virtual_environment_vm.this.vm_id
}
