#!/bin/bash

# Function to run Terraform in a given directory for apply, destroy, output, or upload operation
run_terraform() {
    local dir=$1
    local operation=$2
    local terraform_rc

    echo "Running Terraform $operation in $dir..."
    cd "$dir" || exit 1

    case $operation in
    "apply")
        terraform init
        terraform validate
        terraform apply -auto-approve
        ;;
    "destroy")
        terraform init
        if [ "$dir" == "cis" ] && terraform state list | grep -q "kubernetes_config_map_v1_data.aws_auth"; then
            echo "Removing aws_auth resource from state in cis..."
            terraform state rm "$(terraform state list | grep "kubernetes_config_map_v1_data.aws_auth")"
        fi
        # Destroy still evaluates module variable validation; CDR apply sets TF_VAR_* in CI, generic destroy does not.
        if [ "$dir" == "cdr" ] && [ -z "${TF_VAR_windows_elastic_defend_winrm_ingress_cidr:-}" ]; then
            export TF_VAR_windows_elastic_defend_winrm_ingress_cidr="127.0.0.1/32"
        fi
        terraform destroy -auto-approve && rm terraform.tfstate
        ;;
    "output")
        ../set_cloud_env_params.sh "$dir"
        ;;
    "upload-state")
        ../upload_state.sh "$dir"
        ;;
    *)
        echo "Invalid operation. Use 'apply', 'destroy', 'output', or 'upload-state'." >&2
        false
        ;;
    esac

    terraform_rc=$?
    cd - >/dev/null || exit 1
    return "$terraform_rc"
}

# Check for valid input
if [ "$#" -ne 2 ]; then
    echo "Usage: $0 {elk-stack|cis|cdr|all} {apply|destroy|output|upload-state}"
    exit 1
fi

# Main script logic
action=$2

case $1 in
elk-stack)
    run_terraform "elk-stack" "$action"
    overall_rc=$?
    ;;
cis)
    overall_rc=0
    run_terraform "elk-stack" "$action" || overall_rc=1
    run_terraform "cis" "$action" || overall_rc=1
    ;;
cdr)
    overall_rc=0
    run_terraform "elk-stack" "$action" || overall_rc=1
    run_terraform "cdr" "$action" || overall_rc=1
    ;;
all)
    overall_rc=0
    run_terraform "elk-stack" "$action" || overall_rc=1
    run_terraform "cdr" "$action" || overall_rc=1
    run_terraform "cis" "$action" || overall_rc=1
    ;;
*)
    echo "Usage: $0 {elk-stack|cis|cdr|all} {apply|destroy|output|upload-state}"
    exit 1
    ;;
esac

if [ "$overall_rc" -eq 0 ]; then
    echo "Terraform $action operation completed."
else
    echo "Terraform $action completed with errors (one or more stacks failed)." >&2
fi
exit "$overall_rc"
