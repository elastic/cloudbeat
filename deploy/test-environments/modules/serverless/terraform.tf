terraform {
  required_providers {
    restapi = {
      source                = "mastercard/restapi"
      version               = "~> 3.0.0"
      configuration_aliases = [restapi.elastic_cloud]
    }
    http = {
      source  = "hashicorp/http"
      version = "~> 3.5.0"
    }
  }

  required_version = ">= 1.3, <2.0.0"
}
