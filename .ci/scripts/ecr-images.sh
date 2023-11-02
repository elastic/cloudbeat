#!/usr/bin/env bash
set -euxo pipefail
IMAGE_SUFFIX=$1
CONTAINER_REGISTRY=$2
ELASTIC_AGENT_DOCKER_TAG=$3

load_and_push_image() {
    docker load -i /tmp/"$1"-"${IMAGE_SUFFIX}".tar
    if [ "$1" != "elastic-agent" ]; then
        docker tag "$2" "$CONTAINER_REGISTRY/$2"
    fi
    docker push "$CONTAINER_REGISTRY/$2"
}

load_and_push_image cloudbeat "cloudbeat:latest" &
load_and_push_image pytest "cloudbeat-test:latest" &
load_and_push_image elastic-agent "elastic-agent:$ELASTIC_AGENT_DOCKER_TAG"
