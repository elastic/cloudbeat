#!/bin/bash
set -e

# Accept parameters
PROJECT_ID="$1"
SERVICE_ACCOUNT="$2"
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
    roles/iam.serviceAccountUser
    roles/resourcemanager.projectIamAdmin
    roles/config.admin
    roles/storage.admin
)

echo "Setting up GCP Infrastructure Manager prerequisites..."

# Enable APIs
gcloud services enable "${REQUIRED_APIS[@]}" --quiet

# Create service account if it doesn't exist
if ! gcloud iam service-accounts describe "${SERVICE_ACCOUNT_EMAIL}" >/dev/null 2>&1; then
    gcloud iam service-accounts create "${SERVICE_ACCOUNT}" \
        --display-name="Infra Manager Deployment Account" --quiet
fi

# Grant permissions
for role in "${REQUIRED_ROLES[@]}"; do
    gcloud projects add-iam-policy-binding "${PROJECT_ID}" \
        --member="serviceAccount:${SERVICE_ACCOUNT_EMAIL}" \
        --role="${role}" --condition=None --quiet >/dev/null
done

echo "✓ Setup complete"
