"""
Kubernetes CIS rules verification.
This module verifies correctness of retrieved findings by manipulating audit and remediation actions
"""

from datetime import datetime, timedelta
from functools import partial

import pytest
from commonlib.utils import get_ES_evaluation, res_identifier
from product.tests.data.k8s import k8s_process_cases as k8s_process_tc
from product.tests.parameters import Parameters, register_params

from .data.constants import RES_NAME


@pytest.mark.k8s_process_rules
def test_k8s_object_rules(
    kspm_client,
    cloudbeat_agent,
    rule_tag,
    resource_name,
    expected,
):
    """
    This data driven test verifies rules and findings return by cloudbeat agent.
    In order to add new cases @pytest.mark.parameterize section shall be updated.
    Setup and teardown actions are defined in data method.
    This test creates verifies that cloudbeat returns correct finding.
    @param rule_tag: Name of rule to be verified.
    @param resource_name: resource file name
    @param expected: Result to be found in finding evaluation field.
    @return: None - Test Pass / Fail result is generated.
    """
    process_identifier = partial(res_identifier, RES_NAME, resource_name)

    evaluation = get_ES_evaluation(
        elastic_client=kspm_client,
        timeout=cloudbeat_agent.findings_timeout,
        rule_tag=rule_tag,
        exec_timestamp=datetime.utcnow() - timedelta(minutes=15),
        resource_identifier=process_identifier,
    )

    assert evaluation is not None, f"No evaluation for rule {rule_tag} could be found"
    assert evaluation == expected, f"Rule {rule_tag} verification failed," f"expected: {expected}, got: {evaluation}"


register_params(
    test_k8s_object_rules,
    Parameters(
        ("rule_tag", "resource_name", "expected"),
        [*k8s_process_tc.test_cases_by_config.values()],
        ids=[*k8s_process_tc.test_cases_by_config.keys()],
    ),
)
