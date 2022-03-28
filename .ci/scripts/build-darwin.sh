#!/usr/bin/env bash
set -euox pipefail

# shellcheck disable=SC1091
source ./scripts/make/common.bash

jenkins_setup

./.ci/scripts/build.sh
