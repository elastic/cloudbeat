#!/bin/bash
set -e

# Get the directory where this script lives (for Terraform source files)
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"

# Configure GCP project
PROJECT_ID=$(gcloud config get-value core/project)
SERVICE_ACCOUNT="infra-manager-deployer"

# Ensure prerequisites are configured
"${SCRIPT_DIR}/setup.sh" "${PROJECT_ID}" "${SERVICE_ACCOUNT}"

# Required environment variables (no defaults - must be provided)
# FLEET_URL, ENROLLMENT_TOKEN, STACK_VERSION

# Optional environment variables (defaults are in variables.tf)
# ORG_ID          - Set for org-level monitoring
# ZONE            - GCP zone (default: us-central1-a)
# DEPLOYMENT_NAME - Deployment name prefix (default: elastic-agent-deployment)
# ELASTIC_ARTIFACT_SERVER - Artifact server URL

# Generate unique suffix for resource names (8 hex characters)
RESOURCE_SUFFIX=$(openssl rand -hex 4)

# Set deployment name with suffix
DEPLOYMENT_NAME="${DEPLOYMENT_NAME:-elastic-agent-deployment}-${RESOURCE_SUFFIX}"

# Determine zone for location extraction
# We need the zone to derive the region (location) - use default if not set
EFFECTIVE_ZONE="${ZONE:-us-central1-a}"
LOCATION="${EFFECTIVE_ZONE%-?}" # Extract region from zone

# Build input values - only include values that are set
# Defaults are defined in variables.tf (single source of truth)
INPUT_VALUES="project_id=${PROJECT_ID}"
INPUT_VALUES="${INPUT_VALUES},resource_suffix=${RESOURCE_SUFFIX}"

# Required values
INPUT_VALUES="${INPUT_VALUES},fleet_url=${FLEET_URL}"
INPUT_VALUES="${INPUT_VALUES},enrollment_token=${ENROLLMENT_TOKEN}"
INPUT_VALUES="${INPUT_VALUES},elastic_agent_version=${STACK_VERSION}"

# Optional values - only add if explicitly set (let TF use its defaults otherwise)
if [ -n "${ZONE}" ]; then
    INPUT_VALUES="${INPUT_VALUES},zone=${ZONE}"
fi

if [ -n "${ELASTIC_ARTIFACT_SERVER}" ]; then
    # Remove trailing slash if present
    ELASTIC_ARTIFACT_SERVER="${ELASTIC_ARTIFACT_SERVER%/}"
    INPUT_VALUES="${INPUT_VALUES},elastic_artifact_server=${ELASTIC_ARTIFACT_SERVER}"
fi

# Set scope and parent_id based on ORG_ID
if [ -n "${ORG_ID}" ]; then
    INPUT_VALUES="${INPUT_VALUES},scope=organizations"
    INPUT_VALUES="${INPUT_VALUES},parent_id=${ORG_ID}"
else
    INPUT_VALUES="${INPUT_VALUES},scope=projects"
    INPUT_VALUES="${INPUT_VALUES},parent_id=${PROJECT_ID}"
fi

# Deploy from local source (repo already cloned by Cloud Shell)
echo "Starting deployment ${DEPLOYMENT_NAME}..."
if ! gcloud infra-manager deployments apply "${DEPLOYMENT_NAME}" \
    --location="${LOCATION}" \
    --service-account="projects/${PROJECT_ID}/serviceAccounts/${SERVICE_ACCOUNT}@${PROJECT_ID}.iam.gserviceaccount.com" \
    --local-source="${SCRIPT_DIR}" \
    --input-values="${INPUT_VALUES}"; then
    echo ""
    echo "Deployment failed"
    echo ""
    echo "Common failure reasons:"
    echo "  - Wrong artifacts server for pre-release artifact (check ELASTIC_ARTIFACT_SERVER for snapshots/pre-releases)"
    echo "  - Service account permissions missing for ${SERVICE_ACCOUNT}@${PROJECT_ID}.iam.gserviceaccount.com"
    echo "  - Invalid input values (fleet_url, enrollment_token, etc.)"
    echo ""
    echo "Useful debugging commands:"
    echo "  # View deployment status"
    echo "  gcloud infra-manager deployments describe ${DEPLOYMENT_NAME} --location=${LOCATION}"
    echo ""
    echo "  # Verify service account permissions"
    echo "  gcloud projects get-iam-policy ${PROJECT_ID} --flatten='bindings[].members' --filter='bindings.members:serviceAccount:${SERVICE_ACCOUNT}@${PROJECT_ID}.iam.gserviceaccount.com' --format='table(bindings.role)'"
    echo ""
    echo "  # View Cloud Build logs"
    echo "  gsutil cat \$(gcloud infra-manager revisions describe \$(gcloud infra-manager deployments describe ${DEPLOYMENT_NAME} --location=${LOCATION} --format='value(latestRevision)') --location=${LOCATION} --format='value(logs)')/*.txt"
    echo ""
    echo "  # View VM startup script logs"
    echo "  gcloud compute instances get-serial-port-output elastic-agent-vm-${RESOURCE_SUFFIX} --zone=${EFFECTIVE_ZONE} --project=${PROJECT_ID}"
    echo ""
    exit 1
fi

echo ""
echo "Deployment successful!"
