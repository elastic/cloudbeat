#!/usr/bin/env bash
#
# Shuts down the Elastic Cloud deployment created by create_stack.sh.
# Meant for an `if: always()` step. Nothing sensitive is printed.
#
# Inputs (environment):
#   EC_API_KEY        Elastic Cloud API key (secret)
#   DEPLOYMENT_ID     ID exported by create_stack.sh (preferred)
#   DEPLOYMENT_NAME   Fallback: look the deployment up by name if no ID was exported
#   EC_REGION         Region, default gcp-us-west2

set -euo pipefail
set +x

: "${EC_API_KEY:?EC_API_KEY must be set}"
export EC_REGION="${EC_REGION:-gcp-us-west2}"

deployment_id="${DEPLOYMENT_ID:-}"

if [ -z "${deployment_id}" ] && [ -n "${DEPLOYMENT_NAME:-}" ]; then
    # create_stack.sh may have failed before it could export the ID.
    deployment_id="$(ecctl deployment list --output json |
        jq -r --arg name "${DEPLOYMENT_NAME}" '[.deployments[]? | select(.name == $name) | .id][0] // empty')"
fi

if [ -z "${deployment_id}" ]; then
    echo "No deployment to shut down"
    exit 0
fi

echo "Shutting down deployment ${deployment_id}"
# The JSON output is not needed; discard it.
ecctl deployment shutdown "${deployment_id}" --force --track --output json >/dev/null
echo "Deployment ${deployment_id} shut down"
