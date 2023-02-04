terraform {
  required_providers {
    helm = {
      source = "hashicorp/helm"
      version = ">=2.8.0"
    }

    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = "~> 2.17.0"
    }
  }

  required_version = ">= 1.3, <2.0.0"
}
