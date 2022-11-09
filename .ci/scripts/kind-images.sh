#!/usr/bin/env bash
set -euxo pipefail
IMAGE_SUFFIX=$1
CLUSTER_NAME=$2

load_image () {
  kind load image-archive /tmp/"$1"-"${IMAGE_SUFFIX}".tar --name="$CLUSTER_NAME"
}

load_image cloudbeat &
load_image pytest &
load_image elastic-agent
