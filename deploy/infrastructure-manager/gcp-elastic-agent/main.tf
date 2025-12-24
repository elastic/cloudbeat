terraform {
  required_version = ">= 1.0"
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 5.0"
    }
    null = {
      source  = "hashicorp/null"
      version = "~> 3.0"
    }
  }
}

provider "google" {
  project = var.project_id
}

locals {
  # Use suffix from deploy.sh to ensure all resource names stay within GCP limits and allow multiple deployments
  resource_suffix = var.resource_suffix
  sa_name         = "elastic-agent-sa-${local.resource_suffix}"
  sa_email        = module.service_account.email
  network_name    = "elastic-agent-net-${local.resource_suffix}"
  instance_name   = "elastic-agent-vm-${local.resource_suffix}"
}

# Resource suffix for all resource names
variable "resource_suffix" {
  description = "Unique suffix for resource names (8 hex characters)"
  type        = string
}

module "service_account" {
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

module "startup_validation" {
  source = "./modules/startup_validation"

  enabled       = var.startup_validation_enabled
  project_id    = var.project_id
  instance_name = local.instance_name
  instance_id   = module.compute_instance.id
  zone          = var.zone
  timeout       = var.startup_timeout_seconds

  depends_on = [
    module.compute_instance
  ]
}
