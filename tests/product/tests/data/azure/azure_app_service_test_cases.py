"""
This module provides Azure app service rule test cases.
Cases are organized as rules.
Each rule has one or more test cases.
App service identification is performed by resource name.
"""
from ..azure_test_case import AzureServiceCase
from ..constants import RULE_PASS_STATUS, RULE_FAIL_STATUS

CIS_9_2 = "CIS 9.2"
# TODO: Removed rule
# CIS_9_3 = "CIS 9.3"
CIS_9_4 = "CIS 9.4"
CIS_9_5 = "CIS 9.5"
CIS_9_9 = "CIS 9.9"
# TODO: Removed rule
# CIS_9_10 = "CIS 9.10"

cis_azure_9_4_pass = AzureServiceCase(
    rule_tag=CIS_9_4,
    case_identifier="test-app-service-pass",
    expected=RULE_PASS_STATUS,
)

# TODO: https://github.com/elastic/cloudbeat/issues/1828
# cis_azure_9_4_fail = AzureServiceCase(
#     rule_tag=CIS_9_4,
#     case_identifier="test-app-service-fail",
#     expected=RULE_FAIL_STATUS,
# )

cis_azure_9_4 = {
    """9.4 Ensure the web app has 'Client Certificates (Incoming client certificates)'
      set to 'On' expect: passed""": cis_azure_9_4_pass,
    # """9.4 Ensure the web app has 'Client Certificates (Incoming client certificates)'
    #   set to 'On' expect: failed""": cis_azure_9_4_fail,
}

# cis_azure_9_10_pass = AzureServiceCase(
#     rule_tag=CIS_9_10,
#     case_identifier="test-app-service-pass",
#     expected=RULE_PASS_STATUS,
# )

# cis_azure_9_10_fail = AzureServiceCase(
#     rule_tag=CIS_9_10,
#     case_identifier="test-app-service-fail",
#     expected=RULE_FAIL_STATUS,
# )

# cis_azure_9_10 = {
#     "9.10 Ensure FTP deployments are Disabled expect: passed": cis_azure_9_10_pass,
#     "9.10 Ensure FTP deployments are Disabled expect: failed": cis_azure_9_10_fail,
# }

cis_azure_9_2_pass = AzureServiceCase(
    rule_tag=CIS_9_2,
    case_identifier="test-app-service-pass",
    expected=RULE_PASS_STATUS,
)

cis_azure_9_2_fail = AzureServiceCase(
    rule_tag=CIS_9_2,
    case_identifier="test-app-service-fail",
    expected=RULE_FAIL_STATUS,
)

cis_azure_9_2 = {
    """9.2 Ensure Web App Redirects All HTTP traffic to HTTPS
      in Azure App Service expect: passed""": cis_azure_9_2_pass,
    """9.2 Ensure Web App Redirects All HTTP traffic to HTTPS
      in Azure App Service expect: failed""": cis_azure_9_2_fail,
}

# cis_azure_9_3_pass = AzureServiceCase(
#     rule_tag=CIS_9_3,
#     case_identifier="test-app-service-pass",
#     expected=RULE_PASS_STATUS,
# )

# cis_azure_9_3_fail = AzureServiceCase(
#     rule_tag=CIS_9_3,
#     case_identifier="test-app-service-fail",
#     expected=RULE_FAIL_STATUS,
# )

# cis_azure_9_3 = {
#     """9.3 Ensure Web App is using the latest version
#       of TLS encryption expect: passed""": cis_azure_9_3_pass,
#     """9.3 Ensure Web App is using the latest version
#       of TLS encryption expect: failed""": cis_azure_9_3_fail,
# }

cis_azure_9_5_pass = AzureServiceCase(
    rule_tag=CIS_9_5,
    case_identifier="test-app-service-pass",
    expected=RULE_PASS_STATUS,
)

cis_azure_9_5_fail = AzureServiceCase(
    rule_tag=CIS_9_5,
    case_identifier="test-app-service-fail",
    expected=RULE_FAIL_STATUS,
)

cis_azure_9_5 = {
    """9.5 Ensure that Register with Azure Active Directory
    is enabled on App Service expect: passed""": cis_azure_9_5_pass,
    """9.5 Ensure that Register with Azure Active Directory
    is enabled on App Service expect: failed""": cis_azure_9_5_fail,
}

cis_azure_9_9_pass = AzureServiceCase(
    rule_tag=CIS_9_9,
    case_identifier="test-app-service-pass",
    expected=RULE_PASS_STATUS,
)

cis_azure_9_9_fail = AzureServiceCase(
    rule_tag=CIS_9_9,
    case_identifier="test-app-service-fail",
    expected=RULE_FAIL_STATUS,
)

cis_azure_9_9 = {
    """9.9 Ensure that 'HTTP Version' is the Latest,
      if Used to Run the Web App expect: passed""": cis_azure_9_9_pass,
    """9.9 Ensure that 'HTTP Version' is the Latest,
      if Used to Run the Web App expect: failed""": cis_azure_9_9_fail,
}

cis_azure_app_service_cases = {
    **cis_azure_9_2,
    # **cis_azure_9_3,
    **cis_azure_9_4,
    **cis_azure_9_5,
    **cis_azure_9_9,
    # **cis_azure_9_10,
}
