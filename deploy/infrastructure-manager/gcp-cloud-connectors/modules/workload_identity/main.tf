# Workload Identity Pool
resource "google_iam_workload_identity_pool" "elastic" {
  project                   = var.project_id
  workload_identity_pool_id = var.pool_name
  display_name              = "Elastic Workload Identity Pool"
  description               = "Workload Identity Pool for Elastic Cloud Connectors"
  disabled                  = false
}

# AWS Workload Identity Provider (instead of OIDC)
resource "google_iam_workload_identity_pool_provider" "aws" {
  project                            = var.project_id
  workload_identity_pool_id          = google_iam_workload_identity_pool.elastic.workload_identity_pool_id
  workload_identity_pool_provider_id = var.provider_name
  display_name                       = "Elastic AWS Provider"
  description                        = "AWS provider for Elastic Cloud Connectors"
  disabled                           = false

  attribute_mapping = {
    "google.subject"         = "assertion.arn"
    "attribute.aws_role"     = "assertion.arn.extract('assumed-role/{role}/')"
    "attribute.aws_account"  = "assertion.account"
    "attribute.session_name" = "assertion.arn.extract('assumed-role/${var.aws_role_name}/{session}/')"
  }

  # Validate: AWS role session name must be elastic_resource_id-cloud_connector_id (use this as session name when assuming the role)
  # ARN format: arn:aws:sts::ACCOUNT_ID:assumed-role/ROLE_NAME/SESSION_NAME
  attribute_condition = "assertion.arn == 'arn:aws:sts::${var.aws_account_id}:assumed-role/${var.aws_role_name}/${var.elastic_resource_id}-${var.cloud_connector_id}'"

  aws {
    account_id = var.aws_account_id
  }
}
