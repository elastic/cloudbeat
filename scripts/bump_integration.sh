#!/bin/bash
set -euo pipefail

MANIFEST_PATH="packages/cloud_security_posture/manifest.yml"
INTEGRATION_REPO="orouz/integrations"
BRANCH="bump-to-$NEXT_CLOUDBEAT_VERSION"

yq --version

checkout_integration_repo() {
    gh auth setup-git
    gh repo clone $INTEGRATION_REPO
    cd integrations
    git config --global user.email "elasticmachine@users.noreply.github.com"
    git config --global user.name "Elastic Machine"
}

update_manifest_version_vars() {
    git checkout -b "$BRANCH" main

    MINOR_VERSION=$(echo $NEXT_CLOUDBEAT_VERSION | cut -d '.' -f1,2)
    echo "MINOR_VERSION is $MINOR_VERSION"

    PATCH_VERSION=$NEXT_CLOUDBEAT_VERSION
    echo "PATCH_VERSION is $PATCH_VERSION"

    # cis_gcp
    sed -i'' -E "s/cloudshell_git_branch=[0-9]+\.[0-9]+/cloudshell_git_branch=$MINOR_VERSION/g" $MANIFEST_PATH

    # cis_aws + vuln_mgmt_aws
    sed -i'' -E "s/cloudformation-cnvm-[0-9]+\.[0-9]+\.[0-9]+/cloudformation-cnvm-$PATCH_VERSION/g" $MANIFEST_PATH
    sed -i'' -E "s/cloudformation-cspm-ACCOUNT_TYPE-[0-9]+\.[0-9]+\.[0-9]+/cloudformation-cspm-ACCOUNT_TYPE-$PATCH_VERSION/g" $MANIFEST_PATH

    # cis_azure
    sed -i'' -E "s/cloudbeat%2F[0-9]+\.[0-9]+/cloudbeat%2F$MINOR_VERSION/g" $MANIFEST_PATH

    git add $MANIFEST_PATH
    git commit -m "Update manifest template vars"
    git push origin $BRANCH
}

create_integrations_pr() {
  echo 'Creating a PR to update integration'

  PR_URL="$(gh pr create --title "[Cloud Security] Update integration manifest" \
  --body "Automated PR" \
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

update_changelog() {
    export PR=$PR_URL
    local CHANGELOG_PATH="packages/cloud_security_posture/changelog.yml"\
    # TODO: replace the existing preview version?
    yq -i ".[0].version = \"$NEXT_INTEGRATION_VERSION\"" $CHANGELOG_PATH
    # this line below requires single quotes and strenv(PR) to interpolate this env var
    yq -i '.[0].changes += [{"description": "Bump version", "type": "enhancement", "link": env(PR) }]' $CHANGELOG_PATH
    git add $CHANGELOG_PATH
    git commit -m "Update changelog version"
    git push origin $BRANCH
}

checkout_integration_repo
update_manifest_version_vars
create_integrations_pr
update_manifest_version
update_changelog
