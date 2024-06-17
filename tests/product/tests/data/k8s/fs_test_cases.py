"""
This module provides K8s file rule test cases.
Cases are organized as rules.
Each rule has one or more test cases.
File rule identification is performed by node host and file names.
"""

from ..constants import RULE_FAIL_STATUS, RULE_PASS_STATUS
from .k8s_test_case import FileTestCase

CIS_1_1_1 = "CIS 1.1.1"
CIS_1_1_2 = "CIS 1.1.2"
CIS_1_1_3 = "CIS 1.1.3"
CIS_1_1_4 = "CIS 1.1.4"
CIS_1_1_5 = "CIS 1.1.5"
CIS_1_1_6 = "CIS 1.1.6"
CIS_1_1_7 = "CIS 1.1.7"
CIS_1_1_8 = "CIS 1.1.8"
CIS_1_1_11 = "CIS 1.1.11"
CIS_1_1_12 = "CIS 1.1.12"
CIS_1_1_13 = "CIS 1.1.13"
CIS_1_1_14 = "CIS 1.1.14"
CIS_1_1_15 = "CIS 1.1.15"
CIS_1_1_16 = "CIS 1.1.16"
CIS_1_1_17 = "CIS 1.1.17"
CIS_1_1_18 = "CIS 1.1.18"
CIS_1_1_19 = "CIS 1.1.19"
CIS_1_1_20 = "CIS 1.1.20"
CIS_1_1_21 = "CIS 1.1.21"
CIS_4_1_1 = "CIS 4.1.1"
CIS_4_1_2 = "CIS 4.1.2"
CIS_4_1_5 = "CIS 4.1.5"
CIS_4_1_6 = "CIS 4.1.6"
CIS_4_1_9 = "CIS 4.1.9"
CIS_4_1_10 = "CIS 4.1.10"

KUBE_API_SERVER = "/hostfs/etc/kubernetes/manifests/kube-apiserver.yaml"
CONTROLLER_MANAGER = "/hostfs/etc/kubernetes/manifests/kube-controller-manager.yaml"
KUBE_SCHEDULER = "/hostfs/etc/kubernetes/manifests/kube-scheduler.yaml"
ETCD = "/hostfs/etc/kubernetes/manifests/etcd.yaml"
ETCD_DATA_DIR = "/hostfs/var/lib/etcd"
ADMIN_CONF = "/hostfs/etc/kubernetes/admin.conf"
SCHEDULER_CONF = "/hostfs/etc/kubernetes/scheduler.conf"
CONTROLLER_MANAGER_CONF = "/hostfs/etc/kubernetes/controller-manager.conf"
PKI_DIR = "/hostfs/etc/kubernetes/pki"
API_SERVER_CERT = "/hostfs/etc/kubernetes/pki/apiserver.crt"
API_SERVER_KEY = "/hostfs/etc/kubernetes/pki/apiserver.key"
KUBELET_SERVICE = "/hostfs/etc/systemd/system/kubelet.service.d/10-kubeadm.conf"
KUBELET_CONF = "/hostfs/etc/kubernetes/kubelet.conf"
KUBELET_CONFIG = "/hostfs/var/lib/kubelet/config.yaml"

NODE_NAME_1 = "kind-test-files-control-plane"
NODE_NAME_2 = "kind-test-files-control-plane2"

cis_file_1_1_1_pass = FileTestCase(
    rule_tag=CIS_1_1_1,
    node_hostname=NODE_NAME_2,
    resource_name=KUBE_API_SERVER,
    expected=RULE_PASS_STATUS,
)

cis_file_1_1_1_fail = FileTestCase(
    rule_tag=CIS_1_1_1,
    node_hostname=NODE_NAME_1,
    resource_name=KUBE_API_SERVER,
    expected=RULE_FAIL_STATUS,
)

cis_k8s_file_1_1_1 = {
    "1.1.1 Ensure API server pod file permissions are set to 644 expected: passed": cis_file_1_1_1_pass,
    "1.1.1 Ensure API server pod file permissions are set to 700 expected: failed": cis_file_1_1_1_fail,
}

cis_file_1_1_2_pass = FileTestCase(
    rule_tag=CIS_1_1_2,
    node_hostname=NODE_NAME_2,
    resource_name=KUBE_API_SERVER,
    expected=RULE_PASS_STATUS,
)

cis_file_1_1_2_fail = FileTestCase(
    rule_tag=CIS_1_1_2,
    node_hostname=NODE_NAME_1,
    resource_name=KUBE_API_SERVER,
    expected=RULE_FAIL_STATUS,
)

cis_k8s_file_1_1_2 = {
    "1.1.2 Ensure API server pod file ownership is set to root:root: passed": cis_file_1_1_2_pass,
    "1.1.2 Ensure API server pod file ownership is set to daemon:daemon: failed": cis_file_1_1_2_fail,
}

cis_file_1_1_3_pass = FileTestCase(
    rule_tag=CIS_1_1_3,
    node_hostname=NODE_NAME_2,
    resource_name=CONTROLLER_MANAGER,
    expected=RULE_PASS_STATUS,
)

cis_file_1_1_3_fail = FileTestCase(
    rule_tag=CIS_1_1_3,
    node_hostname=NODE_NAME_1,
    resource_name=CONTROLLER_MANAGER,
    expected=RULE_FAIL_STATUS,
)

cis_k8s_file_1_1_3 = {
    "1.1.3 Ensure controller manager pod file permissions are set to 644 expected: passed": cis_file_1_1_3_pass,
    "1.1.3 Ensure controller manager pod file permissions are set to 700 expected: failed": cis_file_1_1_3_fail,
}

cis_file_1_1_4_pass = FileTestCase(
    rule_tag=CIS_1_1_4,
    node_hostname=NODE_NAME_2,
    resource_name=CONTROLLER_MANAGER,
    expected=RULE_PASS_STATUS,
)

cis_file_1_1_4_fail = FileTestCase(
    rule_tag=CIS_1_1_4,
    node_hostname=NODE_NAME_1,
    resource_name=CONTROLLER_MANAGER,
    expected=RULE_FAIL_STATUS,
)

cis_k8s_file_1_1_4 = {
    "1.1.4 Ensure controller manager pod file ownership is set to root:root: passed": cis_file_1_1_4_pass,
    "1.1.4 Ensure controller manager pod file ownership is set to root:daemon: failed": cis_file_1_1_4_fail,
}

cis_file_1_1_5_pass = FileTestCase(
    rule_tag=CIS_1_1_5,
    node_hostname=NODE_NAME_2,
    resource_name=KUBE_SCHEDULER,
    expected=RULE_PASS_STATUS,
)

cis_file_1_1_5_fail = FileTestCase(
    rule_tag=CIS_1_1_5,
    node_hostname=NODE_NAME_1,
    resource_name=KUBE_SCHEDULER,
    expected=RULE_FAIL_STATUS,
)

cis_k8s_file_1_1_5 = {
    "1.1.5 Ensure scheduler pod file permissions are set to 644 expected: passed": cis_file_1_1_5_pass,
    "1.1.5 Ensure scheduler pod file permissions are set to 700 expected: failed": cis_file_1_1_5_fail,
}

cis_file_1_1_6_pass = FileTestCase(
    rule_tag=CIS_1_1_6,
    node_hostname=NODE_NAME_2,
    resource_name=KUBE_SCHEDULER,
    expected=RULE_PASS_STATUS,
)

cis_file_1_1_6_fail = FileTestCase(
    rule_tag=CIS_1_1_6,
    node_hostname=NODE_NAME_1,
    resource_name=KUBE_SCHEDULER,
    expected=RULE_FAIL_STATUS,
)

cis_k8s_file_1_1_6 = {
    "1.1.6 Ensure scheduler pod file ownership is set to root:root: passed": cis_file_1_1_6_pass,
    "1.1.6 Ensure scheduler pod file ownership is set to root:daemon: failed": cis_file_1_1_6_fail,
}

cis_file_1_1_7_pass = FileTestCase(
    rule_tag=CIS_1_1_7,
    node_hostname=NODE_NAME_2,
    resource_name=ETCD,
    expected=RULE_PASS_STATUS,
)

cis_file_1_1_7_fail = FileTestCase(
    rule_tag=CIS_1_1_7,
    node_hostname=NODE_NAME_1,
    resource_name=ETCD,
    expected=RULE_FAIL_STATUS,
)

cis_k8s_file_1_1_7 = {
    "1.1.7 Ensure etcd pod file permissions are set to 700 expected: passed": cis_file_1_1_7_pass,
    "1.1.7 Ensure etcd pod file permissions are set to 600 expected: failed": cis_file_1_1_7_fail,
}

cis_file_1_1_8_pass = FileTestCase(
    rule_tag=CIS_1_1_8,
    node_hostname=NODE_NAME_2,
    resource_name=ETCD,
    expected=RULE_PASS_STATUS,
)

cis_file_1_1_8_fail = FileTestCase(
    rule_tag=CIS_1_1_8,
    node_hostname=NODE_NAME_1,
    resource_name=ETCD,
    expected=RULE_FAIL_STATUS,
)

cis_k8s_file_1_1_8 = {
    "1.1.8 Ensure etcd pod file ownership is set to root:root: passed": cis_file_1_1_8_pass,
    "1.1.8 Ensure etcd pod file ownership is set to root:daemon: failed": cis_file_1_1_8_fail,
}

cis_file_1_1_11_pass = FileTestCase(
    rule_tag=CIS_1_1_1,
    node_hostname=NODE_NAME_2,
    resource_name=ETCD_DATA_DIR,
    expected=RULE_PASS_STATUS,
)

cis_file_1_1_11_fail = FileTestCase(
    rule_tag=CIS_1_1_11,
    node_hostname=NODE_NAME_1,
    resource_name=ETCD_DATA_DIR,
    expected=RULE_FAIL_STATUS,
)

cis_k8s_file_1_1_11 = {
    # TODO: check why this test fails
    # "1.1.11 Ensure etcd data dir permissions are set to 700 expected: passed": cis_file_1_1_11_pass,
    "1.1.11 Ensure etcd data dir permissions are set to 777 expected: failed": cis_file_1_1_11_fail,
}

cis_file_1_1_12_pass = FileTestCase(
    rule_tag=CIS_1_1_12,
    node_hostname=NODE_NAME_2,
    resource_name=ETCD_DATA_DIR,
    expected=RULE_PASS_STATUS,
)

cis_file_1_1_12_fail = FileTestCase(
    rule_tag=CIS_1_1_12,
    node_hostname=NODE_NAME_1,
    resource_name=ETCD_DATA_DIR,
    expected=RULE_FAIL_STATUS,
)

cis_k8s_file_1_1_12 = {
    # TODO: fix etcd configuration tests/test_environments/k8s-cloudbeat-tests/templates/_k8s-file-permission-job.yaml
    # "1.1.12 Ensure etcd pod file ownership is set to etcd:etcd: passed": cis_file_1_1_12_pass,
    "1.1.12 Ensure etcd pod file ownership is set to root:root: failed": cis_file_1_1_12_fail,
}

cis_file_1_1_13_pass = FileTestCase(
    rule_tag=CIS_1_1_13,
    node_hostname=NODE_NAME_2,
    resource_name=ADMIN_CONF,
    expected=RULE_PASS_STATUS,
)

cis_file_1_1_13_fail = FileTestCase(
    rule_tag=CIS_1_1_13,
    node_hostname=NODE_NAME_1,
    resource_name=ADMIN_CONF,
    expected=RULE_FAIL_STATUS,
)

cis_k8s_file_1_1_13 = {
    "1.1.13 Ensure admin.conf file permissions are set to 600 expected: passed": cis_file_1_1_13_pass,
    "1.1.13 Ensure admin.conf file permissions are set to 700 expected: failed": cis_file_1_1_13_fail,
}

cis_file_1_1_14_pass = FileTestCase(
    rule_tag=CIS_1_1_14,
    node_hostname=NODE_NAME_2,
    resource_name=ADMIN_CONF,
    expected=RULE_PASS_STATUS,
)

cis_file_1_1_14_fail = FileTestCase(
    rule_tag=CIS_1_1_14,
    node_hostname=NODE_NAME_1,
    resource_name=ADMIN_CONF,
    expected=RULE_FAIL_STATUS,
)

cis_k8s_file_1_1_14 = {
    "1.1.14 Ensure admin.conf file ownership is set to root:root: passed": cis_file_1_1_14_pass,
    "1.1.14 Ensure admin.conf file ownership is set to daemon:root: failed": cis_file_1_1_14_fail,
}

cis_file_1_1_15_pass = FileTestCase(
    rule_tag=CIS_1_1_15,
    node_hostname=NODE_NAME_2,
    resource_name=SCHEDULER_CONF,
    expected=RULE_PASS_STATUS,
)

cis_file_1_1_15_fail = FileTestCase(
    rule_tag=CIS_1_1_15,
    node_hostname=NODE_NAME_1,
    resource_name=SCHEDULER_CONF,
    expected=RULE_FAIL_STATUS,
)

cis_k8s_file_1_1_15 = {
    "1.1.15 Ensure scheduler.conf file permissions are set to 644 expected: passed": cis_file_1_1_15_pass,
    "1.1.15 Ensure scheduler.conf file permissions are set to 700 expected: failed": cis_file_1_1_15_fail,
}

cis_file_1_1_16_pass = FileTestCase(
    rule_tag=CIS_1_1_16,
    node_hostname=NODE_NAME_2,
    resource_name=SCHEDULER_CONF,
    expected=RULE_PASS_STATUS,
)

cis_file_1_1_16_fail = FileTestCase(
    rule_tag=CIS_1_1_16,
    node_hostname=NODE_NAME_1,
    resource_name=SCHEDULER_CONF,
    expected=RULE_FAIL_STATUS,
)

cis_k8s_file_1_1_16 = {
    "1.1.16 Ensure scheduler.conf file ownership is set to root:root: passed": cis_file_1_1_16_pass,
    "1.1.16 Ensure scheduler.conf file ownership is set to daemon:root: failed": cis_file_1_1_16_fail,
}

cis_file_1_1_17_pass = FileTestCase(
    rule_tag=CIS_1_1_17,
    node_hostname=NODE_NAME_2,
    resource_name=CONTROLLER_MANAGER_CONF,
    expected=RULE_PASS_STATUS,
)

cis_file_1_1_17_fail = FileTestCase(
    rule_tag=CIS_1_1_17,
    node_hostname=NODE_NAME_1,
    resource_name=CONTROLLER_MANAGER_CONF,
    expected=RULE_FAIL_STATUS,
)

cis_k8s_file_1_1_17 = {
    "1.1.17 Ensure controller-manager.conf file permissions are set to 644 expected: passed": cis_file_1_1_17_pass,
    "1.1.17 Ensure controller-manager.conf file permissions are set to 700 expected: failed": cis_file_1_1_17_fail,
}

cis_file_1_1_18_pass = FileTestCase(
    rule_tag=CIS_1_1_18,
    node_hostname=NODE_NAME_2,
    resource_name=CONTROLLER_MANAGER_CONF,
    expected=RULE_PASS_STATUS,
)

cis_file_1_1_18_fail = FileTestCase(
    rule_tag=CIS_1_1_18,
    node_hostname=NODE_NAME_1,
    resource_name=CONTROLLER_MANAGER_CONF,
    expected=RULE_FAIL_STATUS,
)

cis_k8s_file_1_1_18 = {
    "1.1.18 Ensure controller-manager.conf file ownership is set to root:root: passed": cis_file_1_1_18_pass,
    "1.1.18 Ensure controller-manager.conf file ownership is set to root:daemon: failed": cis_file_1_1_18_fail,
}

cis_file_1_1_19_pass = FileTestCase(
    rule_tag=CIS_1_1_19,
    node_hostname=NODE_NAME_2,
    resource_name=PKI_DIR,
    expected=RULE_PASS_STATUS,
)

cis_file_1_1_19_fail = FileTestCase(
    rule_tag=CIS_1_1_19,
    node_hostname=NODE_NAME_1,
    resource_name=PKI_DIR,
    expected=RULE_FAIL_STATUS,
)

cis_k8s_file_1_1_19 = {
    "1.1.19 Ensure Kubernetes PKI dir ownership is set to root:root: passed": cis_file_1_1_19_pass,
    "1.1.19 Ensure Kubernetes PKI dir ownership is set to root:daemon: failed": cis_file_1_1_19_fail,
}

cis_file_1_1_20_pass = FileTestCase(
    rule_tag=CIS_1_1_20,
    node_hostname=NODE_NAME_2,
    resource_name=API_SERVER_CERT,
    expected=RULE_PASS_STATUS,
)

cis_file_1_1_20_fail = FileTestCase(
    rule_tag=CIS_1_1_20,
    node_hostname=NODE_NAME_1,
    resource_name=API_SERVER_CERT,
    expected=RULE_FAIL_STATUS,
)

cis_k8s_file_1_1_20 = {
    "1.1.20 Ensure PKI dir/*.crt file permissions are set to 644 expected: passed": cis_file_1_1_20_pass,
    "1.1.20 Ensure PKI dir/*.crt file permissions are set to 666 expected: failed": cis_file_1_1_20_fail,
}

cis_file_1_1_21_pass = FileTestCase(
    rule_tag=CIS_1_1_21,
    node_hostname=NODE_NAME_2,
    resource_name=API_SERVER_KEY,
    expected=RULE_PASS_STATUS,
)

cis_file_1_1_21_fail = FileTestCase(
    rule_tag=CIS_1_1_21,
    node_hostname=NODE_NAME_1,
    resource_name=API_SERVER_KEY,
    expected=RULE_FAIL_STATUS,
)

cis_k8s_file_1_1_21 = {
    "1.1.21 Ensure PKI dir/*.key file permissions are set to 600 expected: passed": cis_file_1_1_21_pass,
    "1.1.21 Ensure PKI dir/*.key file permissions are set to 644 expected: failed": cis_file_1_1_21_fail,
}

cis_file_4_1_1_pass = FileTestCase(
    rule_tag=CIS_4_1_1,
    node_hostname=NODE_NAME_2,
    resource_name=KUBELET_SERVICE,
    expected=RULE_PASS_STATUS,
)

cis_file_4_1_1_fail = FileTestCase(
    rule_tag=CIS_4_1_1,
    node_hostname=NODE_NAME_1,
    resource_name=KUBELET_SERVICE,
    expected=RULE_FAIL_STATUS,
)

cis_k8s_file_4_1_1 = {
    "4.1.1 Ensure kubelet service file permissions are set to 644 expected: passed": cis_file_4_1_1_pass,
    "4.1.1 Ensure kubelet service file permissions are set to 700 expected: failed": cis_file_4_1_1_fail,
}

cis_file_4_1_2_pass = FileTestCase(
    rule_tag=CIS_4_1_2,
    node_hostname=NODE_NAME_2,
    resource_name=KUBELET_SERVICE,
    expected=RULE_PASS_STATUS,
)

cis_file_4_1_2_fail = FileTestCase(
    rule_tag=CIS_4_1_2,
    node_hostname=NODE_NAME_1,
    resource_name=KUBELET_SERVICE,
    expected=RULE_FAIL_STATUS,
)

cis_k8s_file_4_1_2 = {
    "4.1.2 Ensure kubelet service file ownership is set to root:root: passed": cis_file_4_1_2_pass,
    "4.1.2 Ensure kubelet service file ownership is set to root:daemon: failed": cis_file_4_1_2_fail,
}

cis_file_4_1_5_pass = FileTestCase(
    rule_tag=CIS_4_1_5,
    node_hostname=NODE_NAME_2,
    resource_name=KUBELET_CONF,
    expected=RULE_PASS_STATUS,
)

cis_file_4_1_5_fail = FileTestCase(
    rule_tag=CIS_4_1_5,
    node_hostname=NODE_NAME_1,
    resource_name=KUBELET_CONF,
    expected=RULE_FAIL_STATUS,
)

cis_k8s_file_4_1_5 = {
    "4.1.5 Ensure kubelet conf file permissions are set to 644 expected: passed": cis_file_4_1_5_pass,
    "4.1.5 Ensure kubelet conf file permissions are set to 700 expected: failed": cis_file_4_1_5_fail,
}

cis_file_4_1_6_pass = FileTestCase(
    rule_tag=CIS_4_1_6,
    node_hostname=NODE_NAME_2,
    resource_name=KUBELET_CONF,
    expected=RULE_PASS_STATUS,
)

cis_file_4_1_6_fail = FileTestCase(
    rule_tag=CIS_4_1_6,
    node_hostname=NODE_NAME_1,
    resource_name=KUBELET_CONF,
    expected=RULE_FAIL_STATUS,
)

cis_k8s_file_4_1_6 = {
    "4.1.6 Ensure kubelet conf file ownership is set to root:root: passed": cis_file_4_1_6_pass,
    "4.1.6 Ensure kubelet conf file ownership is set to daemon:root: failed": cis_file_4_1_6_fail,
}

cis_file_4_1_9_pass = FileTestCase(
    rule_tag=CIS_4_1_9,
    node_hostname=NODE_NAME_2,
    resource_name=KUBELET_CONFIG,
    expected=RULE_PASS_STATUS,
)

cis_file_4_1_9_fail = FileTestCase(
    rule_tag=CIS_4_1_9,
    node_hostname=NODE_NAME_1,
    resource_name=KUBELET_CONFIG,
    expected=RULE_FAIL_STATUS,
)

cis_k8s_file_4_1_9 = {
    "4.1.9 Ensure kubelet --config file permissions are set to 644 expected: passed": cis_file_4_1_9_pass,
    "4.1.9 Ensure kubelet --config file permissions are set to 700 expected: failed": cis_file_4_1_9_fail,
}

cis_file_4_1_10_pass = FileTestCase(
    rule_tag=CIS_4_1_10,
    node_hostname=NODE_NAME_2,
    resource_name=KUBELET_CONF,
    expected=RULE_PASS_STATUS,
)

cis_file_4_1_10_fail = FileTestCase(
    rule_tag=CIS_4_1_10,
    node_hostname=NODE_NAME_1,
    resource_name=KUBELET_CONF,
    expected=RULE_FAIL_STATUS,
)

cis_k8s_file_4_1_10 = {
    "4.1.10 Ensure kubelet conf file ownership is set to root:root: passed": cis_file_4_1_10_pass,
    "4.1.10 Ensure kubelet conf file ownership is set to daemon:root: failed": cis_file_4_1_10_fail,
}

# The name of this variable needs to be `tests_cases` in order to CIS Rules coverage stats to be generated
test_cases = {
    **cis_k8s_file_1_1_1,
    **cis_k8s_file_1_1_2,
    **cis_k8s_file_1_1_3,
    **cis_k8s_file_1_1_4,
    **cis_k8s_file_1_1_5,
    **cis_k8s_file_1_1_6,
    **cis_k8s_file_1_1_7,
    **cis_k8s_file_1_1_8,
    **cis_k8s_file_1_1_11,
    **cis_k8s_file_1_1_12,
    **cis_k8s_file_1_1_13,
    **cis_k8s_file_1_1_14,
    **cis_k8s_file_1_1_15,
    **cis_k8s_file_1_1_16,
    **cis_k8s_file_1_1_17,
    **cis_k8s_file_1_1_18,
    **cis_k8s_file_1_1_19,
    **cis_k8s_file_1_1_20,
    **cis_k8s_file_1_1_21,
    **cis_k8s_file_4_1_1,
    **cis_k8s_file_4_1_2,
    **cis_k8s_file_4_1_5,
    **cis_k8s_file_4_1_6,
    **cis_k8s_file_4_1_9,
    **cis_k8s_file_4_1_10,
}
