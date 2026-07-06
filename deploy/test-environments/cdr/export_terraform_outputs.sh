#!/usr/bin/env bash
# Writes CDR Terraform outputs to GITHUB_OUTPUT (null-safe via terraform output -json).
# Used by the CDR composite action when some modules are disabled (e.g. Wiz-only or full CDR).
set -euo pipefail

: "${GITHUB_OUTPUT:?GITHUB_OUTPUT must be set}"

TF_OUT=$(terraform output -json)

# Append "name=value" to GITHUB_OUTPUT; mask non-empty values in logs.
out_masked() {
    local out_name="$1" tf_key="$2"
    local val
    val=$(echo "$TF_OUT" | jq -r --arg k "$tf_key" '.[$k].value // empty')
    if [ -n "$val" ]; then
        echo "::add-mask::$val"
    fi
    printf '%s=%s\n' "$out_name" "$val" >>"$GITHUB_OUTPUT"
}

out_masked aws-ec2-cloudtrail-public-ip ec2_cloudtrail_public_ip
out_masked aws-ec2-cloudtrail-key ec2_cloudtrail_key
out_masked az-vm-activity-logs-public-ip az_vm_activity_logs_public_ip
out_masked az-vm-activity-logs-key az_vm_activity_logs_key
out_masked gcp-audit-logs-public-ip gcp_audit_logs_public_ip
out_masked gcp-audit-logs-key gcp_audit_logs_key
out_masked ec2-asset-inv-key ec2_asset_inventory_key
out_masked asset-inv-public-ip ec2_asset_inventory_public_ip
out_masked ec2-wiz-key ec2_wiz_key
out_masked ec2-wiz-public-ip ec2_wiz_public_ip
out_masked elastic-defend-linux-public-ip ec2_elastic_defend_linux_public_ip
out_masked elastic-defend-linux-key ec2_elastic_defend_linux_key
out_masked elastic-defend-windows-public-ip ec2_elastic_defend_windows_public_ip
out_masked elastic-defend-windows-key ec2_elastic_defend_windows_key
out_masked elastic-defend-windows-instance-id ec2_elastic_defend_windows_instance_id
