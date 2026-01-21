output "email" {
  description = "Email of the target service account"
  value       = google_service_account.target.email
}

output "unique_id" {
  description = "Unique ID of the target service account"
  value       = google_service_account.target.unique_id
}

output "name" {
  description = "Name (resource ID) of the target service account"
  value       = google_service_account.target.name
}
