provider "aws" {
  region = var.region
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
}

resource "random_string" "suffix" {
  length  = 3
  special = false
}

# EC2 + kind deployment
module "aws_ec2_for_kspm" {
  count           = var.deploy_aws_kspm ? 1 : 0
  source          = "../../cloud/modules/ec2"
  providers       = { aws : aws }
  aws_ami         = var.ami_map[var.region]
  deploy_agent    = false # Agent will not be deployed
  deployment_name = "${var.deployment_name}-${random_string.suffix.result}"
  specific_tags   = merge(local.common_tags, { "ec2_type" : "kspm" })
}

module "aws_ec2_for_cspm" {
  count           = var.deploy_aws_cspm ? 1 : 0
  source          = "../../cloud/modules/ec2"
  providers       = { aws : aws }
  aws_ami         = var.ami_map[var.region]
  deploy_k8s      = false
  deploy_agent    = false # Agent will not be deployed
  deployment_name = "${var.deployment_name}-${random_string.suffix.result}"
  specific_tags   = merge(local.common_tags, { "ec2_type" : "cspm" })
}

module "eks" {
  source                      = "../../cloud/modules/provision-eks-cluster"
  region                      = var.region
  cluster_name                = var.deployment_name
  node_group_one_desired_size = 2
  # node_group_two_desired_size = 1
  enable_node_group_two = false
  tags                  = merge(local.common_tags, { "ec2_type" : "kspm_eks" })
}
