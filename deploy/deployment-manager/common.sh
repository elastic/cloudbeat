#!/usr/bin/env bash
set -e

configure_scope() {
    if [ -n "$ORG_ID" ]; then
        SCOPE="organizations"
        PARENT_ID="$ORG_ID"
        ROLE="roles/resourcemanager.organizationAdmin"
    fi

    # If ORG_ID is not set, SCOPE defaults to "projects" and PARENT_ID defaults to PROJECT_NAME
    SCOPE=${SCOPE:-"projects"}
    PARENT_ID=${PARENT_ID:-"$PROJECT_NAME"}
}

# Function to check if a role is assigned to the service account
is_role_not_assigned() {
    local role_assigned
    role_assigned=$(gcloud "${SCOPE}" get-iam-policy "${PARENT_ID}" \
        --flatten="bindings[].members" --format="value(bindings.members)" \
        --filter="bindings.role=${ROLE}" \
        --format="table[no-heading](bindings.members)" |
        grep "${PROJECT_NUMBER}@cloudservices.gserviceaccount.com")

    if [ -n "${role_assigned}" ]; then
        return 1 # Role is assigned
    else
        return 0 # Role is not assigned
    fi
}
