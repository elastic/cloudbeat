locals {
  version                          = var.stack_version
  region                           = var.region
  name_prefix                      = var.deployment_name_prefix
  deployment_template              = var.deployment_template
  es_docker_image                  = lookup(var.docker_image, "elasticsearch", "")
  es_docker_image_tag_override     = lookup(var.docker_image_tag_override, "elasticsearch", "")
  kibana_docker_image              = lookup(var.docker_image, "kibana", "")
  kibana_docker_image_tag_override = lookup(var.docker_image_tag_override, "kibana", "")
  apm_docker_image                 = lookup(var.docker_image, "apm", "")
  apm_docker_image_tag_override    = lookup(var.docker_image_tag_override, "apm", "")
}

data "ec_stack" "deployment_version" {
  version_regex = local.version
  region        = local.region
}

resource "ec_deployment" "deployment" {
  name                   = "${local.name_prefix}-${data.ec_stack.deployment_version.version}"
  version                = data.ec_stack.deployment_version.version
  region                 = local.region
  deployment_template_id = local.deployment_template
  tags                   = var.tags

  elasticsearch = {
    autoscale = var.elasticsearch_autoscale
    strategy  = "rolling_all"
    config = {
      docker_image = local.es_docker_image_tag_override != "" ? "${local.es_docker_image}:${local.es_docker_image_tag_override}" : null
    }

    hot = {
      autoscaling = {
        max_size = "128g"
      }
      size       = var.elasticsearch_size
      zone_count = var.elasticsearch_zone_count
    }
  }

  kibana = {
    config = {
      docker_image = local.kibana_docker_image_tag_override != "" ? "${local.kibana_docker_image}:${local.kibana_docker_image_tag_override}" : null
    }
  }

  integrations_server = {
    config = {
      docker_image = local.apm_docker_image_tag_override != "" ? "${local.apm_docker_image}:${local.apm_docker_image_tag_override}" : null
    }
  }

}