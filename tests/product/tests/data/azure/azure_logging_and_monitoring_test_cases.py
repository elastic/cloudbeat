"""
This module provides Azure logging and monitoring rule test cases.
Cases are organized as rules.
Each rule has one or more test cases.
Logging and monitoring identification is performed by resource name.
"""

from ..azure_test_case import AzureServiceCase
from ..constants import RULE_PASS_STATUS

CIS_5_5 = "CIS 5.5"
# TODO: Alert Rules are per subscription
# TODO: No Alert Rules for evaluation of fail not possible due to having Alert Rules for pass
CIS_5_2_1 = "CIS 5.2.1"
CIS_5_2_10 = "CIS 5.2.10"
CIS_5_2_2 = "CIS 5.2.2"
CIS_5_2_3 = "CIS 5.2.3"
CIS_5_2_4 = "CIS 5.2.4"
CIS_5_2_5 = "CIS 5.2.5"
CIS_5_2_6 = "CIS 5.2.6"
CIS_5_2_7 = "CIS 5.2.7"
CIS_5_2_8 = "CIS 5.2.8"
CIS_5_2_9 = "CIS 5.2.9"
CIS_5_1_2 = "CIS 5.1.2"
CIS_5_3_1 = "CIS 5.3.1"

cis_azure_5_5_pass = AzureServiceCase(
    rule_tag=CIS_5_5,
    case_identifier="testsapass",
    expected=RULE_PASS_STATUS,
)

# TODO: Not sure how to deploy a basic sku on our subscription
# cis_azure_5_5_fail = AzureServiceCase(
#     rule_tag=CIS_5_5,
#     case_identifier="TODO",
#     expected=RULE_FAIL_STATUS,
# )

cis_azure_5_5 = {
    """5.5 Ensure that SKU Basic/Consumption is not used on artifacts that need
      to be monitored (Particularly for Production Workloads) expect: passed""": cis_azure_5_5_pass,
    # """5.5 Ensure that SKU Basic/Consumption is not used on artifacts that need
    #   to be monitored (Particularly for Production Workloads) expect: failed""": cis_azure_5_5_fail,
}

cis_azure_5_2_1_pass = AzureServiceCase(
    rule_tag=CIS_5_2_1,
    case_identifier="azure-activity-log-alert-ef111ee2-6c89-4b09-92c6-5c2321f888df",
    expected=RULE_PASS_STATUS,
)

# cis_azure_5_2_1_fail = AzureServiceCase(
# rule_tag=CIS_5_2_1,
# case_identifier="TODO",
# expected=RULE_FAIL_STATUS,
# )

cis_azure_5_2_1 = {
    """5.2.1 Ensure that Activity Log Alert exists
      for Create Policy Assignment expect: passed""": cis_azure_5_2_1_pass,
    # """5.2.1 Ensure that Activity Log Alert exists
    #   for Create Policy Assignment expect: failed""": cis_azure_5_2_1_fail,
}

cis_azure_5_2_10_pass = AzureServiceCase(
    rule_tag=CIS_5_2_10,
    case_identifier="azure-activity-log-alert-ef111ee2-6c89-4b09-92c6-5c2321f888df",
    expected=RULE_PASS_STATUS,
)

# cis_azure_5_2_10_fail = AzureServiceCase(
# rule_tag=CIS_5_2_10,
# case_identifier="TODO",
# expected=RULE_FAIL_STATUS,
# )

cis_azure_5_2_10 = {
    """5.2.10 Ensure that Activity Log Alert exists
      for Delete Public IP Address rule expect: passed""": cis_azure_5_2_10_pass,
    # """5.2.10 Ensure that Activity Log Alert exists
    #   for Delete Public IP Address rule expect: failed""": cis_azure_5_2_10_fail,
}

cis_azure_5_2_2_pass = AzureServiceCase(
    rule_tag=CIS_5_2_2,
    case_identifier="azure-activity-log-alert-ef111ee2-6c89-4b09-92c6-5c2321f888df",
    expected=RULE_PASS_STATUS,
)

# cis_azure_5_2_2_fail = AzureServiceCase(
# rule_tag=CIS_5_2_2,
# case_identifier="TODO",
# expected=RULE_FAIL_STATUS,
# )

cis_azure_5_2_2 = {
    """5.2.2 Ensure that Activity Log Alert exists
      for Delete Policy Assignment expect: passed""": cis_azure_5_2_2_pass,
    # """5.2.2 Ensure that Activity Log Alert exists
    #   for Delete Policy Assignment expect: failed""": cis_azure_5_2_2_fail,
}

cis_azure_5_2_3_pass = AzureServiceCase(
    rule_tag=CIS_5_2_3,
    case_identifier="azure-activity-log-alert-ef111ee2-6c89-4b09-92c6-5c2321f888df",
    expected=RULE_PASS_STATUS,
)

# cis_azure_5_2_3_fail = AzureServiceCase(
# rule_tag=CIS_5_2_3,
# case_identifier="TODO",
# expected=RULE_FAIL_STATUS,
# )

cis_azure_5_2_3 = {
    """5.2.3 Ensure that Activity Log Alert exists for Create
      or Update Network Security Group expect: passed""": cis_azure_5_2_3_pass,
    # """5.2.3 Ensure that Activity Log Alert exists for Create
    #   or Update Network Security Group expect: failed""": cis_azure_5_2_3_fail,
}

cis_azure_5_2_4_pass = AzureServiceCase(
    rule_tag=CIS_5_2_4,
    case_identifier="azure-activity-log-alert-ef111ee2-6c89-4b09-92c6-5c2321f888df",
    expected=RULE_PASS_STATUS,
)

# cis_azure_5_2_4_fail = AzureServiceCase(
# rule_tag=CIS_5_2_4,
# case_identifier="TODO",
# expected=RULE_FAIL_STATUS,
# )

cis_azure_5_2_4 = {
    """5.2.4 Ensure that Activity Log Alert exists
      for Delete Network Security Group expect: passed""": cis_azure_5_2_4_pass,
    # """5.2.4 Ensure that Activity Log Alert exists
    #   for Delete Network Security Group expect: failed""": cis_azure_5_2_4_fail,
}

cis_azure_5_2_5_pass = AzureServiceCase(
    rule_tag=CIS_5_2_5,
    case_identifier="azure-activity-log-alert-ef111ee2-6c89-4b09-92c6-5c2321f888df",
    expected=RULE_PASS_STATUS,
)

# cis_azure_5_2_5_fail = AzureServiceCase(
# rule_tag=CIS_5_2_5,
# case_identifier="TODO",
# expected=RULE_FAIL_STATUS,
# )

cis_azure_5_2_5 = {
    """5.2.5 Ensure that Activity Log Alert exists for Create
      or Update Security Solution expect: passed""": cis_azure_5_2_5_pass,
    # """5.2.5 Ensure that Activity Log Alert exists for Create
    #   or Update Security Solution expect: failed""": cis_azure_5_2_5_fail,
}

cis_azure_5_2_6_pass = AzureServiceCase(
    rule_tag=CIS_5_2_6,
    case_identifier="azure-activity-log-alert-ef111ee2-6c89-4b09-92c6-5c2321f888df",
    expected=RULE_PASS_STATUS,
)

# cis_azure_5_2_6_fail = AzureServiceCase(
# rule_tag=CIS_5_2_6,
# case_identifier="TODO",
# expected=RULE_FAIL_STATUS,
# )

cis_azure_5_2_6 = {
    """5.2.6 Ensure that Activity Log Alert exists
      for Delete Security Solution expect: passed""": cis_azure_5_2_6_pass,
    # """5.2.6 Ensure that Activity Log Alert exists
    #   for Delete Security Solution expect: failed""": cis_azure_5_2_6_fail,
}

cis_azure_5_2_7_pass = AzureServiceCase(
    rule_tag=CIS_5_2_7,
    case_identifier="azure-activity-log-alert-ef111ee2-6c89-4b09-92c6-5c2321f888df",
    expected=RULE_PASS_STATUS,
)

# cis_azure_5_2_7_fail = AzureServiceCase(
# rule_tag=CIS_5_2_7,
# case_identifier="TODO",
# expected=RULE_FAIL_STATUS,
# )

cis_azure_5_2_7 = {
    """5.2.7 Ensure that Activity Log Alert exists for Create
      or Update SQL Server Firewall Rule expect: passed""": cis_azure_5_2_7_pass,
    # """5.2.7 Ensure that Activity Log Alert exists for Create
    #   or Update SQL Server Firewall Rule expect: failed""": cis_azure_5_2_7_fail,
}

cis_azure_5_2_8_pass = AzureServiceCase(
    rule_tag=CIS_5_2_8,
    case_identifier="azure-activity-log-alert-ef111ee2-6c89-4b09-92c6-5c2321f888df",
    expected=RULE_PASS_STATUS,
)

# cis_azure_5_2_8_fail = AzureServiceCase(
# rule_tag=CIS_5_2_8,
# case_identifier="TODO",
# expected=RULE_FAIL_STATUS,
# )

cis_azure_5_2_8 = {
    """5.2.8 Ensure that Activity Log Alert exists
      for Delete SQL Server Firewall Rule expect: passed""": cis_azure_5_2_8_pass,
    # """5.2.8 Ensure that Activity Log Alert exists
    #   for Delete SQL Server Firewall Rule expect: failed""": cis_azure_5_2_8_fail,
}

cis_azure_5_2_9_pass = AzureServiceCase(
    rule_tag=CIS_5_2_9,
    case_identifier="azure-activity-log-alert-ef111ee2-6c89-4b09-92c6-5c2321f888df",
    expected=RULE_PASS_STATUS,
)

# cis_azure_5_2_9_fail = AzureServiceCase(
# rule_tag=CIS_5_2_9,
# case_identifier="TODO",
# expected=RULE_FAIL_STATUS,
# )

cis_azure_5_2_9 = {
    """5.2.9 Ensure that Activity Log Alert exists for Create
      or Update Public IP Address rule expect: passed""": cis_azure_5_2_9_pass,
    # """5.2.9 Ensure that Activity Log Alert exists for Create
    #   or Update Public IP Address rule expect: failed""": cis_azure_5_2_9_fail,
}

cis_azure_5_1_2_pass = AzureServiceCase(
    rule_tag=CIS_5_1_2,
    case_identifier="azure-diagnostic-settings-ef111ee2-6c89-4b09-92c6-5c2321f888df",
    expected=RULE_PASS_STATUS,
)

# TODO: Diagnostic Settings are per subscription
# TODO: No Diagnostic Settings for evaluation of fail not possible due to having Diagnostic Settings for pass
# cis_azure_5_1_2_fail = AzureServiceCase(
#     rule_tag=CIS_5_1_2,
#     case_identifier="TODO",
#     expected=RULE_FAIL_STATUS,
# )

cis_azure_5_1_2 = {
    "5.1.2 Ensure Diagnostic Setting captures appropriate categories expect: passed": cis_azure_5_1_2_pass,
    # "5.1.2 Ensure Diagnostic Setting captures appropriate categories expect: failed": cis_azure_5_1_2_fail,
}

cis_azure_5_3_1_pass = AzureServiceCase(
    rule_tag=CIS_5_3_1,
    case_identifier="azure-insights-component-ef111ee2-6c89-4b09-92c6-5c2321f888df",
    expected=RULE_PASS_STATUS,
)

# TODO: Application Insights are per subscription
# TODO: No Application Insights for evaluation of fail not possible due to having Application Insights for pass
# cis_azure_5_3_1_fail = AzureServiceCase(
#     rule_tag=CIS_5_3_1,
#     case_identifier="TODO",
#     expected=RULE_FAIL_STATUS,
# )

cis_azure_5_3_1 = {
    "5.3.1 Ensure Application Insights are Configured expect: passed": cis_azure_5_3_1_pass,
    # "5.3.1 Ensure Application Insights are Configured expect: failed": cis_azure_5_3_1_fail,
}

# The name of this variable needs to be `tests_cases` in order to CIS Rules coverage stats to be generated
test_cases = {
    **cis_azure_5_5,
    **cis_azure_5_2_1,
    **cis_azure_5_2_10,
    **cis_azure_5_2_2,
    **cis_azure_5_2_3,
    **cis_azure_5_2_4,
    **cis_azure_5_2_5,
    **cis_azure_5_2_6,
    **cis_azure_5_2_7,
    **cis_azure_5_2_8,
    **cis_azure_5_2_9,
    **cis_azure_5_1_2,
    **cis_azure_5_3_1,
}
