# ========Elastic Cloud Variables Section===============
variable "ec_api_key" {
  description = "Provide Elastic Cloud API key or use export TF_VAR_ec_api_key={TOKEN}"
  type        = string
}

variable "ess_region" {
  default     = "gcp-us-west2"
  description = "Optional ESS region where the deployment will be created. Defaults to gcp-us-west2"
  type        = string
}

variable "stack_version" {
  default     = "latest"
  description = "Optional stack version"
  type        = string
}

variable "pin_version" {
  default     = ""
  description = "Optional pinned stack version for BC reasons"
  type        = string
}

variable "serverless_mode" {
  default     = false
  description = "Set to true to create a serverless security project instead of an ESS deployment"
  type        = bool
}

variable "deployment_template" {
  default     = "gcp-general-purpose"
  description = "Optional deployment template. Defaults to the CPU optimized template for GCP"
  type        = string
}

variable "deployment_name" {
  default     = "test-env-ci-tf"
  description = "Optional set a prefix of the deployment. Defaults to test-env-ci-tf"
}

variable "elasticsearch_size" {
  default     = "8g"
  type        = string
  description = "Optional Elasticsearch instance size"
}

variable "elasticsearch_zone_count" {
  default     = 2
  type        = number
  description = "Optional Elasticsearch zone count"
}

variable "docker_image_tag_override" {
  default = {
    "elasticsearch" = "",
    "kibana"        = "",
    "apm"           = "",
  }
  description = "Optional docker image tag override"
  type        = map(string)
}
# ========End Of Elastic Cloud Variables Section========

# ============= Tags Section============================
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
# ======End Of Tags section=====================
