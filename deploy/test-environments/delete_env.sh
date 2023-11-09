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

# Function to delete Terraform environment
function delete_environment() {
    local ENV=$1
    echo "Deleting Terraform environment: $ENV"
    tfstate="./$ENV-terraform.tfstate"

    # Copy state file
    if aws s3 cp "$BUCKET/$ENV/terraform.tfstate" "$tfstate"; then
        echo "Downloaded Terraform state file from S3."

        # Check if the resource aws_auth exists in the local state file and remove it
        terraform state rm -state "$tfstate" $(terraform state list -state "$tfstate" | grep "kubernetes_config_map_v1_data.aws_auth") || true
        # Destroy environment and remove environment data from S3
        if terraform destroy -var="region=$AWS_REGION" -state "$tfstate" --auto-approve &&
            aws s3 rm "$BUCKET"/"$ENV" --recursive; then
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
: "${TF_VAR_ec_api_key:?Please set TF_VAR_ec_api_key with an Elastic Cloud API Key}"

BUCKET=s3://tf-state-bucket-test-infra
ALL_ENVS=$(aws s3 ls $BUCKET/"$ENV_PREFIX" | awk '{print $2}' | sed 's/\///g')
ALL_STACKS=$(aws cloudformation list-stacks --stack-status-filter "CREATE_COMPLETE" "UPDATE_COMPLETE" --region "$AWS_REGION" | jq -r '.StackSummaries[] | select(.StackName | startswith("'$ENV_PREFIX'") and (if "'$IGNORE_PREFIX'" != "" then .StackName | startswith("'$IGNORE_PREFIX'") | not else true end)) | .StackName')

if [ -n "$IGNORE_PREFIX" ]; then
    # If IGNORE_PREFIX exists and is not empty
    GCP_FILTER="name:'$ENV_PREFIX*' AND NOT name:'$IGNORE_PREFIX*'"
else
    # If IGNORE_PREFIX is empty or does not exist
    GCP_FILTER="name:'$ENV_PREFIX*'"
fi

ALL_GCP_DEPLOYMENTS=$(gcloud deployment-manager deployments list --filter="$GCP_FILTER" --format="value(name)")

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
    delete_environment "$ENV"
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

DELETED_DEPLOYMENTS=()
FAILED_DEPLOYMENTS=()

PROJECT_NAME=$(gcloud config get-value core/project)
PROJECT_NUMBER=$(gcloud projects list --filter="${PROJECT_NAME}" --format="value(PROJECT_NUMBER)")
export PROJECT_NAME
export PROJECT_NUMBER

# Delete GCP Deployments
for DEPLOYMENT in $ALL_GCP_DEPLOYMENTS; do
    # Add the needed roles to delete the templates to the project using the deployment manager
    gcloud projects add-iam-policy-binding "${PROJECT_NAME}" --member=serviceAccount:"${PROJECT_NUMBER}"@cloudservices.gserviceaccount.com --role=roles/iam.roleAdmin --no-user-output-enabled
    gcloud projects add-iam-policy-binding "${PROJECT_NAME}" --member=serviceAccount:"${PROJECT_NUMBER}"@cloudservices.gserviceaccount.com --role=roles/resourcemanager.projectIamAdmin --no-user-output-enabled

    if gcloud deployment-manager deployments delete "$DEPLOYMENT" -q; then
        echo "Successfully deleted GCP deployment: $DEPLOYMENT"
        DELETED_DEPLOYMENTS+=("$DEPLOYMENT")
    else
        echo "Failed to delete GCP deployment: $DEPLOYMENT"
        FAILED_DEPLOYMENTS+=("$DEPLOYMENT")
    fi

    # Remove the roles required to deploy the DM templates
    gcloud projects remove-iam-policy-binding "${PROJECT_NAME}" --member=serviceAccount:"${PROJECT_NUMBER}"@cloudservices.gserviceaccount.com --role=roles/iam.roleAdmin --no-user-output-enabled
    gcloud projects remove-iam-policy-binding "${PROJECT_NAME}" --member=serviceAccount:"${PROJECT_NUMBER}"@cloudservices.gserviceaccount.com --role=roles/resourcemanager.projectIamAdmin --no-user-output-enabled

done

# Print summary of gcp deployments deletions
echo "Successfully deleted GCP deployments (${#DELETED_DEPLOYMENTS[@]}):"
printf "%s\n" "${DELETED_DEPLOYMENTS[@]}"

echo "Failed to delete GCP deployments (${#FAILED_DEPLOYMENTS[@]}):"
printf "%s\n" "${FAILED_DEPLOYMENTS[@]}"

# Final check to exit with an error if any deletions of environments, stacks, or deployments failed
if [ ${#FAILED_ENVS[@]} -gt 0 ] || [ ${#FAILED_STACKS[@]} -gt 0 ] || [ ${#FAILED_DEPLOYMENTS[@]} -gt 0 ]; then
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
    exit 1
fi
