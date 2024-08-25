output "gcp_vm_puglic_ip" {
  description = "GCP VM instance public IP"
  value       = google_compute_instance.vm_instance.network_interface[0].access_config[0].nat_ip
}

output "gcp_vm_ssh_cmd" {
  description = "Use this command to SSH into the GCP VM instance"
  value       = "ssh -i ${local.vm_private_key_file} ${local.vm_username}@${google_compute_instance.vm_instance.network_interface[0].access_config[0].nat_ip}"
}

output "gcp_vm_ssh_key" {
  description = "The path to the private SSH key file."
  value       = local.vm_private_key_file
}
