#!/usr/bin/env bash
set -uox pipefail

export PLATFORMS="linux/amd64,linux/arm64"
export TYPES="tar.gz"
CLOUDBEAT_VERSION=$(grep defaultBeatVersion version/version.go | cut -d'=' -f2 | tr -d '" ')

if [ "$WORKFLOW" = "snapshot" ] ; then
    export SNAPSHOT="true"
fi

source ./bin/activate-hermit

mage pythonEnv

./bin/python3 ./.buildkite/scripts/generate_notice.py --csv build/dependencies-"${CLOUDBEAT_VERSION}"-"${WORKFLOW}".csv

mage package

cp build/dependencies-*.csv build/distributions/.
