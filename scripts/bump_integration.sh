#!/bin/bash
set -euo pipefail

export MANIFEST_PATH="packages/cloud_security_posture/manifest.yml"
export CHANGELOG_PATH="packages/cloud_security_posture/changelog.yml"
export INTEGRATION_REPO="elastic/integrations"
export BRANCH="bump-to-$NEXT_CLOUDBEAT_VERSION"
MAJOR_MINOR_CLOUDBEAT=$(echo "$NEXT_CLOUDBEAT_VERSION" | cut -d. -f1,2)

export MAJOR_MINOR_CLOUDBEAT

checkout_integration_repo() {
    echo "• Checkout integration repo"
    gh auth setup-git
    gh repo clone $INTEGRATION_REPO
    cd integrations
    git checkout -b "$BRANCH" origin/main
}

# reads the last version from changelog.yml version map
# and increments the minor version
get_next_integration_version() {
    echo "• Get next integration version"
    input_line=$(sed -n '3p' $CHANGELOG_PATH) # last version is always on line 3
    first_version=$(echo "$input_line" | cut -d' ' -f2)
    major_minor=$(echo "$first_version" | cut -d'.' -f1-2)
    major=$(echo "$major_minor" | cut -d'.' -f1)
    minor=$(echo "$major_minor" | cut -d'.' -f2)
    next_minor=$((minor + 1))
    export NEXT_INTEGRATION_VERSION="$major.$next_minor.0"
    echo "NEXT_INTEGRATION_VERSION: $NEXT_INTEGRATION_VERSION"
}

update_manifest_version_vars() {
    # cis_gcp
    echo "• Update cloudshell_git_branch in manifest.yml"
    sed -i'' -E "s/cloudshell_git_branch=[0-9]+\.[0-9]+/cloudshell_git_branch=$MAJOR_MINOR_CLOUDBEAT/g" $MANIFEST_PATH

    # cis_aws + vuln_mgmt_aws
    echo "• Update cloudformation-* in manifest.yml"
    sed -i'' -E "s/cloudformation-cnvm-[0-9]+\.[0-9]+\.[0-9]+/cloudformation-cnvm-$NEXT_CLOUDBEAT_VERSION/g" $MANIFEST_PATH
    sed -i'' -E "s/cloudformation-cspm-ACCOUNT_TYPE-[0-9]+\.[0-9]+\.[0-9]+/cloudformation-cspm-ACCOUNT_TYPE-$NEXT_CLOUDBEAT_VERSION/g" $MANIFEST_PATH

    # cis_azure
    echo "• Update cloudshell_git_branch in manifest.yml"
    sed -i'' -E "s/cloudbeat%2F[0-9]+\.[0-9]+/cloudbeat%2F$MAJOR_MINOR_CLOUDBEAT/g" $MANIFEST_PATH

    git add $MANIFEST_PATH
    git commit -m "Update manifest template vars"
    git push origin "$BRANCH"
}

create_integrations_pr() {
    cat <<EOF >pr_body
Bump integration version - \`$NEXT_INTEGRATION_VERSION\`

> [!NOTE]
> This is an automated PR
EOF

    echo '• Create a PR to update integration'
    PR_URL="$(gh pr create --title "[Cloud Security] Bump integration" \
        --body-file pr_body \
        --base "main" \
        --head "$BRANCH" \
        --label "enhancement" \
        --label "Team:Cloud Security" \
        --repo "$INTEGRATION_REPO")"
    echo "$PR_URL"
}

update_manifest_version() {
    echo "• Update manifest version"
    yq -i ".version = \"$NEXT_INTEGRATION_VERSION\"" $MANIFEST_PATH
    git add $MANIFEST_PATH
    git commit -m "Update manifest version"
    git push origin "$BRANCH"
}

update_changelog_version() {
    local PR_URL="$1"
    echo "• Update changelog version"
    yq -i ".[0].version = \"$NEXT_INTEGRATION_VERSION\"" $CHANGELOG_PATH
    # this line below requires single quotes and env(PR) to interpolate this env var
    yq -i '.[0].changes += [{"description": "Bump version", "type": "enhancement", "link": env(PR_URL) }]' $CHANGELOG_PATH
    git add $CHANGELOG_PATH
    git commit -m "Update changelog version"
    git push origin "$BRANCH"
}

update_changelog_version_map() {
    echo "• Update changelog version map"
    next_minor=$(echo "$NEXT_INTEGRATION_VERSION" | cut -d'.' -f1,2)
    new_comment="# ${next_minor}.x - ${MAJOR_MINOR_CLOUDBEAT}.x"
    new_file_content=$(awk -v var="$new_comment" 'NR==3 {print var} {print}' "$CHANGELOG_PATH")
    echo -e "$new_file_content" >temp.yaml
    mv temp.yaml "$CHANGELOG_PATH"
    git add $CHANGELOG_PATH
    git commit -m "Update changelog version map"
    git push origin "$BRANCH"
}

checkout_integration_repo
get_next_integration_version
update_manifest_version_vars
update_manifest_version
update_changelog_version "$(create_integrations_pr)"
update_changelog_version_map
