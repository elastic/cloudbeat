#!/bin/bash

# Set deployment configuration (accepts values from environment or uses defaults)
export ORG_ID="${ORG_ID:-}"  # Optional: Set to your organization ID for org-level monitoring
export ZONE="${ZONE:-us-central1-a}"  # Optional: Set to your desired GCP zone (default: us-central1-a)
export DEPLOYMENT_NAME="${DEPLOYMENT_NAME:-elastic-agent-cspm}" # Optional: Set to your desired deployment name (default: elastic-agent-cspm)
export FLEET_URL="${FLEET_URL:-}"
export ENROLLMENT_TOKEN="${ENROLLMENT_TOKEN:-}"
export ELASTIC_AGENT_VERSION="${ELASTIC_AGENT_VERSION:-}"

# Configure GCP project and location
export PROJECT_ID=$(gcloud config get-value core/project)
export LOCATION=$(echo ${ZONE} | sed 's/-[a-z]$//')  # Extract region from zone

# Set scope and parent_id based on ORG_ID
if [ -n "${ORG_ID}" ]; then
  export SCOPE="organizations"
  export PARENT_ID="${ORG_ID}"
else
  export SCOPE="projects"
  export PARENT_ID="${PROJECT_ID}"
fi

# Deploy from local source (repo already cloned by Cloud Shell)
gcloud infra-manager deployments apply ${DEPLOYMENT_NAME} \
    --location=${LOCATION} \
    --service-account="projects/${PROJECT_ID}/serviceAccounts/infra-manager-deployer@${PROJECT_ID}.iam.gserviceaccount.com" \
    --local-source="." \
    --input-values="\
project_id=${PROJECT_ID},\
deployment_name=${DEPLOYMENT_NAME},\
zone=${ZONE},\
fleet_url=${FLEET_URL},\
enrollment_token=${ENROLLMENT_TOKEN},\
elastic_agent_version=${ELASTIC_AGENT_VERSION},\
scope=${SCOPE},\
parent_id=${PARENT_ID}"