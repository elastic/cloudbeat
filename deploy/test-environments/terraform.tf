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
    ec = {
      source  = "elastic/ec"
      version = ">=0.9.0"
    }

    restapi = {
      source  = "mastercard/restapi"
      version = "~> 1.18.0"
    }

    random = {
      source  = "hashicorp/random"
      version = "~> 3.5.1"
    }

  }

  required_version = ">= 1.3, <2.0.0"
}
