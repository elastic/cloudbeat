"""A template file for deploying a service account"""


def generate_config(context):
    """Generates service account user"""
    deployment_name = context.env["deployment"]
    scope = context.properties["scope"]
    parent_id = context.properties["parentId"]
    sa_name = context.properties["serviceAccountName"]
    roles = ["roles/cloudasset.viewer", "roles/browser"]

    # pylint: disable=duplicate-code
    service_account = {
        "name": sa_name,
        "type": "iam.v1.serviceAccount",
        "properties": {
            "accountId": f"{sa_name}-id",
            "displayName": "Elastic agent service account for CSPM",
            "projectId": context.env["project"],
        },
    }

    service_account_key = {
        "name": f"{deployment_name}-sa-key",
        "type": "iam.v1.serviceAccounts.key",
        "metadata": {
            "dependsOn": [sa_name],
        },
        "properties": {
            "parent": f"$(ref.{sa_name}.name)",
        },
    }

    # pylint: disable=duplicate-code
    bindings = []
    for role in roles:
        bindings.append(
            {
                "name": f"{deployment_name}-iam-binding-{role}",
                "type": f"gcp-types/cloudresourcemanager-v1:virtual.{scope}.iamMemberBinding",
                "properties": {
                    "resource": get_resource_name(scope, parent_id),
                    "role": role,
                    "member": f"serviceAccount:$(ref.{sa_name}.email)",
                },
                "metadata": {
                    "dependsOn": [sa_name],
                },
            },
        )

    resources = [service_account, service_account_key]
    resources.extend(bindings)

    return {
        "outputs": [{"name": "serviceAccountKey", "value": f"$(ref.{deployment_name}-sa-key.privateKeyData)"}],
        "resources": resources,
    }


def get_resource_name(scope, parent_id):
    """return the resource name based on the scope."""
    if scope == "organizations":
        return f"{scope}/{parent_id}"
    return parent_id
