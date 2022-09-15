#!/usr/bin/bash

# This script builds an image from the elastic-agent image
# with a locally built apm-server binary injected. Additional
# flags (e.g. -t <name>) will be passed to `docker build`.

set -eu

REPO_ROOT=$(cd $(dirname $(realpath "$0"))/../../../.. && pwd)

DEFAULT_IMAGE_TAG=$(curl -X GET https://snapshots.elastic.co/latest/master.json | jq -r '.build_id')-SNAPSHOT
BASE_IMAGE="${BASE_IMAGE:-docker.elastic.co/beats/elastic-agent:$DEFAULT_IMAGE_TAG}"
GOARCH="${GOARCH:-$(go env GOARCH)}"

export DOCKER_BUILDKIT=1
docker pull $BASE_IMAGE

STACK_VERSION=$(docker inspect -f '{{index .Config.Labels "org.label-schema.version"}}' $BASE_IMAGE)
VCS_REF=$(docker inspect -f '{{index .Config.Labels "org.label-schema.vcs-ref"}}' $BASE_IMAGE)

# make -C $REPO_ROOT build/apm-server-linux-$GOARCH
GOOS=linux GOARCH=amd64 mage -d $REPO_ROOT build

docker build \
	-f $REPO_ROOT/dev-tools/packaging/docker/elastic-agent/Dockerfile \
	--build-arg ELASTIC_AGENT_IMAGE=$BASE_IMAGE \
	--build-arg STACK_VERSION=$STACK_VERSION \
	--build-arg VCS_REF_SHORT=${VCS_REF:0:6} \
	--platform linux/$GOARCH \
	$* $REPO_ROOT
