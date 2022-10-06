#!/bin/bash

# This script uses the kubectl commands in order to ssh into the cluster defined in the host current-context
# The script lets you to remotely edit the cloudbeat.yml file that configures cloudbeat process

source ./scripts/common.sh

TMP_LOCAL="/tmp/cloudbeat.yml"

PODS=$(kubectl -n kube-system get pod -l app=elastic-agent -o name)
for P in $PODS; do
  POD=$(echo "$P" | cut -d '/' -f 2)
  CONFIG_FILEPATH="$(find_cloudbeat_config $POD)"
  if [ -z "$CONFIG_FILEPATH" ]
  then
    echo "could not find remote config file"
    exit 1
  fi

  kubectl -n kube-system cp "$POD":"$CONFIG_FILEPATH" $TMP_LOCAL
  vi $TMP_LOCAL
  kubectl -n kube-system cp $TMP_LOCAL "$POD":"$CONFIG_FILEPATH"
  kubectl -n kube-system exec "$POD" -- chmod go-w "$CONFIG_FILEPATH"
  kubectl -n kube-system exec "$POD" -- chown root:root "$CONFIG_FILEPATH"
  rm $TMP_LOCAL
done
