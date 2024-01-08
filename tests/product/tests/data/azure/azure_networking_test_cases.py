"""
This module provides Azure networking rule test cases.
Cases are organized as rules.
Each rule has one or more test cases.
Networking identification is performed by resource name.
"""
from ..azure_test_case import AzureServiceCase
from ..constants import RULE_PASS_STATUS, RULE_FAIL_STATUS

CIS_6_6 = "CIS 6.6"
CIS_6_5 = "CIS 6.5"

cis_azure_6_6_pass = AzureServiceCase(
    rule_tag=CIS_6_6,
    case_identifier="test-networking-pass",
    expected=RULE_PASS_STATUS,
)

cis_azure_6_6_fail = AzureServiceCase(
    rule_tag=CIS_6_6,
    case_identifier="test-networking-fail",
    expected=RULE_FAIL_STATUS,
)

cis_azure_6_6 = {
    "6.6 Ensure that Network Watcher is 'Enabled' expect: passed": cis_azure_6_6_pass,
    "6.6 Ensure that Network Watcher is 'Enabled' expect: failed": cis_azure_6_6_fail,
}

cis_azure_6_5_pass = AzureServiceCase(
    rule_tag=CIS_6_5,
    case_identifier="test-networking-pass",
    expected=RULE_PASS_STATUS,
)

cis_azure_6_5_fail = AzureServiceCase(
    rule_tag=CIS_6_5,
    case_identifier="test-networking-fail",
    expected=RULE_FAIL_STATUS,
)

cis_azure_6_5 = {
    """6.5 Ensure that Network Security Group Flow Log retention period
      is 'greater than 90 days' expect: passed""": cis_azure_6_5_pass,
    """6.5 Ensure that Network Security Group Flow Log retention period
      is 'greater than 90 days' expect: failed""": cis_azure_6_5_fail,
}

cis_azure_networking_cases = {
    **cis_azure_6_6,
    **cis_azure_6_5,
}
