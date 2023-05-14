locals {
  fleet_url    = jsondecode(data.http.fleet_url.response_body).item.fleet_server_hosts[0]
  isSemVar     = length(split(var.stack_version, ".")) == 3
  major        = local.isSemVar ? split(var.stack_version, ".")[0] : 99
  minor        = local.isSemVar ? split(var.stack_version, ".")[1] : 0
  is8_7OrAbove = local.isSemVar ? local.major > 8 || (local.major == 8 && local.minor >= 7) : true
}

data "http" "fleet_url" {
  url = "${var.uri}/api/fleet/settings"

  request_headers = {
    kbn-xsrf = true
    Authorization : "Basic ${base64encode("${var.username}:${var.password}")}"
  }
}

module "vanilla" {
  source    = "./modules/vanilla"
  providers = { restapi : restapi }

  depends_on = [data.http.fleet_url]

  fleet_url        = local.fleet_url
  username         = var.username
  password         = var.password
  uri              = var.uri
  role_arn         = var.role_arn
  agent_docker_img = var.agent_docker_img
}

module "eks" {
  source    = "./modules/eks"
  providers = { restapi : restapi }

  depends_on = [data.http.fleet_url]

  fleet_url        = local.fleet_url
  username         = var.username
  password         = var.password
  uri              = var.uri
  role_arn         = var.role_arn
  agent_docker_img = var.agent_docker_img
}

module "cspm_aws" {
  source    = "./modules/cspm_aws"
  providers = { restapi : restapi }

  depends_on = [data.http.fleet_url]
  count      = local.is8_7OrAbove ? 1 : 0 # Supported from 8.7.0

  fleet_url        = local.fleet_url
  username         = var.username
  password         = var.password
  uri              = var.uri
  agent_docker_img = var.agent_docker_img
}
