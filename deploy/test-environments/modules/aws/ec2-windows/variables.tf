variable "deployment_name" {
  description = "EC2 instance name tag"
  type        = string
}

variable "specific_tags" {
  description = "Additional tags for this deployment"
  type        = map(string)
  default     = {}
}

variable "aws_ec2_instance_type" {
  description = "AWS instance type for Windows Server"
  type        = string
  default     = "t3.large"
}

variable "windows_ami_id" {
  description = "Optional override for Windows Server AMI ID. When empty, the latest Amazon Windows Server 2022 Base image for the current region is used."
  type        = string
  default     = ""
}

variable "iam_instance_profile" {
  description = "IAM instance profile name"
  type        = string
  default     = "ec2-role-with-security-audit"
}

variable "winrm_ingress_cidr" {
  description = "Source CIDR for WinRM HTTP (5985)"
  type        = string
  default     = "0.0.0.0/0"
}
