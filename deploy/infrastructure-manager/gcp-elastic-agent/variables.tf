variable "project_id" {
  description = "GCP Project ID"
  type        = string
}

variable "zone" {
  description = "GCP Zone for the compute instance"
  type        = string
  default     = "us-central1-a"
}

variable "fleet_url" {
  description = "Elastic Agent Fleet URL"
  type        = string
}

variable "enrollment_token" {
  description = "Elastic Agent Enrollment Token"
  type        = string
  sensitive   = true
}

variable "elastic_agent_version" {
  description = "Elastic Agent Version (e.g., 8.8.0 or 8.8.0-SNAPSHOT)"
  type        = string
}

variable "elastic_artifact_server" {
  description = "Elastic Artifact Server URL"
  type        = string
  default     = "https://artifacts.elastic.co/downloads/beats/elastic-agent"
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

variable "machine_type" {
  description = "Machine type for the compute instance"
  type        = string
  default     = "n2-standard-4"
}
