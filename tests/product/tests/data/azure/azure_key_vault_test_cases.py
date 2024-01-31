"""
This module provides Azure key vault rule test cases.
Cases are organized as rules.
Each rule has one or more test cases.
Key vault identification is performed by resource name.
"""

from ..azure_test_case import AzureServiceCase
from ..constants import RULE_PASS_STATUS, RULE_FAIL_STATUS

CIS_8_5 = "CIS 8.5"

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

# The name of this variable needs to be `tests_cases` in order to CIS Rules coverage stats to be generated
test_cases = {
    **cis_azure_8_5,
}
