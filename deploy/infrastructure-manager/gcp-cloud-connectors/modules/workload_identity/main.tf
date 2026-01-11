# Workload Identity Pool
resource "google_iam_workload_identity_pool" "elastic" {
  project                   = var.project_id
  workload_identity_pool_id = var.pool_name
  display_name              = "Elastic Workload Identity Pool"
  description               = "Workload Identity Pool for Elastic Cloud Connectors"
  disabled                  = false
}

# OIDC Workload Identity Provider
resource "google_iam_workload_identity_pool_provider" "oidc" {
  project                            = var.project_id
  workload_identity_pool_id          = google_iam_workload_identity_pool.elastic.workload_identity_pool_id
  workload_identity_pool_provider_id = var.provider_name
  display_name                       = "Elastic OIDC Provider"
  description                        = "OIDC provider for Elastic Cloud Connectors"
  disabled                           = false

  attribute_mapping = {
    "google.subject" = "assertion.sub"
    "attribute.aud"  = "assertion.aud"
  }

  oidc {
    issuer_uri        = var.oidc_issuer_uri
    allowed_audiences = ["ElasticCloudConnector"]
  }
}

