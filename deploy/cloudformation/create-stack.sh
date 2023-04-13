#!/bin/bash
STACK_NAME="${1:-$STACK_NAME}"
FLEET_URL="${2:-$FLEET_URL}"
ENROLLMENT_TOKEN="${3:-$ENROLLMENT_TOKEN}"
ELASTIC_AGENT_VERSION="${4:-elastic-agent-8.8.0-SNAPSHOT-linux-arm64}"
TEMPLATE="deploy/cloudformation/elastic-agent-ec2.yml"
if [ -n "${DEV}" ]; then
  python3 deploy/cloudformation/generate_dev.py || { echo 'Dev CloudFormation generation failed' ; exit 1; }
  TEMPLATE="deploy/cloudformation/elastic-agent-ec2-dev.yml"
fi

aws cloudformation create-stack                                                                     \
     --stack-name ${STACK_NAME}                                                                     \
     --template-body file://${TEMPLATE}                                                             \
     --capabilities CAPABILITY_NAMED_IAM --parameters                                               \
     ParameterKey=ElasticAgentVersion,ParameterValue=${ELASTIC_AGENT_VERSION}                       \
     ParameterKey=FleetUrl,ParameterValue=${FLEET_URL}                                              \
     ParameterKey=EnrollmentToken,ParameterValue=${ENROLLMENT_TOKEN}
