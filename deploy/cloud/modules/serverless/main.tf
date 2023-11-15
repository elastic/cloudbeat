provider "restapi" {
  uri      = var.ec_url
  insecure = true

  write_returns_object = true

  headers = {
    Content-type  = "application/json"
    Authorization = "ApiKey ${var.ec_apikey}"
  }
}


resource "restapi_object" "ec_project" {
  path = "/api/v1/serverless/projects/security"
  data = jsonencode({
    region_id = var.region_id
    name      = var.project_name
  })
}

