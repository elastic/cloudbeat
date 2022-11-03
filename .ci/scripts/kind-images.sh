#!/usr/bin/env bash
set -euxo pipefail
IMAGE_SUFFIX=$1
CLUSTER_NAME=$2
IMAGES="cloudbeat pytest elastic-agent"

load_image () {
  kind load image-archive /tmp/"$1"-"${IMAGE_SUFFIX}".tar --name="$CLUSTER_NAME"
}

# for i in $IMAGES
# do
#   load_image "$i"
# done

# load_image cloudbeat &
# load_image pytest &
load_image elastic-agent
