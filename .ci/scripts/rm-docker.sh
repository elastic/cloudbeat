#!/usr/bin/env bash
IMAGE="docker.elastic.co/infra/release-manager:latest"
WORKFLOW="snapshot"
VERSION="8.2.0"
BRANCH="main"

# Allow other users write access to create checksum files
chmod -R 777 ${BASE_DIR}/build/distributions 

# The "branch" here selects which "$BRANCH.gradle" file of release manager is used
# VERSION=$(make get-version)
# MAJOR=$(echo $VERSION | awk -F. '{ print $1 }')
# MINOR=$(echo $VERSION | awk -F. '{ print $2 }')
# if [ -n "$(git ls-remote --heads origin $MAJOR.$MINOR)" ] ; then
#     BRANCH=$MAJOR.$MINOR
# elif [ -n "$(git ls-remote --heads origin $MAJOR.x)" ] ; then
#     BRANCH=$MAJOR.x
# else
#     BRANCH=main
# fi


# Generate checksum files and upload to GCS
docker run --rm \
  --name release-manager \
  -e VAULT_ADDR \
  -e VAULT_ROLE_ID \
  -e VAULT_SECRET_ID \
  --mount type=bind,readonly=false,src="$PWD/${BASE_DIR}",target=/artifacts \
  "$IMAGE" \
    cli collect \
      --project cloudbeat \
      --branch "$BRANCH" \
      --commit `git rev-parse HEAD` \
      --workflow "$WORKFLOW" \
      --artifact-set main