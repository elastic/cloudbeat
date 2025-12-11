#!/usr/bin/env bash
set -euo pipefail

# Script: determine_vault_path.sh
#
# Description:
# Determines the Vault path for Elastic Cloud credentials based on the environment type.
# Requires ENV_TYPE to be set (typically by parse_ess_region.sh).
#
# Usage:
#   export ENV_TYPE="production"
#   source .ci/scripts/determine_vault_path.sh
#   echo "$VAULT_PATH"
#
# Output:
#   Sets VAULT_PATH environment variable and writes to GITHUB_ENV if available

# Validate input
if [[ -z "${ENV_TYPE:-}" ]]; then
    echo "Error: ENV_TYPE environment variable is not set" >&2
    exit 1
fi

# Normalize environment name
case "$ENV_TYPE" in
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
    echo "Error: Unknown environment type: $ENV_TYPE" >&2
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
