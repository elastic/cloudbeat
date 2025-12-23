#!/bin/bash
set -e
set +x

# Token from environment variable (sensitive)
TOKEN="$GCP_ACCESS_TOKEN"

# Arguments passed from Terraform
PROJECT_ID="$1"
ZONE="$2"
INSTANCE_NAME="$3"
TIMEOUT="$4"

MAX_ATTEMPTS=$((TIMEOUT / 10))
ATTEMPT=0

# Function to get guest attribute value
get_guest_attribute() {
  local key=$1
  local response=$(curl -s -H "Authorization: Bearer $TOKEN" \
    "https://compute.googleapis.com/compute/v1/projects/${PROJECT_ID}/zones/${ZONE}/instances/${INSTANCE_NAME}/getGuestAttributes?queryPath=elastic-agent/$key" \
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
    echo "  gcloud compute instances get-serial-port-output ${INSTANCE_NAME} --zone ${ZONE} --project ${PROJECT_ID}" >&2
    echo "" >&2
    echo "STARTUP_VALIDATION_FAILED: $ERROR" >&2
    exit 1
  fi

  echo "Attempt $ATTEMPT/$MAX_ATTEMPTS: Status is '$STATUS', waiting..."
  sleep 10
  ATTEMPT=$((ATTEMPT + 1))
done

echo "✗ Timeout waiting for agent installation (${TIMEOUT}s)"
echo "Check status manually:"
echo "  gcloud compute instances get-guest-attributes ${INSTANCE_NAME} --zone ${ZONE} --project ${PROJECT_ID} --query-path=elastic-agent/"
exit 1
