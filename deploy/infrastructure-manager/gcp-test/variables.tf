variable "project_id" {
  description = "GCP Project ID"
  type        = string
}

variable "deployment_name" {
  description = "Name of the deployment"
  type        = string
}

variable "region" {
  description = "GCP Region"
  type        = string
  default     = "us-central1"
}
