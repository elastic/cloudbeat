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

output "ec2_cloudtrail_ssh_cmd" {
  value     = var.cdr_infra ? module.aws_ec2_for_cloudtrail[0].cloudbeat_ssh_cmd : null
  sensitive = true
}

output "ec2_cloudtrail_public_ip" {
  value     = var.cdr_infra ? module.aws_ec2_for_cloudtrail[0].aws_instance_cloudbeat_public_ip : null
  sensitive = true
}

output "ec2_cloudtrail_key" {
  value     = var.cdr_infra ? module.aws_ec2_for_cloudtrail[0].ec2_ssh_key : null
  sensitive = true
}

output "gcp_audit_logs_ssh_cmd" {
  value     = var.cdr_infra ? module.gcp_audit_logs[0].gcp_vm_ssh_cmd : null
  sensitive = true
}

output "gcp_audit_logs_public_ip" {
  value     = var.cdr_infra ? module.gcp_audit_logs[0].gcp_vm_puglic_ip : null
  sensitive = true
}

output "gcp_audit_logs_key" {
  value     = var.cdr_infra ? module.gcp_audit_logs[0].gcp_vm_ssh_key : null
  sensitive = true
}

output "az_vm_activity_logs_ssh_cmd" {
  value     = var.cdr_infra ? module.azure_vm_activity_logs[0].azure_vm_ssh_cmd : null
  sensitive = true
}

output "az_vm_activity_logs_public_ip" {
  value     = var.cdr_infra ? module.azure_vm_activity_logs[0].azure_vm_public_ip : null
  sensitive = true
}

output "az_vm_activity_logs_key" {
  value     = var.cdr_infra ? module.azure_vm_activity_logs[0].azure_vm_ssh_key : null
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
