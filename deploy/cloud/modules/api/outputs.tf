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
  value       = local.yaml
}

output "yaml_vanilla" {
  description = "Kubernetes deployment YAML"
  value       = local.yaml_vanilla
}

output "manifests" {
  description = "Kubernetes deployment hcl manifests"
  value       = local.manifests
}

output "other_manifests" {
  description = "Kubernetes deployment hcl manifests of all the resources but the service account"
  value       = local.other_manifests
}

output "service_account_manifests" {
  description = "Kubernetes deployment hcl manifests of the service account(s)"
  value       = local.service_account_manifests
}
