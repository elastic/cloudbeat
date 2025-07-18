#!/bin/bash
set -e

# Specify the path to the JSON file
config_file="config.json"

# Check if the config.json file exists
if [ ! -f "$config_file" ]; then
    echo "WARNING: The 'config.json' file does not exist. Please create it."
    return
fi

# Use jq to extract the key-value pairs and format them as "key=value" strings
for kv in $(jq -r "to_entries|map(\"\(.key)=\(.value|tostring)\")|.[]" "$config_file"); do
    export "$kv"
done
