# Service Account
resource "google_service_account" "elastic_agent" {
  account_id   = var.service_account_name
  display_name = "Elastic agent service account for CSPM"
  project      = var.project_id
}

# Service Account Key
resource "google_service_account_key" "elastic_agent_key" {
  count              = var.create_key ? 1 : 0
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
