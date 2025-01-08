locals {
  ec_headers = {
    Content-type  = "application/json"
    Authorization = "ApiKey ${var.ec_apikey}"
  }
  wait_script = <<EOT
  set -e
  echo "Waiting for project to be available..."
  sleep_timeout=15
  for i in {1..20}; do
    response=$(curl -s -H "Content-type: application/json" -H "Authorization: ApiKey ${var.ec_apikey}" "${var.ec_url}/api/v1/serverless/projects/security/${restapi_object.ec_project.api_data.id}/status" 2>/dev/null)
    phase=$(echo $response | jq -r '.phase')
    if [ "$phase" == "initialized" ]; then
      echo "Project is available!"
      exit 0
    else
      echo "Retry $i: Project phase is '$phase'. Waiting..."
      sleep $sleep_timeout
    fi
  done
  echo "Project is not available after retries."
  exit 1
  EOT
}

resource "restapi_object" "ec_project" {
  provider = restapi.elastic_cloud
  path     = "/api/v1/serverless/projects/security"
  data = jsonencode({
    region_id = var.region_id
    name      = var.project_name
  })
}

data "http" "project_credentials" {
  url             = "${var.ec_url}/api/v1/serverless/projects/security/${restapi_object.ec_project.api_data.id}/_reset-internal-credentials"
  method          = "POST"
  request_headers = local.ec_headers
}

resource "null_resource" "wait_for_project" {
  depends_on = [restapi_object.ec_project]

  provisioner "local-exec" {
    command = local.wait_script
  }
}
