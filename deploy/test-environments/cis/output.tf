# Global output
# ============================================================
output "deployment_name" {
  value       = var.deployment_name
  description = "Terraform deployment name"
}

# EC2 output
# ============================================================
output "ec2_kspm_ssh_cmd" {
  value     = module.aws_ec2_for_kspm[0].cloudbeat_ssh_cmd
  sensitive = true
}

output "ec2_kspm_public_ip" {
  value     = module.aws_ec2_for_kspm[0].aws_instance_cloudbeat_public_ip
  sensitive = true
}

output "ec2_kspm_key" {
  value     = module.aws_ec2_for_kspm[0].ec2_ssh_key
  sensitive = true
}

output "ec2_cspm_ssh_cmd" {
  value     = module.aws_ec2_for_cspm[0].cloudbeat_ssh_cmd
  sensitive = true
}

output "ec2_cspm_public_ip" {
  value     = module.aws_ec2_for_cspm[0].aws_instance_cloudbeat_public_ip
  sensitive = true
}

output "ec2_cspm_key" {
  value     = module.aws_ec2_for_cspm[0].ec2_ssh_key
  sensitive = true
}
# =============================================================
