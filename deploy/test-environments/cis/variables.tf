# AWS provider variable
variable "region" {
  description = "AWS region"
  type        = string
  default     = "eu-west-1"
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

variable "deployment_name" {
  default     = "test-env-ci-tf"
  description = "Optional set a prefix of the deployment. Defaults to test-env-ci-tf"
}

variable "deploy_aws_kspm" {
  description = "Deploy AWS KSPM EC2 resources"
  type        = bool
  default     = true
}

variable "deploy_aws_cspm" {
  description = "Deploy AWS CSPM EC2 resources"
  type        = bool
  default     = true
}

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
# ============================================
