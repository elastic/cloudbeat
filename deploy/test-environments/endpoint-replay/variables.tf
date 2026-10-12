variable "deployment_name" {
  description = "Base name for the Endpoint Replay VM (matches the EA environment's deployment name)"
  type        = string
  # Matches elk-stack's default so the generic `manage_infrastructure.sh all destroy`
  # path (used by the daily expired-environment cleanup for every environment,
  # including ones that never provisioned this VM) doesn't error on a missing
  # required variable -- there's nothing in state to destroy in that case anyway.
  default = "test-env-ci-tf"
}

variable "region" {
  description = "AWS region for the Endpoint Replay VM"
  type        = string
  default     = "eu-west-1"
}

variable "ami_map" {
  description = "Mapping of regions to AMI IDs (Ubuntu, x86_64 -- must match the Linux x86_64 Fleet install command)"
  type        = map(any)
  default = {
    "eu-west-1" = "ami-0a5b3305c37e58e04"
    "eu-west-3" = "ami-0532b3f7436b93d52"
  }
}

variable "project" {
  description = "Optional project resource tag"
  type        = string
  default     = "test-environments"
}

variable "owner" {
  description = "Optional owner tag"
  type        = string
  default     = "cloudbeat"
}
