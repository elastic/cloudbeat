"""
This module provides AWS IAM rules test cases.
Cases are organized as rules.
Each rule has one or more test cases.
IAM identification is performed by resource name.
"""

from commonlib.framework.reporting import SkipReportData, skip_param_case

from ..constants import RULE_FAIL_STATUS, RULE_PASS_STATUS
from ..eks_test_case import EksAwsServiceCase

CIS_1_4 = "CIS 1.4"
CIS_1_5 = "CIS 1.5"
CIS_1_6 = "CIS 1.6"
CIS_1_7 = "CIS 1.7"
CIS_1_8 = "CIS 1.8"
CIS_1_9 = "CIS 1.9"
CIS_1_10 = "CIS 1.10"
CIS_1_11 = "CIS 1.11"
CIS_1_13 = "CIS 1.13"
CIS_1_15 = "CIS 1.15"
CIS_1_16 = "CIS 1.16"
CIS_1_17 = "CIS 1.17"
CIS_1_20 = "CIS 1.20"
ROOT_ACCOUNT = "<root_account>"
ROOT_ACCOUNT_2 = "<root_account_2>"

cis_aws_iam_1_4_pass = EksAwsServiceCase(
    rule_tag=CIS_1_4,
    case_identifier=ROOT_ACCOUNT,
    expected=RULE_PASS_STATUS,
)

cis_aws_iam_1_4_fail = EksAwsServiceCase(
    rule_tag=CIS_1_4,
    case_identifier=ROOT_ACCOUNT_2,
    expected=RULE_FAIL_STATUS,
)

cis_aws_iam_1_4 = {
    "1.4 Ensure no access keys: root_account=no access keys expect: passed": cis_aws_iam_1_4_pass,
}

cis_aws_iam_1_4_skip = {
    "1.4 Ensure no access keys: root_account has access keys expect: failed": cis_aws_iam_1_4_fail,
}

cis_aws_iam_1_5_pass = EksAwsServiceCase(
    rule_tag=CIS_1_5,
    case_identifier=ROOT_ACCOUNT,
    expected=RULE_PASS_STATUS,
)

cis_aws_iam_1_5_fail = EksAwsServiceCase(
    rule_tag=CIS_1_5,
    case_identifier=ROOT_ACCOUNT_2,
    expected=RULE_FAIL_STATUS,
)

cis_aws_iam_1_5 = {
    "1.5 Ensure MFA is enabled: root_account=MFA enabled, expect: passed": cis_aws_iam_1_5_pass,
}

cis_aws_iam_1_5_skip = {
    "1.5 Ensure MFA is enabled: root_account MFA disabled, expect: failed": cis_aws_iam_1_5_fail,
}

cis_aws_iam_1_6_fail = EksAwsServiceCase(
    rule_tag=CIS_1_6,
    case_identifier=ROOT_ACCOUNT,
    expected=RULE_FAIL_STATUS,
)

cis_aws_iam_1_6_pass = EksAwsServiceCase(
    rule_tag=CIS_1_6,
    case_identifier=ROOT_ACCOUNT_2,
    expected=RULE_PASS_STATUS,
)

cis_aws_iam_1_6 = {
    "1.6 Ensure hardware MFA is enabled: root_account hardware MFA disabled, expect: failed": cis_aws_iam_1_6_fail,
}

cis_aws_iam_1_6_skip = {
    "1.6 Ensure hardware MFA is enabled: root_account_2 hardware MFA enabled, expect: failed": cis_aws_iam_1_6_pass,
}

cis_aws_iam_1_7_fail = EksAwsServiceCase(
    rule_tag=CIS_1_7,
    case_identifier=ROOT_ACCOUNT_2,
    expected=RULE_FAIL_STATUS,
)

cis_aws_iam_1_7_pass = EksAwsServiceCase(
    rule_tag=CIS_1_7,
    case_identifier=ROOT_ACCOUNT,
    expected=RULE_PASS_STATUS,
)

cis_aws_iam_1_7 = {
    "1.7 Root user eliminate daily tasks: root_account no access keys, expect: passed": cis_aws_iam_1_7_pass,
}

cis_aws_iam_1_7_skip = {
    "1.7 Root user eliminate daily tasks: root_account_2 daily usage, expect: failed": cis_aws_iam_1_7_fail,
}

cis_aws_iam_1_8_fail = EksAwsServiceCase(
    rule_tag=CIS_1_8,
    case_identifier="account-password-policy",
    expected=RULE_FAIL_STATUS,
)

cis_aws_iam_1_8_pass = EksAwsServiceCase(
    rule_tag=CIS_1_8,
    case_identifier="new-account-password-policy",
    expected=RULE_PASS_STATUS,
)

cis_aws_iam_1_8 = {
    "1.8 Account password policy: password length=8, expect: fail": cis_aws_iam_1_8_fail,
}

cis_aws_iam_1_8_skip = {
    "1.8 Account password policy: password length=14, expect: pass": cis_aws_iam_1_8_pass,
}

cis_aws_iam_1_9_fail = EksAwsServiceCase(
    rule_tag=CIS_1_9,
    case_identifier="account-password-policy",
    expected=RULE_FAIL_STATUS,
)

cis_aws_iam_1_9_pass = EksAwsServiceCase(
    rule_tag=CIS_1_9,
    case_identifier="new-account-password-policy",
    expected=RULE_PASS_STATUS,
)

cis_aws_iam_1_9 = {
    "1.9 Account password policy reuse: reuse_count=5, expect: fail": cis_aws_iam_1_9_fail,
}

cis_aws_iam_1_9_skip = {
    "1.9 Account password policy reuse: reuse_count=24, expect: pass": cis_aws_iam_1_9_pass,
}

cis_aws_iam_1_10_pass_1 = EksAwsServiceCase(
    rule_tag=CIS_1_10,
    case_identifier="test-mfa-virtual-pass",
    expected=RULE_PASS_STATUS,
)

cis_aws_iam_1_10_pass_2 = EksAwsServiceCase(
    rule_tag=CIS_1_10,
    case_identifier="test-mfa-virtual-never-used",
    expected=RULE_PASS_STATUS,
)

cis_aws_iam_1_10_fail = EksAwsServiceCase(
    rule_tag=CIS_1_10,
    case_identifier="test-no-mfa",  # test-no-mfa-fail
    expected=RULE_FAIL_STATUS,
)

cis_aws_iam_1_10 = {
    "1.10 MFA enabled: MFA virtual=true, expect: passed": cis_aws_iam_1_10_pass_1,
    "1.10 MFA enabled: MFA virtual=true but account never used, expect: passed": cis_aws_iam_1_10_pass_2,
    "1.10 MFA enabled: MFA enabled=false, expect: failed": cis_aws_iam_1_10_fail,
}

cis_aws_iam_1_11_pass = EksAwsServiceCase(
    rule_tag=CIS_1_11,
    case_identifier="test-mfa-virtual-pass",
    expected=RULE_PASS_STATUS,
)

cis_aws_iam_1_11_fail = EksAwsServiceCase(
    rule_tag=CIS_1_11,
    case_identifier="test-setup-access-keys-during-init-fail",
    expected=RULE_FAIL_STATUS,
)

cis_aws_iam_1_11 = {
    "1.11 Access key during user init: no access key, expect: passed": cis_aws_iam_1_11_pass,
}

cis_aws_iam_1_11_skip = {
    # Skipping this test because creating access keys through user init is not possible.
    "1.11 Access key during user init: access key is not not used, expect: failed": cis_aws_iam_1_11_fail,
}

cis_aws_iam_1_13_fail = EksAwsServiceCase(
    rule_tag=CIS_1_13,
    case_identifier="test-user-2-active-keys",
    expected=RULE_FAIL_STATUS,
)

cis_aws_iam_1_13_pass = EksAwsServiceCase(
    rule_tag=CIS_1_13,
    case_identifier="test-setup-access-keys-during-init",
    expected=RULE_PASS_STATUS,
)

cis_aws_iam_1_13_pass_2 = EksAwsServiceCase(
    rule_tag=CIS_1_13,
    case_identifier="test-user-1-active-1-not-active-keys",
    expected=RULE_PASS_STATUS,
)

cis_aws_iam_1_13 = {
    "1.13 Active key for user: 1 active key, expect: passed": cis_aws_iam_1_13_pass,
    "1.13 Active key for user: 1 active key, 1 deactivated key, expect: passed": cis_aws_iam_1_13_pass_2,
    "1.13 Active key for user: 2 active keys, expect: failed": cis_aws_iam_1_13_fail,
}

cis_aws_iam_1_15_fail = EksAwsServiceCase(
    rule_tag=CIS_1_15,
    case_identifier="test-user-with-inline-policy-fail",
    expected=RULE_FAIL_STATUS,
)

cis_aws_iam_1_15_pass = EksAwsServiceCase(
    rule_tag=CIS_1_15,
    case_identifier="test-mfa-virtual-pass",  # contains only group policy
    expected=RULE_PASS_STATUS,
)

cis_aws_iam_1_15_pass_2 = EksAwsServiceCase(
    rule_tag=CIS_1_15,
    case_identifier="test-mfa-virtual-never-used",  # no groups, no inline policies
    expected=RULE_PASS_STATUS,
)

cis_aws_iam_1_15_skip = {
    "1.15 Permissions through groups: dev group only, expect: passed": cis_aws_iam_1_15_pass,
    "1.15 Permissions through groups: no group permissions, expect: passed": cis_aws_iam_1_15_pass_2,
    "1.15 Permissions through groups: inline policy, expect: failed": cis_aws_iam_1_15_fail,
}

cis_aws_iam_1_16 = {
    "1.16 Built-in EC2 policy, expect: passed": EksAwsServiceCase(
        rule_tag=CIS_1_16,
        case_identifier="AmazonEC2FullAccess",
        expected=RULE_PASS_STATUS,
    ),
    "1.16 Attached AdministratorAccess Policy, expect: failed": EksAwsServiceCase(
        rule_tag=CIS_1_16,
        case_identifier="AdministratorAccess",
        expected=RULE_FAIL_STATUS,
    ),
}

cis_aws_iam_1_17 = {
    "1.17 Attached support access policy, expect: passed": EksAwsServiceCase(
        rule_tag=CIS_1_17,
        case_identifier="AWSSupportAccess",
        expected=RULE_PASS_STATUS,
    ),
}

cis_aws_iam_1_20 = {
    "1.20 Valid access analyzers in all regions, expect: passed": EksAwsServiceCase(
        rule_tag=CIS_1_20,
        case_identifier="account-access-analyzers",
        expected=RULE_PASS_STATUS,
    ),
}

# The name of this variable needs to be `tests_cases` in order to CIS Rules coverage stats to be generated
test_cases = {
    **cis_aws_iam_1_4,
    **skip_param_case(
        cis_aws_iam_1_4_skip,
        data_to_report=SkipReportData(
            skip_reason="Test case data generation issue",
            url_title="security-team: #6204",
            url_link="https://github.com/elastic/security-team/issues/6204",
        ),
    ),
    **cis_aws_iam_1_5,
    **skip_param_case(
        cis_aws_iam_1_5_skip,
        data_to_report=SkipReportData(
            skip_reason="Test case data generation issue",
            url_title="security-team: #6204",
            url_link="https://github.com/elastic/security-team/issues/6204",
        ),
    ),
    **cis_aws_iam_1_6,
    **skip_param_case(
        cis_aws_iam_1_6_skip,
        data_to_report=SkipReportData(
            skip_reason="Test case data generation issue",
            url_title="security-team: #6204",
            url_link="https://github.com/elastic/security-team/issues/6204",
        ),
    ),
    **cis_aws_iam_1_7,
    **skip_param_case(
        cis_aws_iam_1_7_skip,
        data_to_report=SkipReportData(
            skip_reason="Test case data generation issue",
            url_title="security-team: #6204",
            url_link="https://github.com/elastic/security-team/issues/6204",
        ),
    ),
    **cis_aws_iam_1_8,
    **skip_param_case(
        cis_aws_iam_1_8_skip,
        data_to_report=SkipReportData(
            skip_reason="Test case data generation issue",
            url_title="security-team: #6204",
            url_link="https://github.com/elastic/security-team/issues/6204",
        ),
    ),
    **cis_aws_iam_1_9,
    **skip_param_case(
        cis_aws_iam_1_9_skip,
        data_to_report=SkipReportData(
            skip_reason="Test case data generation issue",
            url_title="security-team: #6204",
            url_link="https://github.com/elastic/security-team/issues/6204",
        ),
    ),
    **cis_aws_iam_1_10,
    **cis_aws_iam_1_11,
    **cis_aws_iam_1_13,
    **skip_param_case(
        cis_aws_iam_1_15_skip,
        data_to_report=SkipReportData(
            skip_reason="Test case data generation issue",
            url_title="security-team: #6204",
            url_link="https://github.com/elastic/security-team/issues/6204",
        ),
    ),
    **cis_aws_iam_1_16,
    **cis_aws_iam_1_17,
    **cis_aws_iam_1_20,
}
