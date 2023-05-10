#!/usr/bin/env bash
set -euo pipefail

function die() {
    local msg=${1:-missing error message}
    local code=${2:-1}
    echo "$msg"
    exit "$code"
}

function generate_dev_yml() {
    yq '
        .Parameters.KeyName = {
            "Description": "SSH Keypair to login to the instance",
            "Type": "AWS::EC2::KeyPair::KeyName"
        } |
        .Resources.ElasticAgentEc2Instance.Properties.KeyName = { "Ref": "KeyName" } |
        .Resources.ElasticAgentSecurityGroup.Properties.GroupDescription = "Allow SSH from anywhere" |
        .Resources.ElasticAgentSecurityGroup.Properties.SecurityGroupIngress += {
            "CidrIp": "0.0.0.0/0",
            "FromPort": 22,
            "IpProtocol": "tcp",
            "ToPort": 22
        }
    ' "$yml_file" > elastic-agent-ec2-dev.yml
}

if [[ ! -f .env ]]; then
    die 'Missing .env file, see README.md'
fi

yml_file='elastic-agent-ec2.yml'

# shellcheck disable=SC2046 # Intended splitting of .env arguments
export $(sed 's/#.*//g' .env| xargs)
declare -a parameters

[[ -z "${STACK_NAME-}" ]] && die 'Missing STACK_NAME variable'
[[ -z "${FLEET_URL-}" ]] && die 'Missing FLEET_URL variable' || parameters+=("ParameterKey=FleetUrl,ParameterValue=$FLEET_URL")
[[ -z "${ENROLLMENT_TOKEN-}" ]] && die 'Missing ENROLLMENT_TOKEN variable' || parameters+=("ParameterKey=EnrollmentToken,ParameterValue=$ENROLLMENT_TOKEN")
[[ -z "${ELASTIC_AGENT_VERSION-}" ]] && die 'Missing ELASTIC_AGENT_VERSION variable' || parameters+=("ParameterKey=ElasticAgentVersion,ParameterValue=$ELASTIC_AGENT_VERSION")
# optional parameter
[[ -z "${ELASTIC_ARTIFACT_SERVER-}" ]] || parameters+=("ParameterKey=ElasticArtifactServer,ParameterValue=$ELASTIC_ARTIFACT_SERVER")
if [[ -n "${KEY_NAME-}" ]]; then
    generate_dev_yml
    parameters+=("ParameterKey=KeyName,ParameterValue=$KEY_NAME")
    yml_file='elastic-agent-ec2-dev.yml'
fi

function create_stack() {
    aws cloudformation create-stack --parameters "${parameters[@]}" --stack-name "$STACK_NAME" --template-body file://"$yml_file" --capabilities CAPABILITY_NAMED_IAM
}

function delete_stack() {
    aws cloudformation delete-stack --stack-name "$STACK_NAME"
}

operation=${1:-create}
case "$operation" in
    "delete" ) delete_stack;;
    "create") create_stack;;
    *) die "Usage: $0 [create|delete]" ;;
esac
