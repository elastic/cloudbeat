#!/bin/bash
set -euo pipefail

# versions
CURRENT_MINOR_VERSION=$(echo "$CURRENT_CLOUDBEAT_VERSION" | cut -d '.' -f1,2)
export CURRENT_MINOR_VERSION

# branches
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

update_version_mergify() {
    echo "• Add a new entry to .mergify.yml"
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
    git commit -m "Update .mergify.yml"

    gh label create "backport-v$CURRENT_CLOUDBEAT_VERSION"
}

update_version_arm_template_default_value() {
    echo "• Replace defaultValue for ElasticAgentVersion in ARM templates"
    jq --indent 4 ".parameters.ElasticAgentVersion.defaultValue = \"$NEXT_CLOUDBEAT_VERSION\"" $ARM_SINGLE_ACCOUNT_FILE >tmp.json && mv tmp.json $ARM_SINGLE_ACCOUNT_FILE
    jq --indent 4 ".parameters.ElasticAgentVersion.defaultValue = \"$NEXT_CLOUDBEAT_VERSION\"" $ARM_ORGANIZATION_ACCOUNT_FILE >tmp.json && mv tmp.json $ARM_ORGANIZATION_ACCOUNT_FILE

    echo "• Generate dev ARM templates"
    ./deploy/azure/generate_dev_template.py --template-type single-account
    ./deploy/azure/generate_dev_template.py --template-type organization-account

    git add $ARM_SINGLE_ACCOUNT_FILE $ARM_ORGANIZATION_ACCOUNT_FILE $ARM_SINGLE_ACCOUNT_FILE_DEV $ARM_ORGANIZATION_ACCOUNT_FILE_DEV
    git commit -m "Update ARM templates"
}

update_version_arm_template_file_uris() {
    echo "• Replace fileUris git branch in ARM templates"
    sed -i'' -E "s/cloudbeat\/main/cloudbeat\/$CURRENT_MINOR_VERSION/g" $ARM_SINGLE_ACCOUNT_FILE
    sed -i'' -E "s/cloudbeat\/main/cloudbeat\/$CURRENT_MINOR_VERSION/g" $ARM_ORGANIZATION_ACCOUNT_FILE
    git add $ARM_SINGLE_ACCOUNT_FILE $ARM_ORGANIZATION_ACCOUNT_FILE
    git commit -m "Update ARM templates"
}

update_version_beat() {
    echo "• Update version/version.go with new version"
    sed -i'' -E "s/const defaultBeatVersion = .*/const defaultBeatVersion = \"$NEXT_CLOUDBEAT_VERSION\"/g" version/version.go
    git add version/version.go
    git commit -m "Update version.go"
}

create_cloudbeat_versions_pr_for_main() {
    echo "• Create PR for cloudbeat next version"
    git push origin "$NEXT_CLOUDBEAT_BRANCH"
    cat <<EOF >cloudbeat_pr_body
Bump cloudbeat version - \`$NEXT_CLOUDBEAT_VERSION\`

> [!NOTE]
> This is an automated PR
EOF
    pr_url="$(gh pr create --title "Bump cloudbeat version" \
        --body-file cloudbeat_pr_body \
        --base "main" \
        --head "$NEXT_CLOUDBEAT_BRANCH" \
        --label "backport-skip")"
    # shellcheck disable=SC2086
    echo "[Cloudbeat Version PR to main]($pr_url)" >>$GITHUB_STEP_SUMMARY
}

create_cloudbeat_versions_pr_for_release() {
    echo "• Create PR for cloudbeat release version"
    git push origin "$RELEASE_CLOUDBEAT_BRANCH"
    cat <<EOF >cloudbeat_pr_body_release
Release cloudbeat version - \`$CURRENT_CLOUDBEAT_VERSION\`

> [!NOTE]
> This is an automated PR
EOF
    pr_url="$(gh pr create --title "Release cloudbeat version" \
        --body-file cloudbeat_pr_body_release \
        --base "$CURRENT_MINOR_VERSION" \
        --head "$RELEASE_CLOUDBEAT_BRANCH" \
        --label "backport-skip")"
    # shellcheck disable=SC2086
    echo "[Cloudbeat Version PR to release branch]($pr_url)" >>$GITHUB_STEP_SUMMARY
}

# We need to bump hermit seperately because we need to wait for the snapshot build to be available
bump_hermit() {
    echo "• Bump hermit cloudbeat version"
    sed -i'' -E "s/\"CLOUDBEAT_VERSION\": .*/\"CLOUDBEAT_VERSION\": \"$CURRENT_CLOUDBEAT_VERSION\",/g" $HERMIT_FILE
    git add $HERMIT_FILE
    git commit -m "Bump cloudbeat to $CURRENT_CLOUDBEAT_VERSION"
    git push origin "$NEXT_CLOUDBEAT_HERMIT_BRANCH"

    cat <<EOF >hermit_pr_body
Bump cloudbeat version - \`$CURRENT_CLOUDBEAT_VERSION\`

> [!IMPORTANT]
> to be merged after snapshot build for $CURRENT_CLOUDBEAT_VERSION is available

> [!NOTE]
> This is an automated PR
EOF

    echo "• Create a PR for cloudbeat hermit version"
    pr_url="$(gh pr create --title "Bump hermit cloudbeat version" \
        --body-file hermit_pr_body \
        --base "main" \
        --head "$NEXT_CLOUDBEAT_HERMIT_BRANCH" \
        --label "backport-skip")"
    # shellcheck disable=SC2086
    echo "[Cloudbeat Hermit PR]($pr_url)" >>$GITHUB_STEP_SUMMARY
}

upload_cloud_formation_templates() {
    echo "• Upload cloud formation templates for $CURRENT_CLOUDBEAT_VERSION"
    aws configure set aws_access_key_id "$AWS_ACCESS_KEY_ID"
    aws configure set aws_secret_access_key "$AWS_SECRET_ACCESS_KEY"
    aws configure set region us-east-2
    scripts/publish_cft.sh
}

# make changes to 'main' for next version
run_version_changes_for_main() {
    # create a new branch from the main branch
    git fetch origin main
    git checkout -b "$NEXT_CLOUDBEAT_BRANCH" origin/main

    # commit
    update_version_beat
    update_version_mergify
    update_version_arm_template_default_value

    # push
    create_cloudbeat_versions_pr_for_main

    # create, commit and push a separate PR for hermit
    git checkout -b "$NEXT_CLOUDBEAT_HERMIT_BRANCH" origin/main
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
    create_cloudbeat_versions_pr_for_release

    # upload cloud formation templates for the release version
    upload_cloud_formation_templates
}

run_version_changes_for_main
run_version_changes_for_release_branch
