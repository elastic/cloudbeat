#!/bin/bash
set -euo pipefail

export NEXT_CLOUDBEAT_BRANCH="bump-to-$NEXT_CLOUDBEAT_VERSION"
export NEXT_MINOR_VERSION=$(echo $NEXT_CLOUDBEAT_VERSION | cut -d '.' -f1,2)
export CURRENT_MINOR_VERSION=$(echo $CURRENT_CLOUDBEAT_VERSION | cut -d '.' -f1,2)

echo "NEXT_CLOUDBEAT_VERSION: $NEXT_CLOUDBEAT_VERSION"
echo "NEXT_MINOR_VERSION: $NEXT_MINOR_VERSION"
echo "CURRENT_CLOUDBEAT_VERSION: $CURRENT_CLOUDBEAT_VERSION"
echo "CURRENT_MINOR_VERSION: $CURRENT_MINOR_VERSION"

git config --global user.email "cloudsecmachine@users.noreply.github.com"
git config --global user.name "Cloud Security Machine"

create_release_branch() {
  if git fetch origin "$CURRENT_MINOR_VERSION" 2>/dev/null; then
    echo "release branch '$CURRENT_MINOR_VERSION' already exists"
    git checkout "$CURRENT_MINOR_VERSION"
    git pull origin $CURRENT_MINOR_VERSION
  else
    echo "Create and push a new release branch $CURRENT_MINOR_VERSION from main"
    git checkout -b "$CURRENT_MINOR_VERSION" main
    git push origin $CURRENT_MINOR_VERSION
  fi
}

update_version_mergify() {
  echo "Update .mergify.yml with new version"
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
}

update_version_arm_template() {
  echo "Update ARM template with new version"
  local single_account_file="deploy/azure/ARM-for-single-account.json"
  local organization_account_file="deploy/azure/ARM-for-organization-account.json"
  jq --indent 4 ".parameters.ElasticAgentVersion.defaultValue = \"$NEXT_CLOUDBEAT_VERSION\"" $single_account_file >tmp.json && mv tmp.json $single_account_file
  jq --indent 4 ".parameters.ElasticAgentVersion.defaultValue = \"$NEXT_CLOUDBEAT_VERSION\"" $organization_account_file >tmp.json && mv tmp.json $organization_account_file

  sed -i'' -E "s/cloudbeat\/main/cloudbeat\/$NEXT_MINOR_VERSION/g" $single_account_file
  sed -i'' -E "s/cloudbeat\/main/cloudbeat\/$NEXT_MINOR_VERSION/g" $organization_account_file
}

update_version_beat() {
  echo "Update version/version.go with new version"
  sed -i'' -E "s/const defaultBeatVersion = .*/const defaultBeatVersion = \"$NEXT_CLOUDBEAT_VERSION\"/g" version/version.go
}

create_cloudbeat_versions_pr() {
  git add .
  git commit -m "Bump cloudbeat to $NEXT_CLOUDBEAT_VERSION"
  git push origin "$NEXT_CLOUDBEAT_BRANCH"
  gh pr create --title "Bump cloudbeat version" \
    --body "Bump cloudbeat to new version - $NEXT_CLOUDBEAT_VERSION (Automated PR)" \
    --base "main" \
    --head "$NEXT_CLOUDBEAT_BRANCH"
}

bump_cloudbeat() {
  git checkout -b "$NEXT_CLOUDBEAT_BRANCH" main
  update_version_mergify
  update_version_arm_template
  update_version_beat
  create_cloudbeat_versions_pr
}

bump_hermit() {
  local BRANCH="bump-hermit-to-$NEXT_CLOUDBEAT_VERSION"
  git checkout -b "$BRANCH" main
  sed -i'' -E "s/\"CLOUDBEAT_VERSION\": .*/\"CLOUDBEAT_VERSION\": \"$NEXT_CLOUDBEAT_VERSION\",/g" bin/hermit.hcl
  git add bin/hermit.hcl
  git commit -m "Bump cloudbeat to $NEXT_CLOUDBEAT_VERSION"
  git push origin "$BRANCH"
  gh pr create --title "Bump hermit cloudbeat version" \
    --body "to be merged after snapshot build for $NEXT_CLOUDBEAT_VERSION is available. (Automated PR)" \
    --base "main" \
    --head "$BRANCH"
}

upload_cloud_formation_templates() {
  aws configure set aws_access_key_id $AWS_ACCESS_KEY_ID
  aws configure set aws_secret_access_key $AWS_SECRET_ACCESS_KEY
  aws configure set region us-east-2
  scripts/publish_cft.sh
}

create_release_branch
bump_cloudbeat
bump_hermit
# upload_cloud_formation_templates
