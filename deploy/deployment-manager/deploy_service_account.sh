#!/usr/bin/env bash
set -e

# this script:
# 1. enables necessary APIs for CSPM GCP integration
# 2. applies a Deployment Manager template to create a service account with role and a key
# 3. saves generated key to a file (KEY_FILE.json) and prompts the user to copy-paste it to the GCP integration
# 4. handles deployment failure cleanup

DEPLOYMENT_NAME=${DEPLOYMENT_NAME:-elastic-agent-cspm-user}
SERVICE_ACCOUNT_NAME=${SERVICE_ACCOUNT_NAME:-elastic-agent-cspm-user-sa}
PROJECT_NAME="$(gcloud config get-value core/project)"
PROJECT_NUMBER="$(gcloud projects list --filter="${PROJECT_NAME}" --format="value(PROJECT_NUMBER)")"
ROLE="roles/resourcemanager.projectIamAdmin"

export PROJECT_NAME
export PROJECT_NUMBER

source ./common.sh

configure_scope

# Enable the Google Cloud APIs needed for misconfiguration scanning
gcloud services enable \
    iam.googleapis.com \
    deploymentmanager.googleapis.com \
    cloudresourcemanager.googleapis.com \
    cloudasset.googleapis.com

ADD_ROLE=false
if is_role_not_assigned; then
    gcloud "${SCOPE}" add-iam-policy-binding "${PARENT_ID}" --member="serviceAccount:${PROJECT_NUMBER}@cloudservices.gserviceaccount.com" --role="${ROLE}"
    ADD_ROLE=true
fi

result="$(gcloud deployment-manager deployments create --automatic-rollback-on-error "${DEPLOYMENT_NAME}" --project "${PROJECT_NAME}" \
    --template service_account.py \
    --properties "scope:'${SCOPE}',parentId:'${PARENT_ID}',serviceAccountName:'${SERVICE_ACCOUNT_NAME}'")"

key="$(echo "$result" | awk '/serviceAccountKey/{getline; print}' | awk '{print $2}')"

RED='\033[0;31m'
GREEN='\033[0;32m'
RESET='\033[0m'

if [ -z "$key" ]; then
    echo -e "${RED}Error: Failed to deploy a service account. Exiting...${RESET}"
    exit 1
fi

if [ "$ADD_ROLE" = "true" ]; then
    gcloud "${SCOPE}" remove-iam-policy-binding "${PARENT_ID}" --member=serviceAccount:"${PROJECT_NUMBER}"@cloudservices.gserviceaccount.com --role="${ROLE}"
fi

echo -e "\n${GREEN}Deployment complete.${RESET}\n"

gcloud deployment-manager deployments describe "${DEPLOYMENT_NAME}" --format='table(resources)'

echo "$key" | base64 -d >KEY_FILE.json

echo -e "\n${GREEN}Run 'cat KEY_FILE.json' to view the service account key. Copy and paste it in the CSPM GCP integration."
echo -e "\nYou should also save the key in a secure location for future use.${RESET}"
