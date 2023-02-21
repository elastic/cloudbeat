#!/bin/bash

# This utility script contains functions that are used by the benchmark scripts.

#######################################
# Creates a new agent policy and set the new POLICY_ID as the new integration policy id
# Globals:
#   POLICY_ID
# Arguments:
#   $1: Kibana URL
#   $2: Kibana auth
#   $3: Agent policy
# Returns:
#   None
#######################################
create_a_new_agent_policy() {
  local kibana_url=$1
  local kibana_auth=$2
  local agent_policy=$3

  # Install Agent policy
  local install_agent_response
  install_agent_response="$(curl -X POST \
    --url "${kibana_url}/api/fleet/agent_policies?sys_monitoring=true" \
    -u "${kibana_auth}" \
    -H 'Cache-Control: no-cache' \
    -H 'Connection: keep-alive' \
    -H 'Content-Type: application/json' \
    -H 'kbn-xsrf: true' \
    -d "@${agent_policy}")"

  echo "Install agent response: ${install_agent_response}"
  check_status_code_of_curl "${install_agent_response}"

  POLICY_ID=$(echo "${install_agent_response}" | jq -r '.item.id')
  echo "Creating a new agent policy has completed successfully: New policy id: ${POLICY_ID}"
}

#######################################
# Creates a new vanilla integration on the given policy id
# Arguments:
#   $1: Kibana URL
#   $2: Kibana auth
#   $3: Policy id
#   $4: Integration policy
# Returns:
#   None
#######################################
create_a_new_vanilla_integration() {
  local kibana_url=$1
  local kibana_auth=$2
  local policy_id=$3
  local integration_policy=$4

  # Updating the new integration policy with the policy id
  local updated_policy
  updated_policy="$(jq --arg policy_id "${policy_id}" '.policy_id |= $policy_id' "${integration_policy}")"
  echo "New integration policy: ${updated_policy}"

  package_policy_response="$(curl -X POST \
    --url "${kibana_url}/api/fleet/package_policies" \
    -u "${kibana_auth}" \
    -H 'Cache-Control: no-cache' \
    -H 'Connection: keep-alive' \
    -H 'Content-Type: application/json' \
    -H 'kbn-xsrf: true' \
    -d "${updated_policy}")"

  check_status_code_of_curl "${package_policy_response}"

  echo "Creating a new a new vanilla integration with policy id: ${policy_id} has completed successfully.Integration policy: ${updated_policy}"
}

#######################################
# Creates a new vanilla integration manifest file manifest.yaml
# Globals:
#   MANIFEST_FILE
# Arguments:
#   $1: Kibana URL
#   $2: Kibana auth
#   $3: Policy id
# Returns:
#   None
#######################################
create_new_vanilla_integration_manifest_file() {
  local kibana_url=$1
  local kibana_auth=$2
  local policy_id=$3

  local enrolment_token_response
  enrolment_token_response="$(curl -X GET \
    --url "${kibana_url}/api/fleet/enrollment_api_keys" \
    -u "${kibana_auth}" \
    -H 'Cache-Control: no-cache' \
    -H 'Connection: keep-alive' \
    -H 'Content-Type: application/json' \
    -H 'kbn-xsrf: true')"

  check_status_code_of_curl "${enrolment_token_response}"

  local enrolment_token
  enrolment_token="$(echo "${enrolment_token_response}" | jq --arg policy "${policy_id}" '.list[] | select(.policy_id == $policy)' | jq -r '.api_key')"
  echo "enrolment_token: ${enrolment_token}"

  local fleet_data_response
  fleet_data_response="$(curl -X GET \
    --url "${kibana_url}/api/fleet/settings" \
    -u "${kibana_auth}" \
    -H 'Cache-Control: no-cache' \
    -H 'Connection: keep-alive' \
    -H 'Content-Type: application/json' \
    -H 'kbn-xsrf: true')"

  check_status_code_of_curl "${fleet_data_response}"

  local fleet_server_host
  fleet_server_host="$(echo "${fleet_data_response}" | jq -r '.item.fleet_server_hosts[0]')"
  echo "fleet_server_host: ${fleet_server_host}"

  # Create the manifest file
  local manifest_creation_response
  manifest_creation_response="$(curl -X GET \
    --url "${kibana_url}/api/fleet/kubernetes?fleetServer=${fleet_server_host}&enrolToken=${enrolment_token}" \
    -u "${kibana_auth}" \
    -H 'Cache-Control: no-cache' \
    -H 'Connection: keep-alive' \
    -H 'Content-Type: application/json' \
    -H 'kbn-xsrf: true')"

  check_status_code_of_curl "${manifest_creation_response}"

  # write the manifest file to the file system
  # get the item field from the response
  MANIFEST_FILE=$(echo "$manifest_creation_response" | jq -r '.item')
  echo "$MANIFEST_FILE" > manifest.yaml
}
#######################################
# Creates new alerts for the vanilla integration
# Arguments:
#   $1: Kibana URL
#   $2: Kibana auth
#   $3: Agent policy
#   $4: Integration policy
# Returns:
#   None
#######################################
create_alerts_for_the_vanilla_integration() {
  local kibana_url=$1
  local kibana_auth=$2
  local alerts_file=$3
  local slack_webhook_url=$4
  local slack_configuration_file=$5

  # Imports the slack connector and the rule alerts from a saved object file
  local import_alerts_response
  import_alerts_response=$(curl -X POST \
    "${kibana_url}/api/saved_objects/_import?overwrite=true" \
    -u "${kibana_auth}" \
    -H 'Cache-Control: no-cache' \
    -H 'Connection: keep-alive' \
    -H "kbn-xsrf: true" \
    --form file="@${alerts_file}")
  check_status_code_of_curl "${import_alerts_response}"

  # Get the connector id
  local connector_response
  connector_response=$(curl -X GET \
      "${kibana_url}/api/actions/connectors" \
      -u "${kibana_auth}" \
      -H 'Cache-Control: no-cache' \
      -H 'Connection: keep-alive' \
      -H "kbn-xsrf: true" \
      -H 'Content-Type: application/json')
  check_status_code_of_curl "${connector_response}"

  # Extracts the connector id of the slack connector
  local connector_id
  connector_id="$(echo "${connector_response}" |  jq '.[]  | select(.name == "#cloud-security-posture")' | jq -r '.id')"
  echo "Connector id: ${connector_id}"

  # Updates the slack connector with the webhook url
  local connector_configuration
  connector_configuration="$(jq --arg slack_webhook_url "${slack_webhook_url}" '.secrets.webhookUrl |= $slack_webhook_url' "${slack_configuration_file}")"
  echo "New connector configuration: ${connector_configuration}"
  local update_connector_response
  update_connector_response=$(curl -X PUT \
    "${kibana_url}/api/actions/connector/${connector_id}" \
    -u "${kibana_auth}" \
    -H 'Cache-Control: no-cache' \
    -H 'Connection: keep-alive' \
    -H "kbn-xsrf: true" \
    -H 'Content-Type: application/json' \
    -d "${connector_configuration}")

  check_status_code_of_curl "${update_connector_response}"
}

#######################################
# Checks the status code of the curl response and exits if the status code is not 200
# Globals:
# Arguments:
#   $1: Curl response
# Returns:
#   None
#######################################
check_status_code_of_curl() {
  local curl_response=$1
  error_code=$(echo "$curl_response" | jq -r '.statusCode')
  if [ "$error_code" != "null" ] && [ "$error_code" != "200" ]; then
    echo "Error code: $error_code"
    echo "Error message: $(echo "$curl_response" | jq -r '.message')"
    echo "Error full response: $curl_response"
    exit 1
  fi
}
