# Global output
# ============================================================
output "deployment_name" {
  value       = var.deployment_name
  description = "Terraform deployment name"
}

# EC2 output
# ============================================================
output "ec2_kspm_ssh_cmd" {
  value     = var.deploy_aws_kspm ? module.aws_ec2_for_kspm[0].cloudbeat_ssh_cmd : null
  sensitive = true
}

output "ec2_kspm_public_ip" {
  value     = var.deploy_aws_kspm ? module.aws_ec2_for_kspm[0].aws_instance_cloudbeat_public_ip : null
  sensitive = true
}

output "ec2_kspm_key" {
  value     = var.deploy_aws_kspm ? module.aws_ec2_for_kspm[0].ec2_ssh_key : null
  sensitive = true
}

output "ec2_cspm_ssh_cmd" {
  value     = var.deploy_aws_cspm ? module.aws_ec2_for_cspm[0].cloudbeat_ssh_cmd : null
  sensitive = true
}

output "ec2_cspm_public_ip" {
  value     = var.deploy_aws_cspm ? module.aws_ec2_for_cspm[0].aws_instance_cloudbeat_public_ip : null
  sensitive = true
}

output "ec2_cspm_key" {
  value     = var.deploy_aws_cspm ? module.aws_ec2_for_cspm[0].ec2_ssh_key : null
  sensitive = true
}

# =============================================================
