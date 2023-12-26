output "kibana_url" {
  value       = ec_deployment.deployment.kibana.https_endpoint
  description = "The secure Kibana URL"
}

output "elasticsearch_url" {
  value       = ec_deployment.deployment.elasticsearch.https_endpoint
  description = "The secure Elasticsearch URL"
}

output "elasticsearch_username" {
  value       = ec_deployment.deployment.elasticsearch_username
  sensitive   = true
  description = "The Elasticsearch username"
}

output "elasticsearch_password" {
  value       = ec_deployment.deployment.elasticsearch_password
  sensitive   = true
  description = "The Elasticsearch password"
}

output "stack_version" {
  value       = data.ec_stack.deployment_version.version
  description = "The matching stack pack version from the provided stack_version"
}
