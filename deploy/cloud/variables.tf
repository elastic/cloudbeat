## Deployment configuration

variable "ec_api_key" {
  description = "Elastic cloud API key"
  type        = string
}

variable "ess_region" {
  default     = "gcp-us-central1"
  description = "Optional ESS region where the deployment will be created. Defaults to gcp-us-west2"
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

variable "security_team_repository" {
  default = "github.com/elastic/security-team"
}

variable "deployment_name" {
  description = "Name of the deployment. Example: `john-8-8-0bc1-7May`"
}

variable "eks_region" {
  default     = "eu-west-1"
  description = "Optional AWS region where the EKS cluster will be created. Defaults to eu-west-1"
  type        = string
}

variable "node_group_one_desired_size" {
  default = 3
  type = number
  description = "Node group one default size"
}

variable "node_group_two_desired_size" {
  default = 3
  type = number
  description = "Node group two default size"
}

variable "agent_docker_image_override" {
  default     = "" # When empty, uses the default provided by kibana
  description = "(Optional) Override agent image when deploying KSPM/CSPM. Defaults to stack version."
  type        = string
}
