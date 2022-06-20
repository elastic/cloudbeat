"""
Kubernetes CIS rules verification.
This module verifies correctness of retrieved findings by manipulating audit and remediation actions
"""
from datetime import datetime

import pytest

from commonlib.utils import get_evaluation
from product.tests.tests.file_system.file_system_test_cases import *


@pytest.mark.parametrize(
    ("rule_tag", "command", "param_value", "resource", "expected"),
    [*cis_1_1_1,
     *cis_1_1_2,
     *cis_1_1_3,
     *cis_1_1_4,
     *cis_1_1_5,
     *cis_1_1_6,
     *cis_1_1_7,
     *cis_1_1_8,
     *cis_1_1_11,
     *cis_1_1_12,
     *cis_1_1_13,
     *cis_1_1_14,
     *cis_1_1_15,
     *cis_1_1_16,
     *cis_1_1_17,
     *cis_1_1_18,
     *cis_1_1_19,
     *cis_4_1_1,
     *cis_4_1_2,
     *cis_4_1_5,
     *cis_4_1_9,
     *cis_4_1_10
     ],
)
def test_file_system_configuration(config_node_pre_test,
                                   rule_tag,
                                   command,
                                   param_value,
                                   resource,
                                   expected):
    """
    This data driven test verifies rules and findings return by cloudbeat agent.
    In order to add new cases @pytest.mark.parameterize section shall be updated.
    Setup and teardown actions are defined in data method.
    This test creates cloudbeat agent instance, changes node resources (modes, users, groups) and verifies,
    that cloudbeat returns correct finding.
    @param rule_tag: Name of rule to be verified.
    @param command: Command to be executed, for example chmod / chown
    @param param_value: Value to be used when executing command.
    @param resource: Full path to resource / file
    @param expected: Result to be found in finding evaluation field.
    @return: None - Test Pass / Fail result is generated.
    """
    k8s_client, api_client, cloudbeat_agent = config_node_pre_test
    # Currently, single node is used, in the future may be extended for all nodes.
    node = k8s_client.get_cluster_nodes()[0]
    pods = k8s_client.get_agent_pod_instances(agent_name=cloudbeat_agent.name, namespace=cloudbeat_agent.namespace)

    api_client.exec_command(container_name=node.metadata.name,
                            command=command,
                            param_value=param_value,
                            resource=resource)

    evaluation = get_evaluation(
        k8s=k8s_client,
        timeout=cloudbeat_agent.findings_timeout,
        pod_name=pods[0].metadata.name,
        namespace=cloudbeat_agent.namespace,
        rule_tag=rule_tag,
        exec_timestamp=datetime.utcnow()
    )

    assert evaluation == expected, f"Rule {rule_tag} verification failed."
