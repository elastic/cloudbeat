"""
Kubernetes CIS rules verification.
This module verifies correctness of retrieved findings by manipulating audit and remediation actions
"""
from datetime import datetime

import pytest

from commonlib.utils import get_evaluation, get_resource_identifier
from product.tests.kube_rules import *


def todict(obj):
    if hasattr(obj, 'attribute_map'):
        result = {}
        for k, v in getattr(obj, 'attribute_map').items():
            val = getattr(obj, k)
            if val is not None:
                result[v] = todict(val)
        return result
    elif type(obj) == list:
        return [todict(x) for x in obj]
    elif type(obj) == datetime:
        return str(obj)
    else:
        return obj


@pytest.mark.rules
@pytest.mark.parametrize(
    ("rule_tag", "resource_type", "resource_body", "expected"),
    [
        *cis_5_1_3.values(),
        *cis_5_1_5.values(),
        *cis_5_1_6.values(),
        *cis_5_2_2.values(),
        *cis_5_2_3.values(),
        *cis_5_2_4.values(),
        *cis_5_2_5.values(),
        *cis_5_2_6.values(),
        *cis_5_2_7.values(),
        # *cis_5_2_8.values(),
        # *cis_5_2_9.values(),
        # *cis_5_2_10.values(),
    ],
    ids=[
        *cis_5_1_3.keys(),
        *cis_5_1_5.keys(),
        *cis_5_1_6.keys(),
        *cis_5_2_2.keys(),
        *cis_5_2_3.keys(),
        *cis_5_2_4.keys(),
        *cis_5_2_5.keys(),
        *cis_5_2_6.keys(),
        *cis_5_2_7.keys(),
        # *cis_5_2_8.keys(),
        # *cis_5_2_9.keys(),
        # *cis_5_2_10.keys(),
    ]
)
def test_kube_resource_patch(test_env, rule_tag, resource_type, resource_body, expected):
    """
    Test kube resource
    @param test_env: pre step that set-ups a kube resources to test on
    @param rule_tag: rule tag in the CIS benchmark
    @param resource_type: kube resource type, e.g., Pod, ServiceAccount, etc.
    @param resource_body: a dict to represent the relevant properties of the resource
    @param expected: "failed" or "passed"
    """
    k8s_client, _, agent_config = test_env

    # make sure resource exists
    metadata = resource_body['metadata']
    relevant_metadata = {k: metadata[k] for k in ('name', 'namespace') if k in metadata}
    try:
        resource = k8s_client.get_resource(resource_type=resource_type, **relevant_metadata)
    except TypeError as e:
        print(e)
        resource = k8s_client.get_resource(resource_type=resource_type,
                                           namespace=agent_config.namespace,
                                           **relevant_metadata)

    assert resource, f"Resource {resource_type} not found"

    # patch resource
    k8s_client.patch_resources(
        resource_type=resource_type,
        body=resource_body,
        **relevant_metadata
    )

    # check resource evaluation
    pods = k8s_client.get_agent_pod_instances(agent_name=agent_config.name, namespace=agent_config.namespace)

    evaluation = get_evaluation(
        k8s=k8s_client,
        timeout=agent_config.findings_timeout,
        pod_name=pods[0].metadata.name,
        namespace=agent_config.namespace,
        rule_tag=rule_tag,
        exec_timestamp=datetime.utcnow(),
        resource_identifier=get_resource_identifier(resource_body)
    )

    assert evaluation == expected, f"Rule {rule_tag} verification failed. expected: {expected} actual: {evaluation}"
