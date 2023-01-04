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
  description = "Kubernetes yaml to be applied"
  type        = string

}
