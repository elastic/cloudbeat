"""
This module provides K8s file rule test cases.
Cases are organized as rules.
Each rule has one or more test cases.
File rule identification is performed by node host and file names.
"""

from ..k8s_test_case import FileTestCase
from ..constants import RULE_PASS_STATUS, RULE_FAIL_STATUS

CIS_1_1_1 = "CIS 1.1.1"
CIS_1_1_2 = "CIS 1.1.2"
CIS_1_1_3 = "CIS 1.1.3"
CIS_1_1_4 = "CIS 1.1.4"
CIS_1_1_5 = "CIS 1.1.5"
CIS_1_1_6 = "CIS 1.1.6"
CIS_1_1_7 = "CIS 1.1.7"
CIS_1_1_8 = "CIS 1.1.8"
CIS_1_1_11 = "CIS 1.1.11"

KUBE_API_SERVER = "/hostfs/etc/kubernetes/manifests/kube-apiserver.yaml"
CONTROLLER_MANAGER = "/hostfs/etc/kubernetes/manifests/kube-controller-manager.yaml"
KUBE_SCHEDULER = "/hostfs/etc/kubernetes/manifests/kube-scheduler.yaml"
ETCD = "/hostfs/etc/kubernetes/manifests/etcd.yaml"
ETCD_DATA_DIR = "/hostfs/var/lib/etcd"

cis_file_1_1_1_pass = FileTestCase(
    rule_tag=CIS_1_1_1,
    node_hostname="kind-test-file-control-plane2",
    resource_name=KUBE_API_SERVER,
    expected=RULE_PASS_STATUS,
)

cis_file_1_1_1_fail = FileTestCase(
    rule_tag=CIS_1_1_1,
    node_hostname="kind-test-file-control-plane",
    resource_name=KUBE_API_SERVER,
    expected=RULE_FAIL_STATUS,
)

cis_k8s_file_1_1_1 = {
    "1.1.1 Ensure API server pod file permissions are set to 644 expected: passed": cis_file_1_1_1_pass,
    "1.1.1 Ensure API server pod file permissions are set to 700 expected: failed": cis_file_1_1_1_fail,
}

cis_file_1_1_2_pass = FileTestCase(
    rule_tag=CIS_1_1_2,
    node_hostname="kind-test-file-control-plane2",
    resource_name=KUBE_API_SERVER,
    expected=RULE_PASS_STATUS,
)

cis_file_1_1_2_fail = FileTestCase(
    rule_tag=CIS_1_1_2,
    node_hostname="kind-test-file-control-plane",
    resource_name=KUBE_API_SERVER,
    expected=RULE_FAIL_STATUS,
)

cis_k8s_file_1_1_2 = {
    "1.1.2 Ensure API server pod file ownership is set to root:root: passed": cis_file_1_1_2_pass,
    "1.1.2 Ensure API server pod file ownership is set to daemon:daemon: failed": cis_file_1_1_2_fail,
}

cis_file_1_1_3_pass = FileTestCase(
    rule_tag=CIS_1_1_3,
    node_hostname="kind-test-file-control-plane2",
    resource_name=CONTROLLER_MANAGER,
    expected=RULE_PASS_STATUS,
)

cis_file_1_1_3_fail = FileTestCase(
    rule_tag=CIS_1_1_3,
    node_hostname="kind-test-file-control-plane",
    resource_name=CONTROLLER_MANAGER,
    expected=RULE_FAIL_STATUS,
)

cis_k8s_file_1_1_3 = {
    "1.1.3 Ensure controller manager pod file permissions are set to 644 expected: passed": cis_file_1_1_3_pass,
    "1.1.3 Ensure controller manager pod file permissions are set to 700 expected: failed": cis_file_1_1_3_fail,
}

cis_file_1_1_4_pass = FileTestCase(
    rule_tag=CIS_1_1_4,
    node_hostname="kind-test-file-control-plane2",
    resource_name=CONTROLLER_MANAGER,
    expected=RULE_PASS_STATUS,
)

cis_file_1_1_4_fail = FileTestCase(
    rule_tag=CIS_1_1_4,
    node_hostname="kind-test-file-control-plane",
    resource_name=CONTROLLER_MANAGER,
    expected=RULE_FAIL_STATUS,
)

cis_k8s_file_1_1_4 = {
    "1.1.4 Ensure controller manager pod file ownership is set to root:root: passed": cis_file_1_1_4_pass,
    "1.1.4 Ensure controller manager pod file ownership is set to root:daemon: failed": cis_file_1_1_4_fail,
}

cis_file_1_1_5_pass = FileTestCase(
    rule_tag=CIS_1_1_5,
    node_hostname="kind-test-file-control-plane2",
    resource_name=KUBE_SCHEDULER,
    expected=RULE_PASS_STATUS,
)

cis_file_1_1_5_fail = FileTestCase(
    rule_tag=CIS_1_1_5,
    node_hostname="kind-test-file-control-plane",
    resource_name=KUBE_SCHEDULER,
    expected=RULE_FAIL_STATUS,
)

cis_k8s_file_1_1_5 = {
    "1.1.5 Ensure scheduler pod file permissions are set to 644 expected: passed": cis_file_1_1_5_pass,
    "1.1.5 Ensure scheduler pod file permissions are set to 700 expected: failed": cis_file_1_1_5_fail,
}

cis_file_1_1_6_pass = FileTestCase(
    rule_tag=CIS_1_1_6,
    node_hostname="kind-test-file-control-plane2",
    resource_name=KUBE_SCHEDULER,
    expected=RULE_PASS_STATUS,
)

cis_file_1_1_6_fail = FileTestCase(
    rule_tag=CIS_1_1_6,
    node_hostname="kind-test-file-control-plane",
    resource_name=KUBE_SCHEDULER,
    expected=RULE_FAIL_STATUS,
)

cis_k8s_file_1_1_6 = {
    "1.1.6 Ensure scheduler pod file ownership is set to root:root: passed": cis_file_1_1_6_pass,
    "1.1.6 Ensure scheduler pod file ownership is set to root:daemon: failed": cis_file_1_1_6_fail,
}

cis_file_1_1_7_pass = FileTestCase(
    rule_tag=CIS_1_1_7,
    node_hostname="kind-test-file-control-plane2",
    resource_name=ETCD,
    expected=RULE_PASS_STATUS,
)

cis_file_1_1_7_fail = FileTestCase(
    rule_tag=CIS_1_1_7,
    node_hostname="kind-test-file-control-plane",
    resource_name=ETCD,
    expected=RULE_FAIL_STATUS,
)

cis_k8s_file_1_1_7 = {
    "1.1.7 Ensure etcd pod file permissions are set to 700 expected: passed": cis_file_1_1_7_pass,
    "1.1.7 Ensure etcd pod file permissions are set to 600 expected: failed": cis_file_1_1_7_fail,
}

cis_file_1_1_8_pass = FileTestCase(
    rule_tag=CIS_1_1_8,
    node_hostname="kind-test-file-control-plane2",
    resource_name=ETCD,
    expected=RULE_PASS_STATUS,
)

cis_file_1_1_8_fail = FileTestCase(
    rule_tag=CIS_1_1_8,
    node_hostname="kind-test-file-control-plane",
    resource_name=ETCD,
    expected=RULE_FAIL_STATUS,
)

cis_k8s_file_1_1_8 = {
    "1.1.8 Ensure etcd pod file ownership is set to root:root: passed": cis_file_1_1_8_pass,
    "1.1.8 Ensure etcd pod file ownership is set to root:daemon: failed": cis_file_1_1_8_fail,
}

cis_file_1_1_11_pass = FileTestCase(
    rule_tag=CIS_1_1_1,
    node_hostname="kind-test-file-control-plane2",
    resource_name=ETCD_DATA_DIR,
    expected=RULE_PASS_STATUS,
)

cis_file_1_1_11_fail = FileTestCase(
    rule_tag=CIS_1_1_11,
    node_hostname="kind-test-file-control-plane",
    resource_name=ETCD_DATA_DIR,
    expected=RULE_FAIL_STATUS,
)

cis_k8s_file_1_1_11 = {
    "1.1.11 Ensure etcd data dir permissions are set to 700 expected: passed": cis_file_1_1_11_pass,
    "1.1.11 Ensure etcd data dir permissions are set to 710 expected: failed": cis_file_1_1_11_fail,
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
}
