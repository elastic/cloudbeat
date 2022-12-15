output "agent_policy_id" {
  description = "Agent policy ID"
  value       = restapi_object.agent_policy.id
}

output "enrollment_token" {
  description = "Agent enrollment token"
  value       = local.enrollment_token
}

output "fleet_url" {
  description = "Fleet Server URL"
  value       = local.fleet_url
}

output "yaml" {
  description = "Kubernetes deployment YAML"
  value       = jsondecode(data.http.yaml.response_body).item
}
