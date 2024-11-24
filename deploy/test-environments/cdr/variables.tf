variable "deployment_name" {
  default     = "test-env-ci-tf"
  description = "Optional set a prefix of the deployment. Defaults to test-env-ci-tf"
}

# AWS provider variable
variable "region" {
  description = "AWS region"
  type        = string
  default     = "eu-west-1"
}

# Azure provider variable
variable "location" {
  description = "Azure location"
  type        = string
  default     = "East US"
}

# EC2 variable
variable "ami_map" {
  description = "Mapping of regions to AMI IDs"
  type        = map(any)
  default = {
    "eu-west-1" = "ami-0a5b3305c37e58e04"
    "eu-west-3" = "ami-0532b3f7436b93d52"
    # Add more regions and respective AMI IDs here
  }
}

# GCP project ID
variable "gcp_project_id" {
  description = "GCP project ID"
  type        = string
  default     = "default"
}

variable "gcp_service_account_json" {
  description = "GCP Service Account JSON"
  type        = string
  default     = "default"
  sensitive   = true
}

variable "deploy_az_vm" {
  description = "Deploy Azure VM resources"
  type        = bool
  default     = true
}

variable "deploy_gcp_vm" {
  description = "Deploy GCP VM resources"
  type        = bool
  default     = true
}

variable "deploy_aws_ec2" {
  description = "Deploy AWS EC2 resources"
  type        = bool
  default     = true
}

variable "deploy_aws_ec2_wiz" {
  description = "Deploy AWS EC2 resources"
  type        = bool
  default     = true
}

variable "deploy_aws_asset_inventory" {
  description = "Deploy AWS Asset Inventory EC2 resources"
  type        = bool
  default     = true
}

# ========= Cloud Tags ========================
variable "division" {
  default     = "engineering"
  type        = string
  description = "Optional division resource tag"
}

variable "org" {
  default     = "security"
  type        = string
  description = "Optional org resource tag"
}

variable "team" {
  default     = "cloud-security-posture"
  type        = string
  description = "Optional team resource tag"
}

variable "project" {
  default     = "test-environments"
  type        = string
  description = "Optional project resource tag"
}

variable "owner" {
  default     = "cloudbeat"
  type        = string
  description = "Optional owner tag"
}
# ========= End Of Cloud Tags ===================
