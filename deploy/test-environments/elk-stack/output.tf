# Global output
# ============================================================
output "deployment_name" {
  value       = var.deployment_name
  description = "Terraform deployment name"
}

# Elastic Cloud output
# =============================================================
output "elasticsearch_url" {
  value       = var.serverless_mode ? module.ec_project[0].elasticsearch_url : module.ec_deployment[0].elasticsearch_url
  description = "The secure Elasticsearch URL"
}

output "elasticsearch_username" {
  value       = var.serverless_mode ? module.ec_project[0].elasticsearch_username : module.ec_deployment[0].elasticsearch_username
  description = "The Elasticsearch username"
  sensitive   = true
}

output "elasticsearch_password" {
  value       = var.serverless_mode ? module.ec_project[0].elasticsearch_password : module.ec_deployment[0].elasticsearch_password
  description = "The Elasticsearch password"
  sensitive   = true
}

output "kibana_url" {
  value       = var.serverless_mode ? module.ec_project[0].kibana_url : module.ec_deployment[0].kibana_url
  description = "The secure Kibana URL"
}

# =============================================================
