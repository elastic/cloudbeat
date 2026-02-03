#!/bin/bash
set -e

# Accept parameters
PROJECT_ID="$1"
SERVICE_ACCOUNT="$2"
ORGANIZATION_ID="$3" # Optional: required for organization-scope deployments
SERVICE_ACCOUNT_EMAIL="${SERVICE_ACCOUNT}@${PROJECT_ID}.iam.gserviceaccount.com"

REQUIRED_APIS=(
    iam.googleapis.com
    config.googleapis.com
    cloudresourcemanager.googleapis.com
    cloudasset.googleapis.com
    secretmanager.googleapis.com
)

REQUIRED_ROLES=(
    roles/iam.serviceAccountAdmin
    roles/iam.serviceAccountKeyAdmin
    roles/resourcemanager.projectIamAdmin
    roles/config.admin
    roles/storage.admin
    roles/secretmanager.admin
)

ORG_LEVEL_ROLES=(
    roles/iam.securityAdmin
)

echo "Setting up GCP Infrastructure Manager prerequisites..."

# Enable APIs
gcloud services enable "${REQUIRED_APIS[@]}" --quiet

# Create service account if it doesn't exist
if ! gcloud iam service-accounts describe "${SERVICE_ACCOUNT_EMAIL}" >/dev/null 2>&1; then
    gcloud iam service-accounts create "${SERVICE_ACCOUNT}" \
        --display-name="Infra Manager Deployment Account" --quiet
fi

# Grant project-level permissions
for role in "${REQUIRED_ROLES[@]}"; do
    gcloud projects add-iam-policy-binding "${PROJECT_ID}" \
        --member="serviceAccount:${SERVICE_ACCOUNT_EMAIL}" \
        --role="${role}" --condition=None --quiet >/dev/null
done

# Grant organization-level permissions if ORGANIZATION_ID is provided
if [ -n "${ORGANIZATION_ID}" ]; then
    echo "Granting organization-level permissions for org ${ORGANIZATION_ID}..."
    for role in "${ORG_LEVEL_ROLES[@]}"; do
        gcloud organizations add-iam-policy-binding "${ORGANIZATION_ID}" \
            --member="serviceAccount:${SERVICE_ACCOUNT_EMAIL}" \
            --role="${role}" --condition=None --quiet >/dev/null
    done
fi

echo "✓ Setup complete"
