"""
This module provides AWS Virtual Private Cloud (VPC) rule test cases.
Cases are organized as rules.
Each rule has one or more test cases.
VPC rules identification is performed by resource name.
"""

from ..constants import RULE_FAIL_STATUS, RULE_PASS_STATUS
from ..eks_test_case import EksAwsServiceCase

CIS_5_1 = "CIS 5.1"
CIS_5_2 = "CIS 5.2"
CIS_5_3 = "CIS 5.3"
CIS_5_4 = "CIS 5.4"

cis_aws_vpc_5_1_pass = EksAwsServiceCase(
    rule_tag=CIS_5_1,
    case_identifier="arn:aws:ec2:eu-north-1:391946104644:network-acl/acl-05b25ae6744e862cc",
    expected=RULE_PASS_STATUS,
)

cis_aws_vpc_5_1_fail = EksAwsServiceCase(
    rule_tag=CIS_5_1,
    case_identifier="arn:aws:ec2:eu-west-1:391946104644:network-acl/acl-0deecaa33cc9b0ea2",
    expected=RULE_FAIL_STATUS,
)

cis_aws_vpc_5_1 = {
    "5.1 Ensure no Network ACLs allow ingress, inbound rules=denied expect: passed": cis_aws_vpc_5_1_pass,
    "5.1 Ensure no Network ACLs allow ingress, inbound rule all ports=allowed, expect: failed": cis_aws_vpc_5_1_fail,
}

cis_aws_vpc_5_2_pass = EksAwsServiceCase(
    rule_tag=CIS_5_2,
    case_identifier="arn:aws:ec2:eu-north-1:391946104644:security-group/sg-0fa75b3c7e90cf03a",
    expected=RULE_PASS_STATUS,
)

cis_aws_vpc_5_2_fail = EksAwsServiceCase(
    rule_tag=CIS_5_2,
    case_identifier="arn:aws:ec2:eu-north-1:391946104644:security-group/sg-07dea7d21e0777476",
    expected=RULE_FAIL_STATUS,
)

cis_aws_vpc_5_2 = {
    "5.2 Ensure no Security groups allow ingress, inbound rules=denied expect: passed": cis_aws_vpc_5_2_pass,
    "5.2 Ensure no Security groups allow ingress, inbound rule 0.0.0.0/0, expect: failed": cis_aws_vpc_5_2_fail,
}

cis_aws_vpc_5_3_pass = EksAwsServiceCase(
    rule_tag=CIS_5_3,
    case_identifier="arn:aws:ec2:eu-north-1:391946104644:security-group/sg-0fa75b3c7e90cf03a",
    expected=RULE_PASS_STATUS,
)

cis_aws_vpc_5_3_fail = EksAwsServiceCase(
    rule_tag=CIS_5_3,
    case_identifier="arn:aws:ec2:eu-north-1:391946104644:security-group/sg-07dea7d21e0777476",
    expected=RULE_FAIL_STATUS,
)

cis_aws_vpc_5_3 = {
    "5.3 Ensure no Security groups allow ingress, inbound rules=denied expect: passed": cis_aws_vpc_5_3_pass,
    "5.3 Ensure no Security groups allow ingress, inbound rule ::/0 all ports, expect: failed": cis_aws_vpc_5_3_fail,
}

cis_aws_vpc_5_4_pass = EksAwsServiceCase(
    rule_tag=CIS_5_4,
    case_identifier="arn:aws:ec2:eu-central-1:391946104644:security-group/sg-03f2e4eb9617f108a",
    expected=RULE_PASS_STATUS,
)

cis_aws_vpc_5_4_fail = EksAwsServiceCase(
    rule_tag=CIS_5_4,
    case_identifier="arn:aws:ec2:eu-north-1:391946104644:security-group/sg-0fa75b3c7e90cf03a",
    expected=RULE_FAIL_STATUS,
)

cis_aws_vpc_5_4 = {
    "5.4 Ensure default Security Group, no inbound and outbound expect: passed": cis_aws_vpc_5_4_pass,
    "5.4 Ensure default Security Group, inbound and outbound groups exist expect: failed": cis_aws_vpc_5_4_fail,
}

# The name of this variable needs to be `tests_cases` in order to CIS Rules coverage stats to be generated
test_cases = {
    **cis_aws_vpc_5_1,
    **cis_aws_vpc_5_2,
    **cis_aws_vpc_5_3,
    **cis_aws_vpc_5_4,
}
