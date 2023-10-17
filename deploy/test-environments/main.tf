provider "aws" {
  region = var.region
}

locals {
  common_tags = {
    division = "${var.division}"
    org      = "${var.org}"
    team     = "${var.team}"
    project  = "${var.project}"
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

resource "random_string" "suffix" {
  length  = 3
  special = false
}

provider "ec" {
  apikey = var.ec_api_key
}

# Elastic Cloud (EC) deployment
module "ec_deployment" {
  source = "github.com/elastic/apm-server/testing/infra/terraform/modules/ec_deployment"

  region        = var.ess_region
  stack_version = var.stack_version

  deployment_template    = var.deployment_template
  deployment_name_prefix = "${var.deployment_name}-${random_string.suffix.result}"

  integrations_server = true

  elasticsearch_size       = var.elasticsearch_size
  elasticsearch_zone_count = var.elasticsearch_zone_count

  docker_image = var.docker_image_override
  docker_image_tag_override = {
    "elasticsearch" : "",
    "kibana" : "",
    "apm" : ""
  }
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
