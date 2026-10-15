provider "aws" {
  region = var.region
}

locals {
  common_tags = {
    division   = "engineering"
    org        = "security"
    team       = "contextual-security"
    project    = var.project
    owner      = var.owner
    deployment = var.deployment_name
  }
}

# Bare Ubuntu x86_64 VM for the Insights Testing (project-zero) endpoint replay:
# Fleet + Elastic Defend are enrolled onto it post-apply (see
# tests/integrations_setup/install_elastic_defend_integration.py), it is not
# part of the generic CDR stack, so it gets its own minimal terraform root
# instead of adding more unrelated resources (Azure/GCP/Wiz/Okta/Cloudtrail)
# to every EA environment creation.
module "endpoint_replay_vm" {
  source          = "../modules/aws/ec2"
  aws_ami         = var.ami_map[var.region]
  deploy_k8s      = false
  deploy_agent    = false
  deployment_name = "${var.deployment_name}-endpoint-replay"
  specific_tags   = merge(local.common_tags, { "ec2_type" : "endpoint_replay" })
}
