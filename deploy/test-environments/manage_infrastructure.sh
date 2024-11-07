#!/bin/bash

# Function to run Terraform in a given directory for apply, destroy, output, or upload operation
run_terraform() {
    local dir=$1
    local operation=$2

    echo "Running Terraform $operation in $dir..."
    cd "$dir" || exit

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
        terraform destroy -auto-approve && rm terraform.tfstate
        ;;
    "output")
        ../set_cloud_env_params.sh "$dir"
        ;;
    "upload-state")
        ../upload_state.sh "$dir"
        ;;
    *)
        echo "Invalid operation. Use 'apply', 'destroy', 'output', or 'upload-state'."
        cd - >/dev/null || exit 1
        ;;
    esac

    cd - >/dev/null || exit
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
    ;;
cis)
    run_terraform "elk-stack" "$action"
    run_terraform "cis" "$action"
    ;;
cdr)
    run_terraform "elk-stack" "$action"
    run_terraform "cdr" "$action"
    ;;
all)
    run_terraform "elk-stack" "$action"
    run_terraform "cdr" "$action"
    run_terraform "cis" "$action"
    ;;
*)
    echo "Usage: $0 {elk-stack|cis|cdr|all} {apply|destroy|output|upload-state}"
    exit 1
    ;;
esac

echo "Terraform $action operation completed."
