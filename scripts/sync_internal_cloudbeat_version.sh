#! /bin/bash
set -xeuo pipefail

VERSION_FILE="version/version.go"
HERMIT_FILE="bin/hermit.hcl"

git config --global user.email "cloudsecmachine@users.noreply.github.com"
git config --global user.name "Cloud Security Machine"

find_current_cloudbeat_version() {
    echo "Checking current cloudbeat version"
    CLOUDBEAT_VERSION=$(grep -oE 'defaultBeatVersion\s+=\s+".*"' $VERSION_FILE | grep -oE '[0-9]\.[0-9]\.[0-9]')
    echo "Cloudbeat version is $CLOUDBEAT_VERSION"
}

set_hermit_cloudbeat_version() {
    echo "Setting cloudbeat version for hermit version"
    sed -E -i.tmp "s/CLOUDBEAT_VERSION\": \".*\"/CLOUDBEAT_VERSION\": \"$CLOUDBEAT_VERSION\"/g" $HERMIT_FILE && rm $HERMIT_FILE.tmp
}

handle_version_changes() {
    if git diff --quiet --exit-code $HERMIT_FILE; then
        echo "No changes to $HERMIT_FILE; I'm done"
        return
    fi

    # Get current branch
    current_branch=$(git branch --show-current)
    echo "Current branch is: $current_branch"

    if [ "$current_branch" = "main" ]; then
        branch_name="sync-cloudbeat-version-$(date +%s)"
        echo "Creating new branch: $branch_name"
        git checkout -b $branch_name
        
        echo "Versions changed, commiting changes"
        git add $HERMIT_FILE
        git commit -m "bump CLOUDBEAT_VERSION in $HERMIT_FILE to $CLOUDBEAT_VERSION"
        
        echo "Pushing branch to origin"
        git push origin $branch_name
        
        echo "Creating PR with gh cli"
        gh pr create \
            --title "Sync CLOUDBEAT_VERSION in hermit.hcl to $CLOUDBEAT_VERSION" \
            --body "Automated update of CLOUDBEAT_VERSION in hermit.hcl to match version.go" \
            --base main \
            --head $branch_name
    else
        echo "Not on main branch, committing directly to $current_branch"
        echo "Versions changed, commiting changes"
        git add $HERMIT_FILE
        git commit -m "bump CLOUDBEAT_VERSION in $HERMIT_FILE to $CLOUDBEAT_VERSION"
    fi
}

find_current_cloudbeat_version
set_hermit_cloudbeat_version
handle_version_changes
