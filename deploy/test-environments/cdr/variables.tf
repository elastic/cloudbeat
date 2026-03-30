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

variable "cdr_elastic_defend_only" {
  description = "When true, deploy Wiz EC2 only among log-integration VMs (no CloudTrail, Azure, GCP, Asset Inventory) and skip Elastic Defend Linux/Windows. Wiz/Okta Fleet steps run in CI. When false, full CDR including those VMs and Elastic Defend per deploy_aws_elastic_defend_*."
  type        = bool
  default     = false
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

variable "deploy_aws_elastic_defend_linux" {
  description = "Deploy Ubuntu EC2 for Elastic Defend. Selective for local/terraform applies (-var / .tfvars). The CDR GitHub composite always forces true."
  type        = bool
  default     = true
}

variable "deploy_aws_elastic_defend_windows" {
  description = "Deploy Windows EC2 for Elastic Defend. Selective for local/terraform applies (-var / .tfvars). The CDR GitHub composite always forces true."
  type        = bool
  default     = true
}

variable "windows_elastic_defend_ami_id" {
  description = "Optional Windows AMI override for Elastic Defend host. When empty, the latest Amazon Windows Server 2022 Base image for the region is used."
  type        = string
  default     = ""
}

variable "windows_elastic_defend_instance_type" {
  description = "Instance type for Elastic Defend Windows VM"
  type        = string
  default     = "t3.large"
}

variable "windows_elastic_defend_winrm_ingress_cidr" {
  description = "Source CIDR for WinRM HTTP (5985) on the Elastic Defend Windows host"
  type        = string
  default     = "0.0.0.0/0"
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
  default     = "contextual-security"
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
