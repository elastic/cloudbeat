#!/usr/bin/env bash
#
# Poll Kibana until ready, then (unless serverless) wait for Fleet setup and EPM stability.
#
# Required environment variables:
#   ES_USER, ES_PASSWORD, KIBANA_URL
# Optional:
#   SERVERLESS_MODE — when "true", skip Fleet checks (Elastic Cloud manages Fleet).

set -euo pipefail

: "${ES_USER:?ES_USER must be set}"
: "${ES_PASSWORD:?ES_PASSWORD must be set}"
: "${KIBANA_URL:?KIBANA_URL must be set}"
SERVERLESS_MODE="${SERVERLESS_MODE:-false}"

# -s: silent; no -f so HTTP 4xx/5xx still return exit 0 and we read %{http_code}.
# --connect-timeout / --max-time: avoid hanging; failed connect uses || echo 000 below.
readonly CURL_COMMON=(--connect-timeout 5 --max-time 20 -s)

wait_kibana_ready() {
    echo "Waiting for Kibana to report 'available' at ${KIBANA_URL}/api/status"
    local kibana_ok=""
    local i
    for i in $(seq 1 30); do
        local http_code
        http_code=$(
            curl "${CURL_COMMON[@]}" -u "${ES_USER}:${ES_PASSWORD}" \
                -o /tmp/kibana_status.json \
                -w "%{http_code}" \
                "${KIBANA_URL}/api/status" || echo "000"
        )
        if [[ "$http_code" == "200" ]]; then
            local level state
            level=$(jq -r '.status.overall.level // empty' /tmp/kibana_status.json 2>/dev/null || true)
            state=$(jq -r '.status.overall.state // empty' /tmp/kibana_status.json 2>/dev/null || true)
            if [[ "$level" == "available" || "$state" == "green" ]]; then
                echo "Kibana is ready (level=${level:-n/a}, state=${state:-n/a})"
                kibana_ok=1
                break
            fi
            echo "attempt $i/30: Kibana not ready yet (level=${level:-n/a}, state=${state:-n/a}), sleeping 10s"
        else
            echo "attempt $i/30: Kibana returned HTTP ${http_code}, sleeping 10s"
        fi
        sleep 10
    done
    if [[ -z "$kibana_ok" ]]; then
        echo "Timed out waiting for Kibana to become ready after 300s"
        return 1
    fi
    return 0
}

wait_fleet_setup() {
    echo "Triggering Fleet setup at ${KIBANA_URL}/api/fleet/setup"
    local fleet_ok=""
    local i
    for i in $(seq 1 30); do
        local http_code
        http_code=$(
            curl "${CURL_COMMON[@]}" \
                -X POST \
                -u "${ES_USER}:${ES_PASSWORD}" \
                -H "Content-Type: application/json" \
                -H "kbn-xsrf: true" \
                -o /tmp/fleet_status.json \
                -w "%{http_code}" \
                "${KIBANA_URL}/api/fleet/setup" || echo "000"
        )
        local body
        if [[ "$http_code" == "200" ]]; then
            local is_initialized
            is_initialized=$(jq -r '.isInitialized // false' /tmp/fleet_status.json 2>/dev/null || true)
            if [[ "$is_initialized" == "true" ]]; then
                echo "Fleet setup complete (isInitialized=true)"
                fleet_ok=1
                break
            fi
            body=$(cat /tmp/fleet_status.json 2>/dev/null | head -c 300 || true)
            echo "attempt $i/30: Fleet setup not complete yet, body: ${body}, sleeping 10s"
        else
            body=$(cat /tmp/fleet_status.json 2>/dev/null | head -c 300 || true)
            echo "attempt $i/30: Fleet setup returned HTTP ${http_code}, body: ${body}, sleeping 10s"
        fi
        sleep 10
    done
    if [[ -z "$fleet_ok" ]]; then
        echo "Timed out waiting for Fleet setup to complete after 300s"
        return 1
    fi
    return 0
}

wait_fleet_epm_stable() {
    echo "Waiting for Fleet Server to be stable at ${KIBANA_URL}/api/fleet/epm/packages"
    local consecutive_ok=0
    local required_ok=5
    local i
    for i in $(seq 1 60); do
        local http_code
        http_code=$(
            curl "${CURL_COMMON[@]}" \
                -u "${ES_USER}:${ES_PASSWORD}" \
                -H "Content-Type: application/json" \
                -H "kbn-xsrf: true" \
                -o /tmp/fleet_epm.json \
                -w "%{http_code}" \
                "${KIBANA_URL}/api/fleet/epm/packages" || echo "000"
        )
        if [[ "$http_code" == "200" ]]; then
            consecutive_ok=$((consecutive_ok + 1))
            echo "attempt $i/60: Fleet Server OK (${consecutive_ok}/${required_ok} consecutive)"
            if [[ "$consecutive_ok" -ge "$required_ok" ]]; then
                echo "Fleet Server is stable"
                break
            fi
        else
            consecutive_ok=0
            local body
            body=$(cat /tmp/fleet_epm.json 2>/dev/null | head -c 300 || true)
            echo "attempt $i/60: Fleet Server returned HTTP ${http_code}, body: ${body}, resetting consecutive counter, sleeping 10s"
        fi
        sleep 10
    done
    if [[ "$consecutive_ok" -lt "$required_ok" ]]; then
        echo "Timed out waiting for Fleet Server to stabilise"
        return 1
    fi
    return 0
}

wait_kibana_ready

if [[ "${SERVERLESS_MODE}" == "true" ]]; then
    echo "Serverless mode: skipping Fleet status check (Fleet is managed by Elastic Cloud)"
    exit 0
fi

wait_fleet_setup
wait_fleet_epm_stable

exit 0
