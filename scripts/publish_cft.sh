#!/bin/bash
set -euo pipefail

function upload_file() {
    local local_file=${1:?Missing local file name}
    local remote_file=${2:?Missing remote file name}

    sed --in-place'' "s/CFT_VERSION/$remote_file/g" "$local_file"

    local s3_file="s3://elastic-cspm-cft/$remote_file"
    echo "Uploading $local_file to $s3_file"
    aws s3 cp "$local_file" "$s3_file"
}

version=$(grep defaultBeatVersion version/version.go | cut -f2 -d "\"")
upload_file deploy/cloudformation/elastic-agent-ec2-cnvm.yml "cloudformation-cnvm-$version.yml"
upload_file deploy/cloudformation/elastic-agent-ec2-cspm.yml "cloudformation-cspm-single-account-$version.yml"
upload_file deploy/cloudformation/elastic-agent-ec2-cspm-organization.yml "cloudformation-cspm-organization-account-$version.yml"
