"""
This module provides AWS S3 service rule test cases.
Cases are organized as rules.
Each rule has one or more test cases.
S3 buckets identification is performed by resource name.
"""

from ..eks_test_case import EksAwsServiceCase
from ..constants import RULE_PASS_STATUS, RULE_FAIL_STATUS


cis_aws_s3_2_1_1_pass = EksAwsServiceCase(
    rule_tag="CIS 2.1.1",
    case_identifier="test-aws-sse-s3-pass",
    expected=RULE_PASS_STATUS,
)

cis_aws_s3_2_1_1_pass_2 = EksAwsServiceCase(
    rule_tag="CIS 2.1.1",
    case_identifier="test-aws-kms-key-pass",
    expected=RULE_PASS_STATUS,
)

cis_aws_s3_2_1_1_fail = EksAwsServiceCase(
    rule_tag="CIS 2.1.1",
    case_identifier="test-aws-no-encryption-fail",
    expected=RULE_FAIL_STATUS,
)

cis_aws_s3_2_1_1 = {
    "2.1.1 Ensure S3 bucket encryption: SSEAlgorithm=AES256 expect: passed": cis_aws_s3_2_1_1_pass,
    "2.1.1 Ensure S3 bucket encryption: SSEAlgorithm=aws:kms expect: passed": cis_aws_s3_2_1_1_pass_2,
    "2.1.1 Ensure S3 bucket encryption: encryption disabled - expect: failed": cis_aws_s3_2_1_1_fail,
}

cis_aws_s3_cases = {
    **cis_aws_s3_2_1_1,
}
