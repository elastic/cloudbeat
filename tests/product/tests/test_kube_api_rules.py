"""
Kubernetes CIS rules verification.
This module verifies correctness of retrieved findings by manipulating audit and remediation actions
"""
import pytest

from commonlib.utils import *
from kube_rules import *


@pytest.mark.rules
@pytest.mark.parametrize(
    ("rule_tag", "resource_type", "resource_body", "expected"),
    [
        *cis_5_1_5.values(),
    ],
    ids=[
        *cis_5_1_5.keys(),
    ]
)
def test_kube_resource_patch(setup_busybox_pod, rule_tag, resource_type, resource_body, expected):
    k8s_client, _, agent_config = setup_busybox_pod

    # make sure resource exists
    resource_name = resource_body["metadata"]["name"]
    resource = k8s_client.get_resource(
        resource_type=resource_type,
        name=resource_name,
        namespace=agent_config.namespace
    )

    assert resource, f"Resource {resource_type} not found"

    # patch resource
    k8s_client.patch_resources(
        name=resource_name,
        resource_type=resource_type,
        namespace=agent_config.namespace,
        body=resource_body,
    )

    # check resource evaluation
    pods = k8s_client.get_agent_pod_instances(agent_name=agent_config.name, namespace=agent_config.namespace)

    evaluation = get_evaluation(
        k8s=k8s_client,
        timeout=agent_config.findings_timeout,
        pod_name=pods[0].metadata.name,
        namespace=agent_config.namespace,
        rule_tag=rule_tag,
        resource_identifier=get_resource_identifier(resource_body)
    )

    assert evaluation == expected, f"Rule {rule_tag} verification failed. expected: {expected} actual: {evaluation}"
