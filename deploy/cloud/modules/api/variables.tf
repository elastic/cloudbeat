## Deployment configuration

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
