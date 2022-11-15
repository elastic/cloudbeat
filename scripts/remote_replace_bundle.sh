#!/bin/bash

set -eo pipefail

# This script uses the kubectl commands in order to ssh into the cluster defined in the host current-context
# The script lets you to remotely replace the OPA bundle with a file named bundle.tar.gz, located on the host's running directory

source ./scripts/common.sh

copy_to_agents bundle.tar.gz
restart_agents
