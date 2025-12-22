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

module "compute_instance" {
  source = "./modules/compute_instance"

  deployment_name         = var.deployment_name
  machine_type            = var.machine_type
  zone                    = var.zone
  network_self_link       = google_compute_network.elastic_agent.self_link
  sa_email                = local.sa_email
  elastic_agent_version   = var.elastic_agent_version
  elastic_artifact_server = var.elastic_artifact_server
  install_command         = local.install_command
  fleet_url               = var.fleet_url
  enrollment_token        = var.enrollment_token

  depends_on = [
    module.service_account
  ]
}
