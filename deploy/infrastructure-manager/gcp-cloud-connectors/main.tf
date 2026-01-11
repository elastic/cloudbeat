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
  # Use suffix from deploy.sh to ensure all resource names stay within GCP limits
  resource_suffix = var.resource_suffix
  sa_name         = "${var.target_service_account_name}-${local.resource_suffix}"
  pool_name       = "${var.wif_pool_name}-${local.resource_suffix}"
  provider_name   = "${var.wif_provider_name}-${local.resource_suffix}"
}

# Resource suffix for all resource names
variable "resource_suffix" {
  description = "Unique suffix for resource names (8 hex characters)"
  type        = string
}

# Workload Identity Federation
module "workload_identity" {
  source = "./modules/workload_identity"

  project_id      = var.project_id
  pool_name       = local.pool_name
  provider_name   = local.provider_name
  oidc_issuer_uri = var.oidc_issuer_uri
}

module "target_service_account" {
  source = "./modules/target_service_account"

  project_id                  = var.project_id
  target_service_account_name = local.sa_name
  wif_pool_name               = module.workload_identity.pool_name
  elastic_resource_id         = var.elastic_resource_id
  scope                       = var.scope
  parent_id                   = var.parent_id
}
