#!/bin/bash
set -euo pipefail

function upload_file() {
    local local_file=${1:?Missing local file name}
    local remote_file_prefix=${2:?Missing remote file name}
    local cft_version=${3:?Missing version}
    local remote_file="${remote_file_prefix}-${cft_version}.yml"

    sed --in-place'' "s/CFT_VERSION/$cft_version/g" "$local_file"

    local s3_file="s3://elastic-cspm-cft/$remote_file"
    echo "Uploading $local_file to $s3_file"
    aws s3 cp "$local_file" "$s3_file"
}

version=$(grep defaultBeatVersion version/version.go | cut -f2 -d "\"")

upload_file deploy/cloudformation/elastic-agent-ec2-cnvm.yml \
    "cloudformation-cnvm" \
    "${version}"
upload_file deploy/cloudformation/elastic-agent-ec2-cspm.yml \
    "cloudformation-cspm-single-account" \
    "${version}"
upload_file deploy/cloudformation/elastic-agent-ec2-cspm-organization.yml \
    "cloudformation-cspm-organization-account" \
    "${version}"
upload_file deploy/cloudformation/elastic-agent-direct-access-key-cspm.yml \
    "cloudformation-cspm-direct-access-key-single-account" \
    "${version}"
upload_file deploy/cloudformation/elastic-agent-direct-access-key-cspm-organization.yml \
    "cloudformation-cspm-direct-access-key-organization-account" \
    "${version}"
upload_file deploy/cloudformation/cloud-connectors-remote-role.yml \
    "cloudformation-cloud-connectors-single-account" \
    "${version}"
upload_file deploy/cloudformation/cloud-connectors-remote-role-organization.yml \
    "cloudformation-cloud-connectors-organization-account" \
    "${version}"

upload_file deploy/asset-inventory-cloudformation/elastic-agent-ec2.yml \
    "cloudformation-asset-inventory-single-account" \
    "${version}"
upload_file deploy/asset-inventory-cloudformation/elastic-agent-ec2-organization.yml \
    "cloudformation-asset-inventory-organization-account" \
    "${version}"
