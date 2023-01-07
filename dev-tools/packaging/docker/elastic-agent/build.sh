#!/usr/bin/bash

# This script builds an image from the elastic-agent image
# with a locally built cloudbeat binary injected. Additional
# flags (e.g. -t <name>) will be passed to `docker build`.

set -eu

REPO_ROOT=$(realpath "$(dirname "$(realpath dev-tools/packaging/docker/elastic-agent/build.sh)")"/../../../..)

DEFAULT_IMAGE_TAG="${DEFAULT_IMAGE_TAG:-$(make get-version)-SNAPSHOT}"
BASE_IMAGE="${BASE_IMAGE:-docker.elastic.co/beats/elastic-agent:$DEFAULT_IMAGE_TAG}"
GOARCH="${GOARCH:-$(go env GOARCH)}"

export DOCKER_BUILDKIT=1
docker pull $BASE_IMAGE

STACK_VERSION=$(docker inspect -f '{{index .Config.Labels "org.label-schema.version"}}' $BASE_IMAGE)
VCS_REF=$(docker inspect -f '{{index .Config.Labels "org.label-schema.vcs-ref"}}' $BASE_IMAGE)

docker buildx build \
	-f $REPO_ROOT/dev-tools/packaging/docker/elastic-agent/Dockerfile \
	--build-arg ELASTIC_AGENT_IMAGE=$BASE_IMAGE \
	--build-arg STACK_VERSION=$STACK_VERSION \
	--build-arg VCS_REF_SHORT=${VCS_REF:0:6} \
	--platform linux/$GOARCH \
	--cache-from=type=local,src=/tmp/.buildx-cache \
	--cache-to=type=local,dest=/tmp/.buildx-cache-new \
  --output type=docker,dest=/tmp/elastic-agent-$CONTAINER_SUFFIX.tar \
	$* $REPO_ROOT
