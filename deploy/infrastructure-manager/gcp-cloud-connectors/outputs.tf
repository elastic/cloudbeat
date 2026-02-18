output "cloud_connector_id" {
  description = "Unique identifier for the cloud connector deployment"
  value       = random_string.cloud_connector_id.result
}

output "target_service_account_email" {
  description = "Email of the target service account (use this in Elastic Agent configuration)"
  value       = module.target_service_account.email
}

output "gcp_audience" {
  description = "GCP audience URL for Workload Identity Federation (use this in Elastic Agent configuration)"
  value       = module.workload_identity.audience
}
