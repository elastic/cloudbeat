"""
This module provides Azure storage account rule test cases.
Cases are organized as rules.
Each rule has one or more test cases.
Storage account identification is performed by resource name.
"""

from ..azure_test_case import AzureServiceCase
from ..constants import RULE_FAIL_STATUS, RULE_PASS_STATUS

CIS_3_1 = "CIS 3.1"
CIS_3_2 = "CIS 3.2"
CIS_3_7 = "CIS 3.7"
CIS_3_8 = "CIS 3.8"
CIS_3_9 = "CIS 3.9"
CIS_3_10 = "CIS 3.10"
CIS_3_15 = "CIS 3.15"
CIS_5_1_4 = "CIS 5.1.4"

cis_azure_5_1_4_pass = AzureServiceCase(
    rule_tag=CIS_5_1_4,
    case_identifier="testsapass",
    expected=RULE_PASS_STATUS,
)

cis_azure_5_1_4_fail = AzureServiceCase(
    rule_tag=CIS_5_1_4,
    case_identifier="testsafail",
    expected=RULE_FAIL_STATUS,
)

cis_azure_5_1_4 = {
    """5.1.4 Ensure the storage account containing the container with activity logs is encrypted
      with Customer Managed Key
      expect: passed""": cis_azure_5_1_4_pass,
    """5.1.4 Ensure the storage account containing the container with activity logs is encrypted
      with Customer Managed Key
      expect: failed""": cis_azure_5_1_4_fail,
}

cis_azure_3_1_pass = AzureServiceCase(
    rule_tag=CIS_3_1,
    case_identifier="testsapass",
    expected=RULE_PASS_STATUS,
)

cis_azure_3_1_fail = AzureServiceCase(
    rule_tag=CIS_3_1,
    case_identifier="testsafail",
    expected=RULE_FAIL_STATUS,
)

cis_azure_3_1 = {
    "3.1 Ensure that 'Secure transfer required' is set to 'Enabled' expect: passed": cis_azure_3_1_pass,
    "3.1 Ensure that 'Secure transfer required' is set to 'Enabled' expect: failed": cis_azure_3_1_fail,
}

cis_azure_3_2_pass = AzureServiceCase(
    rule_tag=CIS_3_2,
    case_identifier="testsapass",
    expected=RULE_PASS_STATUS,
)

cis_azure_3_2_fail = AzureServiceCase(
    rule_tag=CIS_3_2,
    case_identifier="testsafail",
    expected=RULE_FAIL_STATUS,
)

cis_azure_3_2 = {
    """3.2 Ensure that 'Enable Infrastructure Encryption' for Each Storage Account
      in Azure Storage is Set to 'enabled' expect: passed""": cis_azure_3_2_pass,
    """3.2 Ensure that 'Enable Infrastructure Encryption' for Each Storage Account
      in Azure Storage is Set to 'enabled' expect: failed""": cis_azure_3_2_fail,
}

cis_azure_3_7_pass = AzureServiceCase(
    rule_tag=CIS_3_7,
    case_identifier="testsapass",
    expected=RULE_PASS_STATUS,
)

cis_azure_3_7_fail = AzureServiceCase(
    rule_tag=CIS_3_7,
    case_identifier="testsafail",
    expected=RULE_FAIL_STATUS,
)

cis_azure_3_7 = {
    """3.7 Ensure that 'Public access level' is disabled
      for storage accounts with blob containers expect: passed""": cis_azure_3_7_pass,
    """3.7 Ensure that 'Public access level' is disabled
      for storage accounts with blob containers expect: failed""": cis_azure_3_7_fail,
}

cis_azure_3_8_pass = AzureServiceCase(
    rule_tag=CIS_3_8,
    case_identifier="testsapass",
    expected=RULE_PASS_STATUS,
)

cis_azure_3_8_fail = AzureServiceCase(
    rule_tag=CIS_3_8,
    case_identifier="testsafail",
    expected=RULE_FAIL_STATUS,
)

cis_azure_3_8 = {
    "3.8 Ensure Default Network Access Rule for Storage Accounts is Set to Deny expect: passed": cis_azure_3_8_pass,
    "3.8 Ensure Default Network Access Rule for Storage Accounts is Set to Deny expect: failed": cis_azure_3_8_fail,
}

cis_azure_3_9_pass = AzureServiceCase(
    rule_tag=CIS_3_9,
    case_identifier="testsansgflow",
    expected=RULE_PASS_STATUS,
)

cis_azure_3_9_fail = AzureServiceCase(
    rule_tag=CIS_3_9,
    case_identifier="testsafail",
    expected=RULE_FAIL_STATUS,
)

cis_azure_3_9 = {
    """3.9 Ensure 'Allow Azure services on the trusted services list to access this storage account'
      is Enabled for Storage Account Access expect: passed""": cis_azure_3_9_pass,
    """3.9 Ensure 'Allow Azure services on the trusted services list to access this storage account'
      is Enabled for Storage Account Access expect: failed""": cis_azure_3_9_fail,
}

cis_azure_3_10_pass = AzureServiceCase(
    rule_tag=CIS_3_10,
    case_identifier="testsapass",
    expected=RULE_PASS_STATUS,
)

cis_azure_3_10_fail = AzureServiceCase(
    rule_tag=CIS_3_10,
    case_identifier="testsafail",
    expected=RULE_FAIL_STATUS,
)

cis_azure_3_10 = {
    "3.10 Ensure Private Endpoints are used to access Storage Accounts expect: passed": cis_azure_3_10_pass,
    "3.10 Ensure Private Endpoints are used to access Storage Accounts expect: failed": cis_azure_3_10_fail,
}

cis_azure_3_15_pass = AzureServiceCase(
    rule_tag=CIS_3_15,
    case_identifier="testsapass",
    expected=RULE_PASS_STATUS,
)

cis_azure_3_15_fail = AzureServiceCase(
    rule_tag=CIS_3_15,
    case_identifier="testsafail",
    expected=RULE_FAIL_STATUS,
)

cis_azure_3_15 = {
    """3.15 Ensure the 'Minimum TLS version' for storage accounts
      is set to 'Version 1.2' expect: passed""": cis_azure_3_15_pass,
    """3.15 Ensure the 'Minimum TLS version' for storage accounts
      is set to 'Version 1.2' expect: failed""": cis_azure_3_15_fail,
}

# The name of this variable needs to be `tests_cases` in order to CIS Rules coverage stats to be generated
test_cases = {
    **cis_azure_3_1,
    **cis_azure_3_2,
    **cis_azure_3_7,
    **cis_azure_3_8,
    **cis_azure_3_9,
    **cis_azure_3_10,
    **cis_azure_3_15,
    **cis_azure_5_1_4,
}
