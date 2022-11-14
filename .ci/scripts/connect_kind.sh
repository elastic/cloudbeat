#!/usr/bin/env bash
set -euo pipefail
KIND=$1
# The name of the network as it created by the elastic-packge
NETWORK="elastic-package-stack_default"
containers=$(docker ps | grep $KIND | awk '{ print $1 }')
for container in $containers; do
    docker network connect $NETWORK  $container
done
