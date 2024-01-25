variable "region" {
  description = "AWS region"
  type        = string
  default     = "eu-west-1"
}

variable "cluster_name" {
  description = "EKS cluster name"
  type        = string
}

variable "node_group_one_desired_size" {
  default     = 2
  type        = number
  description = "Node group one desired size, ensure that the desired size does not exceed the max size"
}

variable "node_group_two_desired_size" {
  default     = 2
  type        = number
  description = "Node group two desired size, ensure that the desired size does not exceed the max size"
}

variable "enable_node_group_two" {
  default     = true
  type        = bool
  description = "Flag to enable/disable deployment of node_group_two"
}

variable "tags" {
  type        = map(string)
  description = "Tags to be applied to the EKS cluster and its resources."
  default     = {}
}
