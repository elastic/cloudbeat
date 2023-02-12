#!/bin/bash

# create a new agent policy and set POLICY_ID to the new agent id
create_a_new_agent_policy() {
  local KIBANA_URL=$1
  local KIBANA_AUTH=$2
  local AGENT_POLICY=$3

  # Install Agent policy
  installAgentResponse=$(curl -X POST \
    --url "${KIBANA_URL}/api/fleet/agent_policies?sys_monitoring=true" \
    -u "$KIBANA_AUTH" \
    -H 'Cache-Control: no-cache' \
    -H 'Connection: keep-alive' \
    -H 'Content-Type: application/json' \
    -H 'kbn-xsrf: true' \
    -d "@$AGENT_POLICY")

  check_status_code_of_curl "$installAgentResponse"

  POLICY_ID=$(echo "$installAgentResponse" | jq -r '.item.id')
  echo "Creating a new agent policy has completed successfully: New policy id: $POLICY_ID"
}

# create a new vanilla integration and set INTEGRATION_ID to the new integration id
create_a_new_vanilla_integration() {
  local KIBANA_URL=$1
  local KIBANA_AUTH=$2
  local POLICY_ID=$3
  local INTEGRATION_POLICY=$4

  local UPDATED_POLICY="$(jq --arg POLICY_ID "$POLICY_ID" '.policy_id |= $POLICY_ID' "$INTEGRATION_POLICY")"
  echo "New integration policy: $UPDATED_POLICY"

  PACKAGE_POLICY_RESPONSE=$(curl -X POST \
    --url "${KIBANA_URL}/api/fleet/package_policies" \
    -u "$KIBANA_AUTH" \
    -H 'Cache-Control: no-cache' \
    -H 'Connection: keep-alive' \
    -H 'Content-Type: application/json' \
    -H 'kbn-xsrf: true' \
    -d "${UPDATED_POLICY}")

  check_status_code_of_curl "$PACKAGE_POLICY_RESPONSE"

  echo "Creating a new a new vanilla integration with policy id: $POLICY_ID has completed successfully.Integration policy: $UPDATED_POLICY "
}

# create a new vanilla integration manifest file named manifest.yaml
create_new_vanilla_integration_manifest_file() {
  local KIBANA_URL=$1
  local KIBANA_AUTH=$2
  local POLICY_ID=$3

  ENROLMENT_TOKEN_RESPONSE=$(curl -X GET \
    --url "${KIBANA_URL}/api/fleet/enrollment_api_keys" \
    -u "$KIBANA_AUTH" \
    -H 'Cache-Control: no-cache' \
    -H 'Connection: keep-alive' \
    -H 'Content-Type: application/json' \
    -H 'kbn-xsrf: true')

  check_status_code_of_curl "$ENROLMENT_TOKEN_RESPONSE"

  ENROLMENT_TOKEN=$(echo "$ENROLMENT_TOKEN_RESPONSE" | jq --arg policy "$POLICY_ID" '.list[] | select(.policy_id == $policy)' | jq -r '.api_key')
  echo "ENROLMENT_TOKEN: $ENROLMENT_TOKEN"

  FLEET_DATA_RESPONSE=$(curl -X GET \
    --url "${KIBANA_URL}/api/fleet/settings" \
    -u "$KIBANA_AUTH" \
    -H 'Cache-Control: no-cache' \
    -H 'Connection: keep-alive' \
    -H 'Content-Type: application/json' \
    -H 'kbn-xsrf: true')

  check_status_code_of_curl "$FLEET_DATA_RESPONSE"
  FLEET_SERVER_HOST=$(echo "$FLEET_DATA_RESPONSE" | jq -r '.item.fleet_server_hosts[0]')
  echo "FLEET_SERVER_HOST: $FLEET_SERVER_HOST"

  # Create the manifest file
  MANIFEST_CREATION_RESPONSE=$(curl -X GET \
    --url "${KIBANA_URL}/api/fleet/kubernetes?fleetServer=${FLEET_SERVER_HOST}&enrolToken=${ENROLMENT_TOKEN}" \
    -u "$KIBANA_AUTH" \
    -H 'Cache-Control: no-cache' \
    -H 'Connection: keep-alive' \
    -H 'Content-Type: application/json' \
    -H 'kbn-xsrf: true')

  check_status_code_of_curl "$MANIFEST_CREATION_RESPONSE"

  # write the manifest file to the file system
  # get the item field from the response
  MANIFEST_FILE=$(echo "$MANIFEST_CREATION_RESPONSE" | jq -r '.item')
  echo "$MANIFEST_FILE" > manifest.yaml
}

check_status_code_of_curl() {
  local CURL_RESPONSE=$1
  error_code=$(echo "$CURL_RESPONSE" | jq -r '.statusCode')
  if [ "$error_code" != "null" ] && [ "$error_code" != "200" ]; then
    echo "Error code: $error_code"
    echo "Error message: $(echo "$CURL_RESPONSE" | jq -r '.message')"
    echo "Error full response: $CURL_RESPONSE"
    exit 1
  fi
}
