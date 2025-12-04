#!/usr/bin/env bash
set -euo pipefail

# Script: map_ess_region.sh
#
# Description:
# Maps the ess-region input (e.g., production-cft, staging-aws) to actual
# Elastic Cloud regions based on whether it's ESS or Serverless deployment.
#
# Usage:
#   export ESS_REGION_INPUT="production-cft"
#   export SERVERLESS_MODE="true"  # or "false"
#   source .ci/scripts/map_ess_region.sh
#   echo "ESS_REGION=$ESS_REGION"
#   echo "SERVERLESS_REGION_ID=$SERVERLESS_REGION_ID"

# Validate input
if [[ -z "${ESS_REGION_INPUT:-}" ]]; then
    echo "Error: ESS_REGION_INPUT environment variable is not set" >&2
    exit 1
fi

# Define region mappings for ESS deployments
declare -A ECH_REGIONS=(
    ["production-cft"]="gcp-us-west2"
    ["staging-gcp"]="gcp-us-central1"
    ["staging-aws"]="aws-eu-central-1"
    ["staging-azure"]="azure-eastus2"
    ["qa-gcp"]="gcp-us-central1"
    ["qa-aws"]="aws-eu-west-1"
    ["qa-azure"]="azure-eastus2"
)

# Define region mappings for Serverless deployments
declare -A SERVERLESS_REGIONS=(
    ["production-cft"]="aws-us-east-1"
    ["staging-gcp"]="gcp-us-central1"
    ["staging-aws"]="aws-eu-central-1"
    ["staging-azure"]="azure-canadacentral"
    ["qa-gcp"]="gcp-us-central1"
    ["qa-aws"]="aws-us-west-1"
    ["qa-azure"]="azure-eastus2"
)

# Map region based on deployment type
if [[ "${SERVERLESS_MODE:-false}" == "true" ]]; then
    ESS_REGION="${SERVERLESS_REGIONS[$ESS_REGION_INPUT]:-}"
    DEPLOYMENT_TYPE="Serverless"
else
    ESS_REGION="${ECH_REGIONS[$ESS_REGION_INPUT]:-}"
    DEPLOYMENT_TYPE="ESS"
fi

# Validate region was found
if [[ -z "$ESS_REGION" ]]; then
    echo "Error: Invalid ess-region for $DEPLOYMENT_TYPE: $ESS_REGION_INPUT" >&2
    if [[ "${SERVERLESS_MODE:-false}" == "true" ]]; then
        echo "Valid options: ${!SERVERLESS_REGIONS[*]}" >&2
    else
        echo "Valid options: ${!ECH_REGIONS[*]}" >&2
    fi
    exit 1
fi

# Export variable for use in calling script
export ESS_REGION

# Output for GitHub Actions
if [[ -n "${GITHUB_ENV:-}" ]]; then
    echo "ESS_REGION=$ESS_REGION" >>"$GITHUB_ENV"
    echo "TF_VAR_ess_region=$ESS_REGION" >>"$GITHUB_ENV"
fi

# Also output to stdout for debugging
echo "Mapped $ESS_REGION_INPUT -> $DEPLOYMENT_TYPE: $ESS_REGION"
