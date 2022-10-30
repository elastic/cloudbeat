#!/bin/bash

set -euo pipefail

# This script uses the kubectl commands in order to ssh into the cluster defined in the host current-context
# The script lets you to remotely replace the cloudbeat binary with a file named cloudbeat, located on the host's running directory

source ./scripts/common.sh

ARCH=$(find_target_arch)
if [ "$ARCH" = "amd64" ]; then
    ARCH="x86_64"
fi
OS=$(find_target_os)
VERSION="8.6.0" # Read from elsewhere
SNAPSHOT="SNAPSHOT"

LOCAL_DIR=cloudbeat-$VERSION-$SNAPSHOT-$OS-$ARCH
echo "Looking for build distribution: $LOCAL_DIR"
tar -xvf build/distributions/$LOCAL_DIR.tar.gz > /dev/null 2>&1

# cloudbeat.spec.yml should be built in the agent
curl -o cloudbeat.spec.yml https://raw.githubusercontent.com/elastic/elastic-agent/feature-arch-v2/specs/cloudbeat.spec.yml > /dev/null 2>&1

for P in $(get_agents); do
  POD=$(echo $P | cut -d '/' -f 2)
  out=$(exec_pod $POD "elastic-agent version --yaml --binary-only")
  SHA=$(echo $out | cut -d ":" -f4 | awk '{$1=$1};1'|  awk '{ print substr($0, 0, 6) }')
  echo "Found sha=$SHA in pod=$POD"
  
  ROOT=/usr/share/elastic-agent/data/elastic-agent-$SHA
  DEST=$ROOT/components
  cp_to_pod $POD $LOCAL_DIR/cloudbeat $DEST
  cp_to_pod $POD $LOCAL_DIR/cloudbeat.yml $DEST/cloudbeat.yml
  cp_to_pod $POD cloudbeat.spec.yml $DEST/cloudbeat.spec.yml

  BUNDLE="cis_k8s-default" 
  if [ ! -z "$(is_eks)" ]
  then
        BUNDLE="cis_eks-default"
  fi
  BUNDLE_DIR=$ROOT/run/cloudbeat/$BUNDLE
  exec_pod $POD "mkdir -p $BUNDLE_DIR"
  cp_to_pod $POD $LOCAL_DIR/bundle.tar.gz $BUNDLE_DIR/bundle.tar.gz
  echo "Copied all the assets, restarting the agent"
  exec_pod $POD "elastic-agent restart"
done
