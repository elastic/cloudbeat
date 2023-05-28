# EC2 output
# ============================================================
output "ec2_kspm_ssh_cmd" {
  value = module.aws_ec2_for_kspm.cloudbeat_ssh_cmd
}

output "ec2_kspm_public_ip" {
  value = module.aws_ec2_for_kspm.aws_instance_cloudbeat_public_ip
}

output "ec2_cspm_ssh_cmd" {
  value = module.aws_ec2_for_cspm.cloudbeat_ssh_cmd
}

output "ec2_cspm_public_ip" {
  value = module.aws_ec2_for_cspm.aws_instance_cloudbeat_public_ip
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

