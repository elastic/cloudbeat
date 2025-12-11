locals {
  common_tags = {
    division   = "${var.division}"
    org        = "${var.org}"
    team       = "${var.team}"
    project    = "${var.project}"
    owner      = "${var.owner}"
    deployment = "${var.deployment_name}"
  }

  ec_url = var.ec_url != "" ? var.ec_url : "https://cloud.elastic.co"

  # Set default ess_region based on deployment type if not provided
  ess_region = var.ess_region != "" ? var.ess_region : (
    var.serverless_mode ? "aws-us-east-1" : "gcp-us-west2"
  )

  ec_headers = {
    Content-type  = "application/json"
    Authorization = "ApiKey ${var.ec_api_key}"
  }
}

provider "ec" {
  apikey   = var.ec_api_key
  endpoint = local.ec_url
}

provider "restapi" {
  alias                = "ec"
  uri                  = local.ec_url
  write_returns_object = true
  headers              = local.ec_headers
}

resource "random_string" "suffix" {
  length  = 3
  special = false
}

# Elastic Cloud (EC) deployment
module "ec_deployment" {
  count = var.serverless_mode ? 0 : 1

  source        = "../modules/ec"
  ec_api_key    = var.ec_api_key
  region        = local.ess_region
  stack_version = var.stack_version
  tags          = local.common_tags

  deployment_template    = var.deployment_template
  deployment_name_prefix = "${var.deployment_name}-${random_string.suffix.result}"
  max_size               = var.max_size != "" ? var.max_size : "128g"

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
  source       = "../modules/serverless"
  ec_apikey    = var.ec_api_key
  ec_url       = local.ec_url
  project_name = "${var.deployment_name}-${random_string.suffix.result}"
  region_id    = local.ess_region
}
