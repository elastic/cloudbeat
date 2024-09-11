"""
CIS Azure Key Vault rules verification.
This module verifies correctness of retrieved findings by manipulating audit actions
"""

from datetime import datetime, timedelta
from functools import partial

import pytest
from commonlib.utils import get_ES_evaluation, res_identifier
from product.tests.data.azure import azure_key_vault_test_cases as azure_key_vault_tc
from product.tests.parameters import Parameters, register_params

from .data.constants import RES_NAME


@pytest.mark.cspm_azure_key_vault_rules
def test_azure_key_vault_rules(
    cspm_client,
    cloudbeat_agent,
    rule_tag,
    case_identifier,
    expected,
):
    """
    This data driven test verifies rules and findings return by cloudbeat agent.
    In order to add new cases @pytest.mark.parameterize section shall be updated.
    Setup and teardown actions are defined in data method.
    This test verifies that cloudbeat returns correct finding.
    @param rule_tag: Name of rule to be verified.
    @param case_identifier: Resource unique identifier
    @param expected: Result to be found in finding evaluation field.
    @return: None - Test Pass / Fail result is generated.
    """
    key_vault_identifier = partial(res_identifier, RES_NAME, case_identifier)

    evaluation = get_ES_evaluation(
        elastic_client=cspm_client,
        timeout=cloudbeat_agent.azure_findings_timeout,
        rule_tag=rule_tag,
        exec_timestamp=datetime.utcnow() - timedelta(minutes=30),
        resource_identifier=key_vault_identifier,
    )

    assert evaluation is not None, f"No evaluation for rule {rule_tag} could be found"
    assert evaluation == expected, f"Rule {rule_tag} verification failed," f"expected: {expected}, got: {evaluation}"


register_params(
    test_azure_key_vault_rules,
    Parameters(
        ("rule_tag", "case_identifier", "expected"),
        [*azure_key_vault_tc.test_cases.values()],
        ids=[*azure_key_vault_tc.test_cases.keys()],
    ),
)
