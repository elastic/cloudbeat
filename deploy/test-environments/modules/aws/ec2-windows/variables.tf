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
  default     = "t3.medium"
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
  default     = ""

  validation {
    condition     = var.winrm_ingress_cidr != "" && var.winrm_ingress_cidr != "0.0.0.0/0"
    error_message = "winrm_ingress_cidr must be set to a restrictive CIDR and must not be 0.0.0.0/0, as this would expose WinRM to the public internet."
  }
}
