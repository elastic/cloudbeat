"""
This module provides eks file system rule test cases.
Cases are organized as rules.
Each rule has one or more test cases.
"""

# from commonlib.framework.reporting import SkipReportData, skip_param_case
from configuration import eks

from ..constants import RULE_FAIL_STATUS, RULE_PASS_STATUS
from ..eks_test_case import EksKubeObjectCase

cis_eks_4_2_1_pass = EksKubeObjectCase(
    rule_tag="CIS 4.2.1",
    test_resource_id="eks-psp-pass",
    expected=RULE_PASS_STATUS,
)

cis_eks_4_2_1_fail = EksKubeObjectCase(
    rule_tag="CIS 4.2.1",
    test_resource_id="eks-psp-failures",
    expected=RULE_FAIL_STATUS,
)

cis_eks_4_2_1 = {
    "4.2.1 PSP spec.securityContext.privileged==false eval passed": cis_eks_4_2_1_pass,
    "4.2.1 PSP spec.securityContext.privileged==true eval failed": cis_eks_4_2_1_fail,
}

cis_eks_4_2_2_pass = EksKubeObjectCase(
    rule_tag="CIS 4.2.2",
    test_resource_id="eks-psp-pass",
    expected=RULE_PASS_STATUS,
)

cis_eks_4_2_2_fail = EksKubeObjectCase(
    rule_tag="CIS 4.2.2",
    test_resource_id="eks-psp-failures",
    expected=RULE_FAIL_STATUS,
)

cis_eks_4_2_2 = {
    "4.2.2 PSP spec.hostPID==false eval passed": cis_eks_4_2_2_pass,
    "4.2.2 PSP spec.hostPID==true eval failed": cis_eks_4_2_2_fail,
}

cis_eks_4_2_3_pass = EksKubeObjectCase(
    rule_tag="CIS 4.2.3",
    test_resource_id="eks-psp-pass",
    expected=RULE_PASS_STATUS,
)

cis_eks_4_2_3_fail = EksKubeObjectCase(
    rule_tag="CIS 4.2.3",
    test_resource_id="eks-psp-failures",
    expected=RULE_FAIL_STATUS,
)

cis_eks_4_2_3 = {
    "4.2.3 PSP spec.hostIPC==false eval passed": cis_eks_4_2_3_pass,
    "4.2.3 PSP spec.hostIPC==true eval failed": cis_eks_4_2_3_fail,
}

cis_eks_4_2_4_pass = EksKubeObjectCase(
    rule_tag="CIS 4.2.4",
    test_resource_id="eks-psp-pass",
    expected=RULE_PASS_STATUS,
)

cis_eks_4_2_4_fail = EksKubeObjectCase(
    rule_tag="CIS 4.2.4",
    test_resource_id="eks-psp-failures",
    expected=RULE_FAIL_STATUS,
)

cis_eks_4_2_4 = {
    "4.2.4 PSP spec.hostNetwork==false eval passed": cis_eks_4_2_4_pass,
    "4.2.4 PSP spec.hostNetwork==true eval failed": cis_eks_4_2_4_fail,
}

cis_eks_4_2_5_pass = EksKubeObjectCase(
    rule_tag="CIS 4.2.5",
    test_resource_id="eks-psp-pass",
    expected=RULE_PASS_STATUS,
)

cis_eks_4_2_5_fail = EksKubeObjectCase(
    rule_tag="CIS 4.2.5",
    test_resource_id="eks-psp-failures",
    expected=RULE_FAIL_STATUS,
)

cis_eks_4_2_5 = {
    "4.2.5 PSP spec.securityContext.privileged==false eval passed": cis_eks_4_2_5_pass,
    "4.2.5 PSP spec.securityContext.privileged==true eval failed": cis_eks_4_2_5_fail,
}

cis_eks_4_2_6_pass = EksKubeObjectCase(
    rule_tag="CIS 4.2.6",
    test_resource_id="eks-psp-pass",
    expected=RULE_PASS_STATUS,
)

cis_eks_4_2_6_fail = EksKubeObjectCase(
    rule_tag="CIS 4.2.6",
    test_resource_id="eks-psp-failures",
    expected=RULE_FAIL_STATUS,
)

cis_eks_4_2_6 = {
    "4.2.6 PSP spec.securityContext.runAsNonRoot==true eval passed": cis_eks_4_2_6_pass,
    "4.2.6 PSP spec.securityContext.runAsUser==0 eval failed": cis_eks_4_2_6_fail,
}

cis_eks_4_2_7_pass = EksKubeObjectCase(
    rule_tag="CIS 4.2.7",
    test_resource_id="eks-psp-pass",
    expected=RULE_PASS_STATUS,
)

cis_eks_4_2_7_fail = EksKubeObjectCase(
    rule_tag="CIS 4.2.7",
    test_resource_id="eks-psp-failures",
    expected=RULE_FAIL_STATUS,
)

cis_eks_4_2_7 = {
    "4.2.7 PSP spec.securityContext.capabilities default == not presented eval passed": cis_eks_4_2_7_pass,
    '4.2.7 PSP spec.securityContext.capabilities.add==["NET_RAW"] eval failed': cis_eks_4_2_7_fail,
}

cis_eks_4_2_8_pass = EksKubeObjectCase(
    rule_tag="CIS 4.2.8",
    test_resource_id="eks-psp-pass",
    expected=RULE_PASS_STATUS,
)

cis_eks_4_2_8_fail = EksKubeObjectCase(
    rule_tag="CIS 4.2.8",
    test_resource_id="eks-psp-failures",
    expected=RULE_FAIL_STATUS,
)

cis_eks_4_2_8 = {
    "4.2.8 PSP spec.securityContext.capabilities default == not presented eval passed": cis_eks_4_2_8_pass,
    '4.2.8 PSP spec.securityContext.capabilities.add==["NET_ADMIN", "SYS_TIME"] eval failed': cis_eks_4_2_8_fail,
}

cis_eks_4_2_9_pass = EksKubeObjectCase(
    rule_tag="CIS 4.2.9",
    test_resource_id="eks-psp-pass",
    expected=RULE_PASS_STATUS,
)

cis_eks_4_2_9_fail = EksKubeObjectCase(
    rule_tag="CIS 4.2.9",
    test_resource_id="eks-psp-failures",
    expected=RULE_FAIL_STATUS,
)

cis_eks_4_2_9 = {
    '4.2.9 PSP spec.securityContext.capabilities.drop==["ALL"] eval passed': cis_eks_4_2_9_pass,
    '4.2.9 PSP spec.securityContext.capabilities.add==["NET_ADMIN", "SYS_TIME"] eval failed': cis_eks_4_2_9_fail,
}

k8s_object_config_1 = {
    # **cis_eks_4_2_7,
    # **skip_param_case(
    #     cis_eks_4_2_8,
    #     data_to_report=SkipReportData(
    #         skip_reason="Retest after testing configuration will be fixed.",
    #         url_title="cloudbeat: #500",
    #         url_link="https://github.com/elastic/cloudbeat/issues/500",
    #     ),
    # ),
    # **cis_eks_4_2_9,
}

k8s_object_config_2 = {
    # **cis_eks_4_2_1,
    # **cis_eks_4_2_2,
    # **cis_eks_4_2_3,
    # **cis_eks_4_2_4,
    # **cis_eks_4_2_5,
    # **cis_eks_4_2_6,
}

cis_eks_all = {
    "test-eks-config-1": k8s_object_config_1,
    "test-eks-config-2": k8s_object_config_2,
}

cis_eks_k8s_object_cases = cis_eks_all.get(eks.current_config, {})
