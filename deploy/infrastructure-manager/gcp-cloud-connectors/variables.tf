variable "project_id" {
  description = "Customer's GCP Project ID"
  type        = string
}

variable "oidc_issuer_uri" {
  description = "OIDC issuer URI"
  type        = string
  default     = "https://elastic-cloud-connector.s3.us-east-1.amazonaws.com"
}

variable "wif_pool_name" {
  description = "Name of the Workload Identity Pool"
  type        = string
  default     = "elastic-pool"
}

variable "wif_provider_name" {
  description = "Name of the Workload Identity Provider"
  type        = string
  default     = "elastic-oidc-provider"
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
  description = "Unique identifier for the Elastic deployment"
  type        = string
}
