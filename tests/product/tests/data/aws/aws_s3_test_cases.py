"""
This module provides AWS S3 service rule test cases.
Cases are organized as rules.
Each rule has one or more test cases.
S3 buckets identification is performed by resource name.
"""
from commonlib.framework.reporting import skip_param_case, SkipReportData
from ..eks_test_case import EksAwsServiceCase
from ..constants import RULE_PASS_STATUS, RULE_FAIL_STATUS

CIS_2_1_1 = "CIS 2.1.1"
CIS_2_1_2 = "CIS 2.1.2"
CIS_2_1_3 = "CIS 2.1.3"

cis_aws_s3_2_1_1_pass = EksAwsServiceCase(
    rule_tag=CIS_2_1_1,
    case_identifier="test-aws-sse-s3-pass",
    expected=RULE_PASS_STATUS,
)

cis_aws_s3_2_1_1_pass_2 = EksAwsServiceCase(
    rule_tag=CIS_2_1_1,
    case_identifier="test-aws-kms-key-pass",
    expected=RULE_PASS_STATUS,
)

cis_aws_s3_2_1_1_fail = EksAwsServiceCase(
    rule_tag=CIS_2_1_1,
    case_identifier="test-aws-no-encryption-fail",
    expected=RULE_FAIL_STATUS,
)

cis_aws_s3_2_1_1 = {
    "2.1.1 Ensure S3 bucket encryption: SSEAlgorithm=AES256 expect: passed": cis_aws_s3_2_1_1_pass,
    "2.1.1 Ensure S3 bucket encryption: SSEAlgorithm=aws:kms expect: passed": cis_aws_s3_2_1_1_pass_2,
}

cis_aws_s3_2_1_1_skip = {
    "2.1.1 Ensure S3 bucket encryption: encryption disabled - expect: failed": cis_aws_s3_2_1_1_fail,
}

cis_aws_s3_2_1_2_pass = EksAwsServiceCase(
    rule_tag=CIS_2_1_2,
    case_identifier="test-aws-sse-s3-pass",
    expected=RULE_PASS_STATUS,
)

cis_aws_s3_2_1_2_fail = EksAwsServiceCase(
    rule_tag=CIS_2_1_2,
    case_identifier="test-aws-sec-transport-fail",
    expected=RULE_FAIL_STATUS,
)

cis_aws_s3_2_1_2_fail_2 = EksAwsServiceCase(
    rule_tag=CIS_2_1_2,
    case_identifier="test-aws-sec-transport-no-condition-fail",
    expected=RULE_FAIL_STATUS,
)

cis_aws_s3_2_1_2 = {
    "2.1.2 Ensure S3 bucket policy: aws:SecureTransport: false -> expect: passed": cis_aws_s3_2_1_2_pass,
    "2.1.2 Ensure S3 bucket policy: aws:SecureTransport: true -> expect: failed": cis_aws_s3_2_1_2_fail,
    "2.1.2 Ensure S3 bucket policy: Policy exists, no SecurityTransport -> expect: failed": cis_aws_s3_2_1_2_fail_2,
}

cis_aws_s3_2_1_3_fail = EksAwsServiceCase(
    rule_tag=CIS_2_1_3,
    case_identifier="test-aws-mfa-disabled-fail",
    expected=RULE_FAIL_STATUS,
)

cis_aws_s3_2_1_3_pass = EksAwsServiceCase(
    rule_tag=CIS_2_1_3,
    case_identifier="test-aws-mfa-enabled-pass",
    expected=RULE_FAIL_STATUS,
)

cis_aws_s3_2_1_3 = {
    "2.1.2 Ensure MFA Delete is enabled: default -> disabled -> expect: failed": cis_aws_s3_2_1_3_fail,
    "2.1.2 Ensure MFA Delete is enabled, expect: failed": cis_aws_s3_2_1_3_pass,
}

cis_aws_s3_2_1_3_skip = {
    "2.1.2 Ensure MFA Delete is enabled, expect: failed": cis_aws_s3_2_1_3_pass,
}

cis_aws_s3_cases = {
    **cis_aws_s3_2_1_1,
    **skip_param_case(
        cis_aws_s3_2_1_1_skip,
        data_to_report=SkipReportData(
            skip_reason="Test case data generation issue",
            url_title="security-team: #6204",
            url_link="https://github.com/elastic/security-team/issues/6204",
        ),
    ),
    **cis_aws_s3_2_1_2,
    **cis_aws_s3_2_1_3,
    **cis_aws_s3_2_1_3_skip,
    **skip_param_case(
        cis_aws_s3_2_1_3_skip,
        data_to_report=SkipReportData(
            skip_reason="Test case data generation issue",
            url_title="security-team: #6204",
            url_link="https://github.com/elastic/security-team/issues/6204",
        ),
    ),
}
