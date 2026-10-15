#!/bin/bash
# AlertZero QA data seeder: seeds detection coverage knowledge indicators
# (one per verdict type) for exercising the coverage worker.
# Seed detection coverage KIs for local dev testing.
# Usage: ./seed-coverage-kis.sh [command]
#
# Commands:
#   seed      — discover your rules then seed one KI per verdict type (default)
#   reset     — mark all coverage KIs pending (re-run sweep)
#   clear     — mark all coverage KIs processed (start fresh)
#   status    — list all pending coverage KIs
#   rules     — show your installed rules (disabled / enabled / prebuilt status)

set -euo pipefail

KB_HOST="${KB_HOST:-http://localhost:5601}"
ES_HOST="${ES_HOST:-http://localhost:9200}"
ES_CREDS="${ES_CREDS:-elastic:changeme}"
SPACE="${SPACE:-default}"
INDEX="ai-index-idx-security-investigations"
TS=$(date -u +%Y-%m-%dT%H:%M:%SZ)

cmd="${1:-seed}"

es() { curl -s -u "$ES_CREDS" "$@"; }
kb() { curl -s -u "$ES_CREDS" -H "kbn-xsrf: true" -H "elastic-api-version: 2023-10-31" "$@"; }

ki() {
    local title="$1" content="$2" technique="$3"
    es -X POST "$ES_HOST/$INDEX/_doc" \
        -H "Content-Type: application/json" -d "{
    \"@timestamp\": \"$TS\",
    \"type\": \"security.coverage\",
    \"title\": \"$title\",
    \"content\": \"$content\",
    \"attributes\": { \"status\": \"pending\", \"space_id\": \"$SPACE\", \"technique\": \"$technique\" }
  }" | jq '{id: ._id, result: .result, title: "'"$title"'"}'
}

case "$cmd" in
seed)
    echo "Discovering rules on your stack..."

    # Pick a disabled rule for covered_disabled
    disabled_name=$(kb "$KB_HOST/api/detection_engine/rules/_find?filter=alert.attributes.enabled:false&per_page=1" |
        jq -r '.data[0].name // empty')

    # Pick an enabled rule for covered_enabled
    enabled_name=$(kb "$KB_HOST/api/detection_engine/rules/_find?filter=alert.attributes.enabled:true&per_page=1" |
        jq -r '.data[0].name // empty')

    # Check if there are uninstalled prebuilt rules
    prebuilt_not_installed=$(kb "$KB_HOST/api/detection_engine/rules/prepackaged/_status" |
        jq -r '.rules_not_installed // 0')

    echo ""
    echo "Seeding 4 coverage KIs into space '$SPACE'..."

    # covered_disabled
    if [ -n "$disabled_name" ]; then
        ki \
            "Detection gap: activity covered by a disabled rule" \
            "Suspicious activity was observed that matches the detection logic of the rule '$disabled_name', but this rule is currently disabled and not generating alerts." \
            "T1071"
    else
        echo "⚠️  No disabled rules found — skipping covered_disabled KI. Disable a rule and re-run."
    fi

    # covered_enabled
    if [ -n "$enabled_name" ]; then
        ki \
            "Detection gap: activity already covered by an enabled rule" \
            "Suspicious activity was observed that matches the detection logic of the rule '$enabled_name', which is currently enabled and actively monitoring for this pattern." \
            "T1110"
    else
        echo "⚠️  No enabled rules found — skipping covered_enabled KI. Enable a rule and re-run."
    fi

    # prebuilt_available
    if [ "$prebuilt_not_installed" -gt 0 ]; then
        ki \
            "No rule detecting LSASS memory access for credential dumping" \
            "Suspicious access to lsass.exe process memory observed. No rule is currently installed to detect this credential dumping technique, but Elastic provides a prebuilt rule for it." \
            "T1003.001"
    else
        echo "⚠️  All prebuilt rules already installed — skipping prebuilt_available KI."
    fi

    # no_coverage — always seed this one, obscure enough to reliably get no_coverage
    ki \
        "Kubernetes audit log: exec into privileged pod" \
        "kubectl exec was used to open an interactive shell in a pod running with hostPID: true. No detection rule covers this activity." \
        "T1610"

    echo ""
    echo "Done. Trigger system-security-coverage-worker in /app/alertzero."
    ;;

reset)
    echo "Resetting all coverage KIs to pending..."
    es -X POST "$ES_HOST/$INDEX/_update_by_query" \
        -H "Content-Type: application/json" -d '{
      "query": { "term": { "type": "security.coverage" } },
      "script": { "source": "ctx._source.attributes.status = \"pending\"" }
    }' | jq '{updated: .updated}'
    ;;

clear)
    echo "Marking all coverage KIs processed..."
    es -X POST "$ES_HOST/$INDEX/_update_by_query" \
        -H "Content-Type: application/json" -d '{
      "query": { "term": { "type": "security.coverage" } },
      "script": { "source": "ctx._source.attributes.status = \"processed\"" }
    }' | jq '{updated: .updated}'
    ;;

status)
    echo "Pending coverage KIs:"
    es "$ES_HOST/$INDEX/_search" \
        -H "Content-Type: application/json" -d '{
      "query": {
        "bool": {
          "filter": [
            { "term": { "type": "security.coverage" } },
            { "term": { "attributes.status": "pending" } }
          ]
        }
      }
    }' | jq '.hits.hits[]._source | {title, status: .attributes.status, technique: .attributes.technique}'
    ;;

rules)
    echo "=== Disabled rules ==="
    kb "$KB_HOST/api/detection_engine/rules/_find?filter=alert.attributes.enabled:false&per_page=10" |
        jq '.data[] | {name, rule_id}'
    echo "=== Enabled rules (first 10) ==="
    kb "$KB_HOST/api/detection_engine/rules/_find?filter=alert.attributes.enabled:true&per_page=10" |
        jq '.data[] | {name, rule_id}'
    echo "=== Prebuilt rule status ==="
    kb "$KB_HOST/api/detection_engine/rules/prepackaged/_status" |
        jq '{rules_installed, rules_not_installed}'
    ;;

*)
    echo "Unknown command: $cmd"
    echo "Usage: $0 [seed|reset|clear|status|rules]"
    exit 1
    ;;
esac
