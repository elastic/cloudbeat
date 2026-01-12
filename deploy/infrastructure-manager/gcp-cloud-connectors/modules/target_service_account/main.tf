# Target Service Account (Customer-owned)
resource "google_service_account" "target" {
  account_id   = var.target_service_account_name
  display_name = "Elastic Agent Service Account"
  description  = "Service account for Elastic Agent security monitoring"
  project      = var.project_id
}

# Project-level IAM bindings (for single-account)
resource "google_project_iam_member" "cloudasset_viewer" {
  count   = var.scope == "projects" ? 1 : 0
  project = var.parent_id
  role    = "roles/cloudasset.viewer"
  member  = "serviceAccount:${google_service_account.target.email}"
}

resource "google_project_iam_member" "browser" {
  count   = var.scope == "projects" ? 1 : 0
  project = var.parent_id
  role    = "roles/browser"
  member  = "serviceAccount:${google_service_account.target.email}"
}

# Organization-level IAM bindings (for organization-account)
resource "google_organization_iam_member" "cloudasset_viewer_org" {
  count  = var.scope == "organizations" ? 1 : 0
  org_id = var.parent_id
  role   = "roles/cloudasset.viewer"
  member = "serviceAccount:${google_service_account.target.email}"
}

resource "google_organization_iam_member" "browser_org" {
  count  = var.scope == "organizations" ? 1 : 0
  org_id = var.parent_id
  role   = "roles/browser"
  member = "serviceAccount:${google_service_account.target.email}"
}

# Allow Workload Identity Federation to impersonate Target Service Account
# Only the specific AWS role can impersonate this SA
resource "google_service_account_iam_member" "workload_identity_user" {
  service_account_id = google_service_account.target.name
  role               = "roles/iam.workloadIdentityUser"
  # Trust the AWS role from Elastic's account
  member             = "principalSet://iam.googleapis.com/${var.wif_pool_name}/attribute.aws_role/${var.aws_role_name}"
}
