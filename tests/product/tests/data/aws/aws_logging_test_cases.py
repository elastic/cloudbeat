"""
This module provides AWS logging service rule test cases.
Cases are organized as rules.
Each rule has one or more test cases.
Logging identification is performed by resource name.
"""

from commonlib.framework.reporting import SkipReportData, skip_param_case

from ..constants import RULE_FAIL_STATUS, RULE_PASS_STATUS
from ..eks_test_case import EksAwsServiceCase

CIS_3_1 = "CIS 3.1"
CIS_3_2 = "CIS 3.2"
CIS_3_3 = "CIS 3.3"
CIS_3_4 = "CIS 3.4"
CIS_3_6 = "CIS 3.6"
CIS_3_7 = "CIS 3.7"
CIS_3_9 = "CIS 3.9"
CIS_3_10 = "CIS 3.10"
CIS_3_11 = "CIS 3.11"

cis_aws_log_3_1_pass = EksAwsServiceCase(
    rule_tag=CIS_3_1,
    case_identifier="cloudtrail-391946104644",
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
    case_identifier="cloudtrail-391946104644",
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

cis_aws_log_3_2_pass = EksAwsServiceCase(
    rule_tag=CIS_3_2,
    case_identifier="elastic-eng-org-cloudtrail",
    expected=RULE_PASS_STATUS,
)

cis_aws_log_3_2_fail = EksAwsServiceCase(
    rule_tag=CIS_3_2,
    case_identifier="test-aws-bench-trail",
    expected=RULE_FAIL_STATUS,
)

cis_aws_log_3_2 = {
    "3.2 Ensure CloudTrail log file validation is enabled, validation=Enabled expect: passed": cis_aws_log_3_2_pass,
    "3.2 Ensure CloudTrail log file validation is enabled, validation=Disabled expect: passed": cis_aws_log_3_2_fail,
}

cis_aws_log_3_3_pass = EksAwsServiceCase(
    rule_tag=CIS_3_3,
    case_identifier="test-aws-bench-trail",
    expected=RULE_PASS_STATUS,
)

cis_aws_log_3_3_fail = EksAwsServiceCase(
    rule_tag=CIS_3_3,
    case_identifier="test-aws-file-validation-off-failed",
    expected=RULE_FAIL_STATUS,
)

cis_aws_log_3_3 = {
    "3.3 Ensure S3 bucket is not publicly accessible: Effect=Deny, expected passed": cis_aws_log_3_3_pass,
}

cis_aws_log_3_3_skip = {
    "3.3 Ensure S3 bucket is not publicly accessible: accessible=true, expected failed ": cis_aws_log_3_3_fail,
}

cis_aws_log_3_4_pass = EksAwsServiceCase(
    rule_tag=CIS_3_4,
    case_identifier="test-aws-bench-trail",
    expected=RULE_PASS_STATUS,
)

cis_aws_log_3_4_fail = EksAwsServiceCase(
    rule_tag=CIS_3_4,
    case_identifier="elastic-eng-org-cloudtrail",
    expected=RULE_FAIL_STATUS,
)

cis_aws_log_3_4 = {
    "3.4 Ensure CloudTrail integration with CloudWatch, no integration expected failed": cis_aws_log_3_4_fail,
    "3.4 Ensure CloudTrail integration with CloudWatch, integration exists expected passed": cis_aws_log_3_4_pass,
}

cis_aws_log_3_6_pass = EksAwsServiceCase(
    rule_tag=CIS_3_6,
    case_identifier="test-aws-bench-trail",
    expected=RULE_PASS_STATUS,
)

cis_aws_log_3_6_fail = EksAwsServiceCase(
    rule_tag=CIS_3_6,
    case_identifier="elastic-eng-org-cloudtrail",
    expected=RULE_FAIL_STATUS,
)

cis_aws_log_3_6 = {
    "3.6 Ensure CloudTrail access logging, enabled=false expected failed": cis_aws_log_3_6_fail,
    "3.6 Ensure CloudTrail access logging, enabled=true expected passed": cis_aws_log_3_6_pass,
}

cis_aws_log_3_7_pass = EksAwsServiceCase(
    rule_tag=CIS_3_7,
    case_identifier="elastic-eng-org-cloudtrail",
    expected=RULE_PASS_STATUS,
)

cis_aws_log_3_7_fail = EksAwsServiceCase(
    rule_tag=CIS_3_7,
    case_identifier="test-aws-bench-trail",
    expected=RULE_FAIL_STATUS,
)

cis_aws_log_3_7 = {
    "3.7 Ensure CloudTrail KMS encrypted, enabled=true expected passed": cis_aws_log_3_7_pass,
    "3.7 Ensure CloudTrail KMS encrypted, enabled=false expected failed": cis_aws_log_3_7_fail,
}


# VPC location is eu-north-1
cis_aws_log_3_9_pass = EksAwsServiceCase(
    rule_tag=CIS_3_9,
    case_identifier="vpc-0370ad7241170e623",
    expected=RULE_PASS_STATUS,
)

# VPC location is eu-west-1
cis_aws_log_3_9_fail_1 = EksAwsServiceCase(
    rule_tag=CIS_3_9,
    case_identifier="vpc-0cabbadbef30124b0",
    expected=RULE_FAIL_STATUS,
)

# VPC location is eu-west-2
cis_aws_log_3_9_fail_2 = EksAwsServiceCase(
    rule_tag=CIS_3_9,
    case_identifier="vpc-03c8060bbdc893af6",
    expected=RULE_FAIL_STATUS,
)

cis_aws_log_3_9 = {
    "3.9 Ensure VPC flow logging, enabled=true expected passed": cis_aws_log_3_9_pass,
    "3.9 Ensure VPC flow logging, enabled=false region=eu-west-1 expected failed": cis_aws_log_3_9_fail_1,
    "3.9 Ensure VPC flow logging, enabled=false, region=eu-west-2 expected failed": cis_aws_log_3_9_fail_2,
}

cis_aws_log_3_10_pass = EksAwsServiceCase(
    rule_tag=CIS_3_10,
    case_identifier="test-aws-bench-trail",
    expected=RULE_PASS_STATUS,
)

cis_aws_log_3_10_fail = EksAwsServiceCase(
    rule_tag=CIS_3_10,
    case_identifier="test-aws-bench-trail",
    expected=RULE_FAIL_STATUS,
)

cis_aws_log_3_10_skip = {
    "3.10 Ensure Object-level logging, enabled=false, expected failed ": cis_aws_log_3_10_fail,
}

cis_aws_log_3_10 = {
    "3.10 Ensure Object-level logging, enabled=true, expected passed": cis_aws_log_3_10_pass,
}

cis_aws_log_3_11_pass = EksAwsServiceCase(
    rule_tag=CIS_3_11,
    case_identifier="test-aws-bench-trail",
    expected=RULE_PASS_STATUS,
)

cis_aws_log_3_11_fail = EksAwsServiceCase(
    rule_tag=CIS_3_11,
    case_identifier="elastic-eng-org-cloudtrail",
    expected=RULE_FAIL_STATUS,
)

cis_aws_log_3_11 = {
    "3.11 Ensure Object-level logging read events, enabled=false, expected failed ": cis_aws_log_3_11_fail,
}

cis_aws_log_3_11_skip = {
    "3.11 Ensure Object-level logging read events, enabled=true, expected passed": cis_aws_log_3_11_pass,
}


# The name of this variable needs to be `tests_cases` in order to CIS Rules coverage stats to be generated
test_cases = {
    **cis_aws_log_3_1,
    **skip_param_case(
        cis_aws_log_3_1_skip,
        data_to_report=SkipReportData(
            skip_reason="Test case data generation issue",
            url_title="security-team: #6204",
            url_link="https://github.com/elastic/security-team/issues/6204",
        ),
    ),
    **cis_aws_log_3_2,
    **cis_aws_log_3_3,
    **skip_param_case(
        cis_aws_log_3_3_skip,
        data_to_report=SkipReportData(
            skip_reason="Test case data generation issue",
            url_title="security-team: #6204",
            url_link="https://github.com/elastic/security-team/issues/6204",
        ),
    ),
    **cis_aws_log_3_4,
    **cis_aws_log_3_6,
    **cis_aws_log_3_7,
    **cis_aws_log_3_9,
    **cis_aws_log_3_10,
    **skip_param_case(
        cis_aws_log_3_10_skip,
        data_to_report=SkipReportData(
            skip_reason="When object-level logging for write/read events is enabled for S3 bucket evaluation is failed",
            url_title="cloudbeat: #811",
            url_link="https://github.com/elastic/cloudbeat/issues/811",
        ),
    ),
    **cis_aws_log_3_11,
}
