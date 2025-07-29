#!/usr/bin/env bash
set -e

# This script:
# 1. Enables necessary APIs for CSPM GCP integration
# 2. Applies a Deployment Manager template to create a service account with roles and key
# 3. Saves generated key to KEY_FILE.json
# 4. Handles retry for deployment and failure cleanup

DEPLOYMENT_NAME=${DEPLOYMENT_NAME:-elastic-agent-cspm-user}
SERVICE_ACCOUNT_NAME=${SERVICE_ACCOUNT_NAME:-elastic-agent-cspm-user-sa}
PROJECT_NAME="$(gcloud config get-value core/project)"
PROJECT_NUMBER="$(gcloud projects list --filter="${PROJECT_NAME}" --format="value(PROJECT_NUMBER)")"
ROLE="roles/resourcemanager.projectIamAdmin"
MAX_RETRIES=3
DELAY=5

export PROJECT_NAME
export PROJECT_NUMBER

source ./common.sh

configure_scope

# Enable required APIs
gcloud services enable \
    iam.googleapis.com \
    deploymentmanager.googleapis.com \
    cloudresourcemanager.googleapis.com \
    cloudasset.googleapis.com

ADD_ROLE=false
if is_role_not_assigned; then
    gcloud "${SCOPE}" add-iam-policy-binding "${PARENT_ID}" \
        --member="serviceAccount:${PROJECT_NUMBER}@cloudservices.gserviceaccount.com" \
        --role="${ROLE}"
    ADD_ROLE=true
fi

RED='\033[0;31m'
GREEN='\033[0;32m'
RESET='\033[0m'

echo -e "${GREEN}Creating deployment '${DEPLOYMENT_NAME}'...${RESET}"

attempt=1
while true; do
    set +e
    result="$(gcloud deployment-manager deployments create --automatic-rollback-on-error "${DEPLOYMENT_NAME}" --project "${PROJECT_NAME}" \
        --template service_account.py \
        --properties "scope:'${SCOPE}',parentId:'${PARENT_ID}',serviceAccountName:'${SERVICE_ACCOUNT_NAME}'" 2>&1)"
    status=$?
    set -e

    if [ $status -eq 0 ]; then
        echo -e "${GREEN}Deployment succeeded on attempt ${attempt}.${RESET}"
        break
    else
        echo -e "${RED}Attempt ${attempt} failed: ${result}${RESET}"
        if [[ $attempt -ge $MAX_RETRIES ]]; then
            echo -e "${RED}Max retries reached. Deployment failed.${RESET}"
            exit 1
        fi
        sleep_time=$((DELAY * attempt))
        echo -e "${GREEN}Retrying in ${sleep_time} seconds...${RESET}"
        sleep $sleep_time
        attempt=$((attempt + 1))
    fi
done

# Wait for deployment to become available and return outputs
describe_attempt=1
while true; do
    key="$(gcloud deployment-manager deployments describe "${DEPLOYMENT_NAME}" \
        --project="${PROJECT_NAME}" \
        --format="value(outputs[0].finalValue)" 2>/dev/null)"

    if [ -n "$key" ]; then
        break
    fi

    if [[ $describe_attempt -ge $MAX_RETRIES ]]; then
        echo -e "${RED}Error: Service account key not found in deployment outputs after ${describe_attempt} attempts. Exiting...${RESET}"
        exit 1
    fi

    echo -e "${GREEN}Waiting for deployment outputs to be available... attempt ${describe_attempt}${RESET}"
    sleep 5
    describe_attempt=$((describe_attempt + 1))
done

# Remove temporary role if it was added
if [ "$ADD_ROLE" = "true" ]; then
    gcloud "${SCOPE}" remove-iam-policy-binding "${PARENT_ID}" \
        --member="serviceAccount:${PROJECT_NUMBER}@cloudservices.gserviceaccount.com" \
        --role="${ROLE}"
fi

# Save decoded key to file
echo "$key" | base64 -d >KEY_FILE.json

echo -e "\n${GREEN}Deployment complete.${RESET}"
gcloud deployment-manager deployments describe "${DEPLOYMENT_NAME}" --format='table(resources)'

echo -e "\n${GREEN}Run 'cat KEY_FILE.json' to view the service account key. Copy and paste it in the CSPM GCP integration."
echo -e "Save the key securely for future use.${RESET}"
