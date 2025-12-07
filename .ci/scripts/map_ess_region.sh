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

# Extract cloud provider from ESS_REGION_INPUT and map to deployment template
# Format: {env}-{cloud} (e.g., production-cft, qa-azure, staging-aws)
IFS='-' read -r _env_type cloud_provider <<<"$ESS_REGION_INPUT"

# Map cloud provider to deployment template and max_size
case "$cloud_provider" in
azure)
    DEPLOYMENT_TEMPLATE="azure-storage-optimized"
    MAX_SIZE="60g"
    ;;
aws)
    DEPLOYMENT_TEMPLATE="aws-storage-optimized-faster-warm"
    MAX_SIZE="58g"
    ;;
gcp | cft)
    DEPLOYMENT_TEMPLATE="gcp-storage-optimized"
    MAX_SIZE="128g"
    ;;
*)
    echo "Error: Unknown cloud provider: $cloud_provider" >&2
    echo "Valid cloud providers: azure, aws, gcp, cft" >&2
    exit 1
    ;;
esac

# Export variables for use in calling script
export ESS_REGION
export DEPLOYMENT_TEMPLATE
export MAX_SIZE

# Output for GitHub Actions
if [[ -n "${GITHUB_ENV:-}" ]]; then
    {
        echo "ESS_REGION=$ESS_REGION"
        echo "TF_VAR_ess_region=$ESS_REGION"
        echo "DEPLOYMENT_TEMPLATE=$DEPLOYMENT_TEMPLATE"
        echo "TF_VAR_deployment_template=$DEPLOYMENT_TEMPLATE"
        echo "MAX_SIZE=$MAX_SIZE"
        echo "TF_VAR_max_size=$MAX_SIZE"
    } >>"$GITHUB_ENV"
fi

# Also output to stdout for debugging
echo "Mapped $ESS_REGION_INPUT -> $DEPLOYMENT_TYPE: $ESS_REGION"
echo "Deployment template: $DEPLOYMENT_TEMPLATE (from cloud provider: $cloud_provider)"
echo "Max size: $MAX_SIZE (from cloud provider: $cloud_provider)"
