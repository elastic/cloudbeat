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

  # Determine max_size based on cloud provider
  # GCP supports 128g (131072 MB), AWS and Azure support up to 58g (59392 MB) for hot_content
  max_size = startswith(var.region, "gcp-") ? "128g" : "58g"
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
    config = local.es_docker_image_tag_override != "" ? {
      docker_image = "${local.es_docker_image}:${local.es_docker_image_tag_override}"
    } : null

    cold = {
      autoscaling = {}
    }

    frozen = {
      autoscaling = {}
    }

    hot = {
      autoscaling = {
        max_size          = local.max_size
        max_size_resource = "memory"
      }
      size       = var.elasticsearch_size
      zone_count = var.elasticsearch_zone_count
    }

    warm = {
      autoscaling = {}
    }
  }

  kibana = {
    config = local.kibana_docker_image_tag_override != "" ? {
      docker_image = "${local.kibana_docker_image}:${local.kibana_docker_image_tag_override}"
    } : null
  }

  integrations_server = {
    config = local.apm_docker_image_tag_override != "" ? {
      docker_image = "${local.apm_docker_image}:${local.apm_docker_image_tag_override}"
    } : null
  }

}
