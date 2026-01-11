output "pool_name" {
  description = "Full name of the Workload Identity Pool"
  value       = google_iam_workload_identity_pool.elastic.name
}

output "pool_id" {
  description = "ID of the Workload Identity Pool"
  value       = google_iam_workload_identity_pool.elastic.workload_identity_pool_id
}

output "provider_name" {
  description = "Full name of the Workload Identity Provider"
  value       = google_iam_workload_identity_pool_provider.oidc.name
}

output "provider_id" {
  description = "ID of the Workload Identity Provider"
  value       = google_iam_workload_identity_pool_provider.oidc.workload_identity_pool_provider_id
}

output "audience" {
  description = "GCP audience URL for Workload Identity Federation (use in Elastic Agent config)"
  value       = "//iam.googleapis.com/${google_iam_workload_identity_pool_provider.oidc.name}"
}

