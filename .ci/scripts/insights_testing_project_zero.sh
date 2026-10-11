#!/usr/bin/env bash
#
# Load the "Insights Testing" (project-zero) dataset into a freshly created
# environment: import + enable the "Endpoint Security [Insights]" detection
# rule, create its ingest pipeline, then bulk-load the vendored endpoint
# event/alert episode fixtures (ep1-ep8, plus noise1/noise2 if present).
#
# This is the non-interactive, CI-safe rewrite of the original local dev
# script: no prompts, no hardcoded hosts/credentials (all connection info
# comes from required env vars), every HTTP call's status is checked and a
# failure aborts the run, and every identifier placed into an index name or
# URL is either a fixed literal or validated as digits-only before use.
#
# Required: KIBANA_URL, ES_URL, ES_USER, ES_PASSWORD, DATA_DIR
#   DATA_DIR must contain (as synced from the project-zero S3 dataset):
#     attacks/mapping.json
#     attacks/endpoint_alert.ndjson
#     attacks/ep<N>data.ndjson[.gz]   / ep<N>alerts.ndjson[.gz]   (N = 1..8)
#     noise/noise<N>data.ndjson       / noise<N>alerts.ndjson     (N = 1,2; optional)
#
# Requires: jq, curl

set -euo pipefail

command -v jq >/dev/null 2>&1 || {
    echo "insights_testing_project_zero.sh: jq is required" >&2
    exit 1
}

: "${KIBANA_URL:?KIBANA_URL must be set}"
: "${ES_URL:?ES_URL must be set}"
: "${ES_USER:?ES_USER must be set}"
: "${ES_PASSWORD:?ES_PASSWORD must be set}"
: "${DATA_DIR:?DATA_DIR must be set}"

readonly KIBANA_BASE="${KIBANA_URL%/}"
readonly ES_BASE="${ES_URL%/}"
readonly ATTACKS_DIR="${DATA_DIR%/}/attacks"
readonly NOISE_DIR="${DATA_DIR%/}/noise"
readonly MAPPING_FILE="${ATTACKS_DIR}/mapping.json"
readonly RULE_FILE="${ATTACKS_DIR}/endpoint_alert.ndjson"
readonly PIPELINE_NAME="insights_pipeline"
readonly RULE_NAME="Endpoint Security [Insights]"
readonly CURL_COMMON=(--connect-timeout 10 --max-time 120 -sS)
INDEX_SUFFIX="$(date -u +%Y.%m.%d)"
readonly INDEX_SUFFIX

WORK_DIR="$(mktemp -d)"
trap 'rm -rf "${WORK_DIR}"' EXIT

[[ -f "${MAPPING_FILE}" ]] || { echo "Mapping file not found: ${MAPPING_FILE}" >&2; exit 1; }
[[ -f "${RULE_FILE}" ]] || { echo "Rule file not found: ${RULE_FILE}" >&2; exit 1; }

# Credentials are passed via curl's -u flag (same convention as
# install_entity_risk.sh in this repo) and never echoed or written to logs.
kibana_curl() {
    curl "${CURL_COMMON[@]}" -u "${ES_USER}:${ES_PASSWORD}" -H 'kbn-xsrf: true' "$@"
}

es_curl() {
    curl "${CURL_COMMON[@]}" -u "${ES_USER}:${ES_PASSWORD}" "$@"
}

import_and_enable_rule() {
    echo "Checking for existing rule: ${RULE_NAME}"
    local find_out="${WORK_DIR}/rule_find.json"
    local http_code

    http_code=$(kibana_curl -o "${find_out}" -w "%{http_code}" -G \
        "${KIBANA_BASE}/api/detection_engine/rules/_find" \
        --data-urlencode "page=1" --data-urlencode "per_page=5" \
        --data-urlencode "filter=alert.attributes.name:\"${RULE_NAME}\"")
    if [[ "${http_code}" != "200" ]]; then
        echo "Rule lookup failed (HTTP ${http_code})" >&2
        head -c 2000 "${find_out}" >&2 || true
        exit 1
    fi

    local total
    total=$(jq -r '.total // 0' "${find_out}")

    if [[ "${total}" -eq 0 ]]; then
        echo "Rule not found; importing from $(basename "${RULE_FILE}")"
        local import_out="${WORK_DIR}/rule_import.json"
        http_code=$(kibana_curl -o "${import_out}" -w "%{http_code}" -X POST \
            "${KIBANA_BASE}/api/detection_engine/rules/_import" \
            -H 'Content-Type: multipart/form-data' --form "file=@${RULE_FILE}")
        if [[ "${http_code}" != "200" ]]; then
            echo "Rule import failed (HTTP ${http_code})" >&2
            head -c 2000 "${import_out}" >&2 || true
            exit 1
        fi

        http_code=$(kibana_curl -o "${find_out}" -w "%{http_code}" -G \
            "${KIBANA_BASE}/api/detection_engine/rules/_find" \
            --data-urlencode "page=1" --data-urlencode "per_page=5" \
            --data-urlencode "filter=alert.attributes.name:\"${RULE_NAME}\"")
        if [[ "${http_code}" != "200" ]]; then
            echo "Rule lookup after import failed (HTTP ${http_code})" >&2
            exit 1
        fi
        total=$(jq -r '.total // 0' "${find_out}")
        if [[ "${total}" -eq 0 ]]; then
            echo "Rule still not found after import" >&2
            exit 1
        fi
    fi

    local rule_id enabled
    rule_id=$(jq -r '.data[0].id' "${find_out}")
    enabled=$(jq -r '.data[0].enabled' "${find_out}")

    if [[ "${enabled}" == "true" ]]; then
        echo "Rule already enabled."
        return 0
    fi

    echo "Enabling rule ${rule_id}"
    local enable_out="${WORK_DIR}/rule_enable.json"
    http_code=$(kibana_curl -o "${enable_out}" -w "%{http_code}" -X POST \
        "${KIBANA_BASE}/api/detection_engine/rules/_bulk_action" \
        -H 'Content-Type: application/json' \
        -d "$(jq -n --arg id "${rule_id}" '{action:"enable", ids:[$id]}')")
    if [[ "${http_code}" != "200" ]]; then
        echo "Rule enable failed (HTTP ${http_code})" >&2
        head -c 2000 "${enable_out}" >&2 || true
        exit 1
    fi
    echo "Rule enabled."
}

create_pipeline() {
    if es_curl -o /dev/null -f "${ES_BASE}/_ingest/pipeline/${PIPELINE_NAME}" 2>/dev/null; then
        echo "Ingest pipeline ${PIPELINE_NAME} already exists."
        return 0
    fi

    echo "Creating ingest pipeline ${PIPELINE_NAME}"
    local body_file="${WORK_DIR}/pipeline_body.json"
    cat >"${body_file}" <<'EOF'
{
  "description": "Insights testing (project-zero) ingest pipeline",
  "processors": [
    { "date": { "field": "@timestamp", "formats": ["ISO8601"], "output_format": "yyyy-MM-dd'T'HH:mm:ss.SSSSSSSSSXXX" } },
    { "set": { "field": "event.ingested", "value": "{{_ingest.timestamp}}" } },
    { "set": { "field": "@timestamp", "value": "{{_ingest.timestamp}}" } }
  ]
}
EOF

    local out="${WORK_DIR}/pipeline_out.json"
    local http_code
    http_code=$(es_curl -o "${out}" -w "%{http_code}" -X PUT "${ES_BASE}/_ingest/pipeline/${PIPELINE_NAME}" \
        -H 'Content-Type: application/json' --data-binary "@${body_file}")
    if [[ "${http_code}" != "200" ]]; then
        echo "Pipeline creation failed (HTTP ${http_code})" >&2
        head -c 2000 "${out}" >&2 || true
        exit 1
    fi
}

create_index() {
    local index_name=$1

    if es_curl -o /dev/null -f "${ES_BASE}/${index_name}" 2>/dev/null; then
        echo "Index ${index_name} already exists; skipping create."
        return 0
    fi

    echo "Creating index ${index_name}"
    local body_file="${WORK_DIR}/create_index_body.json"
    jq -n --slurpfile mapping "${MAPPING_FILE}" \
        '{settings: {"index.mapping.total_fields.limit": "6000"}, mappings: $mapping[0]}' \
        >"${body_file}"

    local out="${WORK_DIR}/create_index_out.json"
    local http_code
    http_code=$(es_curl -o "${out}" -w "%{http_code}" -X PUT "${ES_BASE}/${index_name}" \
        -H 'Content-Type: application/json' --data-binary "@${body_file}")
    if [[ "${http_code}" != "200" ]]; then
        echo "Index creation failed for ${index_name} (HTTP ${http_code})" >&2
        head -c 2000 "${out}" >&2 || true
        exit 1
    fi
}

# index_name is always a value we constructed ourselves (fixed prefix + a
# digit-validated episode number + a UTC date), never attacker-influenced,
# so interpolating it into the awk program here is safe.
bulk_load_file() {
    local src_file=$1
    local index_name=$2

    local bulk_file="${WORK_DIR}/bulk.ndjson"
    if [[ "${src_file}" == *.gz ]]; then
        zcat "${src_file}"
    else
        cat "${src_file}"
    fi | awk -v idx="${index_name}" \
        'length($0) > 0 { print "{\"index\":{\"_index\":\"" idx "\"}}"; print }' \
        >"${bulk_file}"

    if [[ ! -s "${bulk_file}" ]]; then
        echo "No documents extracted from $(basename "${src_file}"); skipping." >&2
        return 0
    fi

    echo "Bulk loading $(basename "${src_file}") into ${index_name}"
    local out="${WORK_DIR}/bulk_out.json"
    local http_code
    http_code=$(es_curl -o "${out}" -w "%{http_code}" -X POST "${ES_BASE}/_bulk?pipeline=${PIPELINE_NAME}" \
        -H 'Content-Type: application/x-ndjson' --data-binary "@${bulk_file}")
    if [[ "${http_code}" != "200" ]]; then
        echo "Bulk load failed for $(basename "${src_file}") (HTTP ${http_code})" >&2
        head -c 2000 "${out}" >&2 || true
        exit 1
    fi
    if jq -e '.errors == true' "${out}" >/dev/null 2>&1; then
        echo "Bulk load reported item errors for $(basename "${src_file}"):" >&2
        jq -c '.items[] | select(.index.error != null) | .index.error' "${out}" | head -20 >&2
        exit 1
    fi
    echo "Loaded $(basename "${src_file}") successfully."
}

load_episode() {
    local ep_num=$1
    [[ "${ep_num}" =~ ^[0-9]+$ ]] || { echo "Invalid episode number: ${ep_num}" >&2; exit 1; }

    local data_index="logs-endpoint.events.insights.ep${ep_num}.${INDEX_SUFFIX}"
    local alerts_index="insights-alerts-ep${ep_num}-${INDEX_SUFFIX}"

    local data_file alerts_file
    data_file=$(find "${ATTACKS_DIR}" -maxdepth 1 -type f \
        \( -name "ep${ep_num}data.ndjson" -o -name "ep${ep_num}data.ndjson.gz" \) | head -1)
    alerts_file=$(find "${ATTACKS_DIR}" -maxdepth 1 -type f \
        \( -name "ep${ep_num}alerts.ndjson" -o -name "ep${ep_num}alerts.ndjson.gz" \) | head -1)

    if [[ -z "${data_file}" || -z "${alerts_file}" ]]; then
        echo "Episode ${ep_num}: data or alerts file not found under ${ATTACKS_DIR}" >&2
        exit 1
    fi

    create_index "${data_index}"
    bulk_load_file "${data_file}" "${data_index}"

    create_index "${alerts_index}"
    bulk_load_file "${alerts_file}" "${alerts_index}"
}

load_noise() {
    if [[ ! -d "${NOISE_DIR}" ]]; then
        echo "No noise directory found; skipping noise fixtures."
        return 0
    fi

    local noise_data_index="logs-endpoint.events.insights.noise.${INDEX_SUFFIX}"
    local noise_alerts_index="insights-alerts-noise-${INDEX_SUFFIX}"
    local found=0 f

    for f in "${NOISE_DIR}"/*data.ndjson; do
        [[ -e "${f}" ]] || continue
        [[ "${found}" -eq 0 ]] && create_index "${noise_data_index}"
        bulk_load_file "${f}" "${noise_data_index}"
        found=1
    done
    for f in "${NOISE_DIR}"/*alerts.ndjson; do
        [[ -e "${f}" ]] || continue
        create_index "${noise_alerts_index}"
        bulk_load_file "${f}" "${noise_alerts_index}"
        found=1
    done

    [[ "${found}" -eq 1 ]] || echo "No noise fixtures found; skipping."
}

echo "=== Insights Testing (project-zero): detection rule ==="
import_and_enable_rule

echo "=== Insights Testing (project-zero): ingest pipeline ==="
create_pipeline

echo "=== Insights Testing (project-zero): episodes ==="
for ep in 1 2 3 4 5 6 7 8; do
    load_episode "${ep}"
done

echo "=== Insights Testing (project-zero): noise ==="
load_noise

echo "Insights Testing (project-zero) data load complete."
