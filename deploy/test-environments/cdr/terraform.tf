terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 4.67.0"
    }

    google = {
      source  = "hashicorp/google"
      version = ">= 5.0.0"
    }

    azurerm = {
      source  = "hashicorp/azurerm"
      version = "< 4.71"
    }

    random = {
      source  = "hashicorp/random"
      version = "~> 3.5.1"
    }

    tls = {
      source  = "hashicorp/tls"
      version = "~> 4.0"
    }

    local = {
      source  = "hashicorp/local"
      version = "~> 2.4"
    }

  }

  required_version = ">= 1.3, <2.0.0"
}
