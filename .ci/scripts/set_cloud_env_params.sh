#!/bin/bash

KIBANA_URL="$(terraform output -raw kibana_url)"
echo "KIBANA_URL=$KIBANA_URL" >>"$GITHUB_ENV"
ES_URL="$(terraform output -raw elasticsearch_url)"
echo "ES_URL=$ES_URL" >>"$GITHUB_ENV"
ES_USER="$(terraform output -raw elasticsearch_username)"
echo "ES_USER=$ES_USER" >>"$GITHUB_ENV"

ES_PASSWORD=$(terraform output -raw elasticsearch_password)
echo "::add-mask::$ES_PASSWORD"
echo "ES_PASSWORD=$ES_PASSWORD" >>"$GITHUB_ENV"

# Remove 'https://' from the URLs
KIBANA_URL_STRIPPED="${KIBANA_URL//https:\/\//}"
ES_URL_STRIPPED="${ES_URL//https:\/\//}"

# Create test URLs with credentials
TEST_KIBANA_URL="https://${ES_USER}:${ES_PASSWORD}@${KIBANA_URL_STRIPPED}"
echo "::add-mask::${TEST_KIBANA_URL}"
echo "TEST_KIBANA_URL=${TEST_KIBANA_URL}" >>"$GITHUB_ENV"

TEST_ES_URL="https://${ES_USER}:${ES_PASSWORD}@${ES_URL_STRIPPED}"
echo "::add-mask::${TEST_ES_URL}"
echo "TEST_ES_URL=${TEST_ES_URL}" >>"$GITHUB_ENV"

EC2_CSPM=$(terraform output -raw ec2_cspm_ssh_cmd)
echo "::add-mask::$EC2_CSPM"
echo "EC2_CSPM=$EC2_CSPM" >>"$GITHUB_ENV"

EC2_KSPM=$(terraform output -raw ec2_kspm_ssh_cmd)
echo "::add-mask::$EC2_KSPM"
echo "EC2_KSPM=$EC2_KSPM" >>"$GITHUB_ENV"

EC2_CSPM_KEY=$(terraform output -raw ec2_cspm_key)
echo "::add-mask::$EC2_CSPM_KEY"
echo "EC2_CSPM_KEY=$EC2_CSPM_KEY" >>"$GITHUB_ENV"

EC2_KSPM_KEY=$(terraform output -raw ec2_kspm_key)
echo "::add-mask::$EC2_KSPM_KEY"
echo "EC2_KSPM_KEY=$EC2_KSPM_KEY" >>"$GITHUB_ENV"

KSPM_PUBLIC_IP=$(terraform output -raw ec2_kspm_public_ip)
echo "::add-mask::$KSPM_PUBLIC_IP"
echo "KSPM_PUBLIC_IP=$KSPM_PUBLIC_IP" >>"$GITHUB_ENV"

CSPM_PUBLIC_IP=$(terraform output -raw ec2_cspm_public_ip)
echo "::add-mask::$CSPM_PUBLIC_IP"
echo "CSPM_PUBLIC_IP=$CSPM_PUBLIC_IP" >>"$GITHUB_ENV"

if [[ ${TF_VAR_cdr_infra:-} == "true" ]]; then
    ec2_cloudtrail_public_ip=$(terraform output -raw ec2_cloudtrail_public_ip)
    echo "::add-mask::$ec2_cloudtrail_public_ip"
    echo "CLOUDTRAIL_PUBLIC_IP=$ec2_cloudtrail_public_ip" >>"$GITHUB_ENV"

    ec2_cloudtrail_key=$(terraform output -raw ec2_cloudtrail_key)
    echo "::add-mask::$ec2_cloudtrail_key"
    echo "CLOUDTRAIL_KEY=$ec2_cloudtrail_key" >>"$GITHUB_ENV"

    az_vm_activity_logs_public_ip=$(terraform output -raw az_vm_activity_logs_public_ip)
    echo "::add-mask::$az_vm_activity_logs_public_ip"
    echo "ACTIVITY_LOGS_PUBLIC_IP=$az_vm_activity_logs_public_ip" >>"$GITHUB_ENV"

    az_vm_activity_logs_key=$(terraform output -raw az_vm_activity_logs_key)
    echo "::add-mask::$az_vm_activity_logs_key"
    echo "ACTIVITY_LOGS_KEY=$az_vm_activity_logs_key" >>"$GITHUB_ENV"

    gcp_audit_logs_public_ip=$(terraform output -raw gcp_audit_logs_public_ip)
    echo "::add-mask::$gcp_audit_logs_public_ip"
    echo "AUDIT_LOGS_PUBLIC_IP=$gcp_audit_logs_public_ip" >>"$GITHUB_ENV"

    gcp_audit_logs_key=$(terraform output -raw gcp_audit_logs_key)
    echo "::add-mask::$gcp_audit_logs_key"
    echo "AUDIT_LOGS_KEY=$gcp_audit_logs_key" >>"$GITHUB_ENV"
fi
