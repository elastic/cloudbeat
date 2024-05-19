#!/bin/bash
set -euo pipefail

if [ "$#" -lt 3 ]; then
    echo "Missing params. Usage: $0 PROJECT_NAME PROJECT_NUMBER DEPLOYMENT1 DEPLOYMENT2,..."
    exit 1
fi

DELETED_DEPLOYMENTS=()
FAILED_DEPLOYMENTS=()
PROJECT_NAME=$1
PROJECT_NUMBER=$2
shift 2
GCP_DEPLOYMENTS=("$@")

# Add the needed roles to delete the templates to the project using the deployment manager
gcloud projects add-iam-policy-binding "${PROJECT_NAME}" --member=serviceAccount:"${PROJECT_NUMBER}"@cloudservices.gserviceaccount.com --role=roles/iam.roleAdmin --no-user-output-enabled
gcloud projects add-iam-policy-binding "${PROJECT_NAME}" --member=serviceAccount:"${PROJECT_NUMBER}"@cloudservices.gserviceaccount.com --role=roles/resourcemanager.projectIamAdmin --no-user-output-enabled

DELETED_DEPLOYMENTS=()
FAILED_DEPLOYMENTS=()

for DEPLOYMENT in "${GCP_DEPLOYMENTS[@]}"; do
    echo "Deleting GCP deployment: $DEPLOYMENT"
    if gcloud deployment-manager deployments delete "$DEPLOYMENT" -q; then
        echo "Successfully deleted GCP deployment: $DEPLOYMENT"
        DELETED_DEPLOYMENTS+=("$DEPLOYMENT")
    else
        echo "Failed to delete GCP deployment: $DEPLOYMENT"
        FAILED_DEPLOYMENTS+=("$DEPLOYMENT")
    fi

done

# Remove the roles required to deploy the DM templates
gcloud projects remove-iam-policy-binding "${PROJECT_NAME}" --member=serviceAccount:"${PROJECT_NUMBER}"@cloudservices.gserviceaccount.com --role=roles/iam.roleAdmin --no-user-output-enabled
gcloud projects remove-iam-policy-binding "${PROJECT_NAME}" --member=serviceAccount:"${PROJECT_NUMBER}"@cloudservices.gserviceaccount.com --role=roles/resourcemanager.projectIamAdmin --no-user-output-enabled

echo "Successfully deleted GCP deployments (${#DELETED_DEPLOYMENTS[@]}):"
# Print summary of gcp deployments deletions
if [ ${#DELETED_DEPLOYMENTS[@]} -gt 0 ]; then
    printf "%s\n" "${DELETED_DEPLOYMENTS[@]}"
fi

echo "Failed to delete GCP deployments (${#FAILED_DEPLOYMENTS[@]}):"
if [ ${#FAILED_DEPLOYMENTS[@]} -gt 0 ]; then
    printf "%s\n" "${FAILED_DEPLOYMENTS[@]}"
    exit 1
fi
