"""
This module provides AWS logging service rule test cases.
Cases are organized as rules.
Each rule has one or more test cases.
Logging identification is performed by resource name.
"""
from commonlib.framework.reporting import skip_param_case, SkipReportData
from ..eks_test_case import EksAwsServiceCase
from ..constants import RULE_PASS_STATUS, RULE_FAIL_STATUS

CIS_3_1 = "CIS 3.1"

cis_aws_log_3_1_pass = EksAwsServiceCase(
    rule_tag=CIS_3_1,
    case_identifier="cloudtrail-704479110758",
    expected=RULE_PASS_STATUS,
)

"""
cis_aws_log_3_1_fail_1:
No cloudtrail enabled for the account -> expect failure
New account
"""
cis_aws_log_3_1_fail_1 = EksAwsServiceCase(
    rule_tag=CIS_3_1,
    case_identifier="cloudtrail-account-1",
    expected=RULE_FAIL_STATUS,
)

"""
cis_aws_log_3_1_fail_2
Cloudtrail is not enabled in all regions -> expect failure
New account -> single cloudtrail
"""
cis_aws_log_3_1_fail_2 = EksAwsServiceCase(
    rule_tag=CIS_3_1,
    case_identifier="cloudtrail-account-2",
    expected=RULE_FAIL_STATUS,
)

"""
Cloudtrail is enabled in all regions but Logging is set to OFF -> expect failure
New account -> single cloudtrail with all regions enabled -> logging is OFF
"""
cis_aws_log_3_1_fail_3 = EksAwsServiceCase(
    rule_tag=CIS_3_1,
    case_identifier="cloudtrail-account-3",
    expected=RULE_FAIL_STATUS,
)

"""
Cloudtrail is enabled -> all regions -> logging ON ->Read/Write Events None -> expect failure
New account -> single cloudtrail all regions
"""
cis_aws_log_3_1_fail_4 = EksAwsServiceCase(
    rule_tag=CIS_3_1,
    case_identifier="cloudtrail-account-4",
    expected=RULE_FAIL_STATUS,
)

cis_aws_log_3_1 = {
    "3.1 Ensure CloudTrail is enabled in all regions expect: passed": cis_aws_log_3_1_pass,
}

cis_aws_log_3_1_skip = {
    "3.1 Ensure CloudTrail is enabled in all regions: no cloudtrail enabled, expect: failed": cis_aws_log_3_1_fail_1,
    "3.1 Ensure CloudTrail is enabled in all regions: not all regions enabled, expect: failed": cis_aws_log_3_1_fail_2,
    "3.1 Ensure CloudTrail is enabled in all regions: logging is off, expect: failed": cis_aws_log_3_1_fail_3,
    "3.1 Ensure CloudTrail is enabled in all regions: Read/Write event is None, expect: failed": cis_aws_log_3_1_fail_4,
}

cis_aws_log_cases = {
    **cis_aws_log_3_1,
    **skip_param_case(
        cis_aws_log_3_1_skip,
        data_to_report=SkipReportData(
            skip_reason="Test case data generation issue",
            url_title="security-team: #6204",
            url_link="https://github.com/elastic/security-team/issues/6204",
        ),
    ),
}
