output "email" {
  description = "Service account email"
  value       = google_service_account.elastic_agent.email
}
