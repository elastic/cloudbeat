#!/bin/bash

set -euo pipefail

# Define a default Slack user
# TODO: replacy DefaultSlackUser with cloudbeat-eng-team
default_slack_user="DefaultSlackUser"

# Check if BUILDKITE_BUILD_CREATOR_EMAIL is set, defaulting to an empty string
build_creator="${BUILDKITE_BUILD_CREATOR_EMAIL:-"MissingEmail"}"

slack_user=$(vault kv get -field="$build_creator" secret/ci/elastic-cloudbeat/slack-users 2>/dev/null || echo "$default_slack_user")

# Output the YAML configuration
cat <<EOF
steps:
  - label: "Slack notification"
    command: echo "Notification test"
    notify:
      - slack:
          channels:
            - "#cloud-sec-ci"
          message: "Test: ping @$slack_user"
EOF
