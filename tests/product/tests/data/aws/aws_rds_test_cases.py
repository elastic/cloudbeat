"""
This module provides AWS Relational Database Service (RDS) rule test cases.
Cases are organized as rules.
Each rule has one or more test cases.
RDS rules identification is performed by resource name.
"""

from ..constants import RULE_FAIL_STATUS, RULE_PASS_STATUS
from ..eks_test_case import EksAwsServiceCase

CIS_2_3_1 = "CIS 2.3.1"
CIS_2_3_2 = "CIS 2.3.2"
CIS_2_3_3 = "CIS 2.3.3"

cis_aws_rds_2_3_1_pass = EksAwsServiceCase(
    rule_tag=CIS_2_3_1,
    case_identifier="test-rds-instance-1",
    expected=RULE_PASS_STATUS,
)

cis_aws_rds_2_3_1_fail = EksAwsServiceCase(
    rule_tag=CIS_2_3_1,
    case_identifier="test-rds-fail-instance-1",
    expected=RULE_FAIL_STATUS,
)

cis_aws_rds_2_3_1 = {
    "2.3.1 Ensure RDS Instances encryption, Encryption Enabled=true expect: passed": cis_aws_rds_2_3_1_pass,
    "2.3.1 Ensure RDS Instances encryption, Encryption Enabled=false expect: failed": cis_aws_rds_2_3_1_fail,
}

cis_aws_rds_2_3_2_pass = EksAwsServiceCase(
    rule_tag=CIS_2_3_2,
    case_identifier="test-rds-instance-1",
    expected=RULE_PASS_STATUS,
)

cis_aws_rds_2_3_2_fail = EksAwsServiceCase(
    rule_tag=CIS_2_3_2,
    case_identifier="test-rds-fail-instance-1",
    expected=RULE_FAIL_STATUS,
)

cis_aws_rds_2_3_2 = {
    "2.3.2 Ensure Auto Minor Version Enabled, AutoMinorVersionUpgrade=true expect: passed": cis_aws_rds_2_3_2_pass,
    "2.3.2 Ensure Auto Minor Version Enabled, AutoMinorVersionUpgrade=false expect: failed": cis_aws_rds_2_3_2_fail,
}

cis_aws_rds_2_3_3_pass = EksAwsServiceCase(
    rule_tag=CIS_2_3_3,
    case_identifier="test-rds-instance-1",
    expected=RULE_PASS_STATUS,
)

cis_aws_rds_2_3_3_fail = EksAwsServiceCase(
    rule_tag=CIS_2_3_3,
    case_identifier="test-rds-fail-instance-1",
    expected=RULE_FAIL_STATUS,
)

cis_aws_rds_2_3_3 = {
    "2.3.3 Ensure no public access, no public access, expect: passed": cis_aws_rds_2_3_3_pass,
    "2.3.3 Ensure no public access, public access allowed, expect: failed": cis_aws_rds_2_3_3_fail,
}

# The name of this variable needs to be `tests_cases` in order to CIS Rules coverage stats to be generated
test_cases = {
    **cis_aws_rds_2_3_1,
    **cis_aws_rds_2_3_2,
    **cis_aws_rds_2_3_3,
}
