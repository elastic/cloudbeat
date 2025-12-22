variable "instance_name" {
  description = "Name of the compute instance"
  type        = string
}

variable "network_name" {
  description = "Name of the VPC network"
  type        = string
}

variable "machine_type" {
  description = "Machine type for the compute instance"
  type        = string
}

variable "zone" {
  description = "GCP Zone for the compute instance"
  type        = string
}

variable "sa_email" {
  description = "Email of the service account to be used by the instance"
  type        = string
}

variable "elastic_agent_version" {
  description = "Elastic Agent Version"
  type        = string
}

variable "elastic_artifact_server" {
  description = "Elastic Artifact Server URL"
  type        = string
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
