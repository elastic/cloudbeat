#!/bin/bash

ELASTIC_AGENT_VERSION=$1

curl -L -O "https://artifacts.elastic.co/downloads/beats/elastic-agent/elastic-agent-${ELASTIC_AGENT_VERSION}-linux-x86_64.tar.gz"
tar xzvf "elastic-agent-${ELASTIC_AGENT_VERSION}-linux-x86_64.tar.gz"
cp ./tests/test_environments/agents/elastic-agent-8-13.yml "elastic-agent-${ELASTIC_AGENT_VERSION}-linux-x86_64/elastic-agent.yml"
cd "elastic-agent-${ELASTIC_AGENT_VERSION}-linux-x86_64" || exit

# Install the Elastic Agent
sudo ./elastic-agent install -f
