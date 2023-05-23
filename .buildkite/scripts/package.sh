#!/usr/bin/env bash
set -euox pipefail

export PLATFORMS="linux/amd64,linux/arm64"
export TYPES="tar.gz"

source ./bin/activate-hermit
if [ $WORKFLOW = "staging" ] ; then
    make release-manager-release
else
    export SNAPSHOT="true"
    make release-manager-snapshot
fi

cp build/dependencies-*.csv build/distributions/.
