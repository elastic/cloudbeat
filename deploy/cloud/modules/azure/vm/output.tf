output "azure_vm_ssh_cmd" {
  description = "Use this command SSH into the azure vm instance"
  value       = "ssh -i ${local.vm_private_key_file} ${local.vm_username}@${data.azurerm_public_ip.data-pip.ip_address}"
}

output "azure_vm_ssh_key" {
  description = "Azure ssh key path"
  value       = local.vm_private_key_file
}

output "azure_vm_public_ip" {
  description = "Azure VM public IP"
  value       = data.azurerm_public_ip.data-pip.ip_address
}
