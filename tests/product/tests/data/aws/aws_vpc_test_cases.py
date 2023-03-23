"""
This module provides AWS Virtual Private Cloud (VPC) rule test cases.
Cases are organized as rules.
Each rule has one or more test cases.
VPC rules identification is performed by resource name.
"""
from ..eks_test_case import EksAwsServiceCase
from ..constants import RULE_PASS_STATUS, RULE_FAIL_STATUS

CIS_5_1 = "CIS 5.1"
CIS_5_2 = "CIS 5.2"
CIS_5_3 = "CIS 5.3"
CIS_5_4 = "CIS 5.4"

cis_aws_vpc_5_1_pass = EksAwsServiceCase(
    rule_tag=CIS_5_1,
    case_identifier="acl-0919ec1794cb66140",
    expected=RULE_PASS_STATUS,
)

cis_aws_vpc_5_1_fail = EksAwsServiceCase(
    rule_tag=CIS_5_1,
    case_identifier="acl-053fe94e40a49c818",
    expected=RULE_FAIL_STATUS,
)

cis_aws_vpc_5_1 = {
    "5.1 Ensure no Network ACLs allow ingress, inbound rules=denied expect: passed": cis_aws_vpc_5_1_pass,
    "5.1 Ensure no Network ACLs allow ingress, inbound rule all ports=allowed, expect: failed": cis_aws_vpc_5_1_fail,
}

cis_aws_vpc_cases = {
    **cis_aws_vpc_5_1,
}
