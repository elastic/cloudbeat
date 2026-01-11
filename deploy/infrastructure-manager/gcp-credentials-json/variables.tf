variable "project_id" {
  description = "GCP Project ID"
  type        = string
}

variable "resource_suffix" {
  description = "Unique suffix for resource names (8 hex characters)"
  type        = string
}

variable "scope" {
  description = "Scope for IAM bindings (projects or organizations)"
  type        = string
  default     = "projects"

  validation {
    condition     = contains(["projects", "organizations"], var.scope)
    error_message = "Scope must be either 'projects' or 'organizations'."
  }
}

variable "parent_id" {
  description = "Parent ID (project ID or organization ID depending on scope)"
  type        = string
}
