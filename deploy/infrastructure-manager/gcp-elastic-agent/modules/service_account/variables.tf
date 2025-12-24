variable "project_id" {
  description = "GCP Project ID"
  type        = string
}

variable "service_account_name" {
  description = "Service account name"
  type        = string
}

variable "scope" {
  description = "Scope for IAM bindings (projects or organizations)"
  type        = string
  default     = "projects"
}

variable "parent_id" {
  description = "Parent ID (project ID or organization ID)"
  type        = string
}

variable "create_key" {
  description = "Whether to create a service account key"
  type        = bool
  default     = false
}
