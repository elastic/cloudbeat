"""
This module provides Azure virtual machine rule test cases.
Cases are organized as rules.
Each rule has one or more test cases.
Virtual machine identification is performed by resource name.
"""
from ..azure_test_case import AzureServiceCase
from ..constants import RULE_PASS_STATUS, RULE_FAIL_STATUS

CIS_7_1 = "CIS 7.1"
CIS_7_2 = "CIS 7.2"
CIS_7_3 = "CIS 7.3"
CIS_7_4 = "CIS 7.4"

cis_azure_7_1_pass = AzureServiceCase(
    rule_tag=CIS_7_1,
    case_identifier="test-virtual-machine-pass",
    expected=RULE_PASS_STATUS,
)

cis_azure_7_1_fail = AzureServiceCase(
    rule_tag=CIS_7_1,
    case_identifier="test-virtual-machine-pass",
    expected=RULE_FAIL_STATUS,
)

cis_azure_7_1 = {
    "7.1 Ensure an Azure Bastion Host Exists expect: passed": cis_azure_7_1_pass,
    "7.1 Ensure an Azure Bastion Host Exists expect: failed": cis_azure_7_1_fail,
}

cis_azure_7_2_pass = AzureServiceCase(
    rule_tag=CIS_7_2,
    case_identifier="test-virtual-machine-pass",
    expected=RULE_PASS_STATUS,
)

cis_azure_7_2_fail = AzureServiceCase(
    rule_tag=CIS_7_2,
    case_identifier="test-virtual-machine-fail",
    expected=RULE_FAIL_STATUS,
)

cis_azure_7_2 = {
    "7.2 Ensure Virtual Machines are utilizing Managed Disks expect: passed": cis_azure_7_2_pass,
    "7.2 Ensure Virtual Machines are utilizing Managed Disks expect: failed": cis_azure_7_2_fail,
}

cis_azure_7_3_pass = AzureServiceCase(
    rule_tag=CIS_7_3,
    case_identifier="test-virtual-machine-pass",
    expected=RULE_PASS_STATUS,
)

cis_azure_7_3_fail = AzureServiceCase(
    rule_tag=CIS_7_3,
    case_identifier="test-virtual-machine-fail",
    expected=RULE_FAIL_STATUS,
)

cis_azure_7_3 = {
    """7.3 Ensure that 'OS and Data' disks are encrypted with
      Customer Managed Key (CMK) expect: passed""": cis_azure_7_3_pass,
    """7.3 Ensure that 'OS and Data' disks are encrypted with
      Customer Managed Key (CMK) expect: failed""": cis_azure_7_3_fail,
}

cis_azure_7_4_pass = AzureServiceCase(
    rule_tag=CIS_7_4,
    case_identifier="test-virtual-machine-pass",
    expected=RULE_PASS_STATUS,
)

cis_azure_7_4_fail = AzureServiceCase(
    rule_tag=CIS_7_4,
    case_identifier="test-virtual-machine-fail",
    expected=RULE_FAIL_STATUS,
)

cis_azure_7_4 = {
    """7.4 Ensure that 'Unattached disks' are encrypted
      with 'Customer Managed Key' (CMK) expect: passed""": cis_azure_7_4_pass,
    """7.4 Ensure that 'Unattached disks' are encrypted
      with 'Customer Managed Key' (CMK) expect: failed""": cis_azure_7_4_fail,
}

cis_azure_virtual_machine_cases = {
    **cis_azure_7_1,
    **cis_azure_7_2,
    **cis_azure_7_3,
    **cis_azure_7_4,
}
