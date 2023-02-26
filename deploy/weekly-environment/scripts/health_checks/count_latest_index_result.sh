#!/bin/bash

ELASTICSEARCH_URL=$1
KIBANA_PASSWORD=$2
REQUEST_COUNT=$3
REQUEST_INTERVAL=$4
INDEX_NAME=logs-cloud_security_posture.findings_latest-default
KIBANA_AUTH=elastic:${KIBANA_PASSWORD}
MINIMAL_VALUE=200


for i in $(seq 1 "$REQUEST_COUNT"); do
  response="$(curl -X GET \
    --url "${ELASTICSEARCH_URL}/${INDEX_NAME}/_count" \
    -u "${KIBANA_AUTH}" \
    -H 'Cache-Control: no-cache' \
    -H 'Connection: keep-alive' \
    -H 'Content-Type: application/json' \
    -H 'kbn-xsrf: true')"
  count=$(echo "${response}" | jq -r '.count')
  echo "Request $i: $count results"
  if [ "$count" != "null" ] && [ "$count" -ge "$MINIMAL_VALUE" ]; then
    echo "The latest elastic index has at least $MINIMAL_VALUE results"
    exit 0
  fi
  sleep "$REQUEST_INTERVAL"
done

echo "The latest elastic index has less than $MINIMAL_VALUE results for $REQUEST_COUNT consecutive requests made within $REQUEST_INTERVAL"
exit
