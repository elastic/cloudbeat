variable "project_id" {
  description = "Customer's GCP Project ID"
  type        = string
}

variable "elastic_role_arn" {
  description = "ARN of Elastic's AWS IAM role to trust (e.g., arn:aws:iam::254766567737:role/cloud_connectors)"
  type        = string
  default     = "arn:aws:iam::254766567737:role/cloud_connectors"

  validation {
    condition     = can(regex("^arn:aws:iam::[0-9]+:role/.+$", var.elastic_role_arn))
    error_message = "elastic_role_arn must be a valid AWS IAM role ARN (e.g., arn:aws:iam::123456789012:role/my-role)"
  }
}

variable "wif_pool_name" {
  description = "Name of the Workload Identity Pool"
  type        = string
  default     = "elastic-pool"
}

variable "wif_provider_name" {
  description = "Name of the Workload Identity Provider"
  type        = string
  default     = "elastic-aws-provider"
}

variable "target_service_account_name" {
  description = "Name of the target service account"
  type        = string
  default     = "elastic-agent-sa"
}

variable "scope" {
  description = "Scope for IAM bindings (projects or organizations)"
  type        = string
  default     = "projects"

  validation {
    condition     = contains(["projects", "organizations"], var.scope)
    error_message = "scope must be either 'projects' or 'organizations'"
  }
}

variable "parent_id" {
  description = "Parent ID (project ID for projects scope, organization ID for organizations scope)"
  type        = string
}

variable "elastic_resource_id" {
  description = "Unique identifier for the Elastic deployment (must match the AWS role session name)"
  type        = string
}
