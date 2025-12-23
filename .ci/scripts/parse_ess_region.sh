#!/usr/bin/env bash
set -euo pipefail

# Script: parse_ess_region.sh
#
# Description:
# Parses the ESS_REGION_INPUT to extract environment type and cloud provider.
# Format: {env}-{cloud} (e.g., production-cft, qa-gcp, staging-aws)
#
# Usage:
#   export ESS_REGION_INPUT="production-cft"
#   source .ci/scripts/parse_ess_region.sh
#   echo "ENV_TYPE=$ENV_TYPE"
#   echo "CLOUD_PROVIDER=$CLOUD_PROVIDER"
#
# Output:
#   Sets ENV_TYPE and CLOUD_PROVIDER environment variables and writes to GITHUB_ENV if available

# Validate input
if [[ -z "${ESS_REGION_INPUT:-}" ]]; then
    echo "Error: ESS_REGION_INPUT environment variable is not set" >&2
    exit 1
fi

# Parse environment type and cloud provider from ESS_REGION_INPUT
# Format: {env}-{cloud} (e.g., production-cft, qa-gcp, staging-aws)
ENV_TYPE="${ESS_REGION_INPUT%%-*}"
CLOUD_PROVIDER="${ESS_REGION_INPUT#*-}"

# Export variables for use in current shell (when sourced)
export ENV_TYPE
export CLOUD_PROVIDER

# Output for GitHub Actions (persists across steps)
if [[ -n "${GITHUB_ENV:-}" ]]; then
    echo "ENV_TYPE=$ENV_TYPE" >>"$GITHUB_ENV"
    echo "CLOUD_PROVIDER=$CLOUD_PROVIDER" >>"$GITHUB_ENV"
fi
