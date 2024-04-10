#!/bin/bash

# Deploying a service account for the CIS GCP integration using Deployment Manager

## Script Overview
# The script you are about to execute automates the creation of a service account with a builtin roles that holds the necessary permissions for running the CIS GCP integration.
# The following steps are involved in the script execution:

### Step 1: Enabling Required APIs
# Before deploying the CIS GCP integration, certain APIs must be enabled. These APIs include:
# 1. IAM: Required for creating a service account and a custom role to be attached to the compute instance.
# 2. Deployment Manager: Necessary for creating the deployment itself.
# 3. Cloud Resource Manager: Enables querying information about the project being analyzed.
# 4. Cloud Asset Inventory: Facilitates data collection about assets linked to the project.

### Step 2: Applying the Deployment Manager Templates
# Upon executing the script, the following resources will be created:
# 1. service account with bindings for roles/cloudasset.viewer" and "roles/browser"
# 2. service account key

### Step 3: Saving the Service Account Key
# After the script execution is complete, the user is prompted to copy the service account key and paste it in the CIS GCP integration.

# In case the deployment encounters any issues and fails, the script will attempt to delete the deployment along with all the associated resources that were created during the process.

DEPLOYMENT_LABELS=${DEPLOYMENT_LABELS:-type=cspm-gcp}
DEPLOYMENT_NAME=${DEPLOYMENT_NAME:-elastic-agent-cspm-user}

# Set environment variables with the name and number of your project.
PROJECT_NAME="$(gcloud config get-value core/project)"
PROJECT_NUMBER="$(gcloud projects list --filter="${PROJECT_NAME}" --format="value(PROJECT_NUMBER)")"

export PROJECT_NAME
export PROJECT_NUMBER

# Function to run a gcloud command and check its exit code
run_command() {
    eval "$1"
    local status=$?
    if [ "$status" -ne 0 ]; then
        echo "Error: Command \"$1\" failed with exit code $status. Exiting..."
        exit "$status"
    fi
}

configure_scope() {
    if [ -n "$ORG_ID" ]; then
        SCOPE="organizations"
        PARENT_ID="$ORG_ID"
    fi

    # If ORG_ID is not set, SCOPE defaults to "projects" and PARENT_ID defaults to PROJECT_NAME
    SCOPE=${SCOPE:-"projects"}
    PARENT_ID=${PARENT_ID:-"$PROJECT_NAME"}
}

configure_scope

# Enable the Google Cloud APIs needed for misconfiguration scanning
run_command "gcloud services enable \
    iam.googleapis.com \
    deploymentmanager.googleapis.com \
    cloudresourcemanager.googleapis.com \
    cloudasset.googleapis.com"

# Apply the deployment manager templates
run_command "gcloud deployment-manager deployments create --automatic-rollback-on-error ${DEPLOYMENT_NAME} --project ${PROJECT_NAME} \
    --template service_account.py \
    --properties scope:${SCOPE},parentId:${PARENT_ID}"

run_command "gcloud iam service-accounts keys create KEY_FILE \
    --iam-account=${DEPLOYMENT_NAME}-sa@${PROJECT_NAME}.iam.gserviceaccount.com"

run_command "gcloud deployment-manager deployments describe ${DEPLOYMENT_NAME}"

echo -e "\nDeployment completed successfully, the service account has been created. Copy the key below and paste it in the CIS GCP integration:"
cat KEY_FILE # Print the key to the console
echo -e "\n\nIt is also recommended to save the key in a secure location for future use."
