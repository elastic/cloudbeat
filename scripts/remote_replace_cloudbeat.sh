#!/bin/bash

set -eo pipefail

# This script uses the kubectl commands in order to ssh into the cluster defined in the host current-context
# The script lets you to remotely replace the cloudbeat binary with a file named cloudbeat, located on the host's running directory

source ./scripts/common.sh

for P in $(get_agents); do
  POD=$(echo $P | cut -d '/' -f 2)
  SHA=$(get_agent_sha $POD)
  echo "Found sha=$SHA in pod=$POD"

  DEST=/usr/share/elastic-agent/data/elastic-agent-$SHA/components
  cp_to_pod $POD ./cloudbeat $DEST
  cp_to_pod $POD ./cloudbeat.yml $DEST

  echo "Copied all the assets to $POD"
  # exec_pod $POD "elastic-agent restart" # https://github.com/elastic/cloudbeat/pull/458#issuecomment-1308837098
done
