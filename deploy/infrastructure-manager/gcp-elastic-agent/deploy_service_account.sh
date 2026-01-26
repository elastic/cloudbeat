#!/bin/bash
set -e

# Shortcut to gcp-credentials-json/deploy.sh
# Creates a service account with roles and key for Elastic Agent GCP integration.
#
# See ../gcp-credentials-json/README.md for details.

"$(dirname "$0")/../gcp-credentials-json/deploy.sh" "$@"
