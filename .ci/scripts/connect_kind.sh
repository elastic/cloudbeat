#!/usr/bin/env bash
set -euo pipefail
kind=$1
action=${2:-connect}
# The name of the network as it created by the elastic-packge
network="elastic-package-stack_default"
containers=$(docker ps | grep $kind | awk '{ print $1 }')
for container in $containers; do
    docker network $action $network $container
done
