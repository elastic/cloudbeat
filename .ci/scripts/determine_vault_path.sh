#!/usr/bin/env bash
set -euo pipefail

# Script: determine_vault_path.sh
#
# Description:
# Determines the Vault path for Elastic Cloud credentials based on the ess-region input.
# Parses the environment from the format {env}-{cloud} (e.g., production-cft, qa-gcp).
#
# Usage:
#   export ESS_REGION_INPUT="production-cft"
#   source .ci/scripts/determine_vault_path.sh
#   echo "$VAULT_PATH"
#
# Output:
#   Sets VAULT_PATH environment variable and writes to GITHUB_ENV if available

# Validate input
if [[ -z "${ESS_REGION_INPUT:-}" ]]; then
    echo "Error: ESS_REGION_INPUT environment variable is not set" >&2
    exit 1
fi

# Parse environment from ess-region-input
# Format: {env}-{cloud} (e.g., production-cft, qa-gcp, staging-aws)
IFS='-' read -r env_type _cloud_provider <<<"$ESS_REGION_INPUT"

# Normalize environment name
case "$env_type" in
production)
    ENV_NAME="production"
    ;;
staging)
    ENV_NAME="staging"
    ;;
qa)
    ENV_NAME="qa"
    ;;
*)
    echo "Error: Unknown environment type: $env_type" >&2
    echo "Valid environments: production, staging, qa" >&2
    exit 1
    ;;
esac

VAULT_PATH="secret/csp-team/ci/elastic-cloud/${ENV_NAME}"

# Export variable for use in current shell (when sourced)
export VAULT_PATH

# Output for GitHub Actions
if [[ -n "${GITHUB_ENV:-}" ]]; then
    echo "VAULT_PATH=$VAULT_PATH" >>"$GITHUB_ENV"
fi
