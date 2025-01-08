#!/bin/bash

set -e
echo "Waiting for project to be available..."
sleep_timeout=15
for i in {1..20}; do
    response=$(curl -s -H "Content-type: application/json" -H "Authorization: ApiKey ${API_KEY}" "${EC_URL}/api/v1/serverless/projects/security/${PROJECT_ID}/status")
    phase=$(echo $response | jq -r '.phase')
    if [ "$phase" = "initialized" ]; then
        echo "Project is available!"
        exit 0
    else
        echo "Project phase is '$phase'. Waiting..."
        sleep $sleep_timeout
    fi
done
echo "Project is not available after retries."
exit 1