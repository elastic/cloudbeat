terraform {
  required_version = ">= 1.0"
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 5.0"
    }
    random = {
      source  = "hashicorp/random"
      version = "~> 3.0"
    }
  }
}

provider "google" {
  project = var.project_id
}

locals {
  create_service_account = var.service_account_name == ""
  # Use random suffix to ensure all resource names stay within GCP limits and allow multiple deployments
  resource_suffix        = random_id.resource_suffix.hex
  sa_name                = local.create_service_account ? "elastic-agent-sa-${local.resource_suffix}" : var.service_account_name
  sa_email               = local.create_service_account ? module.service_account[0].email : "${var.service_account_name}@${var.project_id}.iam.gserviceaccount.com"
  network_name           = "elastic-agent-net-${local.resource_suffix}"
  instance_name          = "elastic-agent-vm-${local.resource_suffix}"
}

# Generate random suffix for all resource names
# Each deployment (terraform state) gets a unique suffix, allowing multiple deployments per project
resource "random_id" "resource_suffix" {
  byte_length = 4

  keepers = {
    # Unique per deployment - use deployment_name + timestamp to ensure uniqueness
    # This allows multiple deployments with same name over time
    deployment_id = "${var.deployment_name}-${formatdate("YYYYMMDDhhmmss", timestamp())}"
  }

  lifecycle {
    ignore_changes = [keepers]
  }
}

module "service_account" {
  count = local.create_service_account ? 1 : 0

  source = "./modules/service_account"

  project_id           = var.project_id
  service_account_name = local.sa_name
  scope                = var.scope
  parent_id            = var.parent_id
}

module "compute_instance" {
  source = "./modules/compute_instance"

  instance_name           = local.instance_name
  network_name            = local.network_name
  machine_type            = var.machine_type
  zone                    = var.zone
  sa_email                = local.sa_email
  elastic_agent_version   = var.elastic_agent_version
  elastic_artifact_server = var.elastic_artifact_server
  fleet_url               = var.fleet_url
  enrollment_token        = var.enrollment_token

  depends_on = [
    module.service_account
  ]
}
