#!/bin/bash
set -e

# This script:
# 1. Enables necessary APIs for Elastic Agent GCP integration
# 2. Deploys Terraform via GCP Infrastructure Manager to create a service account with roles and key
# 3. Saves generated key to KEY_FILE.json
# 4. Outputs instructions for using the key in Elastic Agent GCP integration

# Configure GCP project
PROJECT_ID=$(gcloud config get-value core/project)
SERVICE_ACCOUNT="infra-manager-deployer"

# Ensure prerequisites are configured
"$(dirname "$0")/setup.sh" "${PROJECT_ID}" "${SERVICE_ACCOUNT}"

# Optional environment variables (defaults are in variables.tf or below)
# ORG_ID          - Set for org-level monitoring
# DEPLOYMENT_NAME - Deployment name prefix (default: elastic-agent-credentials)
# LOCATION        - GCP region for deployment (default: us-central1)

# Generate unique suffix for resource names (8 hex characters)
RESOURCE_SUFFIX=$(openssl rand -hex 4)

# Set deployment name with suffix
DEPLOYMENT_NAME="${DEPLOYMENT_NAME:-elastic-agent-credentials}-${RESOURCE_SUFFIX}"

# Set location (not a TF variable, only used by gcloud)
LOCATION="${LOCATION:-us-central1}"

RED='\033[0;31m'
GREEN='\033[0;32m'
RESET='\033[0m'

# Build input values - only include values that are set
# Defaults are defined in variables.tf (single source of truth)
INPUT_VALUES="project_id=${PROJECT_ID}"
INPUT_VALUES="${INPUT_VALUES},resource_suffix=${RESOURCE_SUFFIX}"

# Set scope and parent_id based on ORG_ID
if [ -n "${ORG_ID}" ]; then
    INPUT_VALUES="${INPUT_VALUES},scope=organizations"
    INPUT_VALUES="${INPUT_VALUES},parent_id=${ORG_ID}"
else
    INPUT_VALUES="${INPUT_VALUES},scope=projects"
    INPUT_VALUES="${INPUT_VALUES},parent_id=${PROJECT_ID}"
fi

echo -e "${GREEN}Starting deployment '${DEPLOYMENT_NAME}'...${RESET}"

# Deploy from local source
gcloud infra-manager deployments apply "${DEPLOYMENT_NAME}" \
    --location="${LOCATION}" \
    --service-account="projects/${PROJECT_ID}/serviceAccounts/${SERVICE_ACCOUNT}@${PROJECT_ID}.iam.gserviceaccount.com" \
    --local-source="." \
    --input-values="${INPUT_VALUES}"

EXIT_CODE=$?
if [ $EXIT_CODE -ne 0 ]; then
    echo ""
    echo -e "${RED}Deployment failed with exit code $EXIT_CODE${RESET}"
    echo ""
    echo "Common failure reasons:"
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

# Get the latest revision name from the deployment
echo -e "${GREEN}Retrieving service account key from Secret Manager...${RESET}"
REVISION=$(gcloud infra-manager deployments describe "${DEPLOYMENT_NAME}" \
    --location="${LOCATION}" \
    --format='value(latestRevision)')

if [ -z "$REVISION" ]; then
    echo -e "${RED}Error: Could not find deployment revision.${RESET}"
    exit 1
fi

# Extract the secret name from revision outputs (outputs are on revisions, not deployments)
SECRET_NAME=$(gcloud infra-manager revisions describe "${REVISION}" \
    --location="${LOCATION}" \
    --format='value(applyResults.outputs.secret_name.value)')

if [ -z "$SECRET_NAME" ]; then
    echo -e "${RED}Error: Secret name not found in revision outputs.${RESET}"
    exit 1
fi

# Access the secret from Secret Manager and save to file
gcloud secrets versions access latest --secret="${SECRET_NAME}" --project="${PROJECT_ID}" | base64 -d > KEY_FILE.json

echo ""
echo -e "${GREEN}Deployment complete.${RESET}"
gcloud infra-manager deployments describe "${DEPLOYMENT_NAME}" --location="${LOCATION}" --format='table(resources)'

echo ""
echo -e "${GREEN}Run 'cat KEY_FILE.json' to view the service account key."
echo -e "Copy and paste it in the Elastic Agent GCP integration."
echo -e "Save the key securely for future use.${RESET}"

