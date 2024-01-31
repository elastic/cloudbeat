"""
This module provides Azure identity access management rule test cases.
Cases are organized as rules.
Each rule has one or more test cases.
Identity access management identification is performed by resource name.
"""

# from ..azure_test_case import AzureServiceCase
# from ..constants import RULE_PASS_STATUS, RULE_FAIL_STATUS

# TODO: Conflicts due to single subscription
# Also no permissions to create custom roles
# CIS_1_23 = "CIS 1.23"

# cis_azure_1_23_pass = AzureServiceCase(
#     rule_tag=CIS_1_23,
#     case_identifier="TODO",
#     expected=RULE_PASS_STATUS,
# )

# cis_azure_1_23_fail = AzureServiceCase(
#     rule_tag=CIS_1_23,
#     case_identifier="test-identity-access-management-fail",
#     expected=RULE_FAIL_STATUS,
# )

# cis_azure_1_23 = {
#     "1.23 Ensure That No Custom Subscription Administrator Roles Exist expect: passed": cis_azure_1_23_pass,
#     "1.23 Ensure That No Custom Subscription Administrator Roles Exist expect: failed": cis_azure_1_23_fail,
# }

# The name of this variable needs to be `tests_cases` in order to CIS Rules coverage stats to be generated
test_cases = {
    # **cis_azure_1_23,
}
