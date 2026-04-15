#!/usr/bin/env bash
#
# Ensure Entity Store v2 is installed (public API), poll until running, init
# maintainers, then POST risk-score maintainer run (internal API).
#
# Required: KIBANA_URL, ES_USER, ES_PASSWORD
# Optional: ENTITY_STORE_STATUS_TIMEOUT_SEC (default 180), ENTITY_STORE_POLL_INTERVAL_SEC (default 5)
#           RISK_SCORE_MAINTAINERS_MAX_ATTEMPTS (default 10), RISK_SCORE_MAINTAINERS_SLEEP_SEC (default 10)
#
# Requires: jq

set -euo pipefail

command -v jq >/dev/null 2>&1 || {
    echo "install_entity_risk.sh: jq is required" >&2
    exit 1
}

: "${KIBANA_URL:?KIBANA_URL must be set}"
: "${ES_USER:?ES_USER must be set}"
: "${ES_PASSWORD:?ES_PASSWORD must be set}"

readonly BASE="${KIBANA_URL%/}"
readonly STATUS_URL="${BASE}/api/security/entity_store/status?apiVersion=2023-10-31"
readonly INSTALL_URL="${BASE}/api/security/entity_store/install?apiVersion=2023-10-31"
readonly INIT_URL="${BASE}/internal/security/entity_store/entity_maintainers/init?apiVersion=2"
readonly RUN_URL="${BASE}/internal/security/entity_store/entity_maintainers/run/risk-score?apiVersion=2"

readonly STATUS_FILE="${TMPDIR:-/tmp}/install_entity_risk_status.json"
readonly CURL_COMMON=(--connect-timeout 10 --max-time 120 -sS)
readonly POLL_TIMEOUT="${ENTITY_STORE_STATUS_TIMEOUT_SEC:-180}"
readonly POLL_INTERVAL="${ENTITY_STORE_POLL_INTERVAL_SEC:-5}"

readonly PUBLIC_HEADERS=(
    -H "Content-Type: application/json"
    -H "kbn-xsrf: true"
)
readonly INTERNAL_HEADERS=(
    -H "Content-Type: application/json"
    -H "kbn-xsrf: true"
    -H "x-elastic-internal-origin: kibana"
)

curl_auth() {
    curl "${CURL_COMMON[@]}" -u "${ES_USER}:${ES_PASSWORD}" "$@"
}

fetch_status_json() {
    local http_code
    http_code="$(
        curl_auth "${PUBLIC_HEADERS[@]}" -o "${STATUS_FILE}" -w "%{http_code}" "${STATUS_URL}" || echo "000"
    )"
    if [[ "${http_code}" != "200" ]]; then
        echo "GET entity store status failed (HTTP ${http_code})" >&2
        head -c 2000 "${STATUS_FILE}" 2>/dev/null >&2 || true
        echo >&2
        exit 1
    fi
}

global_status() {
    jq -r '.status // empty' "${STATUS_FILE}"
}

is_fully_started() {
    local s bad
    s=$(global_status)
    [[ "${s}" == "running" ]] || return 1
    bad=$(jq '[.engines[]? | select(.status != "started")] | length' "${STATUS_FILE}")
    [[ "${bad}" == "0" ]]
}

poll_until_started() {
    local deadline
    deadline=$(($(date +%s) + POLL_TIMEOUT))
    echo "Polling entity store status (timeout ${POLL_TIMEOUT}s, interval ${POLL_INTERVAL}s)…"
    while (($(date +%s) < deadline)); do
        fetch_status_json
        case "$(global_status)" in
        error)
            echo "Entity store status is error. Response preview:" >&2
            head -c 4000 "${STATUS_FILE}" >&2
            echo >&2
            exit 1
            ;;
        esac
        if is_fully_started; then
            echo "Entity store is running and engines are started."
            return 0
        fi
        echo "Entity store not ready yet (status=$(global_status)); sleeping ${POLL_INTERVAL}s…"
        sleep "${POLL_INTERVAL}"
    done
    echo "Timed out waiting for entity store to be fully started." >&2
    head -c 2000 "${STATUS_FILE}" >&2
    echo >&2
    exit 1
}

post_install_if_needed() {
    fetch_status_json
    local s
    s=$(global_status)
    echo "Entity store global status: ${s}"

    case "${s}" in
    error)
        echo "Entity store status is error. Response preview:" >&2
        head -c 4000 "${STATUS_FILE}" >&2
        echo >&2
        exit 1
        ;;
    stopped)
        echo "Entity store status is stopped; install path not attempted. Response preview:" >&2
        head -c 2000 "${STATUS_FILE}" >&2
        echo >&2
        exit 1
        ;;
    esac

    if is_fully_started; then
        echo "Entity store already running; skipping install."
        return 0
    fi

    if [[ "${s}" == "not_installed" ]]; then
        echo "POST entity store install…"
        local http_code
        http_code="$(
            curl_auth -X POST "${PUBLIC_HEADERS[@]}" -d '{}' \
                -o "${TMPDIR:-/tmp}/install_entity_risk_install.json" -w "%{http_code}" "${INSTALL_URL}" || echo "000"
        )"
        if [[ "${http_code}" != "200" && "${http_code}" != "201" && "${http_code}" != "204" ]]; then
            echo "POST install failed (HTTP ${http_code})" >&2
            head -c 2000 "${TMPDIR:-/tmp}/install_entity_risk_install.json" 2>/dev/null >&2 || true
            echo >&2
            exit 1
        fi
    fi

    poll_until_started
}

post_init_maintainers() {
    echo "POST entity maintainers init…"
    local http_code
    http_code="$(
        curl_auth -X POST "${INTERNAL_HEADERS[@]}" -d '{}' \
            -o "${TMPDIR:-/tmp}/install_entity_risk_init.json" -w "%{http_code}" "${INIT_URL}" || echo "000"
    )"
    if [[ "${http_code}" != "200" && "${http_code}" != "201" && "${http_code}" != "204" ]]; then
        echo "POST entity maintainers init failed (HTTP ${http_code})" >&2
        head -c 2000 "${TMPDIR:-/tmp}/install_entity_risk_init.json" 2>/dev/null >&2 || true
        echo >&2
        exit 1
    fi
    echo "Entity maintainers init succeeded (HTTP ${http_code})."
}

run_risk_score_maintainer() {
    local run_out="${TMPDIR:-/tmp}/install_entity_risk_run.json"
    local max_attempts="${RISK_SCORE_MAINTAINERS_MAX_ATTEMPTS:-10}"
    local sleep_sec="${RISK_SCORE_MAINTAINERS_SLEEP_SEC:-10}"
    local attempt=1 http_code body_preview

    while [[ "${attempt}" -le "${max_attempts}" ]]; do
        echo "Risk-score maintainer run: attempt ${attempt}/${max_attempts} POST ${RUN_URL}"
        http_code="$(
            curl_auth -X POST "${INTERNAL_HEADERS[@]}" -d '{}' \
                -o "${run_out}" -w "%{http_code}" "${RUN_URL}" || echo "000"
        )"

        if [[ "${http_code}" == "200" || "${http_code}" == "201" || "${http_code}" == "204" ]]; then
            echo "Risk-score maintainer run succeeded (HTTP ${http_code})"
            head -c 2000 "${run_out}" 2>/dev/null || true
            echo
            return 0
        fi

        body_preview=$(head -c 800 "${run_out}" 2>/dev/null || true)
        echo "HTTP ${http_code} response preview: ${body_preview}"

        if [[ "${http_code}" == "502" || "${http_code}" == "503" || "${http_code}" == "504" || "${http_code}" == "000" ]]; then
            echo "Transient error; sleeping ${sleep_sec}s before retry"
            sleep "${sleep_sec}"
            attempt=$((attempt + 1))
            continue
        fi

        echo "Risk-score maintainer run failed (non-retryable HTTP ${http_code})" >&2
        exit 1
    done

    echo "Risk-score maintainer run failed after ${max_attempts} attempts" >&2
    exit 1
}

post_install_if_needed
post_init_maintainers
run_risk_score_maintainer
