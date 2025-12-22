#!/bin/bash

# Set deployment configuration (accepts values from environment or uses defaults)
export DEPLOYMENT_NAME="${DEPLOYMENT_NAME:-infra-manager-amir-test}"
export REGION="${REGION:-us-central1}"
export GIT_REPO="${GIT_REPO:-https://github.com/amirbenun/cloudbeat.git}"
export GIT_REF="${GIT_REF:-infra-manager-agent}"
export GIT_SOURCE_DIR="${GIT_SOURCE_DIR:-deploy/infrastructure-manager/gcp-test}"

# Configure GCP project
export PROJECT_ID=$(gcloud config get-value core/project)

echo "Deploying Infrastructure Manager test from Git..."
echo "Project: ${PROJECT_ID}"
echo "Deployment: ${DEPLOYMENT_NAME}"
echo "Region: ${REGION}"
echo "Git Repo: ${GIT_REPO}"
echo "Git Ref: ${GIT_REF}"
echo "Source Dir: ${GIT_SOURCE_DIR}"

# Deploy from git source
gcloud infra-manager deployments apply ${DEPLOYMENT_NAME} \
    --location=${REGION} \
    --service-account="projects/${PROJECT_ID}/serviceAccounts/infra-manager-deployer@${PROJECT_ID}.iam.gserviceaccount.com" \
    --git-source-repo="${GIT_REPO}" \
    --git-source-ref="${GIT_REF}" \
    --git-source-directory="${GIT_SOURCE_DIR}" \
    --input-values="\
project_id=${PROJECT_ID},\
deployment_name=${DEPLOYMENT_NAME},\
region=${REGION}"
