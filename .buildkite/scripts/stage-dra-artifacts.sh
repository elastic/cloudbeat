#!/usr/bin/env bash
##
##  Downloads Buildkite build artifacts and stages the workflow's slice into
##  artifacts/ for elastic/dra-prep.
##
##  On release branches, cloudbeat builds both snapshot and staging in the same
##  build, so filenames distinguish them: snapshot files contain "-SNAPSHOT"
##  (e.g. cloudbeat-9.6.0-SNAPSHOT-linux-x86_64.tar.gz, dependencies-9.6.0-SNAPSHOT.csv);
##  staging files do not.
##

set -euo pipefail

WORKFLOW="${DRA_WORKFLOW:?DRA_WORKFLOW is required}"

echo "--- Restoring artifacts"
buildkite-agent artifact download "build/distributions/*" .

echo "--- Preparing ${WORKFLOW} artifacts"
mkdir -p artifacts

if [[ "${WORKFLOW}" == "snapshot" ]]; then
    find build/distributions -type f -name "*SNAPSHOT*" -exec cp {} artifacts/ \;
else
    find build/distributions -type f ! -name "*SNAPSHOT*" -exec cp {} artifacts/ \;
fi

if ! ls artifacts/* >/dev/null 2>&1; then
    echo "ERROR: no ${WORKFLOW} artifacts found in artifacts/" >&2
    exit 1
fi

echo "Staged artifacts:"
ls -1 artifacts/
