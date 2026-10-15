#!/usr/bin/env bash
#
# Set the Kibana "solution view" for a space (default: the `default` space)
# so it opens straight into the Security experience, instead of Classic.
#
# The solution view is per-space data (the `solution` field on the Space
# saved object), not a kibana.yml setting, so it has to be set via the
# Spaces API against the already-running Kibana instance. GET the existing
# space first and PUT the full object back with `solution` changed, rather
# than guessing at the other fields the PUT endpoint expects.
#
# Required: KIBANA_URL, ES_USER, ES_PASSWORD
# Optional: SPACE_ID (default "default"), SOLUTION (default "security")

set -euo pipefail

: "${KIBANA_URL:?KIBANA_URL must be set}"
: "${ES_USER:?ES_USER must be set}"
: "${ES_PASSWORD:?ES_PASSWORD must be set}"

readonly KIBANA_BASE="${KIBANA_URL%/}"
readonly SPACE_ID="${SPACE_ID:-default}"
readonly SOLUTION="${SOLUTION:-security}"
readonly CURL_COMMON=(--connect-timeout 10 --max-time 60 -sS)

WORK_DIR="$(mktemp -d)"
trap 'rm -rf "${WORK_DIR}"' EXIT

kibana_curl() {
    curl "${CURL_COMMON[@]}" -u "${ES_USER}:${ES_PASSWORD}" -H 'kbn-xsrf: true' "$@"
}

echo "Fetching space '${SPACE_ID}'"
get_out="${WORK_DIR}/space_get.json"
http_code=$(kibana_curl -o "${get_out}" -w "%{http_code}" "${KIBANA_BASE}/api/spaces/space/${SPACE_ID}")
if [[ "${http_code}" != "200" ]]; then
    echo "Fetching space '${SPACE_ID}' failed (HTTP ${http_code})" >&2
    head -c 2000 "${get_out}" >&2 || true
    exit 1
fi

current_solution=$(jq -r '.solution // empty' "${get_out}")
if [[ "${current_solution}" == "${SOLUTION}" ]]; then
    echo "Space '${SPACE_ID}' already defaults to solution '${SOLUTION}'."
    exit 0
fi

echo "Setting space '${SPACE_ID}' solution: '${current_solution:-<unset>}' -> '${SOLUTION}'"
put_body="${WORK_DIR}/space_put.json"
jq --arg solution "${SOLUTION}" '.solution = $solution' "${get_out}" >"${put_body}"

put_out="${WORK_DIR}/space_put_out.json"
http_code=$(kibana_curl -o "${put_out}" -w "%{http_code}" -X PUT "${KIBANA_BASE}/api/spaces/space/${SPACE_ID}" \
    -H 'Content-Type: application/json' --data-binary "@${put_body}")
if [[ "${http_code}" != "200" ]]; then
    echo "Updating space '${SPACE_ID}' failed (HTTP ${http_code})" >&2
    head -c 2000 "${put_out}" >&2 || true
    exit 1
fi
echo "Space '${SPACE_ID}' now defaults to solution '${SOLUTION}'."
