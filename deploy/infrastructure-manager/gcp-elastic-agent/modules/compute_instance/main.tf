locals {
  # Determine install command based on version
  install_command = startswith(var.elastic_agent_version, "9.") ? "sudo ./elastic-agent install --non-interactive --install-servers" : "sudo ./elastic-agent install --non-interactive"
}

# VPC Network
resource "google_compute_network" "elastic_agent" {
  name                    = var.network_name
  auto_create_subnetworks = true
  routing_mode            = "REGIONAL"
}

# Compute Instance
resource "google_compute_instance" "elastic_agent" {
  name         = var.instance_name
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
    network = google_compute_network.elastic_agent.self_link

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

  metadata = {
    enable-guest-attributes = "TRUE"
    startup-script          = <<-EOT
    #!/bin/bash
    set -x  # Enable debug output

    # Logging function - logs to both stdout and Cloud Logging
    log() {
      echo "[$(date +'%Y-%m-%d %H:%M:%S')] $1"
      logger -t elastic-agent-setup "$1"
    }

    # Function to set guest attribute
    set_guest_attribute() {
      local key=$1
      local value=$2
      curl -X PUT --data "$value" \
        -H "Metadata-Flavor: Google" \
        "http://metadata.google.internal/computeMetadata/v1/instance/guest-attributes/elastic-agent/$key" \
        || log "WARNING: Failed to set guest attribute $key"
    }

    # Function to report failure
    report_failure() {
      local error_msg="$1"
      log "ERROR: $error_msg"
      set_guest_attribute "startup-status" "failed"
      set_guest_attribute "startup-error" "$error_msg"
      set_guest_attribute "startup-timestamp" "$(date -u +%Y-%m-%dT%H:%M:%SZ)"
      exit 1
    }

    log "Starting Elastic Agent installation"
    set_guest_attribute "startup-status" "in-progress"

    # Download Elastic Agent
    ElasticAgentArtifact=elastic-agent-${var.elastic_agent_version}-linux-x86_64
    ARTIFACT_URL="${var.elastic_artifact_server}/$ElasticAgentArtifact.tar.gz"

    log "Downloading Elastic Agent from $ARTIFACT_URL"
    if ! curl -f -L -O --connect-timeout 30 --max-time 300 "$ARTIFACT_URL"; then
      report_failure "Failed to download Elastic Agent from $ARTIFACT_URL"
    fi
    log "Download successful"

    # Verify download
    if [ ! -f "$ElasticAgentArtifact.tar.gz" ]; then
      report_failure "Downloaded file not found: $ElasticAgentArtifact.tar.gz"
    fi

    # Extract archive
    log "Extracting $ElasticAgentArtifact.tar.gz"
    if ! tar xzvf "$ElasticAgentArtifact.tar.gz"; then
      report_failure "Failed to extract $ElasticAgentArtifact.tar.gz"
    fi

    # Verify extraction
    if [ ! -d "$ElasticAgentArtifact" ]; then
      report_failure "Extracted directory not found: $ElasticAgentArtifact"
    fi

    cd "$ElasticAgentArtifact"

    # Install Elastic Agent
    log "Installing Elastic Agent with command: ${local.install_command}"
    if ! ${local.install_command} --url=${var.fleet_url} --enrollment-token=${var.enrollment_token}; then
      report_failure "Elastic Agent installation command failed"
    fi

    # Verify installation
    log "Verifying Elastic Agent installation"
    if systemctl is-active --quiet elastic-agent; then
      log "SUCCESS: Elastic Agent is running"
      set_guest_attribute "startup-status" "success"
      set_guest_attribute "startup-timestamp" "$(date -u +%Y-%m-%dT%H:%M:%SZ)"
    else
      report_failure "Elastic Agent service is not running after installation"
    fi

    # Cleanup downloaded files
    cd ..
    rm -rf "$ElasticAgentArtifact.tar.gz" "$ElasticAgentArtifact"
    log "Cleanup completed"
    EOT
  }
}
