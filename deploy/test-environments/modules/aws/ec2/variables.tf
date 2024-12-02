variable "aws_ami" {
  description = "AWS AMI to be used as a base for the instance"
  type        = string
  default     = "ami-0a5b3305c37e58e04"
}

variable "aws_ec2_instance_type" {
  description = "AWS instance type to be used"
  type        = string
  default     = "c5.4xlarge"
}

variable "yml" {
  description = "Kubernetes vanilla yaml to be applied"
  type        = string
  default     = ""
}

variable "deployment_name" {
  description = "EC2 instance name"
  type        = string
}

variable "deploy_agent" {
  description = "Deploy agent flag"
  type        = bool
  default     = true # Supporting original behaviour when agent is deployed by default
}

variable "deploy_k8s" {
  description = "Deploy kubernetes flag"
  type        = bool
  default     = true
}

variable "specific_tags" {
  description = "Specific tags for this deployment"
  type        = map(string)
  default     = {}
}

variable "cspm_aws_docker_cmd" {
  default     = ""
  description = "Optional. Docker command to enroll CSPM AWS agent"
  type        = string
}
