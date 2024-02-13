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
