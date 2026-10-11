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

  # Entity Analytics Kibana user_settings_yaml overrides (ESS only).
  entity_analytics_overrides = {
    "feature_flags.overrides" = {
      "securitySolution.useNewEntityAnalyticsPage" = true
    }
  }

  # Insights Testing (project-zero) Kibana user_settings_yaml overrides (ESS only).
  insights_testing_overrides = {
    "uiSettings.overrides" = {
      "agentBuilder:experimentalFeatures"               = true
      "workflows:ui:showManagedWorkflows"               = true
      "securitySolution:enableAlertZero"                = true
      "contextEngine:enabled"                           = true
      "securitySolution:enableAttackDiscoveryWorkflows" = true
    }
    "xpack.securitySolution.enableExperimental" = [
      "leadGenerationEnabled",
      "endpointForensicAnalysisSkill",
    ]
    "xpack.alertzero.enabled"                = true
    "xpack.evals.enabled"                    = true
    "xpack.agenticInvestigations.enabled"    = true
    "xpack.proposals.enabled"                = true
    "xpack.security.serviceAccounts.enabled" = true
  }

  # Terraform can't unify `cond ? <heterogeneous object> : {}` into one type,
  # so build a tuple of optional overrides (null when disabled), drop the
  # nulls, then spread what's left into merge(). Both sets use disjoint
  # top-level keys today, so a shallow merge is enough -- if a future
  # addition needs the same top-level key from both sides, this merge won't
  # combine their nested values, the later argument wins for that whole key.
  kibana_overrides_enabled = [
    for overrides in [
      var.kibana_enable_entity_analytics_settings ? local.entity_analytics_overrides : null,
      var.kibana_enable_insights_testing_settings ? local.insights_testing_overrides : null,
    ] : overrides if overrides != null
  ]

  # merge() always gets a guaranteed {} first argument so it's never called
  # with zero arguments (an empty spread) -- and, unlike a ternary, merge()
  # itself doesn't require its arguments' types to unify statically.
  kibana_overrides_merged = merge({}, local.kibana_overrides_enabled...)

  kibana_docker_config = local.kibana_docker_image_tag_override != "" ? {
    docker_image = "${local.kibana_docker_image}:${local.kibana_docker_image_tag_override}"
  } : {}

  kibana_user_settings_yaml = length(local.kibana_overrides_merged) > 0 ? yamlencode(local.kibana_overrides_merged) : ""

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
