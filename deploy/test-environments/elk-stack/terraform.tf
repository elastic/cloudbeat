terraform {
  required_providers {
    ec = {
      source  = "elastic/ec"
      version = ">=0.9.0"
    }

    restapi = {
      source  = "mastercard/restapi"
      version = "~> 1.20.0"
    }

    random = {
      source  = "hashicorp/random"
      version = "~> 3.8.0"
    }

  }

  required_version = ">= 1.3, <2.0.0"
}
