terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 6.41.0"
    }

    random = {
      source  = "hashicorp/random"
      version = "~> 3.8.0"
    }

  }

  required_version = ">= 1.3, <2.0.0"
}
