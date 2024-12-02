#!/bin/bash

# This script creates a Buildkite notification step and pings the user who triggered the build.
# It assumes the existence of the GitHub user's email.
# If the email doesn't exist, the script attempts to extract it from the build message.
# If no email is found, the default option will be used.

set -euo pipefail

default_team="cloudbeat-eng-team"
# Check if BUILDKITE_BUILD_CREATOR_EMAIL is set, defaulting to user not found
build_creator="${BUILDKITE_BUILD_CREATOR_EMAIL:-$default_team}"
build_message="${BUILDKITE_MESSAGE:-}"
slack_channel="${SLACK_CHANNEL:-"#cloud-sec-ci"}"

default_group_id=$(vault kv get -field=$default_team secret/ci/elastic-cloudbeat/slack-users)
user_id=$(vault kv get -field="$build_creator" secret/ci/elastic-cloudbeat/slack-users 2>/dev/null || echo "$default_group_id")
# If the Slack user is the default one, try to extract email from the BUILD_MESSAGE
if [[ "$user_id" == "$default_group_id" ]]; then
    # Extract email from BUILDKITE_MESSAGE
    email=$(echo "$build_message" | grep -oE '[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,6}' || echo $default_team)
    user_id=$(vault kv get -field="$email" secret/ci/elastic-cloudbeat/slack-users 2>/dev/null || echo "$default_group_id")
fi

# Output the YAML configuration
cat <<EOF
notify:
  - slack:
      channels:
        - "$slack_channel"
      message: "<$user_id>"
    if: build.state != "passed"
EOF
