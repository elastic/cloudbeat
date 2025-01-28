#!/usr/bin/env bash

# Allow other users write access to create checksum files

# The "branch" here selects which "$BRANCH.gradle" file of release manager is used
export VERSION=$(grep defaultBeatVersion version/version.go | cut -f2 -d "\"")
MAJOR=$(echo $VERSION | awk -F. '{ print $1 }')
MINOR=$(echo $VERSION | awk -F. '{ print $2 }')
if [ -n "$(git ls-remote --heads origin $MAJOR.$MINOR)" ]; then
    BRANCH=$MAJOR.$MINOR
elif [ -n "$(git ls-remote --heads origin $MAJOR.x)" ]; then
    BRANCH=$MAJOR.x
else
    BRANCH=main
fi

source .buildkite/scripts/qualifier.sh
echo "VERSION_QUALIFIER: ${VERSION_QUALIFIER}"

# Download artifacts from other stages
echo "Downloading artifacts..."
buildkite-agent artifact download "build/distributions/*" "." --step package-"${WORKFLOW}"
chmod -R 777 build/distributions

# Shared secret path containing the dra creds for project teams
DRA_CREDS=$(vault kv get -field=data -format=json kv/ci-shared/release/dra-role)

# Run release-manager
echo "Running release-manager container..."
IMAGE="docker.elastic.co/infra/release-manager:latest"
docker run --rm \
    --name release-manager \
    -e VAULT_ADDR=$(echo $DRA_CREDS | jq -r '.vault_addr') \
    -e VAULT_ROLE_ID=$(echo $DRA_CREDS | jq -r '.role_id') \
    -e VAULT_SECRET_ID=$(echo $DRA_CREDS | jq -r '.secret_id') \
    --mount type=bind,readonly=false,src="${PWD}",target=/artifacts \
    "$IMAGE" \
    cli collect \
    --project cloudbeat \
    --branch "${BRANCH}" \
    --commit "${BUILDKITE_COMMIT}" \
    --workflow "${WORKFLOW}" \
    --version "${VERSION}" \
    --artifact-set main \
    --qualifier "${VERSION_QUALIFIER}"
