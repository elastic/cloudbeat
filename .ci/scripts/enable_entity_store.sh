#!/usr/bin/env bash
#
# Ensure Entity Store v2 is installed (public API) and poll until running.
# Split out of install_entity_risk.sh so it can run independently of the
# risk-score maintainer: Entity Store should be on whenever Entity Analytics
# or Insights Testing data is being loaded, not only when the risk-score
# maintainer (which needs org-data-generated alerts) also runs.
#
# Required: KIBANA_URL, ES_USER, ES_PASSWORD
# Optional: ENTITY_STORE_STATUS_TIMEOUT_SEC (default 180), ENTITY_STORE_POLL_INTERVAL_SEC (default 5)
#
# Requires: jq

set -euo pipefail

command -v jq >/dev/null 2>&1 || {
    echo "enable_entity_store.sh: jq is required" >&2
    exit 1
}

: "${KIBANA_URL:?KIBANA_URL must be set}"
: "${ES_USER:?ES_USER must be set}"
: "${ES_PASSWORD:?ES_PASSWORD must be set}"

readonly BASE="${KIBANA_URL%/}"
readonly STATUS_URL="${BASE}/api/security/entity_store/status?apiVersion=2023-10-31"
readonly INSTALL_URL="${BASE}/api/security/entity_store/install?apiVersion=2023-10-31"

readonly STATUS_FILE="${TMPDIR:-/tmp}/enable_entity_store_status.json"
readonly CURL_COMMON=(--connect-timeout 10 --max-time 120 -sS)
readonly POLL_TIMEOUT="${ENTITY_STORE_STATUS_TIMEOUT_SEC:-180}"
readonly POLL_INTERVAL="${ENTITY_STORE_POLL_INTERVAL_SEC:-5}"

readonly PUBLIC_HEADERS=(
    -H "Content-Type: application/json"
    -H "kbn-xsrf: true"
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
                -o "${TMPDIR:-/tmp}/enable_entity_store_install.json" -w "%{http_code}" "${INSTALL_URL}" || echo "000"
        )"
        if [[ "${http_code}" != "200" && "${http_code}" != "201" && "${http_code}" != "204" ]]; then
            echo "POST install failed (HTTP ${http_code})" >&2
            head -c 2000 "${TMPDIR:-/tmp}/enable_entity_store_install.json" 2>/dev/null >&2 || true
            echo >&2
            exit 1
        fi
    fi

    poll_until_started
}

post_install_if_needed
