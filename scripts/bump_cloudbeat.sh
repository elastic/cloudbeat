#!/bin/bash
set -xeuo pipefail

# versions
CURRENT_MINOR_VERSION=$(echo "$CURRENT_CLOUDBEAT_VERSION" | cut -d '.' -f1,2)
export CURRENT_MINOR_VERSION

# branches
export BASE_BRANCH="${GIT_BASE_BRANCH:-main}"
export NEXT_CLOUDBEAT_BRANCH="bump-to-$NEXT_CLOUDBEAT_VERSION"
export NEXT_CLOUDBEAT_HERMIT_BRANCH="bump-hermit-to-$CURRENT_CLOUDBEAT_VERSION"
export RELEASE_CLOUDBEAT_BRANCH="release-$CURRENT_MINOR_VERSION"

# paths
export ARM_SINGLE_ACCOUNT_FILE="deploy/azure/ARM-for-single-account.json"
export ARM_SINGLE_ACCOUNT_FILE_DEV="deploy/azure/ARM-for-single-account.dev.json"
export ARM_ORGANIZATION_ACCOUNT_FILE="deploy/azure/ARM-for-organization-account.json"
export ARM_ORGANIZATION_ACCOUNT_FILE_DEV="deploy/azure/ARM-for-organization-account.dev.json"
export HERMIT_FILE="bin/hermit.hcl"

echo "NEXT_CLOUDBEAT_VERSION: $NEXT_CLOUDBEAT_VERSION"
echo "CURRENT_CLOUDBEAT_VERSION: $CURRENT_CLOUDBEAT_VERSION"
echo "CURRENT_MINOR_VERSION: $CURRENT_MINOR_VERSION"

# clear branches if they exists
branches=("$NEXT_CLOUDBEAT_BRANCH" "$NEXT_CLOUDBEAT_HERMIT_BRANCH" "$RELEASE_CLOUDBEAT_BRANCH")
for branch in "${branches[@]}"; do
    if git ls-remote --exit-code --heads origin "$branch"; then
        git push origin --delete "$branch"
    fi
done

update_version_mergify() {
    echo "Add a new entry to .mergify.yml"
    cat <<EOF >>.mergify.yml
  - name: backport patches to $CURRENT_MINOR_VERSION branch
    conditions:
      - merged
      - label=backport-v$CURRENT_CLOUDBEAT_VERSION
    actions:
      backport:
        assignees:
          - "{{ author }}"
        branches:
          - "$CURRENT_MINOR_VERSION"
        labels:
          - "backport"
        title: "[{{ destination_branch }}](backport #{{ number }}) {{ title }}"
EOF
    git add .mergify.yml
    if git diff --cached --quiet; then
        echo "No changes to commit in .mergify.yml"
    else
        git commit -m "Update .mergify.yml"
        gh label create "backport-v$CURRENT_CLOUDBEAT_VERSION" --force
    fi
}

update_version_arm_template_default_value() {
    echo "Replace defaultValue for ElasticAgentVersion in ARM templates"
    jq --indent 4 ".parameters.ElasticAgentVersion.defaultValue = \"$NEXT_CLOUDBEAT_VERSION\"" $ARM_SINGLE_ACCOUNT_FILE >tmp.json && mv tmp.json $ARM_SINGLE_ACCOUNT_FILE
    jq --indent 4 ".parameters.ElasticAgentVersion.defaultValue = \"$NEXT_CLOUDBEAT_VERSION\"" $ARM_ORGANIZATION_ACCOUNT_FILE >tmp.json && mv tmp.json $ARM_ORGANIZATION_ACCOUNT_FILE

    echo "Generate dev ARM templates"
    ./deploy/azure/generate_dev_template.py --template-type single-account
    ./deploy/azure/generate_dev_template.py --template-type organization-account

    git add $ARM_SINGLE_ACCOUNT_FILE $ARM_ORGANIZATION_ACCOUNT_FILE $ARM_SINGLE_ACCOUNT_FILE_DEV $ARM_ORGANIZATION_ACCOUNT_FILE_DEV
    if git diff --cached --quiet; then
        echo "No changes to commit in ARM templates"
    else
        git commit -m "Update ARM templates"
    fi
}

update_version_arm_template_file_uris() {
    echo "Replace fileUris git branch in ARM templates"
    sed -i'' -E "s/cloudbeat\/$BASE_BRANCH/cloudbeat\/$CURRENT_MINOR_VERSION/g" $ARM_SINGLE_ACCOUNT_FILE
    sed -i'' -E "s/cloudbeat\/$BASE_BRANCH/cloudbeat\/$CURRENT_MINOR_VERSION/g" $ARM_ORGANIZATION_ACCOUNT_FILE
    git add $ARM_SINGLE_ACCOUNT_FILE $ARM_ORGANIZATION_ACCOUNT_FILE
    if git diff --cached --quiet; then
        echo "No changes to commit in ARM templates"
    else
        git commit -m "Update ARM templates"
    fi
}

update_version_beat() {
    echo "Update version/version.go with new version"
    sed -i'' -E "s/const defaultBeatVersion = .*/const defaultBeatVersion = \"$NEXT_CLOUDBEAT_VERSION\"/g" version/version.go
    git add version/version.go
    if git diff --cached --quiet; then
        echo "No changes to commit in version.go"
    else
        git commit -m "Update version.go"
    fi
}

create_cloudbeat_versions_pr_for_base_branch() {
    echo "Create PR for cloudbeat next version"
    git push origin "$NEXT_CLOUDBEAT_BRANCH"
    cat <<EOF >cloudbeat_pr_body
Bump cloudbeat version - \`$NEXT_CLOUDBEAT_VERSION\`

EOF

    pr_url="$(gh pr create --title "Bump cloudbeat version" \
        --body-file cloudbeat_pr_body \
        --base "$BASE_BRANCH" \
        --head "$NEXT_CLOUDBEAT_BRANCH" \
        --label "backport-skip")"
    # shellcheck disable=SC2086
    echo "[Cloudbeat Version PR to $BASE_BRANCH]($pr_url)" >>$GITHUB_STEP_SUMMARY
    rm cloudbeat_pr_body
}

create_cloudbeat_versions_pr_for_release() {
    echo "Create PR for cloudbeat release version"
    git push origin "$RELEASE_CLOUDBEAT_BRANCH"
    cat <<EOF >cloudbeat_pr_body_release
Release cloudbeat version - \`$CURRENT_CLOUDBEAT_VERSION\`

EOF
    pr_url="$(gh pr create --title "Release cloudbeat version" \
        --body-file cloudbeat_pr_body_release \
        --base "$CURRENT_MINOR_VERSION" \
        --head "$RELEASE_CLOUDBEAT_BRANCH" \
        --label "backport-skip")"
    # shellcheck disable=SC2086
    echo "[Cloudbeat Version PR to release branch]($pr_url)" >>$GITHUB_STEP_SUMMARY
    rm cloudbeat_pr_body_release
}

# We need to bump hermit seperately because we need to wait for the snapshot build to be available
bump_hermit() {
    echo "Bump hermit cloudbeat version"
    sed -i'' -E "s/\"CLOUDBEAT_VERSION\": .*/\"CLOUDBEAT_VERSION\": \"$CURRENT_CLOUDBEAT_VERSION\",/g" $HERMIT_FILE
    git add $HERMIT_FILE
    if git diff --cached --quiet; then
        echo "No changes to commit in $HERMIT_FILE"
    else
        git commit -m "Bump cloudbeat to $CURRENT_CLOUDBEAT_VERSION"
        git push origin "$NEXT_CLOUDBEAT_HERMIT_BRANCH"
        cat <<EOF >hermit_pr_body
Bump cloudbeat version - \`$CURRENT_CLOUDBEAT_VERSION\`

> [!IMPORTANT]
> to be merged after snapshot build for $CURRENT_CLOUDBEAT_VERSION is available

EOF

        echo "Create a PR for cloudbeat hermit version"
        pr_url="$(gh pr create --title "Bump hermit cloudbeat version" \
            --body-file hermit_pr_body \
            --base "$BASE_BRANCH" \
            --head "$NEXT_CLOUDBEAT_HERMIT_BRANCH" \
            --label "backport-skip")"
        # shellcheck disable=SC2086
        echo "[Cloudbeat Hermit PR]($pr_url)" >>$GITHUB_STEP_SUMMARY
        rm hermit_pr_body
    fi
}

upload_cloud_formation_templates() {
    set +x # disable debug log
    echo "Upload cloud formation templates for $CURRENT_CLOUDBEAT_VERSION"
    aws configure set aws_access_key_id "$AWS_ACCESS_KEY_ID"
    aws configure set aws_secret_access_key "$AWS_SECRET_ACCESS_KEY"
    aws configure set region us-east-2
    scripts/publish_cft.sh
    set -x # enable debug log
}

# make changes to '$BASE_BRANCH' for next version
run_version_changes_for_base_branch() {
    # create a new branch from the $BASE_BRANCH branch
    git fetch origin "$BASE_BRANCH"
    git checkout -b "$NEXT_CLOUDBEAT_BRANCH" "origin/$BASE_BRANCH"

    # commit
    update_version_beat
    update_version_mergify
    update_version_arm_template_default_value

    # push
    if git diff "origin/$BASE_BRANCH..HEAD" --quiet; then
        echo "No commits to push to $BASE_BRANCH $NEXT_CLOUDBEAT_BRANCH"
    else
        create_cloudbeat_versions_pr_for_base_branch
    fi

    # create, commit and push a separate PR for hermit
    git checkout -b "$NEXT_CLOUDBEAT_HERMIT_BRANCH" "origin/$BASE_BRANCH"
    bump_hermit
}

# make changes for 'release' version
run_version_changes_for_release_branch() {
    # create a new branch from the current minor version
    git fetch origin "$CURRENT_MINOR_VERSION"
    git checkout -b "$RELEASE_CLOUDBEAT_BRANCH" origin/"$CURRENT_MINOR_VERSION"

    # commit
    update_version_arm_template_file_uris

    # push
    if git diff "origin/$BASE_BRANCH..HEAD" --quiet; then
        echo "No commits to push to release $RELEASE_CLOUDBEAT_BRANCH"
    else
        create_cloudbeat_versions_pr_for_release
    fi

    # upload cloud formation templates for the release version
    upload_cloud_formation_templates
}

bump_snyk_branch_monitoring() {
    # Get cloudbeat target ID
    SNYK_CLOUDBEAT_ID=$(curl -X GET "https://api.snyk.io/rest/orgs/$SNYK_ORG_ID/targets?version=2024-05-23&display_name=cloudbeat" \
        -H "accept: application/vnd.api+json" \
        -H "authorization: $SNYK_API_KEY" | jq -r '.data[0].id')

    # Delete cloudbeat target
    curl -X DELETE "https://api.snyk.io/rest/orgs/$SNYK_ORG_ID/targets/$SNYK_CLOUDBEAT_ID?version=2024-05-23" \
        -H "accept: application/vnd.api+json" \
        -H "authorization: $SNYK_API_KEY"

    # Import cloudbeat/$BASE_BRANCH
    curl -X POST \
        "https://api.snyk.io/v1/org/$SNYK_ORG_ID/integrations/$SNYK_INTEGRATION_ID/import" \
        -H 'Content-Type: application/json; charset=utf-8' \
        -H "Authorization: token $SNYK_API_KEY" \
        -d "{
  \"target\": {
    \"owner\": \"elastic\",
    \"name\": \"cloudbeat\",
    \"branch\": \"$BASE_BRANCH\"
  },
  \"exclusionGlobs\": \"deploy, scripts, tests, security-policies\"
}"
    # Import cloudbeat/$CURRENT_MINOR_VERSION
    curl -X POST \
        "https://api.snyk.io/v1/org/$SNYK_ORG_ID/integrations/$SNYK_INTEGRATION_ID/import" \
        -H 'Content-Type: application/json; charset=utf-8' \
        -H "Authorization: token $SNYK_API_KEY" \
        -d "{
  \"target\": {
    \"owner\": \"elastic\",
    \"name\": \"cloudbeat\",
    \"branch\": \"$CURRENT_MINOR_VERSION\"
  },
  \"exclusionGlobs\": \"deploy, scripts, tests, security-policies\"
}"

}

validate_base_branch() {
    if [[ "$BASE_BRANCH" == "main" || "$BASE_BRANCH" == "8.x" || "$BASE_BRANCH" == "9.x" ]]; then
        echo "Allowed to bump version for $BASE_BRANCH"
        return
    fi

    if echo "$BASE_BRANCH" | grep -qE '^[89]\.[0-9]+\.[0-9]+$'; then
        echo "Allowed to bump version for $BASE_BRANCH"
        return
    fi
    echo "Not allowed to bump version for $BASE_BRANCH"
    exit 1
}

validate_base_branch
run_version_changes_for_base_branch
run_version_changes_for_release_branch
bump_snyk_branch_monitoring
