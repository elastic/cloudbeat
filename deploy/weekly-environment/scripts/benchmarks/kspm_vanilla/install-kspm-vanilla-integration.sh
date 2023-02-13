#!/bin/bash

source ../../utils.sh

KIBANA_URL=$1
KIBANA_PASSWORD=$2
KIBANA_AUTH=elastic:${KIBANA_PASSWORD}
AGENT_POLICY=data/agent_policy_vanilla.json
INTEGRATION_POLICY=data/package_policy_vanilla.json

# Check if input is empty
if [ -z "$KIBANA_URL" ] || [ -z "$KIBANA_PASSWORD" ]; then
  echo "Kibana URL or Kibana password is empty"
  exit 1
fi

## Create a new agent policy And get the agent id
create_a_new_agent_policy "$KIBANA_URL" "$KIBANA_AUTH" "$AGENT_POLICY"

# Create a new vanilla integration
create_a_new_vanilla_integration "$KIBANA_URL" "$KIBANA_AUTH" "$POLICY_ID" "$INTEGRATION_POLICY"

# Create a new agent policy
create_new_vanilla_integration_manifest_file "$KIBANA_URL" "$KIBANA_AUTH" "$POLICY_ID"

