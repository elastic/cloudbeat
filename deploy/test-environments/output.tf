# Global output
# ============================================================
output "deployment_name" {
  value       = var.deployment_name
  description = "Terraform deployment name"
}

# EC2 output
# ============================================================
output "ec2_kspm_ssh_cmd" {
  value     = module.aws_ec2_for_kspm.cloudbeat_ssh_cmd
  sensitive = true
}

output "ec2_kspm_public_ip" {
  value     = module.aws_ec2_for_kspm.aws_instance_cloudbeat_public_ip
  sensitive = true
}

output "ec2_kspm_key" {
  value     = module.aws_ec2_for_kspm.ec2_ssh_key
  sensitive = true
}

output "ec2_cspm_ssh_cmd" {
  value     = module.aws_ec2_for_cspm.cloudbeat_ssh_cmd
  sensitive = true
}

output "ec2_cspm_public_ip" {
  value     = module.aws_ec2_for_cspm.aws_instance_cloudbeat_public_ip
  sensitive = true
}

output "ec2_cspm_key" {
  value     = module.aws_ec2_for_cspm.ec2_ssh_key
  sensitive = true
}

output "ec2_asset_inventory_ssh_cmd" {
  value     = module.aws_ec2_for_asset_inventory.cloudbeat_ssh_cmd
  sensitive = true
}

output "ec2_asset_inventory_public_ip" {
  value     = module.aws_ec2_for_asset_inventory.aws_instance_cloudbeat_public_ip
  sensitive = true
}

output "ec2_asset_inventory_key" {
  value     = module.aws_ec2_for_asset_inventory.ec2_ssh_key
  sensitive = true
}

output "ec2_cloudtrail_ssh_cmd" {
  value     = var.cdr_infra ? module.aws_ec2_for_cloudtrail.cloudbeat_ssh_cmd : null
  sensitive = true
}

output "ec2_cloudtrail_public_ip" {
  value     = var.cdr_infra ? module.aws_ec2_for_cloudtrail.aws_instance_cloudbeat_public_ip : null
  sensitive = true
}

output "ec2_cloudtrail_key" {
  value     = var.cdr_infra ? module.aws_ec2_for_cloudtrail.ec2_ssh_key : null
  sensitive = true
}
# =============================================================

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
