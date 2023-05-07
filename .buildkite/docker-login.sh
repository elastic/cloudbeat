#!/bin/bash
#

# ELASTICSEARCH CONFIDENTIAL
# __________________
#
#  Copyright Elasticsearch B.V. All rights reserved.
#
# NOTICE:  All information contained herein is, and remains
# the property of Elasticsearch B.V. and its suppliers, if any.
# The intellectual and technical concepts contained herein
# are proprietary to Elasticsearch B.V. and its suppliers and
# may be covered by U.S. and Foreign Patents, patents in
# process, and are protected by trade secret or copyright
# law.  Dissemination of this information or reproduction of
# this material is strictly forbidden unless prior written
# permission is obtained from Elasticsearch B.V.
#
# Build and Push all Drivah-compatible container image definitions.

set -euo pipefail
set +x

REGISTRY_USERNAME=$(vault kv get -field=username "${DOCKER_REGISTRY_VAULT_PATH}")
export REGISTRY_USERNAME

REGISTRY_PASSWORD=$(vault kv get -field=password "${DOCKER_REGISTRY_VAULT_PATH}")
export REGISTRY_PASSWORD

echo "${REGISTRY_PASSWORD}" | \
  buildah login --username="${REGISTRY_USERNAME}" \
                --password-stdin \
                docker.elastic.co
