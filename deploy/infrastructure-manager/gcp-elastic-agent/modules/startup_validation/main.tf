resource "null_resource" "validate_startup" {
  count = var.enabled ? 1 : 0

  provisioner "local-exec" {
    command     = <<-EOT
      #!/bin/bash
      set -e

      MAX_ATTEMPTS=$((${var.timeout} / 10))
      ATTEMPT=0

      echo "Waiting for Elastic Agent startup script to complete..."

      while [ $ATTEMPT -lt $MAX_ATTEMPTS ]; do
        # Query guest attribute
        STATUS=$(gcloud compute instances get-guest-attributes ${var.instance_name} \
          --zone ${var.zone} \
          --query-path=elastic-agent/startup-status \
          --format="value(VALUE)" 2>/dev/null || echo "unknown")

        if [ "$STATUS" = "success" ]; then
          TIMESTAMP=$(gcloud compute instances get-guest-attributes ${var.instance_name} \
            --zone ${var.zone} \
            --query-path=elastic-agent/startup-timestamp \
            --format="value(VALUE)" 2>/dev/null || echo "")
          echo "✓ Elastic Agent installation successful (completed at $TIMESTAMP)"
          exit 0
        elif [ "$STATUS" = "failed" ]; then
          ERROR=$(gcloud compute instances get-guest-attributes ${var.instance_name} \
            --zone ${var.zone} \
            --query-path=elastic-agent/startup-error \
            --format="value(VALUE)" 2>/dev/null || echo "Unknown error")
          TIMESTAMP=$(gcloud compute instances get-guest-attributes ${var.instance_name} \
            --zone ${var.zone} \
            --query-path=elastic-agent/startup-timestamp \
            --format="value(VALUE)" 2>/dev/null || echo "")
          echo "✗ Elastic Agent installation failed (at $TIMESTAMP)"
          echo "Error: $ERROR"
          echo ""
          echo "View detailed logs:"
          echo "  gcloud compute instances get-serial-port-output ${var.instance_name} --zone ${var.zone} | grep elastic-agent-setup"
          exit 1
        fi

        echo "Attempt $ATTEMPT/$MAX_ATTEMPTS: Status is '$STATUS', waiting..."
        sleep 10
        ATTEMPT=$((ATTEMPT + 1))
      done

      echo "✗ Timeout waiting for agent installation (${var.timeout}s)"
      echo "Check status manually:"
      echo "  gcloud compute instances get-guest-attributes ${var.instance_name} --zone ${var.zone} --query-path=elastic-agent/"
      exit 1
    EOT
    interpreter = ["bash", "-c"]
  }

  triggers = {
    instance_id = var.instance_id
  }
}
