#!/bin/bash
set -e

function usage() {
    cat <<EOF
Usage: $0 -p|--prefix ENV_PREFIX [-n|--ignore-prefix IGNORE_PREFIX] [-i|--interactive true|false]
Cleanup script to delete environments based on a prefix.
Required flag: -p|--prefix
EOF
}

# Defaults
INTERACTIVE=true
AWS_REGION="eu-west-1" # Add your desired default AWS region here

# Arrays to keep track of successful and failed deletions
DELETED_ENVS=()
FAILED_ENVS=()

# Function to set environment variables from env_config.json
function set_env_vars_from_config() {
    local env=$1
    local bucket_folder=$2

    # Try to download env_config.json
    if aws s3 cp "$BUCKET/$bucket_folder/env_config.json" /tmp/env_config.json 2>/dev/null; then
        echo "Reading environment configuration from env_config.json for $env"

        # Extract values with defaults for backward compatibility
        # Note: If ess_region_mapped doesn't exist, default to gcp-us-west2 (most common case)
        # We can't map ess_region without serverless_mode info, so we use a safe default
        local ess_region_mapped
        ess_region_mapped=$(jq -r 'if .ess_region_mapped then .ess_region_mapped else "gcp-us-west2" end' /tmp/env_config.json)
        local ec_url
        ec_url=$(jq -r '.ec_url // "https://cloud.elastic.co"' /tmp/env_config.json)
        local serverless_mode
        serverless_mode=$(jq -r '.serverless_mode // "false"' /tmp/env_config.json)
        local deployment_template
        deployment_template=$(jq -r '.deployment_template // "gcp-storage-optimized"' /tmp/env_config.json)
        local max_size
        max_size=$(jq -r '.max_size // "128g"' /tmp/env_config.json)

        # Extract ess_region to determine which API key to use
        local ess_region
        ess_region=$(jq -r '.ess_region // "production-cft"' /tmp/env_config.json)

        # Parse environment type from ess_region (format: {env}-{cloud}, e.g., production-cft, qa-gcp, staging-aws)
        local env_type="${ess_region%%-*}"

        # Convert env_type to uppercase and construct the environment variable name
        local env_type_upper
        env_type_upper=$(echo "$env_type" | tr '[:lower:]' '[:upper:]')
        local api_key_var="EC_API_KEY_${env_type_upper}"

        # Set Terraform variables
        export TF_VAR_ess_region="$ess_region_mapped"
        export TF_VAR_ec_url="$ec_url"
        export TF_VAR_serverless_mode="$serverless_mode"
        export TF_VAR_deployment_template="$deployment_template"
        export TF_VAR_max_size="$max_size"

        # Set TF_VAR_ec_api_key based on environment type using indirect variable reference
        if [[ -n "${!api_key_var:-}" ]]; then
            export TF_VAR_ec_api_key="${!api_key_var}"
            echo "Using ${env_type_upper} API key for $env (ess_region: $ess_region)"
        else
            echo "Warning: ${api_key_var} not set, falling back to TF_VAR_ec_api_key if available"
        fi

        echo "Set TF_VAR_ess_region=$TF_VAR_ess_region"
        echo "Set TF_VAR_ec_url=$TF_VAR_ec_url"
        echo "Set TF_VAR_serverless_mode=$TF_VAR_serverless_mode"
        echo "Set TF_VAR_deployment_template=$TF_VAR_deployment_template"
        echo "Set TF_VAR_max_size=$TF_VAR_max_size"
        echo "Set TF_VAR_ec_api_key (from $env_type environment)"

        rm -f /tmp/env_config.json
    else
        # If env_config.json doesn't exist, use defaults for backward compatibility
        echo "env_config.json not found for $env, using defaults"
        export TF_VAR_ess_region="gcp-us-west2"
        export TF_VAR_ec_url="https://cloud.elastic.co"
        export TF_VAR_serverless_mode="false"
        export TF_VAR_deployment_template="gcp-storage-optimized"
        export TF_VAR_max_size="128g"
        # If no env_config.json, rely on existing TF_VAR_ec_api_key if set
        echo "Using existing TF_VAR_ec_api_key if available"
    fi

    # Final validation: ensure TF_VAR_ec_api_key is set
    if [[ -z "${TF_VAR_ec_api_key:-}" ]]; then
        echo "Error: TF_VAR_ec_api_key is not set for environment $env" >&2
        echo "Please ensure EC_API_KEY_PRODUCTION, EC_API_KEY_STAGING, or EC_API_KEY_QA is set, or provide TF_VAR_ec_api_key" >&2
        return 1
    fi
}

# Function to delete all terraform states from given bucket
function delete_all_states() {
    local bucket_folder=$1
    echo "Deleting all Terraform states from bucket: $bucket_folder"

    # Set environment variables from env_config.json
    if ! set_env_vars_from_config "$bucket_folder" "$bucket_folder"; then
        echo "Failed to set environment variables for $bucket_folder"
        FAILED_ENVS+=("$bucket_folder")
        return 1
    fi

    states=("cdr" "cis" "elk-stack")
    # Get all states
    for state in "${states[@]}"; do
        local state_file="./$state/terraform.tfstate"
        aws s3 cp "$BUCKET/$bucket_folder/${state}-terraform.tfstate" "$state_file" || true
    done
    # Destroy all states and remove environment data from S3
    if ./manage_infrastructure.sh "all" "destroy" &&
        aws s3 rm "$BUCKET/$bucket_folder" --recursive; then
        echo "Successfully deleted $bucket_folder"
        DELETED_ENVS+=("$bucket_folder")
    else
        echo "Failed to delete $bucket_folder"
        FAILED_ENVS+=("$bucket_folder")
    fi
}

# Function to delete Terraform environment
function delete_environment() {
    local ENV=$1
    echo "Deleting Terraform environment: $ENV"
    tfstate="./$ENV-terraform.tfstate"

    # Set environment variables from env_config.json
    if ! set_env_vars_from_config "$ENV" "$ENV"; then
        echo "Failed to set environment variables for $ENV"
        FAILED_ENVS+=("$ENV")
        return 1
    fi

    # Copy state file
    if aws s3 cp "$BUCKET/$ENV/terraform.tfstate" "$tfstate"; then
        echo "Downloaded Terraform state file from S3."

        # Check if the resource aws_auth exists in the local state file and remove it
        terraform state rm -state "$tfstate" "$(terraform state list -state "$tfstate" | grep "kubernetes_config_map_v1_data.aws_auth")" || true
        # Destroy environment and remove environment data from S3
        if terraform destroy -var="region=$AWS_REGION" -state "$tfstate" --auto-approve &&
            aws s3 rm "$BUCKET/$ENV" --recursive; then
            echo "Successfully deleted $ENV"
            DELETED_ENVS+=("$ENV")
        else
            echo "Failed to delete $ENV"
            FAILED_ENVS+=("$ENV")
        fi

        rm "$tfstate"
    else
        echo "Failed to download Terraform state file from S3 for $ENV"
        FAILED_ENVS+=("$ENV")
    fi
}

# Function to delete CloudFormation stack
function delete_stack() {
    local STACK_NAME=$1
    echo "Deleting CloudFormation stack: $STACK_NAME"
    aws cloudformation delete-stack --stack-name "$STACK_NAME" --region "$AWS_REGION"
}

# Parsing command-line arguments
while [[ "$#" -gt 0 ]]; do
    case $1 in
    -p | --prefix)
        ENV_PREFIX="$2"
        shift
        ;;
    -n | --ignore-prefix)
        IGNORE_PREFIX="$2"
        shift
        ;;
    -i | --interactive)
        shift
        case $1 in
        true | True | TRUE) INTERACTIVE=true ;;
        false | False | FALSE) INTERACTIVE=false ;;
        *)
            echo "Invalid value for --interactive. Please use true or false"
            usage
            exit 1
            ;;
        esac
        ;;
    *)
        echo "Unknown parameter passed: $1"
        usage
        exit 1
        ;;
    esac
    shift
done

# Ensure required environment variables and parameters are set
: "${ENV_PREFIX:?$(echo "Missing -p|--prefix. Please provide an environment prefix to delete" && usage && exit 1)}"

# Note: TF_VAR_ec_api_key will be set per-environment in set_env_vars_from_config
# But we still need at least one API key available as fallback
if [[ -z "${EC_API_KEY_PRODUCTION:-}" ]] && [[ -z "${EC_API_KEY_STAGING:-}" ]] && [[ -z "${EC_API_KEY_QA:-}" ]] && [[ -z "${TF_VAR_ec_api_key:-}" ]]; then
    echo "Error: At least one of EC_API_KEY_PRODUCTION, EC_API_KEY_STAGING, EC_API_KEY_QA, or TF_VAR_ec_api_key must be set" >&2
    exit 1
fi

BUCKET=s3://tf-state-bucket-test-infra
ALL_ENVS=$(aws s3 ls $BUCKET/"$ENV_PREFIX" | awk '{print $2}' | sed 's/\///g')
ALL_STACKS=$(aws cloudformation list-stacks --stack-status-filter "CREATE_COMPLETE" "UPDATE_COMPLETE" --region "$AWS_REGION" | jq -r '.StackSummaries[] | select(.StackName | startswith("'"$ENV_PREFIX"'") and (if "'"$IGNORE_PREFIX"'" != "" then .StackName | startswith("'"$IGNORE_PREFIX"'") | not else true end)) | .StackName')

if [ -n "$IGNORE_PREFIX" ]; then
    # If IGNORE_PREFIX exists and is not empty
    GCP_FILTER="name:'$ENV_PREFIX*' AND NOT name:'$IGNORE_PREFIX*'"
else
    # If IGNORE_PREFIX is empty or does not exist
    GCP_FILTER="name:'$ENV_PREFIX*'"
fi

while IFS= read -r line; do
    ALL_GCP_DEPLOYMENTS+=("$line")
done < <(gcloud deployment-manager deployments list --filter="$GCP_FILTER" --format="value(name)")

# Divide environments into those to be deleted and those to be skipped
TO_DELETE_ENVS=()
TO_SKIP_ENVS=()

for ENV in $ALL_ENVS; do
    if [[ -n "$IGNORE_PREFIX" && "$ENV" == "$IGNORE_PREFIX"* ]]; then
        TO_SKIP_ENVS+=("$ENV")
    else
        TO_DELETE_ENVS+=("$ENV")
    fi
done

# Print the lists of environments to be deleted and skipped
echo "Environments to delete (${#TO_DELETE_ENVS[@]}):"
printf "%s\n" "${TO_DELETE_ENVS[@]}"

echo "Environments to skip (${#TO_SKIP_ENVS[@]}):"
printf "%s\n" "${TO_SKIP_ENVS[@]}"

# Ask for user confirmation if interactive mode is enabled
if [ "$INTERACTIVE" = true ]; then
    read -r -p "Are you sure you want to delete these environments? (y/n): " confirm && [[ $confirm == [yY] || $confirm == [yY][eE][sS] ]] || exit 1
fi

# Delete the Terraform environments
for ENV in "${TO_DELETE_ENVS[@]}"; do
    if aws s3 ls "$BUCKET/$ENV/terraform.tfstate" >/dev/null 2>&1; then
        delete_environment "$ENV"
    else
        delete_all_states "$ENV"
    fi
done

# Print summary of environment deletions
echo "Successfully deleted environments (${#DELETED_ENVS[@]}):"
printf "%s\n" "${DELETED_ENVS[@]}"

echo "Failed to delete environments (${#FAILED_ENVS[@]}):"
printf "%s\n" "${FAILED_ENVS[@]}"

DELETED_STACKS=()
FAILED_STACKS=()

# Wait for the CloudFormation stacks to be deleted
for STACK in $ALL_STACKS; do
    if delete_stack "$STACK" && aws cloudformation wait stack-delete-complete --stack-name "$STACK" --region "$AWS_REGION"; then
        echo "Successfully deleted CloudFormation stack: $STACK"
        DELETED_STACKS+=("$STACK")
    else
        echo "Failed to delete CloudFormation stack: $STACK"
        FAILED_STACKS+=("$STACK")
    fi
done

# Print summary of stacks deletions
echo "Successfully deleted CloudFormation stacks (${#DELETED_STACKS[@]}):"
printf "%s\n" "${DELETED_STACKS[@]}"

echo "Failed to delete CloudFormation stacks (${#FAILED_STACKS[@]}):"
printf "%s\n" "${FAILED_STACKS[@]}"

# Delete GCP deployments
# Check if ALL_GCP_DEPLOYMENTS is empty
if [ ${#ALL_GCP_DEPLOYMENTS[@]} -eq 0 ]; then
    echo "No GCP deployments to delete."
else
    PROJECT_NAME=$(gcloud config get-value core/project)
    PROJECT_NUMBER=$(gcloud projects list --filter="${PROJECT_NAME}" --format="value(PROJECT_NUMBER)")
    ./delete_gcp_env.sh "$PROJECT_NAME" "$PROJECT_NUMBER" "${ALL_GCP_DEPLOYMENTS[@]}"
fi

# Delete Azure groups
FAILED_AZURE_GROUPS=()
groups=$(az group list --query "[?starts_with(name, '$ENV_PREFIX')].name" -o tsv | awk -v prefix="$IGNORE_PREFIX" '{if (prefix == "") print $0; else if (!($0 ~ "^" prefix)) print $0}')
for group in $groups; do
    az group delete --name "$group" --yes || {
        echo "Failed to delete resource group: $group"
        FAILED_AZURE_GROUPS+=("$group")
        continue
    }
    echo "Resource group $group has been deleted."
done

# Final check to exit with an error if any deletions of environments, stacks, or deployments failed
if [ ${#FAILED_ENVS[@]} -gt 0 ] || [ ${#FAILED_STACKS[@]} -gt 0 ] || [ ${#FAILED_DEPLOYMENTS[@]} -gt 0 ] || [ ${#FAILED_AZURE_GROUPS[@]} -gt 0 ]; then
    echo "There were errors deleting one or more resources."
    # Optionally, provide more details about which resources failed to delete
    if [ ${#FAILED_ENVS[@]} -gt 0 ]; then
        echo "Failed to delete the following environments:"
        printf "%s\n" "${FAILED_ENVS[@]}"
    fi
    if [ ${#FAILED_STACKS[@]} -gt 0 ]; then
        echo "Failed to delete the following CloudFormation stacks:"
        printf "%s\n" "${FAILED_STACKS[@]}"
    fi
    if [ ${#FAILED_DEPLOYMENTS[@]} -gt 0 ]; then
        echo "Failed to delete the following GCP deployments:"
        printf "%s\n" "${FAILED_DEPLOYMENTS[@]}"
    fi
    if [ ${#FAILED_AZURE_GROUPS[@]} -gt 0 ]; then
        echo "Failed to delete the following Azure groups:"
        printf "%s\n" "${FAILED_AZURE_GROUPS[@]}"
    fi
    exit 1
fi
