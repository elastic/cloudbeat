#! /bin/bash
set -xeuo pipefail

VERSION_FILE="version/version.go"
HERMIT_FILE="bin/hermit.hcl"
ARTIFACTS_API_URL="https://artifacts-api.elastic.co/v1/versions"

git config --global user.email "cloudsecmachine@users.noreply.github.com"
git config --global user.name "Cloud Security Machine"

find_current_cloudbeat_version() {
    echo "Checking current cloudbeat version"
    CLOUDBEAT_VERSION=$(grep -oE 'defaultBeatVersion\s+=\s+".*"' $VERSION_FILE | grep -oE '[0-9]\.[0-9]\.[0-9]')
    echo "Cloudbeat version is $CLOUDBEAT_VERSION"
}

set_hermit_cloudbeat_version() {
    echo "Setting cloudbeat version for hermit version"
    sed -E -i "s/CLOUDBEAT_VERSION\": \".*\"/CLOUDBEAT_VERSION\": \"$CLOUDBEAT_VERSION\"/g" $HERMIT_FILE
}

is_snapshot_published() {
    # ELK_VERSION resolves to "${CLOUDBEAT_VERSION}-SNAPSHOT"; only bump once that
    # snapshot actually exists on the Elastic artifacts API, otherwise the test-runner
    # would pin a version whose agent artifacts aren't published yet (see issue #17923).
    local target="${CLOUDBEAT_VERSION}-SNAPSHOT"
    echo "Checking whether $target is published on the Elastic artifacts API"
    local versions
    if ! versions=$(curl -fsS "$ARTIFACTS_API_URL" | jq -r '.versions[]'); then
        echo "Could not query $ARTIFACTS_API_URL; skipping this run to be safe"
        return 1
    fi
    if grep -qxF "$target" <<<"$versions"; then
        echo "$target is available"
        return 0
    fi
    echo "$target is not published yet"
    return 1
}

handle_version_changes() {
    if git diff --quiet --exit-code $HERMIT_FILE; then
        echo "No changes to $HERMIT_FILE; I'm done"
        return
    fi

    # A bump is pending — only open a PR once the target snapshot is actually
    # published, otherwise revert and let the next scheduled run try again.
    if ! is_snapshot_published; then
        echo "Skipping bump; ${CLOUDBEAT_VERSION}-SNAPSHOT not published yet. Will retry on the next run."
        git checkout -- "$HERMIT_FILE"
        return
    fi

    # Get current branch
    current_branch=$(git branch --show-current)
    echo "Current branch is: $current_branch"

    branch_name="sync-cloudbeat-version-$(date +%s)"
    echo "Creating new branch: $branch_name"
    git checkout -b "$branch_name"

    echo "Versions changed, committing changes"
    git add $HERMIT_FILE
    git commit -m "bump CLOUDBEAT_VERSION in $HERMIT_FILE to $CLOUDBEAT_VERSION"

    echo "Pushing branch to origin"
    git push origin "$branch_name"

    echo "Creating PR with gh cli"
    gh pr create \
        --title "Sync CLOUDBEAT_VERSION in hermit.hcl to $CLOUDBEAT_VERSION" \
        --body "Automated update of CLOUDBEAT_VERSION in hermit.hcl to match version.go" \
        --base "$current_branch" \
        --head "$branch_name"
}
find_current_cloudbeat_version
set_hermit_cloudbeat_version
handle_version_changes
