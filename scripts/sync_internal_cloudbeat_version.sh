#! /bin/bash
set -xeuo pipefail

VERSION_FILE="version/version.go"
HERMIT_FILE="bin/hermit.hcl"

find_current_cloudbeat_version() {
    echo "Checking current cloudbeat version"
    CLOUDBEAT_VERSION=$(grep -oE 'defaultBeatVersion\s+=\s+".*"' $VERSION_FILE | grep -oE '[0-9]\.[0-9]\.[0-9]')
    echo "Cloudbeat version is $CLOUDBEAT_VERSION"
}

set_hermit_cloudbeat_version() {
    echo "Setting cloudbeat version for hermit version"
    sed -E -i '' "s/CLOUDBEAT_VERSION\": \".*\"/CLOUDBEAT_VERSION\": \"$CLOUDBEAT_VERSION\"/g" $HERMIT_FILE
}

create_pr_if_different() {
    if git diff --quiet --exit-code $HERMIT_FILE; then
        echo "No changes to $HERMIT_FILE; I'm done"
        return
    fi
    echo "Versions changed, creating a GitHub PR"
    git checkout -b "sync-hermit-version-$(date +%s)"
    git add $HERMIT_FILE
    git commit -m "bump CLOUDBEAT_VERSION in $HERMIT_FILE to $CLOUDBEAT_VERSION"
    git push -u origin

    echo "Updates CLOUDBEAT_VERSION in $HERMIT_FILE to $CLOUDBEAT_VERSION." >pr_body
    gh pr create --title "Bump cloudbeat version in hermit.hcl to $CLOUDBEAT_VERSION" \
        --body-file pr_body \
        --base "$BASE_BRANCH" \
        --label "backport-skip"
    rm pr_body
}

find_current_cloudbeat_version
set_hermit_cloudbeat_version
create_pr_if_different
