#!/usr/bin/env bash

# this file should be sourced from inside the package and publish script.

fetch_elastic_qualifier() {
    local branch="${1}"
    local url="https://storage.googleapis.com/dra-qualifier/${branch}"
    local qualifier=""
    if curl -sf -o /dev/null "${url}"; then
        qualifier=$(curl -s "${url}")
    fi
    echo "${qualifier}"
}

# If the VERSION_QUALIFIER is already set (e.g. buildkite custom run), use that
# else try to fetch from google bucket for the current branch
if [ -z "${VERSION_QUALIFIER+x}" ]; then
    # VERSION_QUALIFIER is not set, get from bucket
    VERSION_QUALIFIER="$(fetch_elastic_qualifier "${BUILDKITE_BRANCH}")"
fi

# If this is a snapshot build, omit VERSION_QUALIFIER
if [ "${WORKFLOW}" = "snapshot" ]; then
    VERSION_QUALIFIER=''
fi
