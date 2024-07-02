#!/usr/bin/env bash
set -uox pipefail

export PLATFORMS="linux/amd64,linux/arm64"
export TYPES="tar.gz"
source ./bin/activate-hermit

CLOUDBEAT_VERSION=$(grep defaultBeatVersion version/version.go | cut -d'=' -f2 | tr -d '" ')
PYTHON_BIN=./build/ve/$(go env GOOS)/bin
PYTHON=$PYTHON_BIN/python

if [ "$WORKFLOW" = "snapshot" ]; then
    export SNAPSHOT="true"
fi

# debug command to verify
ls -lah /proc/sys/fs/binfmt_misc/ || true

mage pythonEnv
mage package

CSV_FILE="build/dependencies-${CLOUDBEAT_VERSION}"
[ -n "${SNAPSHOT+x}" ] && CSV_FILE+="-SNAPSHOT"

echo "Generating $CSV_FILE.csv"
$PYTHON ./.buildkite/scripts/generate_notice.py --csv "$CSV_FILE.csv"
cp build/dependencies-*.csv build/distributions/.
