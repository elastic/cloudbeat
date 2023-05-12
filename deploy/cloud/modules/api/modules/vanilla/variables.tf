## Deployment configuration

variable "fleet_url" {
  description = "Fleet URL"
  type        = string
}

variable "username" {
  description = "Elastic username"
  type        = string
}

variable "password" {
  description = "Elastic password"
  type        = string
}

variable "uri" {
  description = "Kibana URL"
  type        = string
}

variable "role_arn" {
  description = "IAM Role ARN to use"
  type        = string
}

variable "agent_docker_img" {
  description = "Customize agent's docker image (Optional)"
  default     = ""
  type        = string
}
