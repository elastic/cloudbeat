#!/bin/bash

curl -L -O https://staging.elastic.co/8.15.0-3112d02f/downloads/beats/elastic-agent/elastic-agent-8.15.0-linux-x86_64.tar.gz
tar xzvf elastic-agent-8.15.0-linux-x86_64.tar.gz
cd elastic-agent-8.15.0-linux-x86_64
sudo ./elastic-agent install -f --url=https://45ab75f770e103f4a81ca2e2886ebbd7.fleet.us-west2.gcp.elastic-cloud.com:443 --enrollment-token=OEswUHVKRUJlX1AwTC1YTlNTU246LWNOWW9uXzFUUi1fVEtQZlpvSFlvQQ==