#!/bin/bash

# Deploying the CIS GCP integration using Deployment Manager

## Script Overview
# The script you are about to execute automates the creation of a new deployment, a compute instance, and the attachment of a service account with a builtin roles that holds the necessary permissions for running the CIS GCP integration.
# The following steps are involved in the script execution:

#### Step 1: Enabling Required APIs

# Before deploying the CIS GCP integration, certain APIs must be enabled. These APIs include:
#1. IAM: Required for creating a service account and a custom role to be attached to the compute instance.
#2. Deployment Manager: Necessary for creating the deployment itself.
#3. Compute: Used to create a new compute instance to host the Elastic agent.
#4. Cloud Resource Manager: Enables querying information about the project being analyzed.
#5. Cloud Asset Inventory: Facilitates data collection about assets linked to the project.

#### Step 2: Adding Roles to the Default Service Account
#The template used in this deployment creates a service account and service account bindings.
#However, prior to the successful deployment, the Google APIs Service Agent service account must be granted the `Project IAM Admin` role to be able attaching the service account to the compute instance.
#Once the deployment is completed, the script will remove the role from the default service account. If you wish to delete the deployment, you will need to manually add the role back to the default service account.

### Step 3: Applying the Deployment Manager Templates
#Upon executing the script, the following resources will be created:

#1. A compute instance with the Elastic agent installed.
#2. A service account.
#3. A dedicated network for the compute instance.
#4. A service account bindings that associates the builtin roles with the service account.

# In case the deployment encounters any issues and fails, the script will attempt to delete the deployment along with all the associated resources that were created during the process.

DEPLOYMENT_NAME=${DEPLOYMENT_NAME:-elastic-agent-cspm}
ALLOW_SSH=${ALLOW_SSH:-false}
ZONE=${ZONE:-us-central1-a}
ROLE="roles/resourcemanager.projectIamAdmin"

ELASTIC_ARTIFACT_SERVER=${ELASTIC_ARTIFACT_SERVER%/} # Remove trailing slash if present
ELASTIC_ARTIFACT_SERVER=${ELASTIC_ARTIFACT_SERVER:-https://artifacts.elastic.co/downloads/beats/elastic-agent}
SERVICE_ACCOUNT_NAME=${SERVICE_ACCOUNT_NAME:-false}

# Set environment variables with the name and number of your project.
export PROJECT_NAME=$(gcloud config get-value core/project)
export PROJECT_NUMBER=$(gcloud projects list --filter=${PROJECT_NAME} --format="value(PROJECT_NUMBER)")

source ./common.sh

# Function to check if an environment variable is not provided
check_env_not_provided() {
    local var_name="$1"

    if [ -z "${!var_name}" ]; then
        echo "Error: $var_name not provided. Please set the environment variable $var_name."
        exit 1
    fi
}

# Function to run a gcloud command and check its exit code
run_command() {
    eval $1
    local status=$?
    if [ $status -ne 0 ]; then
        echo "Error: Command \"$1\" failed with exit code $status. Exiting..."
        exit $status
    fi
}

# Set environment variables with the name and number of your project.
export PROJECT_NAME=$(gcloud config get-value core/project)
export PROJECT_NUMBER=$(gcloud projects list --filter="${PROJECT_NAME}" --format="value(PROJECT_NUMBER)")

# Fail fast if any of the required environment variables are not provided
check_env_not_provided "FLEET_URL"
check_env_not_provided "ENROLLMENT_TOKEN"
check_env_not_provided "STACK_VERSION"

configure_scope

# Enable the Google Cloud APIs needed for misconfiguration scanning
run_command "gcloud services enable iam.googleapis.com deploymentmanager.googleapis.com compute.googleapis.com cloudresourcemanager.googleapis.com cloudasset.googleapis.com"

## If needed add the required role to apply the templates to the project using the deployment manager
ADD_ROLE=false
if is_role_not_assigned; then
    run_command "gcloud ${SCOPE} add-iam-policy-binding ${PARENT_ID} --member=serviceAccount:${PROJECT_NUMBER}@cloudservices.gserviceaccount.com --role=${ROLE}"
    ADD_ROLE=true
fi

# Apply the deployment manager templates
run_command "gcloud deployment-manager deployments create --automatic-rollback-on-error ${DEPLOYMENT_NAME} --project ${PROJECT_NAME} \
    --template compute_engine.py \
    --properties elasticAgentVersion:${STACK_VERSION},fleetUrl:${FLEET_URL},enrollmentToken:${ENROLLMENT_TOKEN},allowSSH:${ALLOW_SSH},zone:${ZONE},elasticArtifactServer:${ELASTIC_ARTIFACT_SERVER},scope:${SCOPE},parentId:${PARENT_ID},serviceAccountName:${SERVICE_ACCOUNT_NAME}"

## Remove the role required to deploy the DM templates
if [ "$ADD_ROLE" = "true" ]; then
    run_command "gcloud ${SCOPE} remove-iam-policy-binding ${PARENT_ID} --member=serviceAccount:${PROJECT_NUMBER}@cloudservices.gserviceaccount.com --role=${ROLE}"
fi

run_command "gcloud deployment-manager deployments describe ${DEPLOYMENT_NAME}"
echo -e "\nDeployment completed successfully, elastic agent enrollment might take a few minutes."
