variable "s3_bucket_name" {
  description = "Name of the S3 bucket for CloudTrail logs"
  type        = string
  default     = "tf-test-envs-cloudtrail-logs"
}

variable "cloudtrail_name" {
  description = "Name of the CloudTrail"
  type        = string
  default     = "test-envs"
}

variable "kms_alias_name" {
  description = "Alias name for the KMS key"
  type        = string
  default     = "test-envs-kms-key"
}

variable "aws_account_id" {
  description = "AWS account ID"
  type        = string
  // TODO: remove default value
  default = "704479110758"
}

variable "tags" {
  description = "Optional set of tags to use for all deployments"
  type        = map(string)
  default = {
    "deployment"  = "cloud-security",
    "environment" = "test-enviroments",
  }
}

#==============================================================================
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
# ============================================
