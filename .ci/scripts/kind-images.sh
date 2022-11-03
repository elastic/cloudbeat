#!/usr/bin/env bash
set -euxo pipefail
IMAGE_SUFFIX=$1
CLUSTER_NAME=$2
kind load image-archive /tmp/cloudbeat-${IMAGE_SUFFIX}.tar --name=$CLUSTER_NAME & kind load image-archive /tmp/pytest-${IMAGE_SUFFIX}.tar --name=$CLUSTER_NAME & kind load image-archive /tmp/elastic-agent-${IMAGE_SUFFIX}.tar --name=$CLUSTER_NAME
