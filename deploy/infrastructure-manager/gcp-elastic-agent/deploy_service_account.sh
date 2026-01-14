#!/bin/bash
set -e

# Shortcut to gcp-credentials-json/deploy_service_account.sh
# Creates a service account with roles and key for Elastic Agent GCP integration.
#
# See ../gcp-credentials-json/README.md for details.

SCRIPT_DIR="$(dirname "$0")"
cd "${SCRIPT_DIR}/../gcp-credentials-json"
exec "./deploy_service_account.sh" "$@"
