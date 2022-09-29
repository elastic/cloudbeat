"""
Kubernetes CIS rules verification.
This module verifies correctness of retrieved findings by manipulating audit and remediation actions
"""
from datetime import datetime

import time
import pytest

from commonlib.utils import get_ES_evaluation, command_contains_arguments
from product.tests.data.process.process_test_cases import controller_manager_rules
from product.tests.parameters import register_params, Parameters


@pytest.mark.process_controller_manager_rules
def test_process_controller_manager(elastic_client,
                                    config_node_pre_test,
                                    rule_tag,
                                    dictionary,
                                    resource,
                                    expected):
    """
    This data driven test verifies rules and findings return by cloudbeat agent.
    In order to add new cases @pytest.mark.parameterize section shall be updated.
    Setup and teardown actions are defined in data method.
    This test creates cloudbeat agent instance, changes node resources (modes, users, groups) and verifies,
    that cloudbeat returns correct finding.
    @param rule_tag: Name of rule to be verified.
    @param dictionary: Set and Unset dictionary
    @param resource: Full path to resource / file
    @param expected: Result to be found in finding evaluation field.
    @return: None - Test Pass / Fail result is generated.
    """
    k8s_client, api_client, cloudbeat_agent = config_node_pre_test

    if "edit_process_file" not in dir(api_client):
        pytest.skip("skipping process rules run in non-containerized api_client")

    # Currently, single node is used, in the future may be extended for all nodes.
    node = k8s_client.get_cluster_nodes()[0]
    api_client.edit_process_file(container_name=node.metadata.name,
                                 dictionary=dictionary,
                                 resource=resource)

    # Wait for process reboot
    # TODO: Implement a more optimal way of waiting
    time.sleep(60)

    def identifier(eval_resource):
        return command_contains_arguments(eval_resource.command, dictionary)

    evaluation = get_ES_evaluation(
        elastic_client=elastic_client,
        timeout=cloudbeat_agent.findings_timeout,
        rule_tag=rule_tag,
        exec_timestamp=datetime.utcnow(),
        resource_identifier=identifier,
    )

    assert evaluation is not None, f"No evaluation for rule {rule_tag} could be found"
    assert evaluation == expected, f"Rule {rule_tag} verification failed, expected: {expected} actual: {evaluation}"


register_params(test_process_controller_manager, Parameters(
    ("rule_tag", "dictionary", "resource", "expected"),
    controller_manager_rules))
