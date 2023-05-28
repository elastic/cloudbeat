provider "aws" {
  region = var.region
}

# EC2 + kind deployment
module "aws_ec2_kind" {
  source       = "../cloud/modules/ec2"
  providers    = { aws : aws }
  aws_ami      = var.ami_map[var.region]
  deploy_agent = false # Agent will not be deployed
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
  deployment_name_prefix = "${var.deployment_name_prefix}-${random_string.suffix.result}"

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
  cluster_name_prefix         = "${var.deployment_name_prefix}-${random_string.suffix.result}"
  node_group_one_desired_size = 1
  # node_group_two_desired_size = 1
  enable_node_group_two = false
}
