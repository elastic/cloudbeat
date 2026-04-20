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

  security_solution_experimental_yaml = <<-EOT
xpack.securitySolution.enableExperimental:
  - entityAnalyticsEntityStoreV2
  - entityAnalyticsWatchlistEnabled
  - entityAnalyticsNewHomePageEnabled
  - leadGenerationEnabled
EOT

  # Entity Analytics (EA): AI agents, Agent Builder, Entity Store v2 UI, expanded experimental flags.
  entity_analytics_yaml = <<-EOT
feature_flags.overrides:
  aiAssistant.aiAgents.enabled: true

uiSettings.overrides:
  "agentBuilder:experimentalFeatures": true
  "securitySolution:entityStoreEnableV2": true

xpack.securitySolution.enableExperimental:
  - entityAnalyticsNewHomePageEnabled
  - entityAnalyticsWatchlistEnabled
  - securitySolution:entityStoreEnableV2
  - entityAnalyticsEntityStoreV2
  - leadGenerationEnabled
EOT

  kibana_docker_config = local.kibana_docker_image_tag_override != "" ? {
    docker_image = "${local.kibana_docker_image}:${local.kibana_docker_image_tag_override}"
  } : {}

  # EA settings supersede the smaller experimental-only block when enabled.
  kibana_user_settings_yaml = var.kibana_enable_entity_analytics_settings ? local.entity_analytics_yaml : (
    var.kibana_enable_security_solution_experimental ? local.security_solution_experimental_yaml : ""
  )

  kibana_experimental_config = local.kibana_user_settings_yaml != "" ? {
    user_settings_yaml = local.kibana_user_settings_yaml
  } : {}

  kibana_config_merged = merge(local.kibana_docker_config, local.kibana_experimental_config)

  elasticsearch_base = {
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
        max_size          = var.max_size
        max_size_resource = "memory"
      }
      size       = var.elasticsearch_size
      zone_count = var.elasticsearch_zone_count
    }

    warm = {
      autoscaling = {}
    }
  }

  elasticsearch = merge(
    local.elasticsearch_base,
    var.elasticsearch_ml_enabled ? {
      ml = {
        size        = "1g"
        zone_count  = 1
        autoscaling = {}
      }
    } : {}
  )
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

  elasticsearch = local.elasticsearch

  kibana = {
    size   = var.kibana_instance_size
    config = length(local.kibana_config_merged) > 0 ? local.kibana_config_merged : null
  }

  integrations_server = {
    config = local.apm_docker_image_tag_override != "" ? {
      docker_image = "${local.apm_docker_image}:${local.apm_docker_image_tag_override}"
    } : null
  }

}
