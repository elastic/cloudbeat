#!/bin/bash
set -xeuo pipefail

export INTEGRATION_REPO="elastic/integrations"
export BRANCH="bump-to-$CURRENT_CLOUDBEAT_VERSION"
MAJOR_MINOR_CLOUDBEAT=$(echo "$CURRENT_CLOUDBEAT_VERSION" | cut -d. -f1,2)

export MAJOR_MINOR_CLOUDBEAT

checkout_integration_repo() {
    echo "Checkout integration repo"
    gh auth setup-git
    gh repo clone $INTEGRATION_REPO
    cd integrations

    # clear branch if it exists
    if git ls-remote --exit-code --heads origin "$BRANCH"; then
        echo "Delete $BRANCH"
        git push origin --delete "$BRANCH"
        git branch -D "$BRANCH" 2>/dev/null || true
    else
        echo "$BRANCH does not exist"
    fi

    git checkout -b "$BRANCH" origin/main
}

get_next_integration_version() {
    integration_name="${1}"
    changelog_path="packages/$integration_name/changelog.yml"

    current_version=$(yq '.[0].version' "$changelog_path" | tr -d '"')
    preview_number="${current_version##*-preview}"
    preview_number="${preview_number##*(0)}"
    ((next_preview_number = preview_number + 1))
    next_preview_number_formatted=$(printf "%02d" "$next_preview_number")
    NEXT_INTEGRATION_VERSION="${current_version%-*}-preview${next_preview_number_formatted}"
    echo "Next integration version for $integration_name: $NEXT_INTEGRATION_VERSION"
    export NEXT_INTEGRATION_VERSION
}

update_manifest_version_vars() {
    integration_name="${1}"
    manifest_path="packages/$integration_name/manifest.yml"

    # cis_gcp
    echo "Update cloudshell_git_branch in manifest.yml"
    sed -i'' -E "s/cloudshell_git_branch=[0-9]+\.[0-9]+/cloudshell_git_branch=$MAJOR_MINOR_CLOUDBEAT/g" "$manifest_path"

    # cis_aws + vuln_mgmt_aws
    echo "Update cloudformation-* in manifest.yml"
    sed -i'' -E "s/cloudformation-cnvm-[0-9]+\.[0-9]+\.[0-9]+/cloudformation-cnvm-$CURRENT_CLOUDBEAT_VERSION/g" "$manifest_path"
    sed -i'' -E "s/cloudformation-cspm-ACCOUNT_TYPE-[0-9]+\.[0-9]+\.[0-9]+/cloudformation-cspm-ACCOUNT_TYPE-$CURRENT_CLOUDBEAT_VERSION/g" "$manifest_path"
    sed -i'' -E "s/cloudformation-cspm-direct-access-key-ACCOUNT_TYPE-[0-9]+\.[0-9]+\.[0-9]+/cloudformation-cspm-direct-access-key-ACCOUNT_TYPE-$CURRENT_CLOUDBEAT_VERSION/g" "$manifest_path"

    # cis_azure
    echo "Update cloudshell_git_branch in manifest.yml"
    sed -i'' -E "s/cloudbeat%2F[0-9]+\.[0-9]+/cloudbeat%2F$MAJOR_MINOR_CLOUDBEAT/g" "$manifest_path"

    git add "$manifest_path"
    if git diff --cached --quiet; then
        echo "No changes to commit in $manifest_path"
    else
        git commit -m "Update manifest template vars"
        git push origin "$BRANCH"
    fi
}

create_integrations_pr() {
    cat <<EOF >pr_body
Bump integration version - \`$NEXT_INTEGRATION_VERSION\`

EOF

    echo 'Create a PR to update integration'
    PR_URL="$(gh pr create --title "[Cloud Security] Bump integration" \
        --body-file pr_body \
        --base "main" \
        --head "$BRANCH" \
        --label "enhancement" \
        --label "Team:Cloud Security" \
        --repo "$INTEGRATION_REPO")"
    # shellcheck disable=SC2086
    echo "[Integrations PR]($PR_URL)" >>$GITHUB_STEP_SUMMARY
    export PR_URL
}

update_manifest_version() {
    integration_name="${1}"
    manifest_path="packages/$integration_name/manifest.yml"

    echo "Update manifest version for $integration_name"
    yq -i ".version = \"$NEXT_INTEGRATION_VERSION\"" "$manifest_path"
    git add "$manifest_path"
    if git diff --cached --quiet; then
        echo "No changes to commit in $manifest_path"
    else
        git commit -m "Update manifest version"
        git push origin "$BRANCH"
    fi
}

update_changelog_version() {
    integration_name="${1}"
    changelog_path="packages/$integration_name/changelog.yml"

    echo "Update changelog version for $integration_name"
    yq -i ".[0].version = \"$NEXT_INTEGRATION_VERSION\"" "$changelog_path"
    # PR_URL needs to be exported
    yq -i '.[0].changes += [{"description": "Bump version", "type": "enhancement", "link": env(PR_URL) }]' "$changelog_path"
    git add "$changelog_path"
    if git diff --cached --quiet; then
        echo "No changes to commit in $changelog_path"
    else
        git commit -m "Update changelog version"
        git push origin "$BRANCH"
    fi
}

checkout_integration_repo

get_next_integration_version 'cloud_security_posture'
update_manifest_version 'cloud_security_posture'
update_manifest_version_vars 'cloud_security_posture'

get_next_integration_version 'cloud_asset_inventory'
update_manifest_version 'cloud_asset_inventory'
update_manifest_version_vars 'cloud_asset_inventory'

if git diff origin/main..HEAD --quiet; then
    echo "No commits to push to $BRANCH"
else
    create_integrations_pr
    update_changelog_version 'cloud_security_posture'
    update_changelog_version 'cloud_asset_inventory'
fi
