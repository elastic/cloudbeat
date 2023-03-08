"""
This module provides AWS IAM rules test cases.
Cases are organized as rules.
Each rule has one or more test cases.
IAM identification is performed by resource name.
"""
from commonlib.framework.reporting import skip_param_case, SkipReportData
from ..eks_test_case import EksAwsServiceCase
from ..constants import RULE_PASS_STATUS, RULE_FAIL_STATUS

CIS_1_4 = "CIS 1.4"
CIS_1_5 = "CIS 1.5"
CIS_1_6 = "CIS 1.6"
CIS_1_7 = "CIS 1.7"
CIS_1_8 = "CIS 1.8"
CIS_1_9 = "CIS 1.9"
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

cis_aws_iam_cases = {
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
}
