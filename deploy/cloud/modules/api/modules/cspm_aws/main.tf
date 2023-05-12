resource "restapi_object" "agent_policy" {
  provider     = restapi
  path         = "/api/fleet/agent_policies"
  id_attribute = "item/id"
  data         = file("data/agent_policy_cspm_aws.json")

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
  data = templatefile("data/package_policy_cspm_aws.json", {
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

data "http" "yaml" {
  url = "${var.uri}/api/fleet/kubernetes?fleetServer=${var.fleet_url}&enrolToken=${local.enrollment_token}"

  request_headers = {
    kbn-xsrf = true
    Authorization : "Basic ${base64encode("${var.username}:${var.password}")}"
  }

  depends_on = [data.restapi_object.enrollment_token]
}

locals {
  yaml_src                = jsondecode(data.http.yaml.response_body).item
  origin_docker_agent_img = regex("image: ([\\w\\.\\-\\/:]+)", local.yaml_src)[0]
  enrollment_token        = regex("api_key:(.*\\=\\=)", data.restapi_object.enrollment_token.api_data.item)[0]
  docker_cmd = <<-EOT
sudo docker run -d --platform linux/x86_64 \
  --name cspm_aws_agent
  -e "FLEET_URL=${var.fleet_url}" \
  -e "FLEET_ENROLLMENT_TOKEN=${local.enrollment_token}" \
  -e "FLEET_ENROLL=1" \
  ${var.agent_docker_img != "" ? var.agent_docker_img : local.origin_docker_agent_img}
EOT
}
