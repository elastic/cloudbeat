#!/usr/bin/env bash
set -euo pipefail

# Maximum timeout in seconds
max_timeout=300

# Elapsed time in seconds
elapsed_time=0

# Sleep time in seconds
sleep_time=10

# AWS region provided as the first argument
aws_region=${1:?AWS region not provided}

# Stack name provided as the second argument
stack_name=${2:?Stack name not provided}

# Tags provided as the third argument
tags=${3:?Tags not provided}

# Initialize instance_id variable
instance_id=""

# Loop until instance ID is retrieved or timeout is reached
while [ -z "$instance_id" ]; do
    instance_id=$(aws cloudformation describe-stack-resources --region "$aws_region" --stack-name "$stack_name" --query 'StackResources[?ResourceType==`AWS::EC2::Instance`].PhysicalResourceId' --output text)
    echo "Waiting for instance ID..."

    # Check if instance ID is retrieved and break the loop
    if [ -n "$instance_id" ]; then
        break
    fi

    # Check if timeout is reached and exit the loop
    if [ "$elapsed_time" -ge "$max_timeout" ]; then
        echo "Timeout reached. Exiting..."
        break
    fi

    # Sleep for the specified duration
    sleep "$sleep_time"
    elapsed_time=$((elapsed_time + sleep_time))
done

echo "$instance_id"

# If instance ID is retrieved, create tags for the instance
if [ -n "$instance_id" ]; then
    aws ec2 create-tags --region "$aws_region" --resources "$instance_id" --tags $tags
fi
