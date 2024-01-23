#!/bin/bash
set -euo pipefail

export MANIFEST_PATH="packages/cloud_security_posture/manifest.yml"
export CHANGELOG_PATH="packages/cloud_security_posture/changelog.yml"
export INTEGRATION_REPO="orouz/integrations" # TODO: change to elastic/integrations
export BRANCH="bump-to-$NEXT_CLOUDBEAT_VERSION"
export MAJOR_MINOR_CLOUDBEAT=$(echo "$NEXT_CLOUDBEAT_VERSION" | cut -d. -f1,2)

git config --global user.email "cloudsecmachine@users.noreply.github.com"
git config --global user.name "Cloud Security Machine"

checkout_integration_repo() {
    gh auth setup-git
    gh repo clone $INTEGRATION_REPO
    cd integrations
    git checkout -b "$BRANCH" main
}

update_manifest_version_vars() {
    # cis_gcp
    sed -i'' -E "s/cloudshell_git_branch=[0-9]+\.[0-9]+/cloudshell_git_branch=$MAJOR_MINOR_CLOUDBEAT/g" $MANIFEST_PATH

    # cis_aws + vuln_mgmt_aws
    sed -i'' -E "s/cloudformation-cnvm-[0-9]+\.[0-9]+\.[0-9]+/cloudformation-cnvm-$NEXT_CLOUDBEAT_VERSION/g" $MANIFEST_PATH
    sed -i'' -E "s/cloudformation-cspm-ACCOUNT_TYPE-[0-9]+\.[0-9]+\.[0-9]+/cloudformation-cspm-ACCOUNT_TYPE-$NEXT_CLOUDBEAT_VERSION/g" $MANIFEST_PATH

    # cis_azure
    sed -i'' -E "s/cloudbeat%2F[0-9]+\.[0-9]+/cloudbeat%2F$MAJOR_MINOR_CLOUDBEAT/g" $MANIFEST_PATH

    git add $MANIFEST_PATH
    git commit -m "Update manifest template vars"
    git push origin $BRANCH
}

create_integrations_pr() {
  echo 'Creating a PR to update integration'

  export PR_URL="$(gh pr create --title "[Cloud Security] Bump integration" \
  --body "Bumps integration to new version (Automated PR)" \
  --base "main" \
  --head "$BRANCH" \
  --repo "$INTEGRATION_REPO")"
}

update_manifest_version() {
    yq -i ".version = \"$NEXT_INTEGRATION_VERSION\"" $MANIFEST_PATH
    git add $MANIFEST_PATH
    git commit -m "Update manifest version"
    git push origin $BRANCH
}

update_changelog_version() {
    yq -i ".[0].version = \"$NEXT_INTEGRATION_VERSION\"" $CHANGELOG_PATH
    # this line below requires single quotes and env(PR) to interpolate this env var
    yq -i '.[0].changes += [{"description": "Bump version", "type": "enhancement", "link": env(PR_URL) }]' $CHANGELOG_PATH
    git add $CHANGELOG_PATH
    git commit -m "Update changelog version"
    git push origin $BRANCH
}

update_changelog_version_map() {
    # extract current major.minor version from changelog
    input_line=$(sed -n '3p' $CHANGELOG_PATH) # last version is always on line 3
    first_version=$(echo $input_line | cut -d' ' -f2)
    major_minor=$(echo $first_version | cut -d'.' -f1-2)
    major=$(echo $major_minor | cut -d'.' -f1)
    minor=$(echo $major_minor | cut -d'.' -f2)
    next_minor=$((minor + 1))

    # write new version map
    new_comment="# ${next_minor}.x - ${MAJOR_MINOR_CLOUDBEAT}.x"
    file_content=$(<"$CHANGELOG_PATH")
    new_file_content=$(awk -v var="$new_comment" 'NR==3 {print var} {print}' "$CHANGELOG_PATH")
    echo -e "$new_file_content" > temp.yaml
    mv temp.yaml "$CHANGELOG_PATH"
    git add $CHANGELOG_PATH
    git commit -m "Update changelog version map"
    git push origin $BRANCH
}

checkout_integration_repo
update_manifest_version_vars
create_integrations_pr
update_manifest_version
update_changelog_version
update_changelog_version_map