provider "restapi" {
  alias                = "restapi_headers"
  username             = var.username
  password             = var.password
  uri                  = var.url
  debug                = true
  write_returns_object = true

  headers = {
    kbn-xsrf     = true
    content-type = "application/json"
  }
}

resource "restapi_object" "agent_policy" {
  provider     = restapi.restapi_headers
  path         = "/api/fleet/agent_policies"
  id_attribute = "item/id"
  #  data       = data.local_file.agent_policy.content
  data         = file("data/agent_policy.json")

  # destroy does not work :(
  destroy_path   = "/api/fleet/agent_policies/delete"
  destroy_method = "POST"
  destroy_data   = <<EOF
{
  "agentPolicyId" : "{id}"
}
EOF
}

resource "restapi_object" "package_policy" {
  provider   = restapi.restapi_headers
  depends_on = [restapi_object.agent_policy]
  path       = "/api/fleet/package_policies"
  data       = templatefile("data/package_policy.json", {
    agent_policy_id = restapi_object.agent_policy.id
  })
}

data "restapi_object" "enrollment_token" {
  provider     = restapi.restapi_headers
  search_key   = "policy_id"
  search_value = restapi_object.agent_policy.id
  results_key  = "items"
  depends_on   = [restapi_object.agent_policy]
  path         = "/api/fleet/enrollment_api_keys"
}

data "restapi_object" "fleet_url" {
  provider     = restapi.restapi_headers
  depends_on   = [restapi_object.agent_policy]
  path         = "/api/fleet/outputs"
  search_key   = "id"
  search_value = "fleet-default-output"
  results_key  = "items"
}

locals {
  agent_policy_id  = restapi_object.agent_policy.id
  enrollment_token = regex("api_key:(.*\\=\\=)", data.restapi_object.enrollment_token.api_data.item)[0]
  fleet_url        = regex("hosts:\\[(.*)\\] ", data.restapi_object.fleet_url.api_data.item)[0]
}

data "http" "yaml" {
  url = "${var.url}/api/fleet/kubernetes?fleetServer=${local.fleet_url}&enrolToken=${local.enrollment_token}"

  request_headers = {
    kbn-xsrf     = true
    Authorization: "Basic ${base64encode("${var.username}:${var.password}")}"
  }
}

