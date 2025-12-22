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
  region  = var.region
}

# Simple test resource - creates a storage bucket
resource "google_storage_bucket" "test_bucket" {
  name          = "${var.project_id}-infra-manager-test-${var.deployment_name}"
  location      = var.region
  force_destroy = true

  uniform_bucket_level_access = true

  labels = {
    environment = "test"
    created_by  = "infrastructure-manager"
  }
}
