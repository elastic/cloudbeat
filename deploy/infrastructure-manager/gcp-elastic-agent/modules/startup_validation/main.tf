resource "terraform_data" "validate_startup" {
  count = var.enabled ? 1 : 0

  provisioner "local-exec" {
    command     = <<-EOT
      #!/bin/bash
      set -e

      TOKEN="${data.google_client_config.default.access_token}"
      MAX_ATTEMPTS=$((${var.timeout} / 10))
      ATTEMPT=0

      echo "Waiting for Elastic Agent startup script to complete..."

      while [ $ATTEMPT -lt $MAX_ATTEMPTS ]; do
        # Query guest attribute using Compute Engine API
        RESPONSE=$(curl -s -H "Authorization: Bearer $TOKEN" \
          "https://compute.googleapis.com/compute/v1/projects/${var.project_id}/zones/${var.zone}/instances/${var.instance_name}/getGuestAttributes?queryPath=elastic-agent/startup-status" \
          || echo '{}')

        STATUS=$(echo "$RESPONSE" | grep -o '"value":"[^"]*"' | head -n1 | cut -d'"' -f4 || echo "unknown")

        if [ "$STATUS" = "success" ]; then
          TIMESTAMP_RESPONSE=$(curl -s -H "Authorization: Bearer $TOKEN" \
            "https://compute.googleapis.com/compute/v1/projects/${var.project_id}/zones/${var.zone}/instances/${var.instance_name}/getGuestAttributes?queryPath=elastic-agent/startup-timestamp" \
            || echo '{}')
          TIMESTAMP=$(echo "$TIMESTAMP_RESPONSE" | grep -o '"value":"[^"]*"' | head -n1 | cut -d'"' -f4 || echo "")
          echo "✓ Elastic Agent installation successful (completed at $TIMESTAMP)"
          exit 0
        elif [ "$STATUS" = "failed" ]; then
          ERROR_RESPONSE=$(curl -s -H "Authorization: Bearer $TOKEN" \
            "https://compute.googleapis.com/compute/v1/projects/${var.project_id}/zones/${var.zone}/instances/${var.instance_name}/getGuestAttributes?queryPath=elastic-agent/startup-error" \
            || echo '{}')
          ERROR=$(echo "$ERROR_RESPONSE" | grep -o '"value":"[^"]*"' | head -n1 | cut -d'"' -f4 || echo "Unknown error")

          TIMESTAMP_RESPONSE=$(curl -s -H "Authorization: Bearer $TOKEN" \
            "https://compute.googleapis.com/compute/v1/projects/${var.project_id}/zones/${var.zone}/instances/${var.instance_name}/getGuestAttributes?queryPath=elastic-agent/startup-timestamp" \
            || echo '{}')
          TIMESTAMP=$(echo "$TIMESTAMP_RESPONSE" | grep -o '"value":"[^"]*"' | head -n1 | cut -d'"' -f4 || echo "")

          # Write to both stdout and stderr for better visibility
          echo "✗ Elastic Agent installation failed (at $TIMESTAMP)" >&2
          echo "Error: $ERROR" >&2
          echo "" >&2
          echo "View detailed logs:" >&2
          echo "  curl -H \"Authorization: Bearer \$TOKEN\" \"https://compute.googleapis.com/compute/v1/projects/${var.project_id}/zones/${var.zone}/instances/${var.instance_name}/serialPortOutput\"" >&2
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
      echo "  curl -H \"Authorization: Bearer \$TOKEN\" \"https://compute.googleapis.com/compute/v1/projects/${var.project_id}/zones/${var.zone}/instances/${var.instance_name}/getGuestAttributes?queryPath=elastic-agent/\""
      exit 1
    EOT
    interpreter = ["bash", "-c"]
  }

  triggers_replace = {
    instance_id = var.instance_id
  }
}
