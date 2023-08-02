#!/bin/bash

# Set environment variables with the name and number of your project.
export PROJECT_NAME=$(gcloud config get-value core/project)
export PROJECT_NUMBER=$(gcloud projects list --filter=${PROJECT_NAME} --format="value(PROJECT_NUMBER)")

# Enable the Google Cloud APIs needed for misconfiguration scanning
gcloud services enable iam.googleapis.com deploymentmanager.googleapis.com compute.googleapis.com cloudresourcemanager.googleapis.com cloudasset.googleapis.com

# Add the needed roles to apply the templates to the project using the deployment manager
gcloud projects add-iam-policy-binding ${PROJECT_NAME} --member=serviceAccount:${PROJECT_NUMBER}@cloudservices.gserviceaccount.com --role=roles/iam.roleAdmin
gcloud projects add-iam-policy-binding ${PROJECT_NAME} --member=serviceAccount:${PROJECT_NUMBER}@cloudservices.gserviceaccount.com --role=roles/resourcemanager.projectIamAdmin

# Apply the deployment manager templates
gcloud deployment-manager deployments create elastic-agent-cspm --project ${PROJECT_NAME} \
    --template deploy/deployment-manager/compute_engine.py --properties elasticAgentVersion:${STACK_VERSION},fleetUrl:${FLEET_URL},enrollmentToken:${ENROLLMENT_TOKEN}

# Remove the roles required to deploy the DM templates
gcloud projects remove-iam-policy-binding ${PROJECT_NAME} --member=serviceAccount:${PROJECT_NUMBER}@cloudservices.gserviceaccount.com --role=roles/iam.roleAdmin
gcloud projects remove-iam-policy-binding ${PROJECT_NAME} --member=serviceAccount:${PROJECT_NUMBER}@cloudservices.gserviceaccount.com --role=roles/resourcemanager.projectIamAdmin
