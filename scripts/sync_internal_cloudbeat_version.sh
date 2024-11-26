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

commit_if_different() {
    if git diff --quiet --exit-code $HERMIT_FILE; then
        echo "No changes to $HERMIT_FILE; I'm done"
        return
    fi
    echo "Versions changed, commiting changes"
    git add $HERMIT_FILE
    git commit -m "bump CLOUDBEAT_VERSION in $HERMIT_FILE to $CLOUDBEAT_VERSION"
}

find_current_cloudbeat_version
set_hermit_cloudbeat_version
commit_if_different
