"""
This module provides Azure key vault rule test cases.
Cases are organized as rules.
Each rule has one or more test cases.
Key vault identification is performed by resource name.
"""

from ..azure_test_case import AzureServiceCase
from ..constants import RULE_FAIL_STATUS, RULE_PASS_STATUS

CIS_8_5 = "CIS 8.5"
CIS_5_1_5 = "CIS 5.1.5"

cis_azure_8_5_pass = AzureServiceCase(
    rule_tag=CIS_8_5,
    case_identifier="test-key-vault-pass",
    expected=RULE_PASS_STATUS,
)

cis_azure_8_5_fail = AzureServiceCase(
    rule_tag=CIS_8_5,
    case_identifier="test-key-vault-fail-arm",
    expected=RULE_FAIL_STATUS,
)

cis_azure_8_5 = {
    "8.5 Ensure the Key Vault is Recoverable expect: passed": cis_azure_8_5_pass,
    "8.5 Ensure the Key Vault is Recoverable expect: failed": cis_azure_8_5_fail,
}

cis_azure_5_1_5_pass = AzureServiceCase(
    rule_tag=CIS_5_1_5,
    case_identifier="test-key-vault-diag-pass",
    expected=RULE_PASS_STATUS,
)
cis_azure_5_1_5_fail = AzureServiceCase(
    rule_tag=CIS_5_1_5,
    case_identifier="test-key-vault-diag-fail",
    expected=RULE_FAIL_STATUS,
)

cis_azure_5_1_5 = {
    "5.1.5 Ensure that logging for Azure Key Vault is 'Enabled' expect: passed": cis_azure_5_1_5_pass,
    "5.1.5 Ensure that logging for Azure Key Vault is 'Enabled' expect: failed": cis_azure_5_1_5_fail,
}

# The name of this variable needs to be `tests_cases` in order to CIS Rules coverage stats to be generated
test_cases = {
    **cis_azure_8_5,
    **cis_azure_5_1_5,
}
