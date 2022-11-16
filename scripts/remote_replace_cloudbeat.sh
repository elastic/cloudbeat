#!/bin/bash

set -eo pipefail

# This script uses the kubectl commands in order to ssh into the cluster defined in the host current-context
# The script lets you to remotely replace the cloudbeat binary with a file named cloudbeat, located on the host's running directory

source ./scripts/common.sh

ARCH=$(find_target_arch)
if [ "$ARCH" = "amd64" ]; then
    ARCH="x86_64"
fi
OS=$(find_target_os)
VERSION=$(make get-version)

LOCAL_DIR=cloudbeat-$VERSION-SNAPSHOT-$OS-$ARCH
echo "Looking for build distribution: $LOCAL_DIR"
tar -xvf build/distributions/$LOCAL_DIR.tar.gz -C build/ > /dev/null 2>&1

for P in $(get_agents); do
  POD=$(echo $P | cut -d '/' -f 2)
  SHA=$(get_agent_sha $POD)
  echo "Found sha=$SHA in pod=$POD"
  
  ROOT=/usr/share/elastic-agent
  DEST=$ROOT/data/elastic-agent-$SHA/components
  cp_to_pod $POD build/$LOCAL_DIR/cloudbeat $DEST

  # Start with COPY_BUNDLE=true to move also the opa bundle to the agent
  # the bundle can be found later in in /usr/share/elastic-agent/state/data/run/cloudbeat/{BUNDLE_NAME}
  if [[ ! -z "$COPY_BUNDLE" ]]
  then
    BUNDLE="cis_k8s-default" 
    if [ ! -z "$(is_eks)" ]
    then
          BUNDLE="cis_eks-default"
    fi
    BUNDLE_DIR=$ROOT/state/data/run/cloudbeat/$BUNDLE
    exec_pod $POD "mkdir -p $BUNDLE_DIR"
    cp_to_pod $POD build/$LOCAL_DIR/bundle.tar.gz $BUNDLE_DIR/bundle.tar.gz
    echo "copied bundle to $BUNDLE_DIR/bundle.tar.gz"
  fi
  echo "Copied all the assets, restarting the agent $POD"
  PID=$(exec_pod $POD "pidof cloudbeat")
  exec_pod $POD "kill -9 $PID"
  # exec_pod $POD "elastic-agent restart" # https://github.com/elastic/cloudbeat/pull/458#issuecomment-1308837098
done
