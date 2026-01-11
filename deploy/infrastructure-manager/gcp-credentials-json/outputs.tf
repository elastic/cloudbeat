output "service_account_email" {
  description = "Email of the created service account"
  value       = google_service_account.elastic_agent.email
}

output "secret_name" {
  description = "Secret Manager secret ID containing the service account key"
  value       = google_secret_manager_secret.sa_key.secret_id
}
