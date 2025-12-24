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

output "network_name" {
  description = "Name of the VPC network"
  value       = google_compute_network.elastic_agent.name
}

output "instance_self_link" {
  description = "Self-link to the compute instance (check startup status in Cloud Console or Cloud Logging)"
  value       = google_compute_instance.elastic_agent.self_link
}
