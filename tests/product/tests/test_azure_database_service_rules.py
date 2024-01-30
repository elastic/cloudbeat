"""
CIS Azure Database Service rules verification.
This module verifies correctness of retrieved findings by manipulating audit actions
"""

from datetime import datetime, timedelta
from functools import partial
import pytest
from commonlib.utils import get_ES_evaluation, res_identifier

from product.tests.data.azure import azure_database_service_test_cases as azure_database_service_tc
from product.tests.parameters import register_params, Parameters
from .data.constants import RES_NAME


@pytest.mark.azure_database_service_rules
def test_azure_database_service_rules(
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
    database_service_identifier = partial(res_identifier, RES_NAME, case_identifier)

    evaluation = get_ES_evaluation(
        elastic_client=cspm_client,
        timeout=cloudbeat_agent.azure_findings_timeout,
        rule_tag=rule_tag,
        exec_timestamp=datetime.utcnow() - timedelta(minutes=30),
        resource_identifier=database_service_identifier,
    )

    assert evaluation is not None, f"No evaluation for rule {rule_tag} could be found"
    assert evaluation == expected, f"Rule {rule_tag} verification failed," f"expected: {expected}, got: {evaluation}"


register_params(
    test_azure_database_service_rules,
    Parameters(
        ("rule_tag", "case_identifier", "expected"),
        [*azure_database_service_tc.cis_azure_database_service_cases.values()],
        ids=[*azure_database_service_tc.cis_azure_database_service_cases.keys()],
    ),
)
