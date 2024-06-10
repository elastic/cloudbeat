#!/bin/bash

# Function to check if a file exists in a folder
file_exists() {
    local bucket_name="$1"
    local bucket_key="$2"
    # Check if the file exists in the bucket
    if aws s3api head-object --bucket "$bucket_name" --key "$bucket_key" >/dev/null 2>&1; then
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
    if ! file_exists "$s3_bucket" "$folder/env_config.json"; then
        echo "env_config.json file does not exist in $folder"
        continue
    fi

    file_content=$(aws s3 cp s3://"$s3_bucket"/"$folder"/env_config.json -)

    # Read expiration date from env_config.json
    expiration=$(echo "$file_content" | jq -r '.expiration')
    current_date=$(date +%Y-%m-%d)

    # Compare expiration date with current date
    if [[ ! "$expiration" > "$current_date" ]]; then
        # Extract the deployment_name field using jq
        deployment_name=$(echo "$file_content" | jq -r '.deployment_name')

        # Add deployment name to JSON object
        deployment_json=$(echo "$deployment_json" | jq --arg value "$deployment_name" '. += [{"deployment_name": $value}]')
    fi
done

# Print the deployment names to GITHUB_OUTPUT
echo "deployments=$(echo "$deployment_json" | jq .)" >>"$GITHUB_OUTPUT"
