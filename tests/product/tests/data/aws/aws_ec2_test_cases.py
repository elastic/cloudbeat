"""
This module provides AWS Elastic Compute Cloud EC2 service rule test cases.
Cases are organized as rules.
Each rule has one or more test cases.
EC2 rules identification is performed by resource name.
"""
from ..eks_test_case import EksAwsServiceCase
from ..constants import RULE_PASS_STATUS, RULE_FAIL_STATUS

CIS_2_2_1 = "CIS 2.2.1"

cis_aws_ec2_2_2_1_pass = EksAwsServiceCase(
    rule_tag=CIS_2_2_1,
    case_identifier="ebs-encryption-by-default-704479110758-us-east-1",
    expected=RULE_PASS_STATUS,
)

cis_aws_ec2_2_2_1_fail = EksAwsServiceCase(
    rule_tag=CIS_2_2_1,
    case_identifier="ebs-encryption-by-default-704479110758-eu-west-2",
    expected=RULE_FAIL_STATUS,
)

cis_aws_ec2_2_2_1 = {
    "2.2.1 Ensure EBS volume is enabled, EbsEncryptionByDefault=true expect: passed": cis_aws_ec2_2_2_1_pass,
    "2.2.1 Ensure EBS volume is enabled, EbsEncryptionByDefault=false expect: failed": cis_aws_ec2_2_2_1_fail,
}

cis_aws_ec2_cases = {
    **cis_aws_ec2_2_2_1,
}
