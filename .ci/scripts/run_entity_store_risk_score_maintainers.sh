#!/usr/bin/env bash
#
# POST Kibana internal entity_store maintainers run for risk score (apiVersion=2).
#
# Required: KIBANA_URL, ES_USER, ES_PASSWORD
# Optional:
#   RISK_SCORE_MAINTAINERS_MAX_WAIT_SEC — total wall time before giving up (default 300 = 5 minutes)
#   RISK_SCORE_MAINTAINERS_SLEEP_SEC — sleep between transient retries (default 15)
#   RISK_SCORE_MAINTAINERS_INITIAL_SLEEP_SEC — wait before first POST (default 30; 0 to skip)
#
# Retries while time remains on 429/500/502/503/504/000 (transient Kibana errors after ML setup).
#
# Internal routes require x-elastic-internal-origin (see tests/fleet_api/kibana_settings.py).

set -euo pipefail

: "${KIBANA_URL:?KIBANA_URL must be set}"
: "${ES_USER:?ES_USER must be set}"
: "${ES_PASSWORD:?ES_PASSWORD must be set}"

readonly URL="${KIBANA_URL%/}/internal/security/entity_store/entity_maintainers/run/risk-score?apiVersion=2"
readonly CONNECT_TIMEOUT=10
readonly MAX_TOTAL_SEC="${RISK_SCORE_MAINTAINERS_MAX_WAIT_SEC:-300}"
readonly SLEEP_SEC="${RISK_SCORE_MAINTAINERS_SLEEP_SEC:-15}"
readonly INITIAL_SLEEP_SEC="${RISK_SCORE_MAINTAINERS_INITIAL_SLEEP_SEC:-30}"

START=$SECONDS

remaining_sec() {
    local used=$((SECONDS - START))
    local r=$((MAX_TOTAL_SEC - used))
    if ((r < 0)); then
        echo 0
    else
        echo "$r"
    fi
}

init="${INITIAL_SLEEP_SEC}"
if [[ "${init}" != "0" ]]; then
    rem=$(remaining_sec)
    if ((init > rem)); then
        init=$rem
    fi
    if ((init > 0)); then
        echo "Waiting ${init}s before risk-score maintainers POST (${MAX_TOTAL_SEC}s total budget)"
        sleep "${init}"
    fi
fi

attempt=1
while true; do
    rem=$(remaining_sec)
    if ((rem <= 0)); then
        echo "Entity store risk-score maintainers run: exceeded ${MAX_TOTAL_SEC}s total budget" >&2
        exit 1
    fi

    # Keep each curl within remaining budget (min 5s, cap 90s).
    mt=$rem
    ((mt > 90)) && mt=90
    ((mt < 5)) && mt=5

    echo "Entity store risk-score maintainers run: attempt ${attempt} (${rem}s left in ${MAX_TOTAL_SEC}s budget) POST ${URL}"
    http_code=$(
        curl --connect-timeout "${CONNECT_TIMEOUT}" --max-time "${mt}" -sS \
            -X POST \
            -u "${ES_USER}:${ES_PASSWORD}" \
            -H "Content-Type: application/json" \
            -H "kbn-xsrf: true" \
            -H "x-elastic-internal-origin: kibana" \
            -d '{}' \
            -o /tmp/entity_store_risk_score_maintainers.json \
            -w "%{http_code}" \
            "${URL}" || echo "000"
    )
    # Strip whitespace so comparisons stay reliable (proxies/shell quirks).
    http_code="${http_code//[[:space:]]/}"

    if [[ "$http_code" == "200" || "$http_code" == "201" || "$http_code" == "204" ]]; then
        echo "Entity store risk-score maintainers run succeeded (HTTP ${http_code})"
        head -c 2000 /tmp/entity_store_risk_score_maintainers.json 2>/dev/null || true
        echo
        exit 0
    fi

    body_preview=$(head -c 800 /tmp/entity_store_risk_score_maintainers.json 2>/dev/null || true)
    echo "HTTP ${http_code} response preview: ${body_preview}"

    if [[ "$http_code" == "429" || "$http_code" == "500" || "$http_code" == "502" || "$http_code" == "503" || "$http_code" == "504" || "$http_code" == "000" ]]; then
        rem=$(remaining_sec)
        if ((rem <= 0)); then
            echo "Entity store risk-score maintainers run: no time left after transient HTTP ${http_code}" >&2
            exit 1
        fi
        # Sleep up to SLEEP_SEC but not past the deadline.
        sl=$SLEEP_SEC
        ((sl > rem)) && sl=$rem
        if ((sl < 1)); then
            echo "Entity store risk-score maintainers run: budget exhausted after transient HTTP ${http_code}" >&2
            exit 1
        fi
        echo "Transient error (HTTP ${http_code}); sleeping ${sl}s before retry"
        sleep "${sl}"
        attempt=$((attempt + 1))
        continue
    fi

    echo "Entity store risk-score maintainers run failed (non-retryable HTTP ${http_code})" >&2
    exit 1
done
