data "google_client_config" "default" {}

resource "terraform_data" "validate_startup" {
  count = var.enabled ? 1 : 0

  provisioner "local-exec" {
    command     = <<-EOT
      #!/bin/bash
      set -e
      set +x

      TOKEN="${data.google_client_config.default.access_token}"
      MAX_ATTEMPTS=$((${var.timeout} / 10))
      ATTEMPT=0

      # Function to get guest attribute value
      get_guest_attribute() {
        local key=$1
        local response=$(curl -s -H "Authorization: Bearer $TOKEN" \
          "https://compute.googleapis.com/compute/v1/projects/${var.project_id}/zones/${var.zone}/instances/${var.instance_name}/getGuestAttributes?queryPath=elastic-agent/$key" \
          2>/dev/null || echo '{}')
        echo "$response" | sed -n 's/.*"value":[[:space:]]*"\([^"]*\)".*/\1/p' | head -1
      }

      echo "Waiting for Elastic Agent startup script to complete..."

      while [ $ATTEMPT -lt $MAX_ATTEMPTS ]; do
        STATUS=$(get_guest_attribute "startup-status")
        [ -z "$STATUS" ] && STATUS="unknown"

        if [ "$STATUS" = "success" ]; then
          TIMESTAMP=$(get_guest_attribute "startup-timestamp")
          echo "✓ Elastic Agent installation successful (completed at $TIMESTAMP)"
          exit 0
        elif [ "$STATUS" = "failed" ]; then
          ERROR=$(get_guest_attribute "startup-error")
          [ -z "$ERROR" ] && ERROR="Unknown error"
          TIMESTAMP=$(get_guest_attribute "startup-timestamp")

          # Write to both stdout and stderr for better visibility
          echo "✗ Elastic Agent installation failed (at $TIMESTAMP)" >&2
          echo "Error: $ERROR" >&2
          echo "" >&2
          echo "View detailed logs:" >&2
          echo "  gcloud compute instances get-serial-port-output ${var.instance_name} --zone ${var.zone} --project ${var.project_id}" >&2
          echo "" >&2
          echo "STARTUP_VALIDATION_FAILED: $ERROR" >&2
          exit 1
        fi

        echo "Attempt $ATTEMPT/$MAX_ATTEMPTS: Status is '$STATUS', waiting..."
        sleep 10
        ATTEMPT=$((ATTEMPT + 1))
      done

      echo "✗ Timeout waiting for agent installation (${var.timeout}s)"
      echo "Check status manually:"
      echo "  gcloud compute instances get-guest-attributes ${var.instance_name} --zone ${var.zone} --project ${var.project_id} --query-path=elastic-agent/"
      exit 1
    EOT
    interpreter = ["bash", "-c"]
  }

  triggers_replace = {
    instance_id = var.instance_id
  }
}
