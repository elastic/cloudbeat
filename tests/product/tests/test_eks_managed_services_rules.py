"""
Kubernetes CIS rules verification.
This module verifies correctness of retrieved findings by manipulating audit actions
"""

from datetime import datetime, timedelta

import pytest
from commonlib.utils import get_ES_evaluation
from product.tests.data.aws import managed_services_test_cases as ms_tc
from product.tests.parameters import Parameters, register_params


@pytest.mark.eks_service_rules
def test_eks_service_rules(
    kspm_client,
    cloudbeat_agent,
    rule_tag,
    case_identifier,
    expected,
):
    """
    This data driven test verifies rules and findings return by cloudbeat agent.
    In order to add new cases @pytest.mark.parameterize section shall be updated.
    Setup and teardown actions are defined in data method.
    This test creates verifies that cloudbeat returns correct finding.
    @param rule_tag: Name of rule to be verified.
    @param case_identifier: Resource unique identifier
    @param expected: Result to be found in finding evaluation field.
    @return: None - Test Pass / Fail result is generated.
    """
    # pylint: disable=duplicate-code

    def identifier(eval_resource):
        try:
            eval_resource = eval_resource.resource
            return eval_resource.name == case_identifier
        except AttributeError:
            return False

    evaluation = get_ES_evaluation(
        elastic_client=kspm_client,
        timeout=cloudbeat_agent.eks_findings_timeout,
        rule_tag=rule_tag,
        exec_timestamp=datetime.utcnow() - timedelta(hours=1),
        resource_identifier=identifier,
    )

    assert evaluation is not None, f"No evaluation for rule {rule_tag} could be found"
    assert evaluation == expected, f"Rule {rule_tag} verification failed," f"expected: {expected}, got: {evaluation}"


register_params(
    test_eks_service_rules,
    Parameters(
        ("rule_tag", "case_identifier", "expected"),
        [*ms_tc.test_cases.values()],
        ids=[*ms_tc.test_cases.keys()],
    ),
)
