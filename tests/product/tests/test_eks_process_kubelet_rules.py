"""
Kubernetes CIS rules verification.
This module verifies correctness of retrieved findings by manipulating audit and remediation actions
"""

from datetime import datetime, timedelta

import pytest
from commonlib.utils import get_ES_evaluation
from product.tests.data.eks import eks_process_test_cases as eks_proc_tc
from product.tests.parameters import Parameters, register_params


@pytest.mark.eks_process_rules
def test_eks_process_rules(
    kspm_client,
    cloudbeat_agent,
    rule_tag,
    node_hostname,
    expected,
):
    """
    This data driven test verifies rules and findings return by cloudbeat agent.
    In order to add new cases @pytest.mark.parameterize section shall be updated.
    Setup and teardown actions are defined in data method.
    This test creates verifies that cloudbeat returns correct finding.
    @param rule_tag: Name of rule to be verified.
    @param node_hostname: EKS node hostname
    @param expected: Result to be found in finding evaluation field.
    @return: None - Test Pass / Fail result is generated.
    """
    # pylint: disable=duplicate-code

    def identifier(eval_resource):
        try:
            return eval_resource.host.name == node_hostname
        except AttributeError:
            return False

    evaluation = get_ES_evaluation(
        elastic_client=kspm_client,
        timeout=cloudbeat_agent.eks_findings_timeout,
        rule_tag=rule_tag,
        exec_timestamp=datetime.utcnow() - timedelta(hours=4),
        resource_identifier=identifier,
    )

    assert evaluation is not None, f"No evaluation for rule {rule_tag} could be found"
    assert evaluation == expected, f"Rule {rule_tag} verification failed," f"expected: {expected}, got: {evaluation}"


register_params(
    test_eks_process_rules,
    Parameters(
        ("rule_tag", "node_hostname", "expected"),
        [*eks_proc_tc.cis_eks_kubelet_cases.values()],
        ids=[*eks_proc_tc.cis_eks_kubelet_cases.keys()],
    ),
)
