"""A template file for deploying a compute engine instance."""

COMPUTE_URL_BASE = 'https://www.googleapis.com/compute/v1/'


def GenerateConfig(context):
    """Generates configuration."""
    project = context.env['project']
    zone = context.properties['zone']
    enrollment_token = context.properties['enrollmentToken']
    fleet_url = context.properties['fleetUrl']
    agent_version = context.properties['elasticAgentVersion']
    artifact_server = context.properties['elasticArtifactServer']

    ssh_fw_rule = {  # firewall
        'name': 'elastic-agent-firewall-rule',
        'type': 'compute.v1.firewall',
        'properties': {
            'network': '$(ref.elastic-agent-network.selfLink)',
            'sourceRanges': ['0.0.0.0/0'],
            'allowed': [{
                'IPProtocol': 'TCP',
                'ports': [22],
            }],
        },
    }

    instance = {
        'name': 'elastic-agent',
        'type': 'compute.v1.instance',
        'properties': {
            'zone': zone,
            'machineType': ''.join([
                COMPUTE_URL_BASE, 'projects/', project,
                '/zones/', zone, '/',
                'machineTypes/n2-standard-4',
            ]),
            'serviceAccounts': [{
                'email': '$(ref.elastic-agent-cspm-sa.email)',
                'scopes': ['https://www.googleapis.com/auth/cloud-platform'],
            }],
            'disks': [{
                'deviceName': 'boot',
                'type': 'PERSISTENT',
                'boot': True,
                'sizeGb': 32,
                'autoDelete': True,
                'initializeParams': {
                    'sourceImage': ''.join([
                        COMPUTE_URL_BASE, 'projects/',
                        'ubuntu-os-cloud/global',
                        '/images/family/ubuntu-minimal-2204-lts',
                    ]),
                },
            }],
            'metadata': {
                'dependsOn': ['elastic-agent-cspm-sa'],
                'items': [
                    {
                        'key': 'startup-script',
                        'value': ''.join([
                            '#!/bin/bash\n',
                            'set -x\n',
                            f'ElasticAgentArtifact=elastic-agent-{agent_version}-linux-x86_64\n',
                            f'curl -L -O {artifact_server}/$ElasticAgentArtifact.tar.gz\n',
                            'tar xzvf $ElasticAgentArtifact.tar.gz\n',
                            'cd $ElasticAgentArtifact\n',
                            f'sudo ./elastic-agent install '
                            f'--non-interactive --url={fleet_url} --enrollment-token={enrollment_token}',
                        ]),
                    },
                ],
            },
            'networkInterfaces': [{
                'network': '$(ref.elastic-agent-network.selfLink)',
                'accessConfigs': [{
                    'name': 'External NAT',
                    'type': 'ONE_TO_ONE_NAT',
                }],
            }],
            'labels': {
                'name': 'elastic-agent',
            },
        },
    }

    service_account = {
        'name': 'elastic-agent-cspm-sa',
        'type': 'iam.v1.serviceAccount',
        'properties': {
            'accountId': 'elastic-agent-cspm-sa',
            'displayName': 'Elastic agent service account for CSPM',
            'projectId': context.env['project'],
        },
    }

    iam_role_binding = {
        'name': 'elastic-agent-iam-binding-cspm',
        'type': 'gcp-types/cloudresourcemanager-v1:virtual.projects.iamMemberBinding',
        'properties': {
            'resource': context.env['project'],
            'role': 'roles/cloudasset.viewer',
            'member': 'serviceAccount:$(ref.elastic-agent-cspm-sa.email)',
        },
        'metadata': {
            'dependsOn': ['elastic-agent-cspm-sa'],
        },
    }

    network = {
        'name': 'elastic-agent-network',
        'type': 'compute.v1.network',
        'properties': {
            'routingConfig': {
                'routingMode': 'REGIONAL',
            },
            'autoCreateSubnetworks': True,
        },
    }

    resources = [instance, service_account, iam_role_binding, network]

    if context.properties['allowSSH']:
        resources.append(ssh_fw_rule)

    return {'resources': resources}
