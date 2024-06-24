#!/bin/bash

# This script is used to deploy Azure resources and apply tags to them using Azure CLI.
# Requiered environment variables:
# - AZURE_TAGS: The default tags to apply to the resources. For example, "key1=value1 key2=value2".
# - DEPLOYMENT_NAME: The name of the deployment.
# - STACK_VERSION: The version of the stack.
# Pre-requisites:
# arm_parameters.json: The parameters file for the ARM template should be present in the same directory as this script.
# Example:
# {
#   "$schema": "https://schema.management.azure.com/schemas/2019-04-01/deploymentParameters.json#",
#   "contentVersion": "1.0.0.0",
#   "parameters": {
#     "EnrollmentToken": {
#       "value": "your-enrollment-token"
#     },
#     "FleetUrl": {
#       "value": "https://your-fleet-url"
#     },
#     "ElasticArtifactServer": {
#       "value": "https://artifacts.elastic.co/downloads/beats/elastic-agent"
#     },
#     "ElasticAgentVersion": {
#       "value": "8.14.1"
#     }
#   }
# }
# Usage:
#   cd ./deploy/azure
#   STACK_VERSION="8.14.1" DEPLOYMENT_NAME="elastic-agent-cspm" AZURE_TAGS="key1=value1 key2=value2" ./install_agent_az_cli.sh

# Exit immediately if a command exits with a non-zero status, print each command before executing it, and fail pipelines if any command fails.
set -euo pipefail

MINOR_VERSION=$(echo "${STACK_VERSION}" | cut -d'.' -f2)
# Check if minor version is less than 12, ie. 8.11 and below
if ((MINOR_VERSION < 12)); then
    echo "Versions 8.11 and below are not supported. Please use versions 8.12+"
    exit 0
fi

# Create a resource group with the name DEPLOYMENT_NAME
az group create --location EastUS --name "${DEPLOYMENT_NAME}"

# Create a group deployment using the ARM-for-single-account.json
az deployment group create --resource-group "${DEPLOYMENT_NAME}" --template-file ARM-for-single-account.json --parameters @arm_parameters.json

# Get the resource group ID of the DEPLOYMENT_NAME and store it in group
group=$(az group show -n "${DEPLOYMENT_NAME}" --query id --output tsv)

# --tags requires the tags to be passed as separate arguments
# shellcheck disable=SC2086
az tag update --resource-id "$group" --operation Merge --tags ${AZURE_TAGS}
# Get all resources within the resource group and store their IDs in resources
mapfile -t resources < <(az resource list --resource-group "${DEPLOYMENT_NAME}" --query "[].id" --output tsv)

# Loop over each resource ID in resources
for resource in "${resources[@]}"; do
    # Merge the deploy_tags with the existing tags of the resource
    # --tags requires the tags to be passed as separate arguments
    # shellcheck disable=SC2086
    az tag update --resource-id "$resource" --operation Merge --tags ${AZURE_TAGS}
done
