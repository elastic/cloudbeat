#!/bin/bash
set -e

# Configure GCP project
PROJECT_ID=$(gcloud config get-value core/project)
SERVICE_ACCOUNT="infra-manager-deployer"

# Ensure prerequisites are configured
"$(dirname "$0")/setup.sh" "${PROJECT_ID}" "${SERVICE_ACCOUNT}"

# Required environment variables (no defaults - must be provided)
# ELASTIC_RESOURCE_ID - Unique identifier for your Elastic deployment

# Optional environment variables (defaults are in variables.tf or below)
# ORGANIZATION_ID     - Set for org-level monitoring
# DEPLOYMENT_NAME     - Deployment name prefix (default: elastic-agent-sa)
# LOCATION            - GCP region for deployment (default: us-central1)
# OIDC_ISSUER_URI     - OIDC issuer URI

# Validate required inputs
if [ -z "${ELASTIC_RESOURCE_ID}" ]; then
	echo "Error: ELASTIC_RESOURCE_ID environment variable is required"
	echo "This is a unique identifier for your Elastic deployment"
	echo "Example: export ELASTIC_RESOURCE_ID='my-deployment-uuid'"
	exit 1
fi

# Generate unique suffix for resource names (8 hex characters)
RESOURCE_SUFFIX=$(openssl rand -hex 4)

# Set deployment name with suffix
DEPLOYMENT_NAME="${DEPLOYMENT_NAME:-elastic-agent-sa}-${RESOURCE_SUFFIX}"

# Set location (not a TF variable, only used by gcloud)
LOCATION="${LOCATION:-us-central1}"

# Build input values - only include values that are set
# Defaults are defined in variables.tf (single source of truth)
INPUT_VALUES="project_id=${PROJECT_ID}"
INPUT_VALUES="${INPUT_VALUES},resource_suffix=${RESOURCE_SUFFIX}"
INPUT_VALUES="${INPUT_VALUES},elastic_resource_id=${ELASTIC_RESOURCE_ID}"

# Set scope and parent_id based on ORGANIZATION_ID
if [ -n "${ORGANIZATION_ID}" ]; then
	INPUT_VALUES="${INPUT_VALUES},scope=organizations"
	INPUT_VALUES="${INPUT_VALUES},parent_id=${ORGANIZATION_ID}"
else
	INPUT_VALUES="${INPUT_VALUES},scope=projects"
	INPUT_VALUES="${INPUT_VALUES},parent_id=${PROJECT_ID}"
fi

# Optional values - only add if explicitly set (let TF use its defaults otherwise)
[ -n "${OIDC_ISSUER_URI}" ] && INPUT_VALUES="${INPUT_VALUES},oidc_issuer_uri=${OIDC_ISSUER_URI}"

# Deploy from local source (repo already cloned by Cloud Shell)
echo "Starting deployment ${DEPLOYMENT_NAME}..."
gcloud infra-manager deployments apply "${DEPLOYMENT_NAME}" \
	--location="${LOCATION}" \
	--service-account="projects/${PROJECT_ID}/serviceAccounts/${SERVICE_ACCOUNT}@${PROJECT_ID}.iam.gserviceaccount.com" \
	--local-source="." \
	--input-values="${INPUT_VALUES}"

EXIT_CODE=$?
if [ $EXIT_CODE -ne 0 ]; then
	echo ""
	echo "Deployment failed with exit code $EXIT_CODE"
	echo ""
	echo "Common failure reasons:"
	echo "  - Invalid oidc_issuer_uri (verify the OIDC issuer URL is correct)"
	echo "  - Service account permissions missing for ${SERVICE_ACCOUNT}@${PROJECT_ID}.iam.gserviceaccount.com"
	echo "  - Organization ID incorrect (if using organization scope)"
	echo ""
	echo "Useful debugging commands:"
	echo "  # View deployment status"
	echo "  gcloud infra-manager deployments describe ${DEPLOYMENT_NAME} --location=${LOCATION}"
	echo ""
	echo "  # Verify service account permissions"
	echo "  gcloud projects get-iam-policy ${PROJECT_ID} --flatten='bindings[].members' --filter='bindings.members:serviceAccount:${SERVICE_ACCOUNT}@${PROJECT_ID}.iam.gserviceaccount.com' --format='table(bindings.role)'"
	echo ""
	exit $EXIT_CODE
fi

echo ""
echo "Deployment successful!"
echo ""
echo "Outputs:"
REVISION=$(gcloud infra-manager deployments describe "${DEPLOYMENT_NAME}" --location="${LOCATION}" --format="value(latestRevision)")
gcloud infra-manager revisions describe "${REVISION}" --location="${LOCATION}" --format="yaml(applyResults.outputs)"
echo ""
echo "Use these outputs in your Elastic Agent configuration:"
echo "  - target_service_account_email: The service account to impersonate"
echo "  - gcp_audience: The GCP audience URL for Workload Identity Federation"
