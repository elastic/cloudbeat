#!/bin/bash

ELASTIC_AGENT_VERSION=$1

curl -L -O "https://artifacts.elastic.co/downloads/beats/elastic-agent/elastic-agent-${ELASTIC_AGENT_VERSION}-linux-x86_64.tar.gz"
tar xzvf "elastic-agent-${ELASTIC_AGENT_VERSION}-linux-x86_64.tar.gz"
cd "elastic-agent-${ELASTIC_AGENT_VERSION}-linux-x86_64" || exit

# TODO: Replace elastic-agent.yml with the correct configuration file
# cp /path/to/elastic-agent.yml /path/to/elastic-agent-${ELASTIC_AGENT_VERSION}-linux-x86_64/elastic-agent.yml

# Install the Elastic Agent
# sudo ./elastic-agent install -f
