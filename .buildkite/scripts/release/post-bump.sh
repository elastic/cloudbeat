#!/usr/bin/env bash
set -euo pipefail

# Post-PR-merge operations for a patch version bump.
# Runs after the bump PR is merged and DRA artifacts for NEW_VERSION are confirmed.
#
# Required env vars from PSI trigger params:
#   BRANCH      — release branch (e.g. 9.3)
#   NEW_VERSION — the version that was bumped to (e.g. 9.3.4)
#   REPO        — repository name (e.g. elastic/cloudbeat)
#   WORKFLOW    — must be "patch"
#
# Secrets (from kv/ci-shared/cloudbeat/* via vault — wired in task-7):
#   AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY  — for CFT upload to S3
#   SNYK_ORG_ID, SNYK_API_KEY, SNYK_INTEGRATION_ID — for Snyk branch monitoring
: "${BRANCH:?BRANCH is required}"
: "${NEW_VERSION:?NEW_VERSION is required}"
: "${REPO:?REPO is required}"
: "${WORKFLOW:?WORKFLOW is required}"

ARM_SINGLE_ACCOUNT_FILE="deploy/azure/ARM-for-single-account.json"
ARM_SINGLE_ACCOUNT_FILE_DEV="deploy/azure/ARM-for-single-account.dev.json"
ARM_ORGANIZATION_ACCOUNT_FILE="deploy/azure/ARM-for-organization-account.json"
ARM_ORGANIZATION_ACCOUNT_FILE_DEV="deploy/azure/ARM-for-organization-account.dev.json"

echo "--- Post-bump ops for ${NEW_VERSION}"

update_arm_templates() {
    echo "--- Update ARM templates to ${NEW_VERSION}"
    jq --indent 4 ".parameters.ElasticAgentVersion.defaultValue = \"${NEW_VERSION}\"" "${ARM_SINGLE_ACCOUNT_FILE}" >tmp.json && mv tmp.json "${ARM_SINGLE_ACCOUNT_FILE}"
    jq --indent 4 ".parameters.ElasticAgentVersion.defaultValue = \"${NEW_VERSION}\"" "${ARM_ORGANIZATION_ACCOUNT_FILE}" >tmp.json && mv tmp.json "${ARM_ORGANIZATION_ACCOUNT_FILE}"
    ./deploy/azure/generate_dev_template.py --template-type single-account
    ./deploy/azure/generate_dev_template.py --template-type organization-account
}

upload_cloud_formation_templates() {
    echo "--- Upload CloudFormation templates for ${NEW_VERSION}"
    aws configure set aws_access_key_id "${AWS_ACCESS_KEY_ID}"
    aws configure set aws_secret_access_key "${AWS_SECRET_ACCESS_KEY}"
    aws configure set region us-east-2
    scripts/publish_cft.sh
}

bump_snyk_branch_monitoring() {
    echo "--- Update Snyk branch monitoring"
    local branches latest_major previous_major latest_major_latest_minor previous_major_latest_minor

    branches=$(git branch -r | grep -Eo '[0-9]+\.[0-9]+' | sort -V | uniq)
    latest_major=$(echo "${branches}" | cut -d. -f1 | uniq | tail -1)
    previous_major=$(echo "${branches}" | cut -d. -f1 | uniq | tail -2 | head -1)
    latest_major_latest_minor=$(echo "${branches}" | grep -E "^${latest_major}\." | tail -1)
    previous_major_latest_minor=$(echo "${branches}" | grep -E "^${previous_major}\." | tail -1)

    echo "  latest:   ${latest_major_latest_minor}"
    echo "  previous: ${previous_major_latest_minor}"

    local cloudbeat_id
    cloudbeat_id=$(curl -sf -X GET \
        "https://api.snyk.io/rest/orgs/${SNYK_ORG_ID}/targets?version=2024-05-23&display_name=cloudbeat" \
        -H "accept: application/vnd.api+json" \
        -H "authorization: ${SNYK_API_KEY}" | jq -r '.data[0].id')

    curl -sf -X DELETE \
        "https://api.snyk.io/rest/orgs/${SNYK_ORG_ID}/targets/${cloudbeat_id}?version=2024-05-23" \
        -H "accept: application/vnd.api+json" \
        -H "authorization: ${SNYK_API_KEY}"

    for branch in main "${latest_major_latest_minor}" "${previous_major_latest_minor}"; do
        curl -sf -X POST \
            "https://api.snyk.io/v1/org/${SNYK_ORG_ID}/integrations/${SNYK_INTEGRATION_ID}/import" \
            -H "Content-Type: application/json; charset=utf-8" \
            -H "Authorization: token ${SNYK_API_KEY}" \
            -d "{\"target\":{\"owner\":\"elastic\",\"name\":\"cloudbeat\",\"branch\":\"${branch}\"},\"exclusionGlobs\":\"deploy, scripts, tests, security-policies\"}"
    done
}

update_arm_templates
upload_cloud_formation_templates
bump_snyk_branch_monitoring
