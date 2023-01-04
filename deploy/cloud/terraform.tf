terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 4.15.0"
    }

    random = {
      source  = "hashicorp/random"
      version = "~> 3.1.0"
    }

    tls = {
      source  = "hashicorp/tls"
      version = "~> 3.4.0"
    }

    cloudinit = {
      source  = "hashicorp/cloudinit"
      version = "~> 2.2.0"
    }

    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = "~> 2.12.1"
    }

    ec = {
      source  = "elastic/ec"
      version = ">=0.5.0"
    }

    restapi = {
      source  = "mastercard/restapi"
    }

    helm = {
      source = "hashicorp/helm"
      version = ">=2.8.0"
    }
  }

  required_version = ">= 1.3, <2.0.0"
}
