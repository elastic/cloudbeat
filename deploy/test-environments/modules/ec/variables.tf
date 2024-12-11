variable "ec_api_key" {
  type = string
}

variable "stack_version" {
  description = "Optional version of the Elastic Cloud deployment"
  type        = string
  default     = "latest"
}

variable "region" {
  description = "Optional region of the Elastic Cloud deployment"
  type        = string
  default     = "gcp-us-west2"
}

variable "deployment_template" {
  description = "Optional defaults to the CPU optimized template for GCP"
  type        = string
  default     = "gcp-storage-optimized"
}

variable "deployment_name_prefix" {
  description = "Prefix for the Elastic Cloud deployment name"
  type        = string
  default     = "cloud-security"
}

variable "tags" {
  type = map(string)
  default = {
    "deployment"  = "cloud-security",
    "environment" = "test-enviroment",
  }
  description = "Optional set of tags to use for all deployments"
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

variable "elasticsearch_autoscale" {
  default     = false
  type        = bool
  description = "Optional autoscale the Elasticsearch cluster"
}

# Docker image overrides

# Docker image tag override is used to override the default docker image tag
# for BC reasons. This is used to test new versions of the cloud deployment
# This option allow to pin the docker image tag to a specific version to prevent
# unexpected changes in the deployment.
variable "docker_image_tag_override" {
  default = {
    "elasticsearch" = "",
    "kibana"        = "",
    "apm"           = "",
  }
  description = "Optional docker image tag overrides, The full map needs to be specified"
  type        = map(string)
}

variable "docker_image" {
  default = {
    "elasticsearch" = "docker.elastic.co/cloud-release/elasticsearch-cloud-ess",
    "kibana"        = "docker.elastic.co/cloud-release/kibana-cloud",
    "apm"           = "docker.elastic.co/cloud-release/elastic-agent-cloud",
  }
  type        = map(string)
  description = "Optional docker image overrides. The full map needs to be specified"
}
