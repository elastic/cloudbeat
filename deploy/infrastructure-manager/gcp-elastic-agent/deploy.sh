#!/bin/bash
set -e

# Configure GCP project
PROJECT_ID=$(gcloud config get-value core/project)
SERVICE_ACCOUNT="infra-manager-deployer"

# Ensure prerequisites are configured
"$(dirname "$0")/setup.sh" "${PROJECT_ID}" "${SERVICE_ACCOUNT}"

# Set deployment configuration (accepts values from environment or uses defaults)
export ORG_ID="${ORG_ID:-}"                                    # Optional: Set to your organization ID for org-level monitoring
export ZONE="${ZONE:-us-central1-a}"                           # Optional: Set to your desired GCP zone (default: us-central1-a)
export DEPLOYMENT_NAME="${DEPLOYMENT_NAME:-elastic-agent-gcp}" # Optional: Set to your desired deployment name (default: elastic-agent-gcp)
export FLEET_URL="${FLEET_URL:-}"
export ENROLLMENT_TOKEN="${ENROLLMENT_TOKEN:-}"
export STACK_VERSION="${STACK_VERSION:-}"
export ELASTIC_ARTIFACT_SERVER="${ELASTIC_ARTIFACT_SERVER%/}" # Remove trailing slash if present
export ELASTIC_ARTIFACT_SERVER="${ELASTIC_ARTIFACT_SERVER:-https://artifacts.elastic.co/downloads/beats/elastic-agent}"

export PROJECT_ID
LOCATION="${ZONE%-?}" # Extract region from zone
export LOCATION

# Generate unique suffix for resource names (8 hex characters)
RESOURCE_SUFFIX=$(openssl rand -hex 4)
export RESOURCE_SUFFIX
export DEPLOYMENT_NAME="${DEPLOYMENT_NAME}-${RESOURCE_SUFFIX}"

# Set scope and parent_id based on ORG_ID
if [ -n "${ORG_ID}" ]; then
    export SCOPE="organizations"
    export PARENT_ID="${ORG_ID}"
else
    export SCOPE="projects"
    export PARENT_ID="${PROJECT_ID}"
fi

# Deploy from local source (repo already cloned by Cloud Shell)
echo "Starting deployment ${DEPLOYMENT_NAME}..."
gcloud infra-manager deployments apply "${DEPLOYMENT_NAME}" \
    --location="${LOCATION}" \
    --service-account="projects/${PROJECT_ID}/serviceAccounts/${SERVICE_ACCOUNT}@${PROJECT_ID}.iam.gserviceaccount.com" \
    --local-source="." \
    --input-values="\
project_id=${PROJECT_ID},\
zone=${ZONE},\
fleet_url=${FLEET_URL},\
enrollment_token=${ENROLLMENT_TOKEN},\
elastic_agent_version=${STACK_VERSION},\
elastic_artifact_server=${ELASTIC_ARTIFACT_SERVER},\
scope=${SCOPE},\
parent_id=${PARENT_ID},\
resource_suffix=${RESOURCE_SUFFIX}"

EXIT_CODE=$?
if [ $EXIT_CODE -ne 0 ]; then
    echo ""
    echo "Deployment failed with exit code $EXIT_CODE"
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
    echo "  gcloud compute instances get-serial-port-output elastic-agent-vm-${RESOURCE_SUFFIX} --zone=${ZONE} --project=${PROJECT_ID}"
    echo ""
    exit $EXIT_CODE
fi

echo ""
echo "Deployment successful!"
