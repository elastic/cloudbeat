variable "location" {
  description = "Azure location"
  type        = string
  default     = "East US"
}

variable "deployment_name" {
  description = "Deployment name"
  type        = string
  default     = "test-environments"
}

variable "specific_tags" {
  description = "Specific tags for this deployment"
  type        = map(string)
  default     = {}
}

variable "size" {
  type        = string
  default     = "Standard_F2"
  description = "The size of the VM"
}