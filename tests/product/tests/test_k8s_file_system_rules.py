"""
Kubernetes CIS rules verification.
This module verifies correctness of retrieved findings by manipulating audit and remediation actions
"""

from datetime import datetime, timedelta

import pytest
from commonlib.utils import get_ES_evaluation
from product.tests.data.k8s import fs_test_cases as k8s_fs_tc
from product.tests.parameters import Parameters, register_params


@pytest.mark.k8s_file_system_rules
def test_k8s_file_system_configuration(
    kspm_client,
    cloudbeat_agent,
    rule_tag,
    node_hostname,
    resource_name,
    expected,
):
    """
    This data driven test verifies rules and findings return by cloudbeat agent.
    In order to add new cases @pytest.mark.parameterize section shall be updated.
    Setup and teardown actions are defined in data method.
    This test creates verifies that cloudbeat returns correct finding.
    @param rule_tag: Name of rule to be verified.
    @param node_hostname: k8s node hostname
    @param resource_name: resource file name
    @param expected: Result to be found in finding evaluation field.
    @return: None - Test Pass / Fail result is generated.
    """
    # pylint: disable=duplicate-code

    # file_identifier = partial(res_identifier, RES_HOST_NAME, node_hostname)
    def file_identifier(eval_resource):
        try:
            return eval_resource.host.name == node_hostname and eval_resource.resource.name == resource_name
        except AttributeError:
            return False

    evaluation = get_ES_evaluation(
        elastic_client=kspm_client,
        timeout=cloudbeat_agent.findings_timeout,
        rule_tag=rule_tag,
        exec_timestamp=datetime.utcnow() - timedelta(minutes=15),
        resource_identifier=file_identifier,
    )

    assert evaluation is not None, f"No evaluation for rule {rule_tag} could be found"
    assert evaluation == expected, f"Rule {rule_tag} verification failed," f"expected: {expected}, got: {evaluation}"


register_params(
    test_k8s_file_system_configuration,
    Parameters(
        ("rule_tag", "node_hostname", "resource_name", "expected"),
        [*k8s_fs_tc.test_cases.values()],
        ids=[*k8s_fs_tc.test_cases.keys()],
    ),
)
