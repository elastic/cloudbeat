#!/bin/bash

# Function to upload state and key files for ELK stack
upload_elk_stack() {
    aws s3 cp "./terraform.tfstate" "${S3_BUCKET}/elk-stack-terraform.tfstate"
}

# Function to upload state and key files for CIS
upload_cis() {
    aws s3 cp "./terraform.tfstate" "${S3_BUCKET}/cis-terraform.tfstate"
    aws s3 cp "${EC2_CSPM_KEY}" "${S3_BUCKET}/cspm.pem"
    aws s3 cp "${EC2_KSPM_KEY}" "${S3_BUCKET}/kspm.pem"
    aws s3 cp "${INTEGRATIONS_SETUP_DIR}/state_data.json" "$S3_BUCKET/state_data.json"
}

# Function to upload additional keys for CDR
upload_cdr() {
    aws s3 cp "./terraform.tfstate" "${S3_BUCKET}/cdr-terraform.tfstate"
    if [ -n "${CLOUDTRAIL_KEY:-}" ] && [ -f "${CLOUDTRAIL_KEY}" ]; then
        aws s3 cp "${CLOUDTRAIL_KEY}" "${S3_BUCKET}/cloudtrail.pem"
    fi
    if [ -n "${ACTIVITY_LOGS_KEY:-}" ] && [ -f "${ACTIVITY_LOGS_KEY}" ]; then
        aws s3 cp "${ACTIVITY_LOGS_KEY}" "${S3_BUCKET}/az_activity_logs.pem"
    fi
    if [ -n "${AUDIT_LOGS_KEY:-}" ] && [ -f "${AUDIT_LOGS_KEY}" ]; then
        aws s3 cp "${AUDIT_LOGS_KEY}" "${S3_BUCKET}/gcp_audit_logs.pem"
    fi
    if [ -n "${EC2_ASSET_INV_KEY:-}" ] && [ -f "${EC2_ASSET_INV_KEY}" ]; then
        aws s3 cp "${EC2_ASSET_INV_KEY}" "${S3_BUCKET}/asset_inv.pem"
    fi
    if [ -n "${EC2_WIZ_KEY:-}" ] && [ -f "${EC2_WIZ_KEY}" ]; then
        aws s3 cp "${EC2_WIZ_KEY}" "${S3_BUCKET}/wiz.pem"
    fi
    aws s3 cp "${INTEGRATIONS_SETUP_DIR}/state_data.json" "$S3_BUCKET/state_data.json"

    if [ -n "${ELASTIC_DEFEND_LINUX_KEY:-}" ] && [ -f "${ELASTIC_DEFEND_LINUX_KEY}" ]; then
        aws s3 cp "${ELASTIC_DEFEND_LINUX_KEY}" "${S3_BUCKET}/elastic_defend_linux.pem"
    fi
    if [ -n "${ELASTIC_DEFEND_WINDOWS_KEY:-}" ] && [ -f "${ELASTIC_DEFEND_WINDOWS_KEY}" ]; then
        aws s3 cp "${ELASTIC_DEFEND_WINDOWS_KEY}" "${S3_BUCKET}/elastic_defend_windows.pem"
    fi
    if [ -n "${WINDOWS_DEFEND_CREDENTIALS_FILE:-}" ] && [ -f "${WINDOWS_DEFEND_CREDENTIALS_FILE}" ]; then
        aws s3 cp "${WINDOWS_DEFEND_CREDENTIALS_FILE}" "${S3_BUCKET}/windows-defend-connection.json"
    fi
}

# Check for valid input
if [ "$#" -ne 1 ]; then
    echo "Usage: $0 {elk-stack|cis|cdr}"
    exit 1
fi

# Determine which function to call based on argument
case $1 in
"elk-stack")
    upload_elk_stack
    ;;
"cis")
    upload_cis
    ;;
"cdr")
    upload_cdr
    ;;
*)
    echo "Usage: $0 {elk-stack|cis|cdr}"
    exit 1
    ;;
esac

echo "Upload operation completed."
