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

# Parsing command-line arguments
while [[ "$#" -gt 0 ]]; do
    case $1 in
        -p|--prefix) ENV_PREFIX="$2"; shift ;;
        -n|--ignore-prefix) IGNORE_PREFIX="$2"; shift ;;
        -i|--interactive)
            shift
            case $1 in
                true|True|TRUE) INTERACTIVE=true ;;
                false|False|FALSE) INTERACTIVE=false ;;
                *) echo "Invalid value for --interactive. Please use true or false"; usage; exit 1 ;;
            esac
            ;;
        *) echo "Unknown parameter passed: $1"; usage; exit 1 ;;
    esac
    shift
done

# Ensure required environment variables and parameters are set
: "${ENV_PREFIX:?$(echo "Missing -p|--prefix. Please provide an environment prefix to delete" && usage && exit 1)}"
: "${TF_VAR_ec_api_key:?Please set TF_VAR_ec_api_key with an Elastic Cloud API Key}"

BUCKET=s3://tf-state-bucket-test-infra
ALL_ENVS=$(aws s3 ls $BUCKET/$ENV_PREFIX | awk '{print $2}' | sed 's/\///g')

# Divide environments into those to be deleted and those to be skipped
TO_DELETE=()
TO_SKIP=()

for ENV in $ALL_ENVS
do
    if [[ -n "$IGNORE_PREFIX" && "$ENV" == "$IGNORE_PREFIX"* ]]; then
        TO_SKIP+=("$ENV")
    else
        TO_DELETE+=("$ENV")
    fi
done

# Print the lists of environments to be deleted and skipped
echo "Environments to delete (${#TO_DELETE[@]}):"
printf "%s\n" "${TO_DELETE[@]}"

echo "Environments to skip (${#TO_SKIP[@]}):"
printf "%s\n" "${TO_SKIP[@]}"

# Ask for user confirmation if interactive mode is enabled
if [ "$INTERACTIVE" = true ]; then
    read -r -p "Are you sure you want to delete these environments? (y/n): " confirm && [[ $confirm == [yY] || $confirm == [yY][eE][sS] ]] || exit 1
fi

# Delete the environments
for ENV in "${TO_DELETE[@]}"
do
    echo "Deleting $ENV"
    local="./$ENV-terraform.tfstate"

    # Copy state file, destroy environment, and remove environment data from S3
    if aws s3 cp $BUCKET/$ENV/terraform.tfstate $local && \
    terraform destroy -var="region=eu-west-1" -state $local --auto-approve && \
    aws s3 rm $BUCKET/$ENV --recursive; then
        echo "Successfully deleted $ENV"
    else
        echo "Failed to delete $ENV"
        exit 1
    fi

    rm $local
done
