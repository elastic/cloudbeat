#!/bin/bash

# Function to check if a file exists in a folder
file_exists() {
    local bucket_path="$1"
    local bucket_exists
    # Check if the file exists in the bucket
    bucket_exists=$(aws s3 ls "$bucket_path")
    if [ -z "$bucket_exists" ]; then
        return 1
    fi
    return 0
}

# Get all folders from the specific S3 bucket
s3_bucket="tf-state-bucket-test-infra"
folders=$(aws s3 ls "s3://${s3_bucket}" | awk '{print $2}')

# JSON array to store deployment names
expired_envs='[]'

# Boolean variable to track if an expired environment was found
expired_env_found=false

# Iterate over each folder
for folder in $folders; do
    # Check if env_config.json file exists
    if ! file_exists "s3://${s3_bucket}/${folder}env_config.json"; then
        # TODO: To consider if we want to log this message
        # echo "env_config.json file does not exist in $folder"
        continue
    fi

    # TODO: To consider if we want to log this message
    # echo "Reading env_config.json file from $folder"

    file_content=$(aws s3 cp "s3://${s3_bucket}/${folder}env_config.json" -)

    # Read expiration date from env_config.json
    expiration=$(echo "$file_content" | jq -r '.expiration')
    current_date=$(date +%Y-%m-%d)

    # Compare expiration date with current date
    if [[ ! "$expiration" > "$current_date" ]]; then
        # Extract the deployment_name field using jq
        deployment_name=$(echo "$file_content" | jq -r '.deployment_name')
        # TODO: To consider if we want to log this message
        echo "Environment $deployment_name is expired."
        # Add deployment name to JSON object
        expired_envs=$(echo "$expired_envs" | jq --arg value "$deployment_name" '. += [{"deployment_name": $value}]')

        # Set found to true
        expired_env_found=true
    else
        # TODO: To consider if we want to log this message
        echo "Environment $(echo "$file_content" | jq -r '.deployment_name') is not expired yet."
    fi
done

# Print the deployment names and found status to GITHUB_OUTPUT
deployments_json=$(printf '%s\n' "${expired_envs[@]}" | jq -R . | jq -s .)
echo "deployments=$deployments_json" >>"$GITHUB_OUTPUT"
echo "expired_env_found=$expired_env_found" >>"$GITHUB_OUTPUT"
