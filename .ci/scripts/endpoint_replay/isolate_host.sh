#!/usr/bin/env bash
#
# Isolate the Endpoint Replay VM's enrolled agent via the Isolate Host
# response action, so the environment demonstrates the full replay ->
# detection -> isolate flow without relying on an AI-generated Attack
# Discovery proposal (non-deterministic, needs a configured GenAI connector,
# too flaky for an unattended CI step).
#
# Required: KIBANA_URL, ES_USER, ES_PASSWORD, ENDPOINT_ID
#
# Note: isolation is applied asynchronously by the agent after it picks up
# the action; a 200 here means the action was accepted, not that the host is
# isolated yet.

set -euo pipefail

: "${KIBANA_URL:?KIBANA_URL must be set}"
: "${ES_USER:?ES_USER must be set}"
: "${ES_PASSWORD:?ES_PASSWORD must be set}"
: "${ENDPOINT_ID:?ENDPOINT_ID must be set}"

readonly KIBANA_BASE="${KIBANA_URL%/}"
readonly CURL_COMMON=(--connect-timeout 10 --max-time 60 -sS)

WORK_DIR="$(mktemp -d)"
trap 'rm -rf "${WORK_DIR}"' EXIT

kibana_curl() {
    curl "${CURL_COMMON[@]}" -u "${ES_USER}:${ES_PASSWORD}" -H 'kbn-xsrf: true' "$@"
}

echo "Isolating endpoint ${ENDPOINT_ID}"
body="${WORK_DIR}/isolate_body.json"
jq -n --arg id "${ENDPOINT_ID}" \
    '{endpoint_ids: [$id], comment: "Endpoint Replay (project-zero) scenario"}' \
    >"${body}"

out="${WORK_DIR}/isolate_out.json"
http_code=$(kibana_curl -o "${out}" -w "%{http_code}" -X POST "${KIBANA_BASE}/api/endpoint/action/isolate" \
    -H 'Content-Type: application/json' --data-binary "@${body}")
if [[ "${http_code}" != "200" ]]; then
    echo "Isolate host failed (HTTP ${http_code})" >&2
    head -c 2000 "${out}" >&2 || true
    exit 1
fi
echo "Isolate host action accepted for endpoint ${ENDPOINT_ID}."
