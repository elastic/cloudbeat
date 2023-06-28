#!/bin/bash

ENV_PREFIX=$1
[ -z "$ENV_PREFIX" ] && echo "Please provide an environment prefix to delete" && exit -1
[ -z "$TF_VAR_ec_api_key" ] && echo "Please set TF_VAR_ec_api_key with an Elastic Cloud API Key" && exit -1

BUCKET=s3://tf-state-bucket-test-infra
ENVS=$(aws s3 ls $BUCKET/$ENV_PREFIX | awk '{print $2}' | sed 's/\///g')
COUNT=$(echo -n "$ENVS" | grep -c '^')
echo "Found ${COUNT} environments:"
echo "${ENVS[@]}"

FAILED=()
SUCCEED=()
SKIPPED=()

for ENV in $ENVS
do
    echo "Are you sure you want to delete $ENV? (yes/no)"
    read approve
    if [ "$approve" != "yes" ]; then
        echo "Skipping $ENV"
        SKIPPED+=($ENV)
        continue
    fi
    # Download the state file, destroy the environment, and delete the environment data from S3
    local="./$ENV-terraform.tfstate"
    aws s3 cp $BUCKET/$ENV/terraform.tfstate $local && \
    terraform destroy -var="region=eu-west-1" -state $local --auto-approve && \
    aws s3 rm $BUCKET/$ENV --recursive && \
    SUCCEED+=($ENV) || FAILED+=($ENV)
    rm $local
done

# Print results
echo
echo "Successfully deleted:"
printf "%s\n" "${SUCCEED[@]}"

echo
echo "Skipped deletion:"
printf "%s\n" "${SKIPPED[@]}"

echo
echo "Failed to delete:"
printf "%s\n" "${FAILED[@]}"
