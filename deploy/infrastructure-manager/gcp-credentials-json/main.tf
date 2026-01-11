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
  # Use suffix from deploy.sh to ensure all resource names stay within GCP limits and allow multiple deployments
  resource_suffix = var.resource_suffix
  sa_name         = "elastic-agent-sa-${local.resource_suffix}"
}

# Service Account
resource "google_service_account" "elastic_agent" {
  account_id   = local.sa_name
  display_name = "Elastic Agent service account"
  project      = var.project_id
}

# Service Account Key
resource "google_service_account_key" "elastic_agent_key" {
  service_account_id = google_service_account.elastic_agent.name
}

# Project-level IAM bindings
resource "google_project_iam_member" "cloudasset_viewer" {
  count   = var.scope == "projects" ? 1 : 0
  project = var.parent_id
  role    = "roles/cloudasset.viewer"
  member  = "serviceAccount:${google_service_account.elastic_agent.email}"
}

resource "google_project_iam_member" "browser" {
  count   = var.scope == "projects" ? 1 : 0
  project = var.parent_id
  role    = "roles/browser"
  member  = "serviceAccount:${google_service_account.elastic_agent.email}"
}

# Organization-level IAM bindings
resource "google_organization_iam_member" "cloudasset_viewer_org" {
  count  = var.scope == "organizations" ? 1 : 0
  org_id = var.parent_id
  role   = "roles/cloudasset.viewer"
  member = "serviceAccount:${google_service_account.elastic_agent.email}"
}

resource "google_organization_iam_member" "browser_org" {
  count  = var.scope == "organizations" ? 1 : 0
  org_id = var.parent_id
  role   = "roles/browser"
  member = "serviceAccount:${google_service_account.elastic_agent.email}"
}

# Secret Manager secret to store the service account key securely
resource "google_secret_manager_secret" "sa_key" {
  secret_id = "elastic-agent-sa-key-${local.resource_suffix}"
  project   = var.project_id

  replication {
    auto {}
  }
}

# Store the service account key in Secret Manager
resource "google_secret_manager_secret_version" "sa_key" {
  secret      = google_secret_manager_secret.sa_key.id
  secret_data = google_service_account_key.elastic_agent_key.private_key
}
