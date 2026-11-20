#!/usr/bin/env bash
set -uox pipefail

export TYPES="tar.gz"
source ./bin/activate-hermit
export GOEXPERIMENT=jsonv2 # required by trivy v0.69.3+ (encoding/json/v2)

CLOUDBEAT_VERSION=$(grep defaultBeatVersion version/version.go | cut -d'=' -f2 | tr -d '" ')
PYTHON_BIN=./build/ve/$(go env GOOS)/bin
PYTHON=$PYTHON_BIN/python

source .buildkite/scripts/qualifier.sh
echo "VERSION_QUALIFIER: ${VERSION_QUALIFIER}"
export VERSION_QUALIFIER

if [ "$WORKFLOW" = "snapshot" ]; then
    export SNAPSHOT="true"
fi

mage pythonEnv
mage package

# Beats mage skips a beat missing from its FIPS list with exit 0 and no package.
# The standard steps upload the dependencies CSV, so FIPS steps stop here.
if [[ "${FIPS:-}" == "true" ]]; then
    if ! compgen -G "build/distributions/cloudbeat-fips-*.tar.gz" >/dev/null; then
        echo "ERROR: FIPS=true mage package produced no cloudbeat-fips tar.gz" >&2
        exit 1
    fi
    echo "Produced artifacts:"
    ls -lahR build/distributions/
    exit 0
fi

CSV_FILE="build/dependencies-${CLOUDBEAT_VERSION}"
[ -n "${SNAPSHOT+x}" ] && CSV_FILE+="-SNAPSHOT"
if [[ -n "${VERSION_QUALIFIER}" ]]; then
    CSV_FILE+="-${VERSION_QUALIFIER}"
fi

echo "Generating $CSV_FILE.csv"
$PYTHON ./.buildkite/scripts/generate_notice.py --csv "$CSV_FILE.csv"
cp build/dependencies-*.csv build/distributions/.

echo "Produced artifacts:"
ls -lahR build/distributions/
