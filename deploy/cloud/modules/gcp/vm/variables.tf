variable "deployment_name" {
  description = "The base name used for resources in this deployment."
  type        = string
}

variable "specific_tags" {
  description = "List of tags in the format 'key=value'"
  type        = map(string)
  default     = {}
}

variable "machine_type" {
  description = "The machine type to use for the VM."
  type        = string
  default     = "n2-standard-4"
}

variable "disk_image" {
  description = "The image to use for the VM's boot disk."
  type        = string
  default     = "ubuntu-os-cloud/ubuntu-2204-lts"
}

variable "zone" {
  description = "The GCP zone where the VM will be deployed."
  type        = string
  default     = "us-central1-a"
}

variable "network" {
  description = "The network to attach the VM to."
  type        = string
}

variable "scopes" {
  description = "The scopes to attach to the service account."
  type        = list(string)
  default     = ["https://www.googleapis.com/auth/cloud-platform", "https://www.googleapis.com/auth/cloudplatformorganizations"]
}
