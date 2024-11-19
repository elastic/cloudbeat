#!/bin/bash

# This script creates a Buildkite notification step and pings the user who triggered the build.
# It assumes the existence of the GitHub user's email.
# If the email doesn't exist, the script attempts to extract it from the build message.
# If no email is found, the default option will be used.

set -euo pipefail

# Define the default user not found message
user_not_found="user_not_found"

# Check if BUILDKITE_BUILD_CREATOR_EMAIL is set, defaulting to user not found
build_creator="${BUILDKITE_BUILD_CREATOR_EMAIL:-$user_not_found}"
build_message="${BUILDKITE_MESSAGE:-}"
slack_channel="${SLACK_CHANNEL:-"#cloud-sec-ci"}"

default_group_id=$(vault kv get -field="cloudbeat-eng-team" secret/ci/elastic-cloudbeat/slack-users)

slack_user_id=$(vault kv get -field="$build_creator" secret/ci/elastic-cloudbeat/slack-users 2>/dev/null || echo "$user_not_found")
# If the Slack user is the default one, try to extract email from the BUILD_MESSAGE
if [[ "$slack_user_id" == "$user_not_found" ]]; then
    # Extract email from BUILDKITE_MESSAGE
    email=$(echo "$build_message" | grep -oE '[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,6}' || echo "$user_not_found")
    # Check if an email was found and use it to attempt to retrieve the Slack user from Vault
    if [[ -n "$email" ]]; then
        slack_user_id=$(vault kv get -field="$email" secret/ci/elastic-cloudbeat/slack-users 2>/dev/null || echo "$user_not_found")
    fi
fi

if [[ "$slack_user_id" == "$user_not_found" ]]; then
    user_id="<!subteam^$default_group_id>"
else
    user_id="<@$slack_user_id>"
fi

# Output the YAML configuration
cat <<EOF
notify:
  - slack:
      channels:
        - "$slack_channel"
      message: "$user_id"
EOF
