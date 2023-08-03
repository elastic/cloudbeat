#!/bin/bash

# Deploying the CIS GCP integration using Deployment Manager

## Script Overview
# The script you are about to execute automates the creation of a new deployment, a compute instance, and the attachment of a service account with a custom role that holds the necessary permissions for running the CIS GCP integration.
# The following steps are involved in the script execution:

#### Step 1: Enabling Required APIs

# Before deploying the CIS GCP integration, certain APIs must be enabled. These APIs include:
#1. IAM: Required for creating a service account and a custom role to be attached to the compute instance.
#2. Deployment Manager: Necessary for creating the deployment itself.
#3. Compute: Used to create a new compute instance to host the Elastic agent.
#4. Cloud Resource Manager: Enables querying information about the project being analyzed.
#5. Cloud Asset Inventory: Facilitates data collection about assets linked to the project.

#### Step 2: Adding Roles to the Default Service Account
#The template used in this deployment creates a service account, custom IAM role, and service account bindings.
#However, prior to the successful deployment, the Google APIs Service Agent service account must be granted the `Role Administrator` and `Project IAM Admin` roles. For more information on this account, you can refer to the [Google-managed service account documentation](https://cloud.google.com/iam/docs/maintain-custom-roles-deployment-manager).
#Once the deployment is completed, the script will remove the roles from the default service account. If you wish to delete the deployment, you will need to manually add the roles back to the default service account.

### Step 3: Applying the Deployment Manager Templates
#Upon executing the script, the following resources will be created:

#1. A compute instance with the Elastic agent installed.
#2. A service account with a custom role attached.
#3. A role with appropriate permissions required for the integration.
#4. A service account binding that associates the custom role with the compute instance.

#In case the deployment encounters any issues and fails, the script will attempt to delete the deployment along with all the associated resources that were created during the process.


DEPLOYMENT_NAME=${DEPLOYMENT_NAME:-elastic-agent-cspm}
ALLOW_SSH=${ALLOW_SSH:-false}
ZONE=${ZONE:-us-central1-a}
ELASTIC_ARTIFACT_SERVER=${ELASTIC_ARTIFACT_SERVER:-https://artifacts.elastic.co/downloads/beats/elastic-agent}

# Function to check if an environment variable is not provided
check_env_not_provided() {
  local var_name="$1"

  if [ -z "${!var_name}" ]; then
    echo "Error: $var_name not provided. Please set the environment variable $var_name."
    exit 1
  fi
}

# Fail fast if any of the required environment variables are not provided
check_env_not_provided "FLEET_URL"
check_env_not_provided "ENROLLMENT_TOKEN"
check_env_not_provided "STACK_VERSION"

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
export PROJECT_NUMBER=$(gcloud projects list --filter=${PROJECT_NAME} --format="value(PROJECT_NUMBER)")

# Enable the Google Cloud APIs needed for misconfiguration scanning
run_command "gcloud services enable iam.googleapis.com deploymentmanager.googleapis.com compute.googleapis.com cloudresourcemanager.googleapis.com cloudasset.googleapis.com"

# Add the needed roles to apply the templates to the project using the deployment manager
run_command "gcloud projects add-iam-policy-binding ${PROJECT_NAME} --member=serviceAccount:${PROJECT_NUMBER}@cloudservices.gserviceaccount.com --role=roles/iam.roleAdmin"
run_command "gcloud projects add-iam-policy-binding ${PROJECT_NAME} --member=serviceAccount:${PROJECT_NUMBER}@cloudservices.gserviceaccount.com --role=roles/resourcemanager.projectIamAdmin"

# Apply the deployment manager templates
run_command "gcloud deployment-manager deployments create --automatic-rollback-on-error ${DEPLOYMENT_NAME} --project ${PROJECT_NAME} \
    --template compute_engine.py \
    --properties elasticAgentVersion:${STACK_VERSION},fleetUrl:${FLEET_URL},enrollmentToken:${ENROLLMENT_TOKEN},allowSSH:${ALLOW_SSH},zone:${ZONE},elasticArtifactServer:${ELASTIC_ARTIFACT_SERVER}"

# Remove the roles required to deploy the DM templates
run_command "gcloud projects remove-iam-policy-binding ${PROJECT_NAME} --member=serviceAccount:${PROJECT_NUMBER}@cloudservices.gserviceaccount.com --role=roles/iam.roleAdmin"
run_command "gcloud projects remove-iam-policy-binding ${PROJECT_NAME} --member=serviceAccount:${PROJECT_NUMBER}@cloudservices.gserviceaccount.com --role=roles/resourcemanager.projectIamAdmin"
