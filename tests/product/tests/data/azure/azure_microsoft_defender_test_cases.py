"""
This module provides Azure microsoft defender rule test cases.
Cases are organized as rules.
Each rule has one or more test cases.
Microsoft Defender identification is performed by resource name.
"""

from ..azure_test_case import AzureServiceCase
from ..constants import RULE_FAIL_STATUS, RULE_PASS_STATUS

CIS_2_1_15 = "CIS 2.1.15"
CIS_2_1_18 = "CIS 2.1.18"
CIS_2_1_19 = "CIS 2.1.19"
CIS_2_1_20 = "CIS 2.1.20"

cis_azure_2_1_15_fail = AzureServiceCase(
    rule_tag=CIS_2_1_15,
    case_identifier="azure-security-auto-provisioning-settings-ef111ee2-6c89-4b09-92c6-5c2321f888df",
    expected=RULE_FAIL_STATUS,
)

cis_azure_2_1_18_pass = AzureServiceCase(
    rule_tag=CIS_2_1_18,
    case_identifier="azure-security-contacts-ef111ee2-6c89-4b09-92c6-5c2321f888df",
    expected=RULE_PASS_STATUS,
)

cis_azure_2_1_19_pass = AzureServiceCase(
    rule_tag=CIS_2_1_19,
    case_identifier="azure-security-contacts-ef111ee2-6c89-4b09-92c6-5c2321f888df",
    expected=RULE_PASS_STATUS,
)

cis_azure_2_1_20_pass = AzureServiceCase(
    rule_tag=CIS_2_1_20,
    case_identifier="azure-security-contacts-ef111ee2-6c89-4b09-92c6-5c2321f888df",
    expected=RULE_PASS_STATUS,
)

cis_azure_2_1_15 = {
    "2.1.15 Ensure auto provisioning of vm log analytics agent expect: failed": cis_azure_2_1_15_fail,
}

cis_azure_2_1_18 = {
    "2.1.18 Ensure security alert emails to subscription owners expect: passed": cis_azure_2_1_18_pass,
}

cis_azure_2_1_19 = {
    "2.1.19 Ensure additional email addresses is configured expect: passed": cis_azure_2_1_19_pass,
}

cis_azure_2_1_20 = {
    "2.1.20 Ensure that notification alert severity is set to 'High' expect: passed": cis_azure_2_1_20_pass,
}

# The name of this variable needs to be `tests_cases` in order to CIS Rules coverage stats to be generated
test_cases = {
    **cis_azure_2_1_15,
    **cis_azure_2_1_18,
    **cis_azure_2_1_19,
    **cis_azure_2_1_20,
}
