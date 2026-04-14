#!/usr/bin/env bash
#
# POST Kibana internal entity_store maintainers run for risk score (apiVersion=2).
#
# Required: KIBANA_URL, ES_USER, ES_PASSWORD
# Optional: RISK_SCORE_MAINTAINERS_MAX_ATTEMPTS (default 10), RISK_SCORE_MAINTAINERS_SLEEP_SEC (default 10)
#
# Internal routes require x-elastic-internal-origin (see tests/fleet_api/kibana_settings.py).

set -euo pipefail

: "${KIBANA_URL:?KIBANA_URL must be set}"
: "${ES_USER:?ES_USER must be set}"
: "${ES_PASSWORD:?ES_PASSWORD must be set}"

readonly URL="${KIBANA_URL%/}/internal/security/entity_store/entity_maintainers/run/risk-score?apiVersion=2"
readonly CURL_COMMON=(--connect-timeout 10 --max-time 120 -sS)
readonly MAX_ATTEMPTS="${RISK_SCORE_MAINTAINERS_MAX_ATTEMPTS:-10}"
readonly SLEEP_SEC="${RISK_SCORE_MAINTAINERS_SLEEP_SEC:-10}"

attempt=1
while [[ "$attempt" -le "$MAX_ATTEMPTS" ]]; do
    echo "Entity store risk-score maintainers run: attempt ${attempt}/${MAX_ATTEMPTS} POST ${URL}"
    http_code=$(
        curl "${CURL_COMMON[@]}" \
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

    if [[ "$http_code" == "200" || "$http_code" == "201" || "$http_code" == "204" ]]; then
        echo "Entity store risk-score maintainers run succeeded (HTTP ${http_code})"
        head -c 2000 /tmp/entity_store_risk_score_maintainers.json 2>/dev/null || true
        echo
        exit 0
    fi

    body_preview=$(head -c 800 /tmp/entity_store_risk_score_maintainers.json 2>/dev/null || true)
    echo "HTTP ${http_code} response preview: ${body_preview}"

    if [[ "$http_code" == "502" || "$http_code" == "503" || "$http_code" == "504" || "$http_code" == "000" ]]; then
        echo "Transient error; sleeping ${SLEEP_SEC}s before retry"
        sleep "${SLEEP_SEC}"
        attempt=$((attempt + 1))
        continue
    fi

    echo "Entity store risk-score maintainers run failed (non-retryable HTTP ${http_code})" >&2
    exit 1
done

echo "Entity store risk-score maintainers run failed after ${MAX_ATTEMPTS} attempts" >&2
exit 1
