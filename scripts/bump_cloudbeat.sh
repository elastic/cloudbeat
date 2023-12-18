#!/bin/bash
set -euo pipefail

export NEXT_MINOR_VERSION=$(echo $NEXT_CLOUDBEAT_VERSION | cut -d '.' -f1,2)
export CURRENT_MINOR_VERSION=$(echo $CURRENT_CLOUDBEAT_VERSION | cut -d '.' -f1,2)

echo "NEXT_CLOUDBEAT_VERSION: $NEXT_CLOUDBEAT_VERSION"
echo "NEXT_MINOR_VERSION: $NEXT_MINOR_VERSION"
echo "CURRENT_CLOUDBEAT_VERSION: $CURRENT_CLOUDBEAT_VERSION"
echo "CURRENT_MINOR_VERSION: $CURRENT_MINOR_VERSION"

create_release_branch() {
    if git show-ref --quiet refs/heads/$CURRENT_MINOR_VERSION; then
      echo "release branch '$CURRENT_MINOR_VERSION' already exists"
    else 
      echo "Create and push a new release branch $CURRENT_MINOR_VERSION from main"
      git checkout -b "$CURRENT_MINOR_VERSION" main
      git push origin "$CURRENT_MINOR_VERSION"
    fi
}

update_version_mergify() {
    echo "Update .mergify.yml with new version"
    cat << EOF >> .mergify.yml
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
}

update_version_arm_template() {
    echo "Update ARM template with new version"
    local single_account_file="deploy/azure/ARM-for-single-account.json"
    local organization_account_file="deploy/azure/ARM-for-organization-account.json"
    jq --indent 4 ".parameters.ElasticAgentVersion.defaultValue = \"$NEXT_CLOUDBEAT_VERSION\"" $single_account_file > tmp.json && mv tmp.json $single_account_file
    jq --indent 4 ".parameters.ElasticAgentVersion.defaultValue = \"$NEXT_CLOUDBEAT_VERSION\"" $organization_account_file > tmp.json && mv tmp.json $organization_account_file
}

update_version_beat() {
    echo "Update version/version.go with new version"
    sed -i'' -E "s/const defaultBeatVersion = .*/const defaultBeatVersion = \"$NEXT_CLOUDBEAT_VERSION\"/g" version/version.go
}

create_cloudbeat_pr() {
    echo "Add changes"
    local BRANCH="bump-to-$NEXT_CLOUDBEAT_VERSION"

    # TODO use obtained service account user
    git config --global user.email "elasticmachine@users.noreply.github.com"
    git config --global user.name "Elastic Machine"

    git checkout -b "$BRANCH" main
    git add .
    git commit -m "Bump cloudbeat to $NEXT_CLOUDBEAT_VERSION"
    git push origin "$BRANCH"

    echo "Create PR to bump cloudbeat version"
    gh pr create --title "Bump cloudbeat version" \
             --body "Automated PR" \
             --base "main" \
             --head "$BRANCH"
}

bump_cloudbeat() {
    update_version_mergify
    update_version_arm_template
    update_version_beat
}

create_release_branch
git checkout main 
bump_cloudbeat
create_cloudbeat_pr

