output "kibana_url" {
  value = jsondecode(restapi_object.ec_project.api_response)["endpoints"]["kibana"]
}

output "elasticsearch_url" {
  value = jsondecode(restapi_object.ec_project.api_response)["endpoints"]["elasticsearch"]
}