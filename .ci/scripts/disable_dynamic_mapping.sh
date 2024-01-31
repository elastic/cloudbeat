#!/bin/bash

# Use the provided value of ELASTICSEARCH_URL, or set a default value if not provided
ELASTICSEARCH_URL=${1:-http://localhost:9200}

curl -s -XPUT "$ELASTICSEARCH_URL/_index_template/common_template" -H 'Content-Type: application/json' -d '{
  "index_patterns": ["logs-cloud_security_posture.findings*"],
  "priority": 500,
  "template": {
    "mappings": {
      "dynamic": false,
      "properties": {
        "resource": {
          "dynamic": false,
          "properties": {
            "type": { "type": "keyword" }
          }
        },
        "result": {
          "dynamic": false,
          "properties": {
            "evaluation": { "type": "keyword" }
          }
        },
        "rule": {
          "dynamic": false,
          "properties": {
            "benchmark": {
              "dynamic": false,
              "properties": {
                "id": { "type": "keyword" }
              }
            },
            "tags": { "type": "keyword" }
          }
        },
        "agent": {
          "dynamic": false,
          "properties": {
            "id": { "type": "keyword" },
            "version": { "type": "keyword" }
          }
        },
        "@timestamp": { "type": "date" }
      }
    }
  }
}'
