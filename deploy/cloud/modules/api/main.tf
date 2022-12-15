resource "restapi_object" "agent_policy" {
  provider     = restapi
  path         = "/api/fleet/agent_policies"
  id_attribute = "item/id"
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
  provider     = restapi
  depends_on   = [restapi_object.agent_policy]
  path         = "/api/fleet/package_policies"
  id_attribute = "item/id"
  data         = templatefile("data/package_policy.json", {
    agent_policy_id = restapi_object.agent_policy.id
    role_arn = var.role_arn
  })
}

data "restapi_object" "enrollment_token" {
  provider     = restapi
  search_key   = "policy_id"
  search_value = restapi_object.agent_policy.id
  results_key  = "items"
  path         = "/api/fleet/enrollment_api_keys"

   depends_on   = [restapi_object.agent_policy]
}

locals {
  agent_policy_id  = restapi_object.agent_policy.id
  enrollment_token = regex("api_key:(.*\\=\\=)", data.restapi_object.enrollment_token.api_data.item)[0]
  fleet_url       = jsondecode(data.http.fleet_url.response_body).item.fleet_server_hosts[0]
}

data "http" "fleet_url" {
  url = "${var.uri}/api/fleet/settings"

  request_headers = {
    kbn-xsrf = true
    Authorization : "Basic ${base64encode("${var.username}:${var.password}")}"
  }
}

data "http" "yaml" {
  url = "${var.uri}/api/fleet/kubernetes?fleetServer=${local.fleet_url}&enrolToken=${local.enrollment_token}"

  request_headers = {
    kbn-xsrf = true
    Authorization : "Basic ${base64encode("${var.username}:${var.password}")}"
  }

  depends_on   = [data.restapi_object.enrollment_token, data.http.fleet_url]
}

