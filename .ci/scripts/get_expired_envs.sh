#!/bin/bash

# Function to check if a file exists in a folder
file_exists() {
    local folder="$1"
    local file="$2"
    if [ -f "$folder/$file" ]; then
        return 0
    fi

    return 1
}

# Get all folders from the specific S3 bucket
s3_bucket="tf-state-bucket-test-infra"
folders=$(aws s3 ls "s3://$s3_bucket" | awk '{print $2}')

# JSON array to store deployment names
deployment_json='[]'

# Iterate over each folder
for folder in $folders; do
    # Check if env_config.json file exists
    if ! file_exists "$folder" "env_config.json"; then
        echo "env_config.json file does not exist in $folder"
        continue
    fi

    # Read expiration date from env_config.json
    expiration=$(jq -r '.expiration' "$folder/env_config.json")
    current_date=$(date +%Y-%m-%d)

    # Compare expiration date with current date
    if [[ ! "$expiration" > "$current_date" ]]; then
        # Read deployment name from env_config.json
        deployment_name=$(jq -r '.deployment_name' "$folder/env_config.json")

        # Add deployment name to JSON object
        deployment_json=$(echo "$deployment_json" | jq --arg value "$deployment_name" '. += [{"deployment_name": $value}]')
    fi
done

# Print the deployment names to GITHUB_OUTPUT
echo "deployments=$(echo "$deployment_json" | jq .)" >>"$GITHUB_OUTPUT"
