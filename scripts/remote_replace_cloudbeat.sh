#!/bin/bash

# This script uses the kubectl commands in order to ssh into the cluster defined in the host current-context
# The script lets you to remotely replace the cloudbeat binary with a file named cloudbeat, located on the host's running directory

source ./scripts/common.sh

LOCAL_BINARY="./cloudbeat"

PODS=$(kubectl -n kube-system get pod -l app=elastic-agent -o name)
for P in $PODS; do
  POD=$(echo $P | cut -d '/' -f 2)
  BINARY_FILEPATH="$(find_cloudbeat_binary $POD)"
  if [ -z "$BINARY_FILEPATH" ]
  then
    echo "could not find remote binary file"
    exit 1
  fi

  kubectl -n kube-system cp "$LOCAL_BINARY" "$POD":"$BINARY_FILEPATH"
  kubectl -n kube-system exec "$POD" -- chown elastic-agent:elastic-agent "$BINARY_FILEPATH"
done
