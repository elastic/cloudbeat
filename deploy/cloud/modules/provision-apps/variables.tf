variable "kube_config" {
  type = string
  default = "~/.kube/config"
}

variable "namespace" {
  type = string
  default = "kube-system"
}

variable "replica_count" {
  type = string
  default = "2"
}