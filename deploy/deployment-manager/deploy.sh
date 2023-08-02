#!/bin/bash

DEPLOYMENT_NAME=${DEPLOYMENT_NAME:-elastic-agent-cspm}
ALLOW_SSH=${ALLOW_SSH:-false}
ZONE=${ZONE:-us-central1-a}
ELASTIC_ARTIFACT_SERVER=${ELASTIC_ARTIFACT_SERVER:-https://artifacts.elastic.co/downloads/beats/elastic-agent}

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
    --template deploy/deployment-manager/compute_engine.py \
    --properties elasticAgentVersion:${STACK_VERSION},fleetUrl:${FLEET_URL},enrollmentToken:${ENROLLMENT_TOKEN},allowSSH:${ALLOW_SSH},zone:${ZONE},elasticArtifactServer:${ELASTIC_ARTIFACT_SERVER}"

# Remove the roles required to deploy the DM templates
run_command "gcloud projects remove-iam-policy-binding ${PROJECT_NAME} --member=serviceAccount:${PROJECT_NUMBER}@cloudservices.gserviceaccount.com --role=roles/iam.roleAdmin"
run_command "gcloud projects remove-iam-policy-binding ${PROJECT_NAME} --member=serviceAccount:${PROJECT_NUMBER}@cloudservices.gserviceaccount.com --role=roles/resourcemanager.projectIamAdmin"
