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
    aws s3 cp "./state_data.json" "$S3_BUCKET/state_data.json"
}

# Function to upload additional keys for CDR
upload_cdr() {
    aws s3 cp "./terraform.tfstate" "${S3_BUCKET}/cdr-terraform.tfstate"
    aws s3 cp "${CLOUDTRAIL_KEY}" "${S3_BUCKET}/cloudtrail.pem"
    aws s3 cp "${ACTIVITY_LOGS_KEY}" "${S3_BUCKET}/az_activity_logs.pem"
    aws s3 cp "${AUDIT_LOGS_KEY}" "${S3_BUCKET}/gcp_audit_logs.pem"
    aws s3 cp "${EC2_ASSET_INV_KEY}" "${S3_BUCKET}/asset_inv.pem"
    aws s3 cp "./state_data.json" "$S3_BUCKET/state_data.json"
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
