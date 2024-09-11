output "deployment_name" {
  value       = var.deployment_name
  description = "Terraform deployment name"
}

output "ec2_cloudtrail_ssh_cmd" {
  value     = var.deploy_aws_ec2 ? module.aws_ec2_for_cloudtrail[0].cloudbeat_ssh_cmd : null
  sensitive = true
}

output "ec2_cloudtrail_public_ip" {
  value     = var.deploy_aws_ec2 ? module.aws_ec2_for_cloudtrail[0].aws_instance_cloudbeat_public_ip : null
  sensitive = true
}

output "ec2_cloudtrail_key" {
  value     = var.deploy_aws_ec2 ? module.aws_ec2_for_cloudtrail[0].ec2_ssh_key : null
  sensitive = true
}

output "gcp_audit_logs_ssh_cmd" {
  value     = var.deploy_gcp_vm ? module.gcp_audit_logs[0].gcp_vm_ssh_cmd : null
  sensitive = true
}

output "gcp_audit_logs_public_ip" {
  value     = var.deploy_gcp_vm ? module.gcp_audit_logs[0].gcp_vm_puglic_ip : null
  sensitive = true
}

output "gcp_audit_logs_key" {
  value     = var.deploy_gcp_vm ? module.gcp_audit_logs[0].gcp_vm_ssh_key : null
  sensitive = true
}

output "az_vm_activity_logs_ssh_cmd" {
  value     = var.deploy_az_vm ? module.azure_vm_activity_logs[0].azure_vm_ssh_cmd : null
  sensitive = true
}

output "az_vm_activity_logs_public_ip" {
  value     = var.deploy_az_vm ? module.azure_vm_activity_logs[0].azure_vm_public_ip : null
  sensitive = true
}

output "az_vm_activity_logs_key" {
  value     = var.deploy_az_vm ? module.azure_vm_activity_logs[0].azure_vm_ssh_key : null
  sensitive = true
}

output "ec2_asset_inventory_ssh_cmd" {
  value     = module.aws_ec2_for_asset_inventory[0].cloudbeat_ssh_cmd
  sensitive = true
}

output "ec2_asset_inventory_public_ip" {
  value     = module.aws_ec2_for_asset_inventory[0].aws_instance_cloudbeat_public_ip
  sensitive = true
}

output "ec2_asset_inventory_key" {
  value     = module.aws_ec2_for_asset_inventory[0].ec2_ssh_key
  sensitive = true
}
