output "email" {
  description = "Service account email"
  value       = google_service_account.elastic_agent.email
}

output "key" {
  description = "Service account private key (base64 encoded)"
  value       = var.create_key ? google_service_account_key.elastic_agent_key[0].private_key : null
  sensitive   = true
}
