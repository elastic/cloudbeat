#!/usr/bin/env bash
set -euo pipefail

function die() {
    local msg=${1:-missing error message}
    local code=${2:-1}
    echo "$msg"
    exit "$code"
}

function setup_key() {
    key_name="$STACK_NAME"

    if ! key_output=$(aws ec2 describe-key-pairs --key-name "$key_name" 2>&1); then
        echo "$key_output" | grep InvalidKeyPair.NotFound &>/dev/null || {
            die "Unexpected aws error: $key_output" 1;
        } || exit 1

        output="$HOME/.ssh/$key_name"
        [[ -f "$output" ]] && die "local file $output already exists, refusing to overwrite, run: rm -f $output"
        echo "Creating key $output"

        mkdir -p ~/.ssh
        aws ec2 create-key-pair --key-name "$key_name" --query 'KeyMaterial' --output text > "$output"
        chmod 0400 "$output"
    else
        echo "Re-using existing key $key_name"
        return 0
    fi
}

if [[ ! -f .env ]]; then
    die 'Missing .env file, see README.md'
fi

# shellcheck disable=SC2046 # Intended splitting of .env arguments
export $(sed 's/#.*//g' .env| xargs)
declare -a parameters

[[ -z "${STACK_NAME-}" ]] && die 'Missing STACK_NAME variable'
[[ -z "${FLEET_URL-}" ]] && die 'Missing FLEET_URL variable' || parameters+=("ParameterKey=FleetUrl,ParameterValue=$FLEET_URL")
[[ -z "${ENROLLMENT_TOKEN-}" ]] && die 'Missing ENROLLMENT_TOKEN variable' || parameters+=("ParameterKey=EnrollmentToken,ParameterValue=$ENROLLMENT_TOKEN")
[[ -z "${ELASTIC_AGENT_VERSION-}" ]] && die 'Missing ELASTIC_AGENT_VERSION variable' || parameters+=("ParameterKey=ElasticAgentVersion,ParameterValue=$ELASTIC_AGENT_VERSION")
# optional parameter
[[ -z "${ELASTIC_ARTIFACT_SERVER-}" ]] || parameters+=("ParameterKey=ElasticArtifactServer,ParameterValue=$ELASTIC_ARTIFACT_SERVER")

function create_stack() {
    [[ "$STACK_NAME" == developeraccess-* ]] && setup_key
    aws cloudformation create-stack --parameters "${parameters[@]}" --stack-name "$STACK_NAME" --template-body file://elastic-agent-ec2.yml --capabilities CAPABILITY_NAMED_IAM
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
