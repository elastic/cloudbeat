#!/bin/bash
set -e

# Configuration
PROJECT_ID=$(gcloud config get-value core/project 2>/dev/null)
SERVICE_ACCOUNT="infra-manager-deployer"
SERVICE_ACCOUNT_EMAIL="${SERVICE_ACCOUNT}@${PROJECT_ID}.iam.gserviceaccount.com"

REQUIRED_APIS=(
    iam.googleapis.com
    config.googleapis.com
    compute.googleapis.com
    cloudresourcemanager.googleapis.com
    cloudasset.googleapis.com
)

REQUIRED_ROLES=(
    roles/compute.admin
    roles/iam.serviceAccountAdmin
    roles/resourcemanager.projectIamAdmin
    roles/config.admin
)

# Check if already configured
if gcloud iam service-accounts describe "${SERVICE_ACCOUNT_EMAIL}" >/dev/null 2>&1; then
    exit 0
fi

echo "Setting up GCP Infrastructure Manager prerequisites..."

# Enable APIs
gcloud services enable "${REQUIRED_APIS[@]}" --quiet

# Create service account
gcloud iam service-accounts create "${SERVICE_ACCOUNT}" \
    --display-name="Infra Manager Deployment Account" --quiet

# Grant permissions
for role in "${REQUIRED_ROLES[@]}"; do
    gcloud projects add-iam-policy-binding "${PROJECT_ID}" \
        --member="serviceAccount:${SERVICE_ACCOUNT_EMAIL}" \
        --role="${role}" --condition=None --quiet >/dev/null
done

echo "✓ Setup complete"
