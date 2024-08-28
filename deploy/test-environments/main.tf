provider "aws" {
  region = var.region
}

provider "google" {
  project = var.gcp_project_id
}

provider "azurerm" {
  features {}
}

locals {
  common_tags = {
    division   = "${var.division}"
    org        = "${var.org}"
    team       = "${var.team}"
    project    = "${var.project}"
    owner      = "${var.owner}"
    deployment = "${var.deployment_name}"
  }

  ec_url = "https://cloud.elastic.co"
  ec_headers = {
    Content-type  = "application/json"
    Authorization = "ApiKey ${var.ec_api_key}"
  }
}

# EC2 + kind deployment
module "aws_ec2_for_kspm" {
  source          = "../cloud/modules/ec2"
  providers       = { aws : aws }
  aws_ami         = var.ami_map[var.region]
  deploy_agent    = false # Agent will not be deployed
  deployment_name = "${var.deployment_name}-${random_string.suffix.result}"
  specific_tags   = merge(local.common_tags, { "ec2_type" : "kspm" })
}

module "aws_ec2_for_cspm" {
  source          = "../cloud/modules/ec2"
  providers       = { aws : aws }
  aws_ami         = var.ami_map[var.region]
  deploy_k8s      = false
  deploy_agent    = false # Agent will not be deployed
  deployment_name = "${var.deployment_name}-${random_string.suffix.result}"
  specific_tags   = merge(local.common_tags, { "ec2_type" : "cspm" })
}

module "gcp_audit_logs" {
  count     = var.cdr_infra ? 1 : 0
  providers = { google : google }
  source    = "../cloud/modules/gcp/vm"
  gcp_service_account_json = var.gcp_service_account_json
  deployment_name = var.deployment_name
  network         = "default"
  specific_tags   = merge(local.common_tags, { "vm_instance" : "audit-logs" })

}

resource "random_string" "suffix" {
  length  = 3
  special = false
}

provider "ec" {
  apikey = var.ec_api_key
}

provider "restapi" {
  alias                = "ec"
  uri                  = local.ec_url
  write_returns_object = true
  headers              = local.ec_headers
}

# Elastic Cloud (EC) deployment
module "ec_deployment" {
  count = var.serverless_mode ? 0 : 1

  source        = "../cloud/modules/ec"
  ec_api_key    = var.ec_api_key
  region        = var.ess_region
  stack_version = var.stack_version
  tags          = local.common_tags

  deployment_template    = var.deployment_template
  deployment_name_prefix = "${var.deployment_name}-${random_string.suffix.result}"

  elasticsearch_autoscale  = true
  elasticsearch_size       = var.elasticsearch_size
  elasticsearch_zone_count = var.elasticsearch_zone_count

  docker_image_tag_override = var.pin_version != "" ? {
    "elasticsearch" = "${var.pin_version}",
    "kibana"        = "${var.pin_version}",
    "apm"           = "${var.pin_version}"
    } : {
    "elasticsearch" = "",
    "kibana"        = "",
    "apm"           = ""
  }
}

module "ec_project" {
  providers = {
    restapi.elastic_cloud = restapi.ec
  }
  count        = var.serverless_mode ? 1 : 0
  source       = "../cloud/modules/serverless"
  ec_apikey    = var.ec_api_key
  ec_url       = local.ec_url
  project_name = "${var.deployment_name}-${random_string.suffix.result}"
  region_id    = "aws-us-east-1" # TODO: replace with var.ess_region when more regions are supported
}

module "eks" {
  source                      = "../cloud/modules/provision-eks-cluster"
  region                      = var.region
  cluster_name                = var.deployment_name
  node_group_one_desired_size = 2
  # node_group_two_desired_size = 1
  enable_node_group_two = false
  tags                  = merge(local.common_tags, { "ec2_type" : "kspm_eks" })
}

# ===== CDR Infrastructure Resources =====
module "aws_ec2_for_cloudtrail" {
  count           = var.cdr_infra ? 1 : 0
  source          = "../cloud/modules/ec2"
  providers       = { aws : aws }
  aws_ami         = var.ami_map[var.region]
  deploy_k8s      = false
  deploy_agent    = false # Agent will not be deployed
  deployment_name = "${var.deployment_name}-${random_string.suffix.result}"
  specific_tags   = merge(local.common_tags, { "ec2_type" : "cloudtrail" })
}

module "azure_vm_activity_logs" {
  count           = var.cdr_infra ? 1 : 0
  source          = "../cloud/modules/azure/vm"
  providers       = { azurerm : azurerm }
  location        = var.location
  deployment_name = var.deployment_name
  specific_tags   = merge(local.common_tags, { "vm_type" : "activity-logs" })
}
