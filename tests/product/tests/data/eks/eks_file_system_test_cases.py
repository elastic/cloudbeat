"""
This module provides eks file system rule test cases.
Cases are organized as rules.
Each rule has one or more test cases.
"""

from configuration import eks

from ..constants import RULE_FAIL_STATUS, RULE_PASS_STATUS
from ..eks_test_case import EksTestCase

config_1_node_1 = eks.config_1_node_1
config_1_node_2 = eks.config_1_node_2
config_2_node_1 = eks.config_2_node_1

cis_eks_3_1_1_pass = EksTestCase(
    rule_tag="CIS 3.1.1",
    node_hostname=config_1_node_1,
    expected=RULE_PASS_STATUS,
)

cis_eks_3_1_1_fail = EksTestCase(
    rule_tag="CIS 3.1.1",
    node_hostname=config_1_node_2,
    expected=RULE_FAIL_STATUS,
)

cis_eks_3_1_1_pass_2 = EksTestCase(
    rule_tag="CIS 3.1.1",
    node_hostname=config_2_node_1,
    expected=RULE_PASS_STATUS,
)

cis_eks_3_1_1 = {
    "3.1.1 Kubeconfig file permissions 644 evaluation passed": cis_eks_3_1_1_pass,
    "3.1.1 Kubeconfig file permissions 700 evaluation failed": cis_eks_3_1_1_fail,
}

cis_eks_3_1_1_conf_2 = {
    "3.1.1 Kubeconfig file permissions 644 evaluation passed": cis_eks_3_1_1_pass_2,
}


cis_eks_3_1_2_user_fail = EksTestCase(
    rule_tag="CIS 3.1.2",
    node_hostname=config_1_node_1,
    expected=RULE_FAIL_STATUS,
)

cis_eks_3_1_2_group_fail = EksTestCase(
    rule_tag="CIS 3.1.2",
    node_hostname=config_1_node_2,
    expected=RULE_FAIL_STATUS,
)

cis_eks_3_1_2 = {
    "3.1.2 Kubeconfig ownership invalid user evaluation failed": cis_eks_3_1_2_user_fail,
    "3.1.2 Kubeconfig ownership invalid group evaluation failed": cis_eks_3_1_2_group_fail,
}

cis_eks_3_1_3_pass = EksTestCase(
    rule_tag="CIS 3.1.3",
    node_hostname=config_1_node_1,
    expected=RULE_PASS_STATUS,
)

cis_eks_3_1_3_fail = EksTestCase(
    rule_tag="CIS 3.1.3",
    node_hostname=config_1_node_2,
    expected=RULE_FAIL_STATUS,
)

cis_eks_3_1_3 = {
    "3.1.3 Kubelet-config file permissions 644 evaluation passed": cis_eks_3_1_3_pass,
    "3.1.3 Kubelet-config file permissions 700 evaluation failed": cis_eks_3_1_3_fail,
}


cis_eks_3_1_4_user_fail = EksTestCase(
    rule_tag="CIS 3.1.4",
    node_hostname=config_1_node_1,
    expected=RULE_FAIL_STATUS,
)

cis_eks_3_1_4_group_fail = EksTestCase(
    rule_tag="CIS 3.1.4",
    node_hostname=config_1_node_2,
    expected=RULE_FAIL_STATUS,
)

cis_eks_3_1_4 = {
    "3.1.4 Kubelet-config ownership invalid user evaluation failed": cis_eks_3_1_4_user_fail,
    "3.1.4 Kubelet-config ownership invalid group evaluation failed": cis_eks_3_1_4_group_fail,
}

file_system_config_1 = {
    **cis_eks_3_1_1,
    **cis_eks_3_1_2,
    **cis_eks_3_1_3,
    **cis_eks_3_1_4,
}

file_system_config_2 = {
    **cis_eks_3_1_1_conf_2,
}

# Each rule may contain several test cases depended on configuration
# The configuration is provided through environment var and known when the test execution starts.
# This dictionary summarizes all cases and all configurations.
# But during runtime only one of them may be used (test-eks-config-1 / test-eks-config-2)
cis_eks_all = {
    "test-eks-config-1": file_system_config_1,
    "test-eks-config-2": file_system_config_2,
}

cis_eks_file_system_cases = cis_eks_all.get(eks.current_config, {})
