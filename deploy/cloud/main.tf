terraform {
  required_version = ">= 1.1.8, < 2.0.0"
  required_providers {
    ec = {
      source  = "elastic/ec"
      version = ">=0.5.0"
    }
  }
}

provider "ec" {}


module "ec_deployment"  {
  source = "github.com/elastic/apm-server/testing/infra/terraform/modules/ec_deployment"

  region        = var.ess_region
  stack_version = var.stack_version

  deployment_template    = var.deployment_template
  deployment_name_prefix = var.deployment_name_prefix

  integrations_server = true

  elasticsearch_size       = var.elasticsearch_size
  elasticsearch_zone_count = var.elasticsearch_zone_count

  docker_image              = var.docker_image_override
  docker_image_tag_override = {
    "elasticsearch" : "",
    "kibana" : "",
    "apm" : ""
  }
}

data "local_file" "dashboard" {
    filename = "data/dashboard.ndjson"
}

resource "null_resource" "store_local_dashboard" {
  provisioner "local-exec" {
    command = "curl -X POST -u ${module.ec_deployment.elasticsearch_username}:${module.ec_deployment.elasticsearch_password} ${module.ec_deployment.kibana_url}/api/saved_objects/_import?overwrite=true -H \"kbn-xsrf: true\" --form file=@data/dashboard.ndjson"
  }
  depends_on = [module.ec_deployment]
  triggers = {
    dashboard_sha1 = sha1(file("data/dashboard.ndjson"))
  }
}


data "local_file" "rules" {
    filename = "data/rules.ndjson"
}

resource "null_resource" "rules" {
  provisioner "local-exec" {
    command = "curl -X POST -u ${module.ec_deployment.elasticsearch_username}:${module.ec_deployment.elasticsearch_password} ${module.ec_deployment.kibana_url}/api/saved_objects/_import?overwrite=true -H \"kbn-xsrf: true\" --form file=@data/rules.ndjson"
  }
  depends_on = [module.ec_deployment]
  triggers = {
    dashboard_sha1 = "${sha1(file("data/rules.ndjson"))}"
  }
}