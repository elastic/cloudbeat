provider "ec" {}

#locals {
#  cluster_name = var.cluster_name_prefix
#  post = "curl -X POST -u ${module.ec_deployment.elasticsearch_username}:${module.ec_deployment.elasticsearch_password} ${module.ec_deployment.kibana_url}/api/"
#  depends_on = [module.ec_deployment]
#}

module "ec_deployment" {
  source = "github.com/elastic/apm-server/testing/infra/terraform/modules/ec_deployment"

  region        = var.ess_region
  stack_version = var.stack_version

  deployment_template    = var.deployment_template
  deployment_name_prefix = "${var.deployment_name_prefix}-${random_string.suffix.result}"

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

module "eks" {
  source = "./modules/provision-eks-cluster"

  region              = var.eks_region
  cluster_name_prefix = "${var.deployment_name_prefix}-${random_string.suffix.result}"
}

data "local_file" "dashboard" {
  filename = "data/dashboard.ndjson"
}

resource "null_resource" "store_local_dashboard" {
  provisioner "local-exec" {
    command = "curl -X POST -u ${module.ec_deployment.elasticsearch_username}:${module.ec_deployment.elasticsearch_password} ${module.ec_deployment.kibana_url}/api/saved_objects/_import?overwrite=true -H \"kbn-xsrf: true\" --form file=@data/dashboard.ndjson"
  }
  depends_on = [module.ec_deployment]
  triggers   = {
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
  triggers   = {
    dashboard_sha1 = sha1(file("data/rules.ndjson"))
  }
}

#resource "null_resource" "create_agent_policy" {
#  provisioner "local-exec" {
#    command = "curl -u ${module.ec_deployment.elasticsearch_username}:${module.ec_deployment.elasticsearch_password} ${module.ec_deployment.kibana_url}/api/fleet/agent_policies?sys_monitoring=true -H \"kbn-xsrf: true\" -H \"content-type: application/json\" -d @data/create_agent_policy.json | jq .item.id -r > agent_policy_id.txt"
#  }
#  depends_on = [module.ec_deployment]
#  triggers   = {
##    always_run = timestamp()
#  }
#}
#
#resource "null_resource" "create_agent_policy" {
#  provisioner "local-exec" {
#    command = "curl -u ${module.ec_deployment.elasticsearch_username}:${module.ec_deployment.elasticsearch_password} ${module.ec_deployment.kibana_url}/api/fleet/agent_policies?sys_monitoring=true -H \"kbn-xsrf: true\" -H \"content-type: application/json\" -d @data/create_agent_policy.json | jq .item.id -r > agent_policy_id.txt"
#  }
#  depends_on = [module.null_resource.create_agent_policy]
#  triggers   = {
#    always_run = timestamp()
#  }
#}

provider "restapi" {
  alias                = "restapi_headers"
  username             = module.ec_deployment.elasticsearch_username
  password             = module.ec_deployment.elasticsearch_password
  uri                  = module.ec_deployment.kibana_url
  id_attribute         = "item/id"
  debug                = true
  write_returns_object = true

  headers = {
    kbn-xsrf      = true
    content-type  = "application/json"
  }
  #  depends_on = [module.ec_deployment]
}

resource "restapi_object" "agent_policy" {
  provider = restapi.restapi_headers
  path     = "/api/fleet/agent_policies"
#  query_string = "?sys_monitoring=true"
  data = "{\"name\":\"my-agent-policy9\",\"description\":\"\",\"namespace\":\"default\",\"monitoring_enabled\":[\"logs\",\"metrics\"]}"
#  create_method = "PUT"
}


#
#resource "null_resource" "get_prometheus_ip" {
#  triggers  =  { always_run = "${timestamp()}" }
#  provisioner "local-exec" {
#    command = "kubectl get svc prometheus-server -n monitoring | awk -F' ' '{print $3}' | tail -1 | tr -d '\n' >> ${path.module}/prometheus_private_ips.txt"
#  }
#  depends_on = ["helm_release.prometheus"]
#}
#
#data "local_file" "prometheus-ip" {
#  filename = "${path.module}/prometheus_private_ips.txt"
#  depends_on = ["null_resource.get_prometheus_ip"]
#}
#
#resource "helm_release" "prometheus-adapter" {
#  name      = "prometheus-adapter"
#  chart     = "${path.module}/helm/charts/stable/prometheus-adapter/"
#  namespace = "default"
#
#  // prometheus URL
#  set {
#    name  = "prometheus.url"
#    value = "http://${data.local_file.prometheus-ip.content}"
#  }
#}
#
resource "random_string" "suffix" {
  length  = 3
  special = false
}
