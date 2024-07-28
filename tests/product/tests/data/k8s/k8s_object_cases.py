"""
This module defines k8s objects and psp test cases
"""

from ..constants import RULE_FAIL_STATUS, RULE_PASS_STATUS
from .k8s_test_case import K8sTestCase

K8S_CIS_5_1_3 = "CIS 5.1.3"
K8S_CIS_5_1_5 = "CIS 5.1.5"
K8S_CIS_5_1_6 = "CIS 5.1.6"
K8S_CIS_5_2_2 = "CIS 5.2.2"
K8S_CIS_5_2_3 = "CIS 5.2.3"
K8S_CIS_5_2_4 = "CIS 5.2.4"
K8S_CIS_5_2_5 = "CIS 5.2.5"
K8S_CIS_5_2_6 = "CIS 5.2.6"
K8S_CIS_5_2_7 = "CIS 5.2.7"
K8S_CIS_5_2_8 = "CIS 5.2.8"
K8S_CIS_5_2_10 = "CIS 5.2.10"

TEST_FAIL_POD = "test-k8s-bad-pod"
TEST_PASS_POD = "test-k8s-good-pod"
TEST_PASS_ROLE = "test-role-pass"
TEST_FAIL_ROLE = "test-role-fail"
TEST_PASS_CLUSTER_ROLE = "test-cluster-role-pass"
TEST_FAIL_CLUSTER_ROLE = "test-cluster-role-fail"
TEST_PASS_SERVICE_ACCOUNT = "test-service-account-pass"
TEST_FAIL_SERVICE_ACCOUNT = "test-service-account-fail"

cis_5_1_3_role_pass = K8sTestCase(
    rule_tag=K8S_CIS_5_1_3,
    resource_name=TEST_PASS_ROLE,
    expected=RULE_PASS_STATUS,
)

cis_5_1_3_cluster_role_pass = K8sTestCase(
    rule_tag=K8S_CIS_5_1_3,
    resource_name=TEST_PASS_CLUSTER_ROLE,
    expected=RULE_PASS_STATUS,
)

cis_5_1_3_role_fail = K8sTestCase(
    rule_tag=K8S_CIS_5_1_3,
    resource_name=TEST_FAIL_ROLE,
    expected=RULE_FAIL_STATUS,
)

cis_5_1_3_cluster_role_fail = K8sTestCase(
    rule_tag=K8S_CIS_5_1_3,
    resource_name=TEST_FAIL_CLUSTER_ROLE,
    expected=RULE_FAIL_STATUS,
)

cis_5_1_3 = {
    "5.1.3 Role with wildcards": cis_5_1_3_role_fail,
    "5.1.3 Role with no wildcards": cis_5_1_3_role_pass,
    "5.1.3 ClusterRole with wildcards": cis_5_1_3_cluster_role_fail,
    "5.1.3 ClusterRole with no wildcards": cis_5_1_3_cluster_role_pass,
}

cis_5_1_5_sa_pass = K8sTestCase(
    rule_tag=K8S_CIS_5_1_5,
    resource_name=TEST_PASS_SERVICE_ACCOUNT,
    expected=RULE_PASS_STATUS,
)

cis_5_1_5_pod_sa_default_fail = K8sTestCase(
    rule_tag=K8S_CIS_5_1_5,
    resource_name="test-pod-sa-default",
    expected=RULE_FAIL_STATUS,
)

cis_5_1_5_pod_sa_name_default_fail = K8sTestCase(
    rule_tag=K8S_CIS_5_1_5,
    resource_name="test-pod-sa-name-default",
    expected=RULE_FAIL_STATUS,
)

cis_5_1_5 = {
    "5.1.5 ServiceAccount not default": cis_5_1_5_sa_pass,
    "5.1.5 Pod.serviceAccount == default": cis_5_1_5_pod_sa_default_fail,
    "5.1.5 Pod.serviceAccountName == default": cis_5_1_5_pod_sa_name_default_fail,
}

cis_5_1_6_sa_fail = K8sTestCase(
    rule_tag=K8S_CIS_5_1_6,
    resource_name=TEST_FAIL_SERVICE_ACCOUNT,
    expected=RULE_FAIL_STATUS,
)

cis_5_1_6_pod_fail = K8sTestCase(
    rule_tag=K8S_CIS_5_1_6,
    resource_name=TEST_FAIL_POD,
    expected=RULE_FAIL_STATUS,
)

cis_5_1_6_sa_pass = K8sTestCase(
    rule_tag=K8S_CIS_5_1_6,
    resource_name=TEST_PASS_SERVICE_ACCOUNT,
    expected=RULE_PASS_STATUS,
)

cis_5_1_6_pod_pass = K8sTestCase(
    rule_tag=K8S_CIS_5_1_6,
    resource_name="test-pod-sa-default",
    expected=RULE_PASS_STATUS,
)

cis_5_1_6 = {
    "5.1.6 Pod.spec.automountServiceAccountToken == true": cis_5_1_6_pod_fail,
    "5.1.6 Pod.spec.automountServiceAccountToken == false": cis_5_1_6_pod_pass,
    "5.1.6 ServiceAccount.automountServiceAccountToken == true": cis_5_1_6_sa_pass,
    "5.1.6 ServiceAccount.automountServiceAccountToken == false": cis_5_1_6_sa_fail,
}

cis_psp_5_2_2_pass = K8sTestCase(
    rule_tag=K8S_CIS_5_2_2,
    resource_name=TEST_PASS_POD,
    expected=RULE_PASS_STATUS,
)

cis_psp_5_2_2_fail = K8sTestCase(
    rule_tag=K8S_CIS_5_2_2,
    resource_name=TEST_FAIL_POD,
    expected=RULE_FAIL_STATUS,
)

cis_psp_5_2_2 = {
    "5.2.2 PSP spec.securityContext.privileged==false eval passed": cis_psp_5_2_2_pass,
    "5.2.2 PSP spec.securityContext.privileged==true eval failed": cis_psp_5_2_2_fail,
}

cis_psp_5_2_3_pass = K8sTestCase(
    rule_tag=K8S_CIS_5_2_3,
    resource_name=TEST_PASS_POD,
    expected=RULE_PASS_STATUS,
)

cis_psp_5_2_3_fail = K8sTestCase(
    rule_tag=K8S_CIS_5_2_3,
    resource_name=TEST_FAIL_POD,
    expected=RULE_FAIL_STATUS,
)

cis_psp_5_2_3 = {
    "5.2.3 PSP Pod.spec.hostPID == true eval passed": cis_psp_5_2_3_pass,
    "5.2.3 PSP Pod.spec.hostPID == false eval failed": cis_psp_5_2_3_fail,
}

cis_psp_5_2_4_pass = K8sTestCase(
    rule_tag=K8S_CIS_5_2_4,
    resource_name=TEST_PASS_POD,
    expected=RULE_PASS_STATUS,
)

cis_psp_5_2_4_fail = K8sTestCase(
    rule_tag=K8S_CIS_5_2_4,
    resource_name=TEST_FAIL_POD,
    expected=RULE_FAIL_STATUS,
)

cis_psp_5_2_4 = {
    "5.2.4 PSP Pod.spec.hostIPC == true eval passed": cis_psp_5_2_4_pass,
    "5.2.4 PSP Pod.spec.hostIPC == false eval failed": cis_psp_5_2_4_fail,
}

cis_psp_5_2_5_pass = K8sTestCase(
    rule_tag=K8S_CIS_5_2_5,
    resource_name=TEST_PASS_POD,
    expected=RULE_PASS_STATUS,
)

cis_psp_5_2_5_fail = K8sTestCase(
    rule_tag=K8S_CIS_5_2_5,
    resource_name=TEST_FAIL_POD,
    expected=RULE_FAIL_STATUS,
)

cis_psp_5_2_5 = {
    "5.2.5 PSP Pod.spec.hostNetwork == true eval passed": cis_psp_5_2_5_pass,
    "5.2.5 PSP Pod.spec.hostNetwork == false eval failed": cis_psp_5_2_5_fail,
}

cis_psp_5_2_6_pass = K8sTestCase(
    rule_tag=K8S_CIS_5_2_6,
    resource_name=TEST_PASS_POD,
    expected=RULE_PASS_STATUS,
)

cis_psp_5_2_6_fail = K8sTestCase(
    rule_tag=K8S_CIS_5_2_6,
    resource_name=TEST_FAIL_POD,
    expected=RULE_FAIL_STATUS,
)

cis_psp_5_2_6 = {
    "5.2.6 PSP Pod.spec.containers.securityContext.allowPrivilegeEscalation == true eval passed": cis_psp_5_2_6_pass,
    "5.2.6 PSP Pod.spec.containers.securityContext.allowPrivilegeEscalation == true eval failed": cis_psp_5_2_6_fail,
}

cis_psp_5_2_7_pass = K8sTestCase(
    rule_tag=K8S_CIS_5_2_7,
    resource_name=TEST_PASS_POD,
    expected=RULE_PASS_STATUS,
)

cis_psp_5_2_7_fail = K8sTestCase(
    rule_tag=K8S_CIS_5_2_7,
    resource_name=TEST_FAIL_POD,
    expected=RULE_FAIL_STATUS,
)

cis_psp_5_2_7 = {
    "5.2.7 PSP Pod.spec.runAsUser forbids root eval passed": cis_psp_5_2_7_pass,
    "5.2.7 PSP Pod.spec.runAsUser allows root eval failed": cis_psp_5_2_7_fail,
}

cis_psp_5_2_8_pass = K8sTestCase(
    rule_tag=K8S_CIS_5_2_8,
    resource_name=TEST_PASS_POD,
    expected=RULE_PASS_STATUS,
)

cis_psp_5_2_8_fail = K8sTestCase(
    rule_tag=K8S_CIS_5_2_8,
    resource_name=TEST_FAIL_POD,
    expected=RULE_FAIL_STATUS,
)

cis_psp_5_2_8 = {
    "5.2.8 PSP Pod.container.spec.securityContext.capabilities drop all eval passed": cis_psp_5_2_8_pass,
    "5.2.8 PSP Pod.container.spec.securityContext.runAsUser == root eval failed": cis_psp_5_2_8_fail,
}

cis_psp_5_2_10_pass = K8sTestCase(
    rule_tag=K8S_CIS_5_2_10,
    resource_name=TEST_PASS_POD,
    expected=RULE_PASS_STATUS,
)

cis_psp_5_2_10_fail = K8sTestCase(
    rule_tag=K8S_CIS_5_2_10,
    resource_name=TEST_FAIL_POD,
    expected=RULE_FAIL_STATUS,
)

cis_psp_5_2_10 = {
    "5.2.10 PSP Pod.container.spec.securityContext.capabilities drop all eval passed": cis_psp_5_2_10_pass,
    "5.2.10 PSP Pod.container.spec.securityContext.capabilities assigned eval failed": cis_psp_5_2_10_fail,
}

# The name of this variable needs to be `tests_cases` in order to CIS Rules coverage stats to be generated
test_cases = {
    **cis_5_1_3,
    **cis_5_1_5,
    **cis_5_1_6,
    # **cis_psp_5_2_2,
    # **cis_psp_5_2_3,
    # **cis_psp_5_2_4,
    # **cis_psp_5_2_5,
    # **cis_psp_5_2_6,
    # **cis_psp_5_2_7,
    # **cis_psp_5_2_8,
    # **cis_psp_5_2_10,
}
