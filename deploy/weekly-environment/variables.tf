## Deployment configuration

variable "ec_api_key" {
  description = "Elastic cloud API key"
  type        = string
}

variable "ess_region" {
  default     = "gcp-us-central1"
  description = "Optional ESS region where the deployment will be created. Defaults to gcp-us-central1"
  type        = string
}

variable "deployment_template" {
  default     = "gcp-compute-optimized-v2"
  description = "Optional deployment template. Defaults to the CPU optimized template for GCP"
  type        = string
}

variable "stack_version" {
  default     = "latest"
  description = "Optional stack version"
  type        = string
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

variable "deployment_name_prefix" {
  default     = "weekly-environment"
  description = "Optional set a prefix of the deployment. Defaults to weekly-environment"
}

variable "endpoint" {
    default     = "https://staging.found.no/"
    description = "Optional endpoint for the Elastic Cloud API"
}
