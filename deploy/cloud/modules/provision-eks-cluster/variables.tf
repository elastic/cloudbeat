variable "region" {
  description = "AWS region"
  type        = string
  default     = "us-east-2"
}

variable "cluster_name_prefix" {
  description = "EKS cluster name prefix"
  type        = string
  default     = "cloudbeat-tf"
}
