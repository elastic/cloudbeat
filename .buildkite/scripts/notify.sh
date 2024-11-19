#!/bin/bash

set -euo pipefail

# Define a default Slack user
# default_slack_user="cloudbeat-eng-team"
default_slack_user="Dima Gurevich"

# Check if BUILDKITE_BUILD_AUTHOR is set, defaulting to an empty string
build_author="${BUILDKITE_BUILD_AUTHOR:-}"

# Map GitHub usernames to Slack usernames using a case statement
case "$build_author" in
"gurevichdmitry")
    slack_user="Dima Gurevich"
    ;;
"newuser")
    slack_user="slack user2"
    ;;
*)
    slack_user="$default_slack_user"
    ;;
esac

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
