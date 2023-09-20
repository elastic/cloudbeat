#!/usr/bin/env bash
# Script: add_cnvm_tags.sh
# Description: This script retrieves the instance ID of an EC2 instance
#              associated with a CloudFormation stack and applies tags to it.
# Usage: ./scripts/add_cnvm_tags.sh <aws_region> <stack_name> <tags>
#
# Arguments:
#   <aws_region>: The AWS region where the CloudFormation stack is deployed.
#   <stack_name>: The name of the CloudFormation stack.
#   <tags>: Tags to be applied to the EC2 instance, provided as a single string
#           with each key-value pair separated by spaces.
#
# Example:
#   ./scripts/add_cnvm_tags.sh eu-west-1 cnvm-sanity-test-stack "Key=division,Value=engineering Key=org,Value=security Key=team,Value=cloud-security-posture"

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
