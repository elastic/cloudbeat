"""
This module provides Azure networking rule test cases.
Cases are organized as rules.
Each rule has one or more test cases.
Networking identification is performed by resource name.
"""

from ..azure_test_case import AzureServiceCase
from ..constants import RULE_FAIL_STATUS, RULE_PASS_STATUS

CIS_6_1 = "CIS 6.1"
CIS_6_2 = "CIS 6.2"
CIS_6_3 = "CIS 6.3"
CIS_6_4 = "CIS 6.4"
CIS_6_6 = "CIS 6.6"
CIS_6_5 = "CIS 6.5"

cis_azure_6_1_pass = AzureServiceCase(
    rule_tag=CIS_6_1,
    case_identifier="test-vm-pass",
    expected=RULE_PASS_STATUS,
)

cis_azure_6_1_fail = AzureServiceCase(
    rule_tag=CIS_6_1,
    case_identifier="test-vm-fail",
    expected=RULE_FAIL_STATUS,
)

cis_azure_6_1 = {
    "6.1 Ensure that RDP access from the Internet is evaluated and restricted expect: passed": cis_azure_6_1_pass,
    "6.1 Ensure that RDP access from the Internet is evaluated and restricted expect: failed": cis_azure_6_1_fail,
}

cis_azure_6_2_pass = AzureServiceCase(
    rule_tag=CIS_6_2,
    case_identifier="test-vm-pass",
    expected=RULE_PASS_STATUS,
)

cis_azure_6_2_fail = AzureServiceCase(
    rule_tag=CIS_6_2,
    case_identifier="test-vm-fail",
    expected=RULE_FAIL_STATUS,
)

cis_azure_6_2 = {
    "6.2 Ensure that SSH access from the Internet is evaluated and restricted expect: passed": cis_azure_6_2_pass,
    "6.2 Ensure that SSH access from the Internet is evaluated and restricted expect: failed": cis_azure_6_2_fail,
}

cis_azure_6_3_pass = AzureServiceCase(
    rule_tag=CIS_6_3,
    case_identifier="test-vm-pass",
    expected=RULE_PASS_STATUS,
)

cis_azure_6_3_fail = AzureServiceCase(
    rule_tag=CIS_6_3,
    case_identifier="test-vm-fail",
    expected=RULE_FAIL_STATUS,
)

cis_azure_6_3 = {
    "6.3 Ensure that UDP access from the Internet is evaluated and restricted expect: passed": cis_azure_6_3_pass,
    "6.3 Ensure that UDP access from the Internet is evaluated and restricted expect: failed": cis_azure_6_3_fail,
}

cis_azure_6_4_pass = AzureServiceCase(
    rule_tag=CIS_6_4,
    case_identifier="test-vm-pass",
    expected=RULE_PASS_STATUS,
)

cis_azure_6_4_fail = AzureServiceCase(
    rule_tag=CIS_6_4,
    case_identifier="test-vm-fail",
    expected=RULE_FAIL_STATUS,
)

cis_azure_6_4 = {
    "6.4 Ensure that HTTP(S) access from the Internet is evaluated and restricted expect: passed": cis_azure_6_4_pass,
    "6.4 Ensure that HTTP(S) access from the Internet is evaluated and restricted expect: failed": cis_azure_6_4_fail,
}

cis_azure_6_6_pass = AzureServiceCase(
    rule_tag=CIS_6_6,
    case_identifier="azure-network-watcher-japaneast-ef111ee2-6c89-4b09-92c6-5c2321f888df",
    expected=RULE_PASS_STATUS,
)

cis_azure_6_6_fail = AzureServiceCase(
    rule_tag=CIS_6_6,
    case_identifier="azure-network-watcher-ukwest-ef111ee2-6c89-4b09-92c6-5c2321f888df",
    expected=RULE_FAIL_STATUS,
)

cis_azure_6_6 = {
    "6.6 Ensure that Network Watcher is 'Enabled' expect: passed": cis_azure_6_6_pass,
    "6.6 Ensure that Network Watcher is 'Enabled' expect: failed": cis_azure_6_6_fail,
}

cis_azure_6_5_pass = AzureServiceCase(
    rule_tag=CIS_6_5,
    case_identifier="test-vm-pass-nsg-azurecloudbeatcitests-flowlog",
    expected=RULE_PASS_STATUS,
)

cis_azure_6_5_fail = AzureServiceCase(
    rule_tag=CIS_6_5,
    case_identifier="test-vm-fail-nsg-azurecloudbeatcitests-flowlog",
    expected=RULE_FAIL_STATUS,
)

cis_azure_6_5 = {
    """6.5 Ensure that Network Security Group Flow Log retention period
      is 'greater than 90 days' expect: passed""": cis_azure_6_5_pass,
    """6.5 Ensure that Network Security Group Flow Log retention period
      is 'greater than 90 days' expect: failed""": cis_azure_6_5_fail,
}

# The name of this variable needs to be `tests_cases` in order to CIS Rules coverage stats to be generated
test_cases = {
    **cis_azure_6_1,
    **cis_azure_6_2,
    **cis_azure_6_3,
    **cis_azure_6_4,
    **cis_azure_6_6,
    **cis_azure_6_5,
}
