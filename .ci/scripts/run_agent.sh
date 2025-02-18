#!/bin/bash

ELASTIC_AGENT_VERSION=$1
PACKAGE_VERSION=$2

curl -L -O "https://artifacts.elastic.co/downloads/beats/elastic-agent/elastic-agent-${ELASTIC_AGENT_VERSION}-linux-x86_64.tar.gz"
tar xzvf "elastic-agent-${ELASTIC_AGENT_VERSION}-linux-x86_64.tar.gz"

# For debugging purposes, copy the original configuration file
cp "elastic-agent-${ELASTIC_AGENT_VERSION}-linux-x86_64/elastic-agent.yml" "elastic-agent-${ELASTIC_AGENT_VERSION}-linux-x86_64/elastic-agent-orig.yml"
# Copy customized configuration file
cp ./tests/test_environments/agents/elastic-agent-standalone.yml "elastic-agent-${ELASTIC_AGENT_VERSION}-linux-x86_64/elastic-agent.yml"

# Array of placeholders and corresponding environment variable names
placeholders=("ES_HOST" "ES_USERNAME" "ES_PASSWORD" "AWS_ACCESS_KEY_ID" "AWS_SECRET_ACCESS_KEY" "GOOGLE_APPLICATION_CREDENTIALS" "AZURE_CLIENT_ID" "AZURE_TENANT_ID" "AZURE_CLIENT_SECRET" "PACKAGE_VERSION")
env_variables=("$ES_HOST" "$ES_USERNAME" "$ES_PASSWORD" "$AWS_ACCESS_KEY_ID" "$AWS_SECRET_ACCESS_KEY" "$GOOGLE_APPLICATION_CREDENTIALS" "$AZURE_CLIENT_ID" "$AZURE_TENANT_ID" "$AZURE_CLIENT_SECRET" "$PACKAGE_VERSION")

agent_config="./elastic-agent-${ELASTIC_AGENT_VERSION}-linux-x86_64/elastic-agent.yml"
# Iterate through each placeholder and replace it with the corresponding environment variable value
for ((i = 0; i < ${#placeholders[@]}; i++)); do
    placeholder="${placeholders[$i]}"
    env_variable="${env_variables[$i]}"
    awk -v placeholder="$placeholder" -v env_variable="$env_variable" '{gsub("\\${" placeholder "}", env_variable)} 1' "$agent_config" >tmp && mv tmp "$agent_config"
done

cd "elastic-agent-${ELASTIC_AGENT_VERSION}-linux-x86_64" || exit

install_servers_opt=""
major_version=$(echo "${ELASTIC_AGENT_VERSION}" | cut -d'.' -f1)
if [ "$major_version" -ge 9 ]; then
    install_servers_opt="--install-servers"
fi

# Install the Elastic Agent
sudo ./elastic-agent install -f ${install_servers_opt:+$install_servers_opt}
