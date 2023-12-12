output "kibana_url" {
  value = jsondecode(restapi_object.ec_project.api_response)["endpoints"]["kibana"]
}

output "elasticsearch_url" {
  value = "${jsondecode(restapi_object.ec_project.api_response).endpoints.elasticsearch}:443"
}

output "elasticsearch_username" {
  value = jsondecode(data.http.project_credentials.response_body)["username"]
}

output "elasticsearch_password" {
  value = jsondecode(data.http.project_credentials.response_body)["password"]
}
