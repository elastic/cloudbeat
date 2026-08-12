#!/usr/bin/env bash
#
# POST Kibana internal ML module setup for security_auth (after org-data or similar).
#
# Required: KIBANA_URL, ES_USER, ES_PASSWORD
# Optional: ML_SECURITY_AUTH_MAX_ATTEMPTS (default 10), ML_SECURITY_AUTH_SLEEP_SEC (default 10)
# Optional: ML_SECURITY_AUTH_JSON — full JSON body; default:
#   {"prefix":"","indexPatternName":"logs-*","useDedicatedIndex":false,"startDatafeed":true}
#
# Internal routes require x-elastic-internal-origin (see tests/fleet_api/kibana_settings.py).

set -euo pipefail

: "${KIBANA_URL:?KIBANA_URL must be set}"
: "${ES_USER:?ES_USER must be set}"
: "${ES_PASSWORD:?ES_PASSWORD must be set}"

readonly URL="${KIBANA_URL%/}/internal/ml/modules/setup/security_auth"
readonly CURL_COMMON=(--connect-timeout 10 --max-time 120 -sS)
readonly MAX_ATTEMPTS="${ML_SECURITY_AUTH_MAX_ATTEMPTS:-10}"
readonly SLEEP_SEC="${ML_SECURITY_AUTH_SLEEP_SEC:-10}"

if [[ -n "${ML_SECURITY_AUTH_JSON:-}" ]]; then
    BODY="${ML_SECURITY_AUTH_JSON}"
else
    BODY='{"prefix":"","indexPatternName":"logs-*","useDedicatedIndex":false,"startDatafeed":true}'
fi

attempt=1
while [[ "$attempt" -le "$MAX_ATTEMPTS" ]]; do
    echo "ML security_auth module setup: attempt ${attempt}/${MAX_ATTEMPTS} POST ${URL}"
    http_code=$(
        curl "${CURL_COMMON[@]}" \
            -X POST \
            -u "${ES_USER}:${ES_PASSWORD}" \
            -H "Content-Type: application/json" \
            -H "kbn-xsrf: true" \
            -H "x-elastic-internal-origin: kibana" \
            -d "${BODY}" \
            -o /tmp/ml_security_auth_response.json \
            -w "%{http_code}" \
            "${URL}" || echo "000"
    )

    if [[ "$http_code" == "200" || "$http_code" == "201" || "$http_code" == "204" ]]; then
        echo "ML security_auth module setup succeeded (HTTP ${http_code})"
        head -c 2000 /tmp/ml_security_auth_response.json 2>/dev/null || true
        echo
        exit 0
    fi

    body_preview=$(head -c 800 /tmp/ml_security_auth_response.json 2>/dev/null || true)
    echo "HTTP ${http_code} response preview: ${body_preview}"

    if [[ "$http_code" == "502" || "$http_code" == "503" || "$http_code" == "504" || "$http_code" == "000" ]]; then
        echo "Transient error; sleeping ${SLEEP_SEC}s before retry"
        sleep "${SLEEP_SEC}"
        attempt=$((attempt + 1))
        continue
    fi

    echo "ML security_auth module setup failed (non-retryable HTTP ${http_code})" >&2
    exit 1
done

echo "ML security_auth module setup failed after ${MAX_ATTEMPTS} attempts" >&2
exit 1
