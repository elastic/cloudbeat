terraform {
  required_providers {
    restapi = {
      source  = "mastercard/restapi"
      version = "~> 1.18.0"
    }
    http = {
      source  = "hashicorp/http"
      version = "~> 3.2.1"
    }
  }

  required_version = ">= 1.3, <2.0.0"
}
