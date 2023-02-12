terraform {
  required_providers {
    random = {
      source  = "hashicorp/random"
      version = "~> 3.1.0"
    }

    ec = {
      source  = "elastic/ec"
      version = ">=0.5.0"
    }
  }
}
