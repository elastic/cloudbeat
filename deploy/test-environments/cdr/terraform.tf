terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 4.15.0"
    }

    google = {
      source  = "hashicorp/google"
      version = ">= 5.0.0"
    }

    azurerm = {
      source  = "hashicorp/azurerm"
      version = ">= 3.11, < 4.0"
    }

    random = {
      source  = "hashicorp/random"
      version = "~> 3.5.1"
    }

  }

  required_version = ">= 1.3, <2.0.0"
}
