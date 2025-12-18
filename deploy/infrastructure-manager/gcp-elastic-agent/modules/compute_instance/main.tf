# Compute Instance
resource "google_compute_instance" "elastic_agent" {
  name         = var.deployment_name
  machine_type = var.machine_type
  zone         = var.zone

  labels = {
    name = "elastic-agent"
  }

  boot_disk {
    initialize_params {
      image = "ubuntu-os-cloud/ubuntu-minimal-2204-lts"
      size  = 32
      type  = "pd-standard"
    }
    auto_delete = true
  }

  network_interface {
    network = var.network_self_link

    access_config {
      # Ephemeral public IP
    }
  }

  service_account {
    email = var.sa_email
    scopes = [
      "https://www.googleapis.com/auth/cloud-platform",
      "https://www.googleapis.com/auth/cloudplatformorganizations",
    ]
  }

  metadata_startup_script = <<-EOT
    #!/bin/bash
    set -x
    ElasticAgentArtifact=elastic-agent-${var.elastic_agent_version}-linux-x86_64
    curl -L -O ${var.elastic_artifact_server}/$ElasticAgentArtifact.tar.gz
    tar xzvf $ElasticAgentArtifact.tar.gz
    cd $ElasticAgentArtifact
    ${var.install_command} --url=${var.fleet_url} --enrollment-token=${var.enrollment_token}
  EOT
}
