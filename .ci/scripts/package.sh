#!/usr/bin/env bash
set -euox pipefail

# linux/amd64 is in the default list already, set here
# to prevent jenkins_release.sh from adding more PLATFORMS
export PLATFORMS="linux/amd64,linux/arm64"
export TYPES="tar.gz"

make release-manager-snapshot
cp build/dependencies-*.csv build/distributions/.