"""
This module provides eks file system rule test cases.
Cases are organized as rules.
Each rule has one or more test cases.
"""

from ..eks_test_case import EksTestCase
from ..constants import RULE_PASS_STATUS, RULE_FAIL_STATUS
from configuration import eks

config_1_node_1 = eks.config_1_node_1
config_1_node_2 = eks.config_1_node_2

cis_eks_3_1_1_pass = EksTestCase(
    rule_tag='CIS 3.1.1',
    node_hostname=config_1_node_1,
    expected=RULE_PASS_STATUS
)

cis_eks_3_1_1_fail = EksTestCase(
    rule_tag='CIS 3.1.1',
    node_hostname=config_1_node_2,
    expected=RULE_FAIL_STATUS
)

cis_eks_3_1_1 = {
    '3.1.1 Kubeconfig file permissions 644 evaluation passed': cis_eks_3_1_1_pass,
    '3.1.1 Kubeconfig file permissions 700 evaluation failed': cis_eks_3_1_1_fail
}

cis_eks_3_1_2_user_fail = EksTestCase(
    rule_tag='CIS 3.1.2',
    node_hostname=config_1_node_1,
    expected=RULE_FAIL_STATUS
)

cis_eks_3_1_2_group_fail = EksTestCase(
    rule_tag='CIS 3.1.2',
    node_hostname=config_1_node_2,
    expected=RULE_FAIL_STATUS
)

cis_eks_3_1_2 = {
    '3.1.2 Kubeconfig ownership invalid user evaluation failed': cis_eks_3_1_2_user_fail,
    '3.1.2 Kubeconfig ownership invalid group evaluation failed': cis_eks_3_1_2_group_fail
}

cis_eks_3_1_3_pass = EksTestCase(
    rule_tag='CIS 3.1.3',
    node_hostname=config_1_node_1,
    expected=RULE_PASS_STATUS
)

cis_eks_3_1_3_fail = EksTestCase(
    rule_tag='CIS 3.1.3',
    node_hostname=config_1_node_2,
    expected=RULE_FAIL_STATUS
)

cis_eks_3_1_3 = {
    '3.1.3 Kubelet-config file permissions 644 evaluation passed': cis_eks_3_1_3_pass,
    '3.1.3 Kubelet-config file permissions 700 evaluation failed': cis_eks_3_1_3_fail
}


cis_eks_3_1_4_user_fail = EksTestCase(
    rule_tag='CIS 3.1.4',
    node_hostname=config_1_node_1,
    expected=RULE_FAIL_STATUS
)

cis_eks_3_1_4_group_fail = EksTestCase(
    rule_tag='CIS 3.1.4',
    node_hostname=config_1_node_2,
    expected=RULE_FAIL_STATUS
)

cis_eks_3_1_4 = {
    '3.1.4 Kubelet-config ownership invalid user evaluation failed': cis_eks_3_1_4_user_fail,
    '3.1.4 Kubelet-config ownership invalid group evaluation failed': cis_eks_3_1_4_group_fail
}
