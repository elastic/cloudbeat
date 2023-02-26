#!/bin/bash

ELASTICSEARCH_URL=https://snapshot-8-7-manifest-checks.es.us-central1.gcp.foundit.no
INDEX_NAME=logs-cloud_security_posture.findings_latest-default
KIBANA_PASSWORD=GlveuWm1gvheHSo5KlzOzpfZ
KIBANA_AUTH=elastic:${KIBANA_PASSWORD}
MINIMAL_VALUE=200

default_index_count_response="$(curl -X GET \
  --url "${ELASTICSEARCH_URL}/${INDEX_NAME}/_search" \
  -u "${KIBANA_AUTH}" \
  -H 'Cache-Control: no-cache' \
  -H 'Connection: keep-alive' \
  -H 'Content-Type: application/json' \
  -H 'kbn-xsrf: true')"

echo "The default index has $default_index_count_response results"
count=$(echo "${default_index_count_response}" | jq -r '.count')
echo "The latest elastic index has $count results"

if [ "$count" == "null" ] || [ "$count" -lt "$MINIMAL_VALUE" ]; then
  echo "The latest elastic index has less than $MINIMAL_VALUE results"
  exit 1
fi
