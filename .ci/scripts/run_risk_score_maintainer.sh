#!/usr/bin/env bash
#
# Init entity maintainers (idempotent), then POST risk-score maintainer sync
# run (internal API). Split out of install_entity_risk.sh; assumes Entity
# Store is already installed and running (see enable_entity_store.sh, which
# must run first) and that org-data-generated alerts already exist to score.
#
# Required: KIBANA_URL, ES_USER, ES_PASSWORD
# Optional: RISK_SCORE_MAINTAINERS_MAX_ATTEMPTS (default 10), RISK_SCORE_MAINTAINERS_SLEEP_SEC (default 10)
#           RISK_SCORE_MAINTAINER_SYNC_TIMEOUT_SEC (default 600) — curl max-time for sync run
#
# Requires: jq

set -euo pipefail

command -v jq >/dev/null 2>&1 || {
    echo "run_risk_score_maintainer.sh: jq is required" >&2
    exit 1
}

: "${KIBANA_URL:?KIBANA_URL must be set}"
: "${ES_USER:?ES_USER must be set}"
: "${ES_PASSWORD:?ES_PASSWORD must be set}"

readonly BASE="${KIBANA_URL%/}"
readonly INIT_URL="${BASE}/internal/security/entity_store/entity_maintainers/init?apiVersion=2"
readonly RUN_URL="${BASE}/internal/security/entity_store/entity_maintainers/run/risk-score?apiVersion=2&sync=true"

_sync_timeout_raw="${RISK_SCORE_MAINTAINER_SYNC_TIMEOUT_SEC:-600}"
if [[ ! "${_sync_timeout_raw}" =~ ^[1-9][0-9]*$ ]]; then
    echo "run_risk_score_maintainer.sh: RISK_SCORE_MAINTAINER_SYNC_TIMEOUT_SEC must be a positive integer (got '${_sync_timeout_raw}')" >&2
    exit 1
fi
readonly SYNC_RUN_TIMEOUT_SEC="${_sync_timeout_raw}"
readonly CURL_SYNC_RUN=(--connect-timeout 10 --max-time "${SYNC_RUN_TIMEOUT_SEC}" -sS)

readonly INTERNAL_HEADERS=(
    -H "Content-Type: application/json"
    -H "kbn-xsrf: true"
    -H "x-elastic-internal-origin: kibana"
)

curl_auth() {
    curl --connect-timeout 10 --max-time 120 -sS -u "${ES_USER}:${ES_PASSWORD}" "$@"
}

post_init_maintainers() {
    echo "POST entity maintainers init…"
    local http_code
    http_code="$(
        curl_auth -X POST "${INTERNAL_HEADERS[@]}" -d '{}' \
            -o "${TMPDIR:-/tmp}/run_risk_score_maintainer_init.json" -w "%{http_code}" "${INIT_URL}" || echo "000"
    )"
    if [[ "${http_code}" != "200" && "${http_code}" != "201" && "${http_code}" != "204" ]]; then
        echo "POST entity maintainers init failed (HTTP ${http_code})" >&2
        head -c 2000 "${TMPDIR:-/tmp}/run_risk_score_maintainer_init.json" 2>/dev/null >&2 || true
        echo >&2
        exit 1
    fi
    echo "Entity maintainers init succeeded (HTTP ${http_code})."
}

run_risk_score_maintainer() {
    local run_out="${TMPDIR:-/tmp}/run_risk_score_maintainer_run.json"
    local max_attempts="${RISK_SCORE_MAINTAINERS_MAX_ATTEMPTS:-10}"
    local sleep_sec="${RISK_SCORE_MAINTAINERS_SLEEP_SEC:-10}"
    local attempt=1 http_code body_preview

    # Sync run executes scoring inline (avoids Task Manager runSoon race after install).
    # HTTP 500 is fail-fast: indicates a server/scoring error, not a transient TM conflict.
    while [[ "${attempt}" -le "${max_attempts}" ]]; do
        echo "Risk-score maintainer sync run: attempt ${attempt}/${max_attempts} POST ${RUN_URL} (timeout ${SYNC_RUN_TIMEOUT_SEC}s)"
        http_code="$(
            curl "${CURL_SYNC_RUN[@]}" -u "${ES_USER}:${ES_PASSWORD}" -X POST "${INTERNAL_HEADERS[@]}" -d '{}' \
                -o "${run_out}" -w "%{http_code}" "${RUN_URL}" || echo "000"
        )"

        if [[ "${http_code}" == "200" || "${http_code}" == "201" || "${http_code}" == "204" ]]; then
            echo "Risk-score maintainer sync run succeeded (HTTP ${http_code})"
            head -c 2000 "${run_out}" 2>/dev/null || true
            echo
            return 0
        fi

        body_preview=$(head -c 800 "${run_out}" 2>/dev/null || true)
        echo "HTTP ${http_code} response preview: ${body_preview}"

        if [[ "${http_code}" == "502" || "${http_code}" == "503" || "${http_code}" == "504" || "${http_code}" == "000" ]]; then
            echo "Transient gateway/connection error; sleeping ${sleep_sec}s before retry"
            sleep "${sleep_sec}"
            attempt=$((attempt + 1))
            continue
        fi

        echo "Risk-score maintainer sync run failed (non-retryable HTTP ${http_code})" >&2
        exit 1
    done

    echo "Risk-score maintainer sync run failed after ${max_attempts} attempts" >&2
    exit 1
}

post_init_maintainers
run_risk_score_maintainer
