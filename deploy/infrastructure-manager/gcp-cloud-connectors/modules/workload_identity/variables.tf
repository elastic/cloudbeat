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

variable "aws_account_id" {
  description = "AWS Account ID that hosts the trusted role"
  type        = string
}

variable "aws_role_name" {
  description = "Name of the AWS role to trust (without ARN prefix)"
  type        = string
}

variable "elastic_resource_id" {
  description = "Unique identifier for the Elastic deployment (must match the AWS role session name)"
  type        = string
}
