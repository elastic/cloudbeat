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

# Elastic Cloud variables
# ===========================================
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

variable "deployment_template" {
  default     = "gcp-compute-optimized-v2"
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
  default     = 1
  type        = number
  description = "Optional Elasticsearch zone count"
}

variable "docker_image_tag_override" {
  default = {
    "elasticsearch" : "",
    "kibana" : "",
    "apm" : "",
  }
  description = "Optional docker image tag override"
  type        = map(string)
}

variable "docker_image_override" {
  default = {
    "elasticsearch" : "docker.elastic.co/cloud-release/elasticsearch-cloud-ess",
    "kibana" : "docker.elastic.co/cloud-release/kibana-cloud",
    "apm" : "docker.elastic.co/cloud-release/elastic-agent-cloud",
  }
  type = map(string)
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
  default     = "cloud-security-posture"
  type        = string
  description = "Optional team resource tag"
}

variable "project" {
  default     = "test-environments"
  type        = string
  description = "Optional project resource tag"
}

# ============================================
