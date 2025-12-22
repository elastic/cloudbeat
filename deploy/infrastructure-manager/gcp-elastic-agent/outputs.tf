output "instance_name" {
  description = "Name of the compute instance"
  value       = module.compute_instance.name
}

output "instance_id" {
  description = "ID of the compute instance"
  value       = module.compute_instance.id
}

output "instance_zone" {
  description = "Zone of the compute instance"
  value       = module.compute_instance.zone
}

output "network_name" {
  description = "Name of the VPC network"
  value       = module.compute_instance.network_name
}

output "service_account_email" {
  description = "Email of the service account used by the instance"
  value       = local.sa_email
}

output "startup_validation_enabled" {
  description = "Whether startup script validation was enabled"
  value       = var.startup_validation_enabled
}
