"""
This module provides AWS EKS service rule test cases.
Cases are organized as rules.
Each rule has one or more test cases.
"""

from commonlib.framework.reporting import SkipReportData, skip_param_case
from configuration import eks

from ..constants import RULE_FAIL_STATUS, RULE_PASS_STATUS
from ..eks_test_case import EksAwsServiceCase

config_1_node_1 = eks.config_1_node_1

cis_eks_2_1_1_fail = EksAwsServiceCase(
    rule_tag="CIS 2.1.1",
    case_identifier="test-eks-config-1",
    expected=RULE_FAIL_STATUS,
)

cis_eks_2_1_1_pass = EksAwsServiceCase(
    rule_tag="CIS 2.1.1",
    case_identifier="test-eks-config-2",
    expected=RULE_PASS_STATUS,
)

cis_eks_2_1_1_config_1 = {
    "2.1.1 EKS Cluster loggers enabled==false evaluation failed": cis_eks_2_1_1_fail,
}

cis_eks_2_1_1_config_2 = {
    "2.1.1 EKS Cluster loggers all enabled==true evaluation passed": cis_eks_2_1_1_pass,
}

cis_eks_5_1_1_pass = EksAwsServiceCase(
    rule_tag="CIS 5.1.1",
    case_identifier="test-eks-scan-true",
    expected=RULE_PASS_STATUS,
)

cis_eks_5_1_1_fail = EksAwsServiceCase(
    rule_tag="CIS 5.1.1",
    case_identifier="test-eks-scan-false",
    expected=RULE_FAIL_STATUS,
)

cis_eks_5_1_1 = {
    "5.1.1 ECR private repo scanOnPush==true evaluation passed": cis_eks_5_1_1_pass,
    "5.1.1 ECR private repo scanOnPush==false evaluation failed": cis_eks_5_1_1_fail,
}

cis_eks_5_4_3_fail = EksAwsServiceCase(
    rule_tag="CIS 5.4.3",
    case_identifier=config_1_node_1,
    expected=RULE_FAIL_STATUS,
)

cis_eks_5_4_3_config_1 = {
    "5.4.3 Network configuration public==true evaluation failed": cis_eks_5_4_3_fail,
}

cis_eks_5_4_5_fail = EksAwsServiceCase(
    rule_tag="CIS 5.4.5",
    case_identifier="a628adbaa057d44c5b7aa777a9e36462",
    expected=RULE_FAIL_STATUS,
)

cis_eks_5_4_5_config_1 = {
    "5.4.5 ELB - TCP traffic no encryption evaluation failed": cis_eks_5_4_5_fail,
}

aws_test_config_1 = {
    **cis_eks_5_1_1,
    **skip_param_case(
        cis_eks_5_4_3_config_1,
        data_to_report=SkipReportData(
            skip_reason="Retest after testing configuration will be fixed.",
            url_title="cloudbeat: #500",
            url_link="https://github.com/elastic/cloudbeat/issues/500",
        ),
    ),
    **skip_param_case(
        cis_eks_5_4_5_config_1,
        data_to_report=SkipReportData(
            skip_reason="Retest after testing configuration will be fixed.",
            url_title="cloudbeat: #500",
            url_link="https://github.com/elastic/cloudbeat/issues/500",
        ),
    ),
    **skip_param_case(
        cis_eks_2_1_1_config_1,
        data_to_report=SkipReportData(
            skip_reason="This rule is implemented partially",
            url_title="security-team: #3929",
            url_link="https://github.com/elastic/security-team/issues/3929",
        ),
    ),
}

aws_test_config_2 = {
    **skip_param_case(
        cis_eks_2_1_1_config_2,
        data_to_report=SkipReportData(
            skip_reason="This rule is implemented partially",
            url_title="security-team: #3929",
            url_link="https://github.com/elastic/security-team/issues/3929",
        ),
    ),
}

cis_eks_all = {
    "test-eks-config-1": aws_test_config_1,
    "test-eks-config-2": aws_test_config_2,
}

test_cases = cis_eks_all.get(eks.current_config, {})
