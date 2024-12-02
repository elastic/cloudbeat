terraform {
  required_providers {

    google = {
      source  = "hashicorp/google"
      version = ">= 5.0.0"
    }
  }

  required_version = ">= 1.3, <2.0.0"
}
