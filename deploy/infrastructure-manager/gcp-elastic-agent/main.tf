terraform {
  required_version = ">= 1.0"
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 5.0"
    }
  }
}

provider "google" {
  project = var.project_id
}

locals {
  create_service_account = var.service_account_name == ""
  sa_name                = local.create_service_account ? "${var.deployment_name}-sa" : var.service_account_name
  sa_email               = local.create_service_account ? module.service_account[0].email : "${var.service_account_name}@${var.project_id}.iam.gserviceaccount.com"
  network_name           = "${var.deployment_name}-network"

  # Determine install command based on version
  install_command = startswith(var.elastic_agent_version, "9.") ? "sudo ./elastic-agent install --non-interactive --install-servers" : "sudo ./elastic-agent install --non-interactive"
}

module "service_account" {
  count = local.create_service_account ? 1 : 0

  source = "./modules/service_account"

  project_id           = var.project_id
  service_account_name = local.sa_name
  scope                = var.scope
  parent_id            = var.parent_id
}

# VPC Network
resource "google_compute_network" "elastic_agent" {
  name                    = local.network_name
  auto_create_subnetworks = true
  routing_mode            = "REGIONAL"
}

# Firewall rule for SSH (optional)
resource "google_compute_firewall" "ssh" {
  count   = var.allow_ssh ? 1 : 0
  name    = "elastic-agent-firewall-rule"
  network = google_compute_network.elastic_agent.self_link

  allow {
    protocol = "tcp"
    ports    = ["22"]
  }

  source_ranges = ["0.0.0.0/0"]
}

# Compute Instance
resource "google_compute_instance" "elastic_agent" {
  name         = var.deployment_name
  machine_type = var.machine_type
  zone         = var.zone

  labels = {
    name = "elastic-agent"
  }

  boot_disk {
    initialize_params {
      image = "ubuntu-os-cloud/ubuntu-minimal-2204-lts"
      size  = 32
      type  = "pd-standard"
    }
    auto_delete = true
  }

  network_interface {
    network = google_compute_network.elastic_agent.self_link

    access_config {
      # Ephemeral public IP
    }
  }

  service_account {
    email = local.sa_email
    scopes = [
      "https://www.googleapis.com/auth/cloud-platform",
      "https://www.googleapis.com/auth/cloudplatformorganizations",
    ]
  }

  metadata_startup_script = <<-EOT
    #!/bin/bash
    set -x
    ElasticAgentArtifact=elastic-agent-${var.elastic_agent_version}-linux-x86_64
    curl -L -O ${var.elastic_artifact_server}/$ElasticAgentArtifact.tar.gz
    tar xzvf $ElasticAgentArtifact.tar.gz
    cd $ElasticAgentArtifact
    ${local.install_command} --url=${var.fleet_url} --enrollment-token=${var.enrollment_token}
  EOT

  depends_on = [
    module.service_account,
  ]
}