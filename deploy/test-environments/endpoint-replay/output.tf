output "deployment_name" {
  value       = var.deployment_name
  description = "Terraform deployment name"
}

output "endpoint_replay_vm_public_ip" {
  value       = module.endpoint_replay_vm.aws_instance_cloudbeat_public_ip
  description = "Endpoint Replay VM public IP"
  sensitive   = true
}

output "endpoint_replay_vm_ssh_key" {
  value       = module.endpoint_replay_vm.ec2_ssh_key
  description = "Endpoint Replay VM SSH key path"
  sensitive   = true
}

output "endpoint_replay_vm_ssh_cmd" {
  value       = module.endpoint_replay_vm.cloudbeat_ssh_cmd
  description = "Use this command to SSH into the Endpoint Replay VM"
  sensitive   = true
}
