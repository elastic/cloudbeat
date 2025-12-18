output "name" {
  description = "Name of the compute instance"
  value       = google_compute_instance.elastic_agent.name
}

output "id" {
  description = "ID of the compute instance"
  value       = google_compute_instance.elastic_agent.id
}

output "zone" {
  description = "Zone of the compute instance"
  value       = google_compute_instance.elastic_agent.zone
}
