#!/usr/bin/env bash
set -euox pipefail

IMAGE="docker.elastic.co/infra/release-manager:latest"
WORKFLOW="staging"

# The "branch" here selects which "$BRANCH.gradle" file of release manager is used
# VERSION=$(make get-version)
VERSION="8.2.0"
MAJOR=$(echo $VERSION | awk -F. '{ print $1 }')
MINOR=$(echo $VERSION | awk -F. '{ print $2 }')
if [ -n "$(git ls-remote --heads origin $MAJOR.$MINOR)" ] ; then
    BRANCH=$MAJOR.$MINOR
elif [ -n "$(git ls-remote --heads origin $MAJOR.x)" ] ; then
    BRANCH=$MAJOR.x
else
    BRANCH=main
fi

# Generate checksum files and upload to GCS
docker run --rm \
  --name release-manager \
  -e VAULT_ADDR \
  -e VAULT_ROLE_ID \
  -e VAULT_SECRET_ID \
  --mount type=bind,readonly=false,src="$PWD",target=/artifacts \
  "$IMAGE" \
    cli collect \
      --project cloudbeat \
      --branch "$BRANCH" \
      --commit `git rev-parse HEAD` \
      --workflow "$WORKFLOW" \
      # --qualifier "$VERSION_QUALIFIER" \
      --artifact-set main