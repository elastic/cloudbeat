resource "restapi_object" "agent_policy" {
  provider     = restapi
  path         = "/api/fleet/agent_policies"
  id_attribute = "item/id"
  data         = file("data/agent_policy_eks.json")

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
  data = templatefile("data/package_policy_eks.json", {
    agent_policy_id = restapi_object.agent_policy.id
    role_arn        = var.role_arn
  })
}

data "restapi_object" "enrollment_token" {
  provider     = restapi
  search_key   = "policy_id"
  search_value = restapi_object.agent_policy.id
  results_key  = "items"
  path         = "/api/fleet/enrollment_api_keys"

  depends_on = [restapi_object.agent_policy]
}

locals {
  agent_policy_id          = restapi_object.agent_policy.id
  enrollment_token         = regex("api_key:(.*\\=\\=)", data.restapi_object.enrollment_token.api_data.item)[0]
}

data "http" "yaml" {
  url = "${var.uri}/api/fleet/kubernetes?fleetServer=${var.fleet_url}&enrolToken=${local.enrollment_token}"

  request_headers = {
    kbn-xsrf = true
    Authorization : "Basic ${base64encode("${var.username}:${var.password}")}"
  }

  depends_on = [data.restapi_object.enrollment_token]
}

locals {
  yaml_src                  = jsondecode(data.http.yaml.response_body).item
  yaml                      = var.agent_docker_img != "" ? replace(local.yaml_src, "/image: docker\\.elastic\\.co\\/beats\\/elastic-agent:\\d+\\.\\d+\\.\\d/", "image: ${var.agent_docker_img}") : local.yaml_src
  yaml_manifests            = compact(split("---\n", local.yaml))
  manifests                 = { for index, manifest in local.yaml_manifests : index => yamldecode(manifest) }
  other_manifests           = { for index, manifest in local.manifests : index => manifest if !(manifest.kind == "ServiceAccount") }
  service_account_manifests = { for index, manifest in local.manifests : index => manifest if(manifest.kind == "ServiceAccount") }
}
