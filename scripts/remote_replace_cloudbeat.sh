#!/bin/bash

set -eo pipefail

# This script uses the kubectl commands in order to ssh into the cluster defined in the host current-context
# The script lets you to remotely replace the cloudbeat binary with a file named cloudbeat, located on the host's running directory

source ./scripts/common.sh

copy_to_agents cloudbeat cloudbeat.yml
restart_agents
