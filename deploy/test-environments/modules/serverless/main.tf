locals {
  ec_headers = {
    Content-type  = "application/json"
    Authorization = "ApiKey ${var.ec_apikey}"
  }
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
    # command = local.wait_script
    command = "./wait_for_project.sh"
    interpreter = ["/bin/bash", "-c"]
    environment = {
      "API_KEY" = var.ec_apikey
      "EC_URL"  = var.ec_url
      "PROJECT_ID" = restapi_object.ec_project.api_data.id
    }
  }
}
