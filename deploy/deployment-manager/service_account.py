"""A template file for deploying a service account for the elastic agent."""


def generate_config(context):
    """Generates configuration."""
    role_id = "elastic_agent_cspm_role"
    project = context.env["project"]

    service_account = {
        "name": "elastic-agent-cspm-sa",
        "type": "iam.v1.serviceAccount",
        "properties": {
            "accountId": "elastic-agent-cspm-sa",
            "displayName": "Elastic agent service account for CSPM",
            "projectId": context.env["project"],
        },
    }

    custom_role = {
        "name": "elastic-cspm-role",
        "type": "gcp-types/iam-v1:projects.roles",
        "properties": {
            "roleId": role_id,
            "parent": f"projects/{project}",
            "role": {
                "title": "Elastic CSPM role",
                "description": "Elastic CSPM role for GCP",
                "includedPermissions": [
                    "cloudasset.assets.listResource",
                    "cloudasset.assets.listIamPolicy",
                ],
            },
        },
    }

    iam_role_binding = {
        "name": "elastic-agent-iam-binding-cspm",
        "type": "gcp-types/cloudresourcemanager-v1:virtual.projects.iamMemberBinding",
        "properties": {
            "resource": context.env["project"],
            "role": f"projects/{project}/roles/{role_id}",
            "member": "serviceAccount:$(ref.elastic-agent-cspm-sa.email)",
        },
        "metadata": {
            "dependsOn": ["elastic-agent-cspm-sa", "elastic-cspm-role"],
        },
    }

    resources = [service_account, custom_role, iam_role_binding]

    return {"resources": resources}
