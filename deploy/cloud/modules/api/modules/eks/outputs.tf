output "agent_policy_id" {
  description = "EKS Agent policy ID"
  value       = restapi_object.agent_policy.id
}

output "enrollment_token" {
  description = "Agent enrollment token"
  value       = local.enrollment_token
}

output "yaml" {
  description = "Kubernetes EKS deployment YAML"
  value       = local.yaml
}

output "manifests" {
  description = "Kubernetes EKS deployment hcl manifests"
  value       = local.manifests
}

output "other_manifests" {
  description = "Kubernetes EKS deployment hcl manifests of all the resources but the service account"
  value       = local.other_manifests
}

output "service_account_manifests" {
  description = "Kubernetes EKS deployment hcl manifests of the service account(s)"
  value       = local.service_account_manifests
}
