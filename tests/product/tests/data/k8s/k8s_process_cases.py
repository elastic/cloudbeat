"""
This module defines k8s process test cases
Kind configuration for k8s process failed cases is defined: deploy/k8s/kind/kind-test-proc-conf1.yml
Kind configuration for k8s process passed cases is defined: deploy/k8s/kind/kind-test-proc-conf2.yml
To add new test cases, create a new configuration file and add it to the mapping or update the existing one.
"""

from configuration import kubernetes
from .k8s_test_case import K8sTestCase
from ..constants import RULE_PASS_STATUS, RULE_FAIL_STATUS

K8S_CIS_1_3_2 = "CIS 1.3.2"
K8S_CIS_1_3_3 = "CIS 1.3.3"
K8S_CIS_1_3_4 = "CIS 1.3.4"
K8S_CIS_1_3_5 = "CIS 1.3.5"
K8S_CIS_1_3_6 = "CIS 1.3.6"
K8S_CIS_1_3_7 = "CIS 1.3.7"
K8S_CIS_1_4_1 = "CIS 1.4.1"
K8S_CIS_1_4_2 = "CIS 1.4.2"
K8S_CIS_2_1 = "CIS 2.1"
K8S_CIS_2_2 = "CIS 2.2"
K8S_CIS_2_3 = "CIS 2.3"
K8S_CIS_2_4 = "CIS 2.4"
K8S_CIS_2_5 = "CIS 2.5"
K8S_CIS_2_6 = "CIS 2.6"

KUBE_SCHEDULER = "kube-scheduler"
ETCD = "etcd"

cis_1_4_1_fail = K8sTestCase(
    rule_tag=K8S_CIS_1_4_1,
    resource_name=KUBE_SCHEDULER,
    expected=RULE_FAIL_STATUS,
)

cis_1_4_1_pass = K8sTestCase(
    rule_tag=K8S_CIS_1_4_1,
    resource_name=KUBE_SCHEDULER,
    expected=RULE_PASS_STATUS,
)

cis_1_4_2_fail = K8sTestCase(
    rule_tag=K8S_CIS_1_4_2,
    resource_name=KUBE_SCHEDULER,
    expected=RULE_FAIL_STATUS,
)

cis_1_4_2_pass = K8sTestCase(
    rule_tag=K8S_CIS_1_4_2,
    resource_name=KUBE_SCHEDULER,
    expected=RULE_PASS_STATUS,
)

cis_2_1_pass = K8sTestCase(
    rule_tag=K8S_CIS_2_1,
    resource_name=ETCD,
    expected=RULE_PASS_STATUS,
)

cis_2_2_fail = K8sTestCase(
    rule_tag=K8S_CIS_2_2,
    resource_name=ETCD,
    expected=RULE_FAIL_STATUS,
)

cis_2_2_pass = K8sTestCase(
    rule_tag=K8S_CIS_2_2,
    resource_name=ETCD,
    expected=RULE_PASS_STATUS,
)

cis_2_3_fail = K8sTestCase(
    rule_tag=K8S_CIS_2_3,
    resource_name=ETCD,
    expected=RULE_FAIL_STATUS,
)

cis_2_3_pass = K8sTestCase(
    rule_tag=K8S_CIS_2_3,
    resource_name=ETCD,
    expected=RULE_PASS_STATUS,
)

cis_2_4_pass = K8sTestCase(
    rule_tag=K8S_CIS_2_4,
    resource_name=ETCD,
    expected=RULE_PASS_STATUS,
)

cis_2_5_fail = K8sTestCase(
    rule_tag=K8S_CIS_2_5,
    resource_name=ETCD,
    expected=RULE_FAIL_STATUS,
)


cis_2_5_pass = K8sTestCase(
    rule_tag=K8S_CIS_2_5,
    resource_name=ETCD,
    expected=RULE_PASS_STATUS,
)

cis_2_6_fail = K8sTestCase(
    rule_tag=K8S_CIS_2_6,
    resource_name=ETCD,
    expected=RULE_FAIL_STATUS,
)


cis_2_6_pass = K8sTestCase(
    rule_tag=K8S_CIS_2_6,
    resource_name=ETCD,
    expected=RULE_PASS_STATUS,
)

k8s_process_config_1 = {
    "1.4.1 kube-scheduler --profiling=true": cis_1_4_1_fail,
    "1.4.2 kube-scheduler --bind-address=0.0.0.0": cis_1_4_2_fail,
    "2.2 etcd --client-cert-auth=false": cis_2_2_fail,
    "2.3 etcd --auto-tls=true": cis_2_3_fail,
    "2.5 etcd --peer-client-cert-auth=false": cis_2_5_fail,
    "2.6 etcd --peer-auto-tls=true": cis_2_6_fail,
}

k8s_process_config_2 = {
    "1.4.1 kube-scheduler --profiling=false": cis_1_4_1_pass,
    "1.4.2 kube-scheduler --bind-address=127.0.0.1": cis_1_4_2_pass,
    "2.1 etcd --cert-file and --key-file are set": cis_2_1_pass,
    "2.2 etcd --client-cert-auth=true": cis_2_2_pass,
    "2.3 etcd --auto-tls=false": cis_2_3_pass,
    "2.4 etcd --peer-cert-file and --peer-key-file are set": cis_2_4_pass,
    "2.5 etcd --peer-client-cert-auth=true": cis_2_5_pass,
    "2.6 etcd --peer-auto-tls=false": cis_2_6_pass,
}

cis_k8s_process_all = {
    "test-k8s-config-1": k8s_process_config_1,
    "test-k8s-config-2": k8s_process_config_2,
}

# The name of this variable needs to be `tests_cases` in order to CIS Rules coverage stats to be generated
test_cases = cis_k8s_process_all[kubernetes.current_config]
