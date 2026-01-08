variable "enabled" {
  description = "Enable startup validation"
  type        = bool
}

variable "project_id" {
  description = "GCP project ID"
  type        = string
}

variable "instance_name" {
  description = "Name of the instance to validate"
  type        = string
}

variable "instance_id" {
  description = "ID of the instance (for triggering re-validation)"
  type        = string
}

variable "zone" {
  description = "Zone of the instance"
  type        = string
}

variable "timeout" {
  description = "Validation timeout in seconds"
  type        = number
}
