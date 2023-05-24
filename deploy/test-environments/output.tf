# EC2 output
# ============================================================
output "ec2_kind_ssh_cmd" {
  value = module.aws_ec2_kind.cloudbeat_ssh_cmd
}

output "ec2_kind_public_ip" {
  value = module.aws_ec2_kind.aws_instance_cloudbeat_public_ip
}
# =============================================================

# Elastic Cloud output
# =============================================================
output "elasticsearch_url" {
  value       = module.ec_deployment.elasticsearch_url
  description = "The secure Elasticsearch URL"
}

output "elasticsearch_username" {
  value       = module.ec_deployment.elasticsearch_username
  description = "The Elasticsearch username"
  sensitive   = true
}

output "elasticsearch_password" {
  value       = module.ec_deployment.elasticsearch_password
  description = "The Elasticsearch password"
  sensitive   = true
}

output "kibana_url" {
  value       = module.ec_deployment.kibana_url
  description = "The secure Kibana URL"
}

output "admin_console_url" {
  value       = module.ec_deployment.admin_console_url
  description = "The admin console URL"
}
# =============================================================

