terraform {
  required_version = ">= 1.3, <2.0.0"

  required_providers {
    ec = {
      source  = "elastic/ec"
      version = ">=0.9.0"
    }
  }
}
