variable "project_id" {
  description = "GCP Project ID"
  type        = string
}

variable "pool_name" {
  description = "Workload Identity Pool name"
  type        = string
}

variable "provider_name" {
  description = "Workload Identity Provider name"
  type        = string
}

variable "oidc_issuer_uri" {
  description = "OIDC issuer URI (e.g., https://sts.elastic.cloud)"
  type        = string
}

