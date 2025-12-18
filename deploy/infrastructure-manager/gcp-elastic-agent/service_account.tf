# This file is used for standalone service account deployment
# It creates a service account with IAM roles and generates a key

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
  project = var.sa_project_id
}

variable "sa_project_id" {
  description = "GCP Project ID"
  type        = string
}

variable "sa_deployment_name" {
  description = "Name of the deployment"
  type        = string
  default     = "elastic-agent-cspm-user"
}

variable "sa_service_account_name" {
  description = "Service account name"
  type        = string
  default     = "elastic-agent-cspm-user-sa"
}

variable "sa_scope" {
  description = "Scope for IAM bindings (projects or organizations)"
  type        = string
  default     = "projects"
}

variable "sa_parent_id" {
  description = "Parent ID (project ID or organization ID)"
  type        = string
}

# Service Account
resource "google_service_account" "cspm_user" {
  account_id   = var.sa_service_account_name
  display_name = "Elastic agent service account for CSPM"
  project      = var.sa_project_id
}

# Service Account Key
resource "google_service_account_key" "cspm_user_key" {
  service_account_id = google_service_account.cspm_user.name
}

# Project-level IAM bindings
resource "google_project_iam_member" "cloudasset_viewer" {
  count   = var.sa_scope == "projects" ? 1 : 0
  project = var.sa_parent_id
  role    = "roles/cloudasset.viewer"
  member  = "serviceAccount:${google_service_account.cspm_user.email}"
}

resource "google_project_iam_member" "browser" {
  count   = var.sa_scope == "projects" ? 1 : 0
  project = var.sa_parent_id
  role    = "roles/browser"
  member  = "serviceAccount:${google_service_account.cspm_user.email}"
}

# Organization-level IAM bindings
resource "google_organization_iam_member" "cloudasset_viewer_org" {
  count  = var.sa_scope == "organizations" ? 1 : 0
  org_id = var.sa_parent_id
  role   = "roles/cloudasset.viewer"
  member = "serviceAccount:${google_service_account.cspm_user.email}"
}

resource "google_organization_iam_member" "browser_org" {
  count  = var.sa_scope == "organizations" ? 1 : 0
  org_id = var.sa_parent_id
  role   = "roles/browser"
  member = "serviceAccount:${google_service_account.cspm_user.email}"
}

# Output the service account key
output "service_account_key" {
  description = "Service account private key (base64 encoded)"
  value       = google_service_account_key.cspm_user_key.private_key
  sensitive   = true
}

output "service_account_email" {
  description = "Service account email"
  value       = google_service_account.cspm_user.email
}
