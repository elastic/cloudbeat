terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 4.15.0"
    }

    random = {
      source  = "hashicorp/random"
      version = "~> 3.8.0"
    }

  }

  required_version = ">= 1.3, <2.0.0"
}
