"""
This module provides AWS Monitoring rule test cases.
Cases are organized as rules.
Each rule has one or more test cases.
Monitoring rules identification is performed by resource name.
"""

from commonlib.framework.reporting import SkipReportData, skip_param_case

from ..constants import RULE_FAIL_STATUS, RULE_PASS_STATUS
from ..eks_test_case import EksAwsServiceCase

CIS_4_1 = "CIS 4.1"
CIS_4_2 = "CIS 4.2"
CIS_4_3 = "CIS 4.3"
CIS_4_4 = "CIS 4.4"
CIS_4_5 = "CIS 4.5"
CIS_4_6 = "CIS 4.6"
CIS_4_7 = "CIS 4.7"
CIS_4_8 = "CIS 4.8"
CIS_4_9 = "CIS 4.9"
CIS_4_10 = "CIS 4.10"
CIS_4_11 = "CIS 4.11"
CIS_4_12 = "CIS 4.12"
CIS_4_13 = "CIS 4.13"
CIS_4_14 = "CIS 4.14"
CIS_4_15 = "CIS 4.15"
CIS_4_16 = "CIS 4.16"

VALID_METRICS_ACCOUNT_ID = "cloudtrail-391946104644"
INVALID_METRICS_ACCOUNT_ID = "to-define-user-account"

cis_aws_monitoring_4_1_pass = EksAwsServiceCase(
    rule_tag=CIS_4_1,
    case_identifier=VALID_METRICS_ACCOUNT_ID,
    expected=RULE_PASS_STATUS,
)

cis_aws_monitoring_4_1_fail = EksAwsServiceCase(
    rule_tag=CIS_4_1,
    case_identifier=INVALID_METRICS_ACCOUNT_ID,
    expected=RULE_FAIL_STATUS,
)

cis_aws_monitoring_4_1 = {
    "4.1 Ensure filter and alarm unauthorized API, exists=true expect: passed": cis_aws_monitoring_4_1_pass,
}

cis_aws_monitoring_4_1_skip = {
    "4.1 Ensure filter and alarm unauthorized API, exists=false expect: failed": cis_aws_monitoring_4_1_fail,
}

cis_aws_monitoring_4_2_pass = EksAwsServiceCase(
    rule_tag=CIS_4_2,
    case_identifier=VALID_METRICS_ACCOUNT_ID,
    expected=RULE_PASS_STATUS,
)

cis_aws_monitoring_4_2_fail = EksAwsServiceCase(
    rule_tag=CIS_4_2,
    case_identifier=INVALID_METRICS_ACCOUNT_ID,
    expected=RULE_FAIL_STATUS,
)

cis_aws_monitoring_4_2 = {
    "4.2 Ensure filter and alarm Console Management - no MFA, exists=true expect: passed": cis_aws_monitoring_4_2_pass,
}

cis_aws_monitoring_4_2_skip = {
    "4.2 Ensure filter and alarm Console Management - no MFA, exists=false expect: failed": cis_aws_monitoring_4_2_fail,
}

cis_aws_monitoring_4_3_pass = EksAwsServiceCase(
    rule_tag=CIS_4_3,
    case_identifier=VALID_METRICS_ACCOUNT_ID,
    expected=RULE_PASS_STATUS,
)

cis_aws_monitoring_4_3_fail = EksAwsServiceCase(
    rule_tag=CIS_4_3,
    case_identifier=INVALID_METRICS_ACCOUNT_ID,
    expected=RULE_FAIL_STATUS,
)

cis_aws_monitoring_4_3 = {
    "4.3 Ensure filter and alarm 'root' account, exists=true expect: passed": cis_aws_monitoring_4_3_pass,
}

cis_aws_monitoring_4_3_skip = {
    "4.3 Ensure filter and alarm 'root' account, exists=false expect: failed": cis_aws_monitoring_4_3_fail,
}

cis_aws_monitoring_4_4_pass = EksAwsServiceCase(
    rule_tag=CIS_4_4,
    case_identifier=VALID_METRICS_ACCOUNT_ID,
    expected=RULE_PASS_STATUS,
)

cis_aws_monitoring_4_4_fail = EksAwsServiceCase(
    rule_tag=CIS_4_4,
    case_identifier=INVALID_METRICS_ACCOUNT_ID,
    expected=RULE_FAIL_STATUS,
)

cis_aws_monitoring_4_4 = {
    "4.4 Ensure filter and alarm 'IAM policy', exists=true expect: passed": cis_aws_monitoring_4_4_pass,
}

cis_aws_monitoring_4_4_skip = {
    "4.4 Ensure filter and alarm 'IAM policy', exists=false expect: failed": cis_aws_monitoring_4_4_fail,
}

cis_aws_monitoring_4_5_pass = EksAwsServiceCase(
    rule_tag=CIS_4_5,
    case_identifier=VALID_METRICS_ACCOUNT_ID,
    expected=RULE_PASS_STATUS,
)

cis_aws_monitoring_4_5_fail = EksAwsServiceCase(
    rule_tag=CIS_4_5,
    case_identifier=INVALID_METRICS_ACCOUNT_ID,
    expected=RULE_FAIL_STATUS,
)

cis_aws_monitoring_4_5 = {
    "4.5 Ensure filter and alarm CloudTrail config, exists=true expect: passed": cis_aws_monitoring_4_5_pass,
}

cis_aws_monitoring_4_5_skip = {
    "4.5 Ensure filter and alarm CloudTrail config, exists=false expect: failed": cis_aws_monitoring_4_5_fail,
}

cis_aws_monitoring_4_6_pass = EksAwsServiceCase(
    rule_tag=CIS_4_6,
    case_identifier=VALID_METRICS_ACCOUNT_ID,
    expected=RULE_PASS_STATUS,
)

cis_aws_monitoring_4_6_fail = EksAwsServiceCase(
    rule_tag=CIS_4_6,
    case_identifier=INVALID_METRICS_ACCOUNT_ID,
    expected=RULE_FAIL_STATUS,
)

cis_aws_monitoring_4_6 = {
    "4.6 Ensure filter and alarm Auth failures, exists=true expect: passed": cis_aws_monitoring_4_6_pass,
}

cis_aws_monitoring_4_6_skip = {
    "4.6 Ensure filter and alarm Auth failures, exists=false expect: failed": cis_aws_monitoring_4_6_fail,
}

cis_aws_monitoring_4_7_pass = EksAwsServiceCase(
    rule_tag=CIS_4_7,
    case_identifier=VALID_METRICS_ACCOUNT_ID,
    expected=RULE_PASS_STATUS,
)

cis_aws_monitoring_4_7_fail = EksAwsServiceCase(
    rule_tag=CIS_4_7,
    case_identifier=INVALID_METRICS_ACCOUNT_ID,
    expected=RULE_FAIL_STATUS,
)

cis_aws_monitoring_4_7 = {
    "4.7 Ensure filter and alarm CMKs deletion, exists=true expect: passed": cis_aws_monitoring_4_7_pass,
}

cis_aws_monitoring_4_7_skip = {
    "4.7 Ensure filter and alarm CMKs deletion, exists=false expect: failed": cis_aws_monitoring_4_7_fail,
}

cis_aws_monitoring_4_8_pass = EksAwsServiceCase(
    rule_tag=CIS_4_8,
    case_identifier=VALID_METRICS_ACCOUNT_ID,
    expected=RULE_PASS_STATUS,
)

cis_aws_monitoring_4_8_fail = EksAwsServiceCase(
    rule_tag=CIS_4_8,
    case_identifier=INVALID_METRICS_ACCOUNT_ID,
    expected=RULE_FAIL_STATUS,
)

cis_aws_monitoring_4_8 = {
    "4.8 Ensure filter and alarm CMKs deletion, exists=true expect: passed": cis_aws_monitoring_4_8_pass,
}

cis_aws_monitoring_4_8_skip = {
    "4.8 Ensure filter and alarm CMKs deletion, exists=false expect: failed": cis_aws_monitoring_4_8_fail,
}

cis_aws_monitoring_4_9_pass = EksAwsServiceCase(
    rule_tag=CIS_4_9,
    case_identifier=VALID_METRICS_ACCOUNT_ID,
    expected=RULE_PASS_STATUS,
)

cis_aws_monitoring_4_9_fail = EksAwsServiceCase(
    rule_tag=CIS_4_9,
    case_identifier=INVALID_METRICS_ACCOUNT_ID,
    expected=RULE_FAIL_STATUS,
)

cis_aws_monitoring_4_9 = {
    "4.9 Ensure filter and alarm AWS config changes, exists=true expect: passed": cis_aws_monitoring_4_9_pass,
}

cis_aws_monitoring_4_9_skip = {
    "4.9 Ensure filter and alarm AWS config changes, exists=false expect: failed": cis_aws_monitoring_4_9_fail,
}

cis_aws_monitoring_4_10_pass = EksAwsServiceCase(
    rule_tag=CIS_4_10,
    case_identifier=VALID_METRICS_ACCOUNT_ID,
    expected=RULE_PASS_STATUS,
)

cis_aws_monitoring_4_10_fail = EksAwsServiceCase(
    rule_tag=CIS_4_10,
    case_identifier=INVALID_METRICS_ACCOUNT_ID,
    expected=RULE_FAIL_STATUS,
)

cis_aws_monitoring_4_10 = {
    "4.10 Ensure filter and alarm Sec Group changes, exists=true expect: passed": cis_aws_monitoring_4_10_pass,
}

cis_aws_monitoring_4_10_skip = {
    "4.10 Ensure filter and alarm Sec Group changes, exists=false expect: failed": cis_aws_monitoring_4_10_fail,
}

cis_aws_monitoring_4_11_pass = EksAwsServiceCase(
    rule_tag=CIS_4_11,
    case_identifier=VALID_METRICS_ACCOUNT_ID,
    expected=RULE_PASS_STATUS,
)

cis_aws_monitoring_4_11_fail = EksAwsServiceCase(
    rule_tag=CIS_4_11,
    case_identifier=INVALID_METRICS_ACCOUNT_ID,
    expected=RULE_FAIL_STATUS,
)

cis_aws_monitoring_4_11 = {
    "4.11 Ensure filter and alarm NACL changes, exists=true expect: passed": cis_aws_monitoring_4_11_pass,
}

cis_aws_monitoring_4_11_skip = {
    "4.11 Ensure filter and alarm NACL changes, exists=false expect: failed": cis_aws_monitoring_4_11_fail,
}

cis_aws_monitoring_4_12_pass = EksAwsServiceCase(
    rule_tag=CIS_4_12,
    case_identifier=VALID_METRICS_ACCOUNT_ID,
    expected=RULE_PASS_STATUS,
)

cis_aws_monitoring_4_12_fail = EksAwsServiceCase(
    rule_tag=CIS_4_12,
    case_identifier=INVALID_METRICS_ACCOUNT_ID,
    expected=RULE_FAIL_STATUS,
)

cis_aws_monitoring_4_12 = {
    "4.12 Ensure filter and alarm Network gateways changes, exists=true expect: passed": cis_aws_monitoring_4_12_pass,
}

cis_aws_monitoring_4_12_skip = {
    "4.12 Ensure filter and alarm Network gateways changes, exists=false expect: failed": cis_aws_monitoring_4_12_fail,
}

cis_aws_monitoring_4_13_pass = EksAwsServiceCase(
    rule_tag=CIS_4_13,
    case_identifier=VALID_METRICS_ACCOUNT_ID,
    expected=RULE_PASS_STATUS,
)

cis_aws_monitoring_4_13_fail = EksAwsServiceCase(
    rule_tag=CIS_4_13,
    case_identifier=INVALID_METRICS_ACCOUNT_ID,
    expected=RULE_FAIL_STATUS,
)

cis_aws_monitoring_4_13 = {
    "4.13 Ensure filter and alarm Route Table changes, exists=true expect: passed": cis_aws_monitoring_4_13_pass,
}

cis_aws_monitoring_4_13_skip = {
    "4.13 Ensure filter and alarm Route Table changes, exists=false expect: failed": cis_aws_monitoring_4_13_fail,
}

cis_aws_monitoring_4_14_pass = EksAwsServiceCase(
    rule_tag=CIS_4_14,
    case_identifier=VALID_METRICS_ACCOUNT_ID,
    expected=RULE_PASS_STATUS,
)

cis_aws_monitoring_4_14_fail = EksAwsServiceCase(
    rule_tag=CIS_4_14,
    case_identifier=INVALID_METRICS_ACCOUNT_ID,
    expected=RULE_FAIL_STATUS,
)

cis_aws_monitoring_4_14 = {
    "4.14 Ensure filter and alarm VPC changes, exists=true expect: passed": cis_aws_monitoring_4_14_pass,
}

cis_aws_monitoring_4_14_skip = {
    "4.14 Ensure filter and alarm VPC changes, exists=false expect: failed": cis_aws_monitoring_4_14_fail,
}

cis_aws_monitoring_4_15_pass = EksAwsServiceCase(
    rule_tag=CIS_4_15,
    case_identifier=VALID_METRICS_ACCOUNT_ID,
    expected=RULE_PASS_STATUS,
)

cis_aws_monitoring_4_15_fail = EksAwsServiceCase(
    rule_tag=CIS_4_15,
    case_identifier=INVALID_METRICS_ACCOUNT_ID,
    expected=RULE_FAIL_STATUS,
)

cis_aws_monitoring_4_15 = {
    "4.15 Ensure filter and alarm AWS ORGs changes, exists=true expect: passed": cis_aws_monitoring_4_15_pass,
}

cis_aws_monitoring_4_15_skip = {
    "4.15 Ensure filter and alarm AWS ORGs changes, exists=false expect: failed": cis_aws_monitoring_4_15_fail,
}

cis_aws_monitoring_4_16_pass = EksAwsServiceCase(
    rule_tag=CIS_4_16,
    case_identifier="arn:aws:securityhub:eu-north-1:391946104644:hub/default",
    expected=RULE_PASS_STATUS,
)

cis_aws_monitoring_4_16_fail = EksAwsServiceCase(
    rule_tag=CIS_4_16,
    case_identifier="securityhub-eu-west-1-391946104644",
    expected=RULE_FAIL_STATUS,
)

cis_aws_monitoring_4_16 = {
    "4.16 Ensure AWS Security Hub is enabled, Hub Enabled=true expect: passed": cis_aws_monitoring_4_16_pass,
    "4.16 Ensure AWS Security Hub is enabled, Hub Enabled=false expect: failed": cis_aws_monitoring_4_16_fail,
}

# The name of this variable needs to be `tests_cases` in order to CIS Rules coverage stats to be generated
test_cases = {
    **cis_aws_monitoring_4_1,
    **skip_param_case(
        cis_aws_monitoring_4_1_skip,
        data_to_report=SkipReportData(
            skip_reason="Test case data generation issue",
            url_title="security-team: #6204",
            url_link="https://github.com/elastic/security-team/issues/6204",
        ),
    ),
    **cis_aws_monitoring_4_2,
    **skip_param_case(
        cis_aws_monitoring_4_2_skip,
        data_to_report=SkipReportData(
            skip_reason="Test case data generation issue",
            url_title="security-team: #6204",
            url_link="https://github.com/elastic/security-team/issues/6204",
        ),
    ),
    **cis_aws_monitoring_4_3,
    **skip_param_case(
        cis_aws_monitoring_4_3_skip,
        data_to_report=SkipReportData(
            skip_reason="Test case data generation issue",
            url_title="security-team: #6204",
            url_link="https://github.com/elastic/security-team/issues/6204",
        ),
    ),
    **cis_aws_monitoring_4_4,
    **skip_param_case(
        cis_aws_monitoring_4_4_skip,
        data_to_report=SkipReportData(
            skip_reason="Test case data generation issue",
            url_title="security-team: #6204",
            url_link="https://github.com/elastic/security-team/issues/6204",
        ),
    ),
    **cis_aws_monitoring_4_5,
    **skip_param_case(
        cis_aws_monitoring_4_5_skip,
        data_to_report=SkipReportData(
            skip_reason="Test case data generation issue",
            url_title="security-team: #6204",
            url_link="https://github.com/elastic/security-team/issues/6204",
        ),
    ),
    **cis_aws_monitoring_4_6,
    **skip_param_case(
        cis_aws_monitoring_4_6_skip,
        data_to_report=SkipReportData(
            skip_reason="Test case data generation issue",
            url_title="security-team: #6204",
            url_link="https://github.com/elastic/security-team/issues/6204",
        ),
    ),
    **cis_aws_monitoring_4_7,
    **skip_param_case(
        cis_aws_monitoring_4_7_skip,
        data_to_report=SkipReportData(
            skip_reason="Test case data generation issue",
            url_title="security-team: #6204",
            url_link="https://github.com/elastic/security-team/issues/6204",
        ),
    ),
    **cis_aws_monitoring_4_8,
    **skip_param_case(
        cis_aws_monitoring_4_8_skip,
        data_to_report=SkipReportData(
            skip_reason="Test case data generation issue",
            url_title="security-team: #6204",
            url_link="https://github.com/elastic/security-team/issues/6204",
        ),
    ),
    **cis_aws_monitoring_4_9,
    **skip_param_case(
        cis_aws_monitoring_4_9_skip,
        data_to_report=SkipReportData(
            skip_reason="Test case data generation issue",
            url_title="security-team: #6204",
            url_link="https://github.com/elastic/security-team/issues/6204",
        ),
    ),
    **cis_aws_monitoring_4_10,
    **skip_param_case(
        cis_aws_monitoring_4_10_skip,
        data_to_report=SkipReportData(
            skip_reason="Test case data generation issue",
            url_title="security-team: #6204",
            url_link="https://github.com/elastic/security-team/issues/6204",
        ),
    ),
    **cis_aws_monitoring_4_11,
    **skip_param_case(
        cis_aws_monitoring_4_11_skip,
        data_to_report=SkipReportData(
            skip_reason="Test case data generation issue",
            url_title="security-team: #6204",
            url_link="https://github.com/elastic/security-team/issues/6204",
        ),
    ),
    **cis_aws_monitoring_4_12,
    **skip_param_case(
        cis_aws_monitoring_4_12_skip,
        data_to_report=SkipReportData(
            skip_reason="Test case data generation issue",
            url_title="security-team: #6204",
            url_link="https://github.com/elastic/security-team/issues/6204",
        ),
    ),
    **cis_aws_monitoring_4_13,
    **skip_param_case(
        cis_aws_monitoring_4_13_skip,
        data_to_report=SkipReportData(
            skip_reason="Test case data generation issue",
            url_title="security-team: #6204",
            url_link="https://github.com/elastic/security-team/issues/6204",
        ),
    ),
    **cis_aws_monitoring_4_14,
    **skip_param_case(
        cis_aws_monitoring_4_14_skip,
        data_to_report=SkipReportData(
            skip_reason="Test case data generation issue",
            url_title="security-team: #6204",
            url_link="https://github.com/elastic/security-team/issues/6204",
        ),
    ),
    **cis_aws_monitoring_4_15,
    **skip_param_case(
        cis_aws_monitoring_4_15_skip,
        data_to_report=SkipReportData(
            skip_reason="Test case data generation issue",
            url_title="security-team: #6204",
            url_link="https://github.com/elastic/security-team/issues/6204",
        ),
    ),
    **cis_aws_monitoring_4_16,
}
