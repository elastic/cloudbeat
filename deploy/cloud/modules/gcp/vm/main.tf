resource "random_id" "id" {
  byte_length = 4
}

locals {
  vm_private_key_file = "${path.module}/gcp-vm-${random_id.id.hex}.pem"
  vm_username         = "ubuntu"
  deploy_name         = "${var.deployment_name}-${random_id.id.hex}"
}

resource "tls_private_key" "gcp_vm_key" {
  algorithm = "RSA"
  rsa_bits  = 4096
}

resource "local_file" "ssh_private_key" {
  filename        = local.vm_private_key_file
  content         = tls_private_key.gcp_vm_key.private_key_pem
  file_permission = 0400
}

resource "google_compute_instance" "vm_instance" {
  name         = local.deploy_name
  machine_type = var.machine_type
  zone         = var.zone
  labels       = var.specific_tags

  boot_disk {
    initialize_params {
      image = var.disk_image
    }
  }

  network_interface {
    network = var.network

    access_config {
      // Ephemeral public IP
    }
  }

  metadata = {
    ssh-keys = "${local.vm_username}:${tls_private_key.gcp_vm_key.public_key_openssh}"
  }

  service_account {
    scopes = var.scopes
  }
}
