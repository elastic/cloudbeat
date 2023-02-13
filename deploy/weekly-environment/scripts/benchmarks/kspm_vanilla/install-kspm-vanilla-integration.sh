#!/bin/bash

source ../../utils.sh

# This script is used to install a vanilla integration for the KSPM vanilla benchmark.
# It will create a new agent policy, a new vanilla integration and a new vanilla integration manifest file.
# The script requires two arguments:
# 1. Kibana URL
# 2. Kibana password

KIBANA_URL=$1
KIBANA_PASSWORD=$2
KIBANA_AUTH=elastic:${KIBANA_PASSWORD}

readonly AGENT_POLICY=data/agent_policy_vanilla.json
readonly INTEGRATION_POLICY=data/package_policy_vanilla.json

# Check if input is empty
if [ -z "$KIBANA_URL" ] || [ -z "$KIBANA_PASSWORD" ]; then
  echo "Kibana URL or Kibana password is empty"
  exit 1
fi

## Create a new agent policy And get the agent id
create_a_new_agent_policy "$KIBANA_URL" "$KIBANA_AUTH" "$AGENT_POLICY"
if [ -z "$POLICY_ID" ]; then
  echo "Agent policy id is empty"
  exit 1
fi

# Create a new vanilla integration
create_a_new_vanilla_integration "$KIBANA_URL" "$KIBANA_AUTH" "$POLICY_ID" "$INTEGRATION_POLICY"

# Create a new agent policy
create_new_vanilla_integration_manifest_file "$KIBANA_URL" "$KIBANA_AUTH" "$POLICY_ID"
if [ -z "$MANIFEST_FILE" ]; then
  echo "Manifest file is empty"
  exit 1
fi
