variable "project_id" {
  description = "Customer's GCP Project ID"
  type        = string
}

variable "target_service_account_name" {
  description = "Name of the target service account"
  type        = string
}

variable "wif_pool_name" {
  description = "Full name of Elastic's Workload Identity Pool (e.g., projects/PROJECT_NUMBER/locations/global/workloadIdentityPools/POOL_ID)"
  type        = string
}

variable "scope" {
  description = "Scope for IAM bindings (projects or organizations)"
  type        = string
}

variable "parent_id" {
  description = "Parent ID (project ID or organization ID)"
  type        = string
}

variable "aws_role_name" {
  description = "Name of the AWS role that can impersonate this service account"
  type        = string
}
