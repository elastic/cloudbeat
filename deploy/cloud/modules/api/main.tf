locals {
  fleet_url = jsondecode(data.http.fleet_url.response_body).item.fleet_server_hosts[0]
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
