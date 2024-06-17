"""
This module defines k8s process test cases
Kind configuration for k8s process failed cases is defined: deploy/k8s/kind/kind-test-proc-conf1.yml
Kind configuration for k8s process passed cases is defined: deploy/k8s/kind/kind-test-proc-conf2.yml
To add new test cases, create a new configuration file and add it to the mapping or update the existing one.
"""

from configuration import kubernetes

from ..constants import RULE_FAIL_STATUS, RULE_PASS_STATUS
from .k8s_test_case import K8sTestCase

K8S_CIS_1_2_2 = "CIS 1.2.2"
K8S_CIS_1_2_3 = "CIS 1.2.3"
K8S_CIS_1_2_4 = "CIS 1.2.4"
K8S_CIS_1_2_5 = "CIS 1.2.5"
K8S_CIS_1_2_6 = "CIS 1.2.6"
K8S_CIS_1_2_7 = "CIS 1.2.7"
K8S_CIS_1_2_8 = "CIS 1.2.8"
K8S_CIS_1_2_9 = "CIS 1.2.9"
K8S_CIS_1_2_10 = "CIS 1.2.10"
K8S_CIS_1_2_11 = "CIS 1.2.11"
K8S_CIS_1_2_12 = "CIS 1.2.12"
K8S_CIS_1_2_13 = "CIS 1.2.13"
K8S_CIS_1_2_14 = "CIS 1.2.14"
K8S_CIS_1_2_15 = "CIS 1.2.15"
K8S_CIS_1_2_16 = "CIS 1.2.16"
K8S_CIS_1_2_17 = "CIS 1.2.17"
K8S_CIS_1_2_18 = "CIS 1.2.18"
K8S_CIS_1_2_19 = "CIS 1.2.19"
K8S_CIS_1_2_20 = "CIS 1.2.20"
K8S_CIS_1_2_21 = "CIS 1.2.21"
K8S_CIS_1_2_22 = "CIS 1.2.22"
K8S_CIS_1_2_23 = "CIS 1.2.23"
K8S_CIS_1_2_24 = "CIS 1.2.24"
K8S_CIS_1_2_25 = "CIS 1.2.25"
K8S_CIS_1_2_26 = "CIS 1.2.26"
K8S_CIS_1_2_27 = "CIS 1.2.27"
K8S_CIS_1_2_28 = "CIS 1.2.28"
K8S_CIS_1_2_29 = "CIS 1.2.29"
K8S_CIS_1_2_32 = "CIS 1.2.32"
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
K8S_CIS_4_2_1 = "CIS 4.2.1"
K8S_CIS_4_2_2 = "CIS 4.2.2"
K8S_CIS_4_2_3 = "CIS 4.2.3"
K8S_CIS_4_2_4 = "CIS 4.2.4"
K8S_CIS_4_2_5 = "CIS 4.2.5"
K8S_CIS_4_2_6 = "CIS 4.2.6"
K8S_CIS_4_2_7 = "CIS 4.2.7"
K8S_CIS_4_2_9 = "CIS 4.2.9"
K8S_CIS_4_2_10 = "CIS 4.2.10"
K8S_CIS_4_2_11 = "CIS 4.2.11"
K8S_CIS_4_2_12 = "CIS 4.2.12"
K8S_CIS_4_2_13 = "CIS 4.2.13"

KUBE_SCHEDULER = "kube-scheduler"
# /etc/kubernetes/manifests/kube-apiserver.yaml

ETCD = "etcd"
# /etc/kubernetes/manifests/etcd.yaml

KUBE_CONTROLLER = "kube-controller"
# /etc/kubernetes/manifests/kube-controller-manager.yaml

KUBELET = "kubelet"
# /var/lib/kubelet/config.yaml

KUBE_APISERVER = "kube-apiserver"
# /etc/kubernetes/manifests/kube-apiserver.yaml


cis_1_2_2_pass = K8sTestCase(
    rule_tag=K8S_CIS_1_2_2,
    resource_name=KUBE_APISERVER,
    expected=RULE_PASS_STATUS,
)

# TODO: Uncomment when rule 1.2.3 is implemented
# cis_1_2_3_pass = K8sTestCase(
#     rule_tag=K8S_CIS_1_2_3,
#     resource_name=KUBE_APISERVER,
#     expected=RULE_PASS_STATUS,
# )

cis_1_2_4_pass = K8sTestCase(
    rule_tag=K8S_CIS_1_2_4,
    resource_name=KUBE_APISERVER,
    expected=RULE_PASS_STATUS,
)

# https://github.com/kubernetes/kubernetes/pull/101178
# --kubelet-https flag has been deprecated and removed
# cis_1_2_4_fail = K8sTestCase(
#     rule_tag=K8S_CIS_1_2_4,
#     resource_name=KUBE_APISERVER,
#     expected=RULE_FAIL_STATUS,
# )

cis_1_2_5_pass = K8sTestCase(
    rule_tag=K8S_CIS_1_2_5,
    resource_name=KUBE_APISERVER,
    expected=RULE_PASS_STATUS,
)

cis_1_2_6_fail = K8sTestCase(
    rule_tag=K8S_CIS_1_2_6,
    resource_name=KUBE_APISERVER,
    expected=RULE_FAIL_STATUS,
)

cis_1_2_7_pass = K8sTestCase(
    rule_tag=K8S_CIS_1_2_7,
    resource_name=KUBE_APISERVER,
    expected=RULE_PASS_STATUS,
)

# Rules: 1.2.7, 1.2.8, 1.2.9 are about authorization-mode
# Case authorization-mode": "AlwaysAllow" is not configurable through kind-config

cis_1_2_8_pass = K8sTestCase(
    rule_tag=K8S_CIS_1_2_8,
    resource_name=KUBE_APISERVER,
    expected=RULE_PASS_STATUS,
)

cis_1_2_9_pass = K8sTestCase(
    rule_tag=K8S_CIS_1_2_9,
    resource_name=KUBE_APISERVER,
    expected=RULE_PASS_STATUS,
)

cis_1_2_10_fail = K8sTestCase(
    rule_tag=K8S_CIS_1_2_10,
    resource_name=KUBE_APISERVER,
    expected=RULE_FAIL_STATUS,
)

# Rule 1.2.10 cannot set to pass through kind-config

cis_1_2_11_fail = K8sTestCase(
    rule_tag=K8S_CIS_1_2_11,
    resource_name=KUBE_APISERVER,
    expected=RULE_FAIL_STATUS,
)

cis_1_2_11_pass = K8sTestCase(
    rule_tag=K8S_CIS_1_2_11,
    resource_name=KUBE_APISERVER,
    expected=RULE_PASS_STATUS,
)

cis_1_2_12_fail = K8sTestCase(
    rule_tag=K8S_CIS_1_2_12,
    resource_name=KUBE_APISERVER,
    expected=RULE_FAIL_STATUS,
)

# deploy/k8s/kind/kind-test-proc-conf2.yml
# AlwaysPullImages - requires pulling images from external registry
# Cloudbeat and cloudbeat-tests are loaded to local registry
# and cannot be pulled from external registry
# cis_1_2_12_pass = K8sTestCase(
#     rule_tag=K8S_CIS_1_2_12,
#     resource_name=KUBE_APISERVER,
#     expected=RULE_PASS_STATUS,
# )

cis_1_2_13_fail = K8sTestCase(
    rule_tag=K8S_CIS_1_2_13,
    resource_name=KUBE_APISERVER,
    expected=RULE_FAIL_STATUS,
)

# deploy/k8s/kind/kind-test-proc-conf2.yml
# SecurityContextDeny - does not allow deployment of pods with security context SecurityContext.as_user
# This case requires special modification of the deployment
# cis_1_2_13_pass = K8sTestCase(
#     rule_tag=K8S_CIS_1_2_13,
#     resource_name=KUBE_APISERVER,
#     expected=RULE_PASS_STATUS,
# )

# Rule 1.2.14 when set disable-admission-plugins=ServiceAccount kind cluster has many issues
# cis_1_2_14_fail = K8sTestCase(
#     rule_tag=K8S_CIS_1_2_14,
#     resource_name=KUBE_APISERVER,
#     expected=RULE_FAIL_STATUS,
# )


cis_1_2_15_fail = K8sTestCase(
    rule_tag=K8S_CIS_1_2_15,
    resource_name=KUBE_APISERVER,
    expected=RULE_FAIL_STATUS,
)

# The rule says: Ensure that the admission control plugin NamespaceLifecycle is set
# The remediation step is: Edit the API server pod specification file
# /etc/kubernetes/manifests/kube-apiserver.yaml on the Control Plane node and
# set the --disable-admission-plugins parameter to ensure it does not include NamespaceLifecycle.
# deploy/k8s/kind/kind-test-proc-conf2.yml has the following configuration, which should pass the test:
# enable-admission-plugins: "NodeRestriction,NamespaceLifecycle"
# disable-admission-plugins is not set
# It seems like a bug, because the rule is evaluated as failed
# cis_1_2_15_pass = K8sTestCase(
#     rule_tag=K8S_CIS_1_2_15,
#     resource_name=KUBE_APISERVER,
#     expected=RULE_PASS_STATUS,
# )

cis_1_2_16_fail = K8sTestCase(
    rule_tag=K8S_CIS_1_2_16,
    resource_name=KUBE_APISERVER,
    expected=RULE_FAIL_STATUS,
)

cis_1_2_16_pass = K8sTestCase(
    rule_tag=K8S_CIS_1_2_16,
    resource_name=KUBE_APISERVER,
    expected=RULE_PASS_STATUS,
)

cis_1_2_17_pass = K8sTestCase(
    rule_tag=K8S_CIS_1_2_17,
    resource_name=KUBE_APISERVER,
    expected=RULE_PASS_STATUS,
)

cis_1_2_18_fail = K8sTestCase(
    rule_tag=K8S_CIS_1_2_18,
    resource_name=KUBE_APISERVER,
    expected=RULE_FAIL_STATUS,
)

cis_1_2_18_pass = K8sTestCase(
    rule_tag=K8S_CIS_1_2_18,
    resource_name=KUBE_APISERVER,
    expected=RULE_PASS_STATUS,
)

cis_1_2_19_fail = K8sTestCase(
    rule_tag=K8S_CIS_1_2_19,
    resource_name=KUBE_APISERVER,
    expected=RULE_FAIL_STATUS,
)

cis_1_2_20_fail = K8sTestCase(
    rule_tag=K8S_CIS_1_2_20,
    resource_name=KUBE_APISERVER,
    expected=RULE_FAIL_STATUS,
)

cis_1_2_20_pass = K8sTestCase(
    rule_tag=K8S_CIS_1_2_20,
    resource_name=KUBE_APISERVER,
    expected=RULE_PASS_STATUS,
)

cis_1_2_21_fail = K8sTestCase(
    rule_tag=K8S_CIS_1_2_21,
    resource_name=KUBE_APISERVER,
    expected=RULE_FAIL_STATUS,
)

cis_1_2_21_pass = K8sTestCase(
    rule_tag=K8S_CIS_1_2_21,
    resource_name=KUBE_APISERVER,
    expected=RULE_PASS_STATUS,
)

cis_1_2_22_fail = K8sTestCase(
    rule_tag=K8S_CIS_1_2_22,
    resource_name=KUBE_APISERVER,
    expected=RULE_FAIL_STATUS,
)

cis_1_2_22_pass = K8sTestCase(
    rule_tag=K8S_CIS_1_2_22,
    resource_name=KUBE_APISERVER,
    expected=RULE_PASS_STATUS,
)

cis_1_2_23_fail = K8sTestCase(
    rule_tag=K8S_CIS_1_2_23,
    resource_name=KUBE_APISERVER,
    expected=RULE_FAIL_STATUS,
)

cis_1_2_23_pass = K8sTestCase(
    rule_tag=K8S_CIS_1_2_23,
    resource_name=KUBE_APISERVER,
    expected=RULE_PASS_STATUS,
)

cis_1_2_24_fail = K8sTestCase(
    rule_tag=K8S_CIS_1_2_24,
    resource_name=KUBE_APISERVER,
    expected=RULE_FAIL_STATUS,
)

cis_1_2_24_pass = K8sTestCase(
    rule_tag=K8S_CIS_1_2_24,
    resource_name=KUBE_APISERVER,
    expected=RULE_PASS_STATUS,
)

cis_1_2_25_pass = K8sTestCase(
    rule_tag=K8S_CIS_1_2_25,
    resource_name=KUBE_APISERVER,
    expected=RULE_PASS_STATUS,
)

cis_1_2_26_pass = K8sTestCase(
    rule_tag=K8S_CIS_1_2_26,
    resource_name=KUBE_APISERVER,
    expected=RULE_PASS_STATUS,
)

cis_1_2_27_pass = K8sTestCase(
    rule_tag=K8S_CIS_1_2_27,
    resource_name=KUBE_APISERVER,
    expected=RULE_PASS_STATUS,
)

cis_1_2_28_pass = K8sTestCase(
    rule_tag=K8S_CIS_1_2_28,
    resource_name=KUBE_APISERVER,
    expected=RULE_PASS_STATUS,
)

cis_1_2_29_pass = K8sTestCase(
    rule_tag=K8S_CIS_1_2_29,
    resource_name=KUBE_APISERVER,
    expected=RULE_PASS_STATUS,
)

cis_1_2_32_fail = K8sTestCase(
    rule_tag=K8S_CIS_1_2_32,
    resource_name=KUBE_APISERVER,
    expected=RULE_FAIL_STATUS,
)

cis_1_2_32_pass = K8sTestCase(
    rule_tag=K8S_CIS_1_2_32,
    resource_name=KUBE_APISERVER,
    expected=RULE_PASS_STATUS,
)

cis_1_3_2_fail = K8sTestCase(
    rule_tag=K8S_CIS_1_3_2,
    resource_name=KUBE_CONTROLLER,
    expected=RULE_FAIL_STATUS,
)

cis_1_3_2_pass = K8sTestCase(
    rule_tag=K8S_CIS_1_3_2,
    resource_name=KUBE_CONTROLLER,
    expected=RULE_PASS_STATUS,
)

cis_1_3_3_pass = K8sTestCase(
    rule_tag=K8S_CIS_1_3_3,
    resource_name=KUBE_CONTROLLER,
    expected=RULE_PASS_STATUS,
)

cis_1_3_4_pass = K8sTestCase(
    rule_tag=K8S_CIS_1_3_4,
    resource_name=KUBE_CONTROLLER,
    expected=RULE_PASS_STATUS,
)

cis_1_3_5_pass = K8sTestCase(
    rule_tag=K8S_CIS_1_3_5,
    resource_name=KUBE_CONTROLLER,
    expected=RULE_PASS_STATUS,
)

cis_1_3_6_fail = K8sTestCase(
    rule_tag=K8S_CIS_1_3_6,
    resource_name=KUBE_CONTROLLER,
    expected=RULE_FAIL_STATUS,
)

cis_1_3_6_pass = K8sTestCase(
    rule_tag=K8S_CIS_1_3_6,
    resource_name=KUBE_CONTROLLER,
    expected=RULE_PASS_STATUS,
)

cis_1_3_7_fail = K8sTestCase(
    rule_tag=K8S_CIS_1_3_7,
    resource_name=KUBE_CONTROLLER,
    expected=RULE_FAIL_STATUS,
)

cis_1_3_7_pass = K8sTestCase(
    rule_tag=K8S_CIS_1_3_7,
    resource_name=KUBE_CONTROLLER,
    expected=RULE_PASS_STATUS,
)

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

cis_4_2_1_fail = K8sTestCase(
    rule_tag=K8S_CIS_4_2_1,
    resource_name=KUBELET,
    expected=RULE_FAIL_STATUS,
)


cis_4_2_1_pass = K8sTestCase(
    rule_tag=K8S_CIS_4_2_1,
    resource_name=KUBELET,
    expected=RULE_PASS_STATUS,
)

cis_4_2_2_fail = K8sTestCase(
    rule_tag=K8S_CIS_4_2_2,
    resource_name=KUBELET,
    expected=RULE_FAIL_STATUS,
)


cis_4_2_2_pass = K8sTestCase(
    rule_tag=K8S_CIS_4_2_2,
    resource_name=KUBELET,
    expected=RULE_PASS_STATUS,
)

cis_4_2_3_pass = K8sTestCase(
    rule_tag=K8S_CIS_4_2_3,
    resource_name=KUBELET,
    expected=RULE_PASS_STATUS,
)

cis_4_2_4_fail = K8sTestCase(
    rule_tag=K8S_CIS_4_2_4,
    resource_name=KUBELET,
    expected=RULE_FAIL_STATUS,
)

# kind-test-proc-conf2.yml: config issue
# cis_4_2_4_pass = K8sTestCase(
#     rule_tag=K8S_CIS_4_2_4,
#     resource_name=KUBELET,
#     expected=RULE_PASS_STATUS,
# )

cis_4_2_5_fail = K8sTestCase(
    rule_tag=K8S_CIS_4_2_5,
    resource_name=KUBELET,
    expected=RULE_FAIL_STATUS,
)


cis_4_2_5_pass = K8sTestCase(
    rule_tag=K8S_CIS_4_2_5,
    resource_name=KUBELET,
    expected=RULE_PASS_STATUS,
)

cis_4_2_6_fail = K8sTestCase(
    rule_tag=K8S_CIS_4_2_6,
    resource_name=KUBELET,
    expected=RULE_FAIL_STATUS,
)


cis_4_2_6_pass = K8sTestCase(
    rule_tag=K8S_CIS_4_2_6,
    resource_name=KUBELET,
    expected=RULE_PASS_STATUS,
)

cis_4_2_7_fail = K8sTestCase(
    rule_tag=K8S_CIS_4_2_7,
    resource_name=KUBELET,
    expected=RULE_FAIL_STATUS,
)


cis_4_2_7_pass = K8sTestCase(
    rule_tag=K8S_CIS_4_2_7,
    resource_name=KUBELET,
    expected=RULE_PASS_STATUS,
)

cis_4_2_9_fail = K8sTestCase(
    rule_tag=K8S_CIS_4_2_9,
    resource_name=KUBELET,
    expected=RULE_FAIL_STATUS,
)


cis_4_2_9_pass = K8sTestCase(
    rule_tag=K8S_CIS_4_2_9,
    resource_name=KUBELET,
    expected=RULE_PASS_STATUS,
)

cis_4_2_10_fail = K8sTestCase(
    rule_tag=K8S_CIS_4_2_10,
    resource_name=KUBELET,
    expected=RULE_FAIL_STATUS,
)

# kind-test-proc-conf1.yml: config issue
# cis_4_2_11_fail = K8sTestCase(
#     rule_tag=K8S_CIS_4_2_11,
#     resource_name=KUBELET,
#     expected=RULE_FAIL_STATUS,
# )

cis_4_2_11_pass = K8sTestCase(
    rule_tag=K8S_CIS_4_2_11,
    resource_name=KUBELET,
    expected=RULE_PASS_STATUS,
)

# kind-test-proc-conf1.yml: config issue
# cis_4_2_12_fail = K8sTestCase(
#     rule_tag=K8S_CIS_4_2_12,
#     resource_name=KUBELET,
#     expected=RULE_FAIL_STATUS,
# )


cis_4_2_12_pass = K8sTestCase(
    rule_tag=K8S_CIS_4_2_12,
    resource_name=KUBELET,
    expected=RULE_PASS_STATUS,
)

cis_4_2_13_pass = K8sTestCase(
    rule_tag=K8S_CIS_4_2_13,
    resource_name=KUBELET,
    expected=RULE_PASS_STATUS,
)

k8s_process_config_1 = {
    "1.2.6 kube-apiserver --kubelet-certificate-authority is not set": cis_1_2_6_fail,
    "1.2.10 kube-apiserver --enable-admission-plugins=EventRateLimit is not set": cis_1_2_10_fail,
    "1.2.11 kube-apiserver --enable-admission-plugins=AlwaysAdmit": cis_1_2_11_fail,
    "1.2.12 kube-apiserver --enable-admission-plugins=AlwaysPullImages": cis_1_2_12_fail,
    "1.2.13 kube-apiserver --enable-admission-plugins=SecurityContextDeny is not set": cis_1_2_13_fail,
    # On disable-admission-plugins=ServiceAccount kind cluster has many issues
    # "1.2.14 kube-apiserver --disable-admission-plugins=ServiceAccount is not set": cis_1_2_14_fail,
    "1.2.15 kube-apiserver --disable-admission-plugins=NamespaceLifecycle is set": cis_1_2_15_fail,
    "1.2.16 kube-apiserver --disable-admission-plugins=NodeRestriction is not set": cis_1_2_16_fail,
    "1.2.18 kube-controller --profiling=true": cis_1_2_18_fail,
    "1.2.19 kube-controller --audit-log-path is not set": cis_1_2_19_fail,
    "1.2.20 kube-controller --audit-log-maxage is not set": cis_1_2_20_fail,
    "1.2.21 kube-controller --audit-log-maxage is not set": cis_1_2_21_fail,
    "1.2.22 kube-controller --audit-log-maxsize is not set": cis_1_2_22_fail,
    "1.2.23 kube-controller --request-timeout=59s": cis_1_2_23_fail,
    "1.2.24 kube-controller --service-account-lookup=false": cis_1_2_24_fail,
    "1.2.32 kube-controller --tls-cipher-suites is not set": cis_1_2_32_fail,
    "1.3.2 kube-controller --profiling=true": cis_1_3_2_fail,
    "1.3.6 kube-controller --feature-gates=RotateKubeletServerCertificate=false": cis_1_3_6_fail,
    "1.3.7 kube-controller --bind-address=0.0.0.0": cis_1_3_7_fail,
    "1.4.1 kube-scheduler --profiling=true": cis_1_4_1_fail,
    "1.4.2 kube-scheduler --bind-address=0.0.0.0": cis_1_4_2_fail,
    "2.2 etcd --client-cert-auth=false": cis_2_2_fail,
    "2.3 etcd --auto-tls=true": cis_2_3_fail,
    "2.5 etcd --peer-client-cert-auth=false": cis_2_5_fail,
    "2.6 etcd --peer-auto-tls=true": cis_2_6_fail,
    "4.2.1 kubelet authentication.anonymous.enabled=true": cis_4_2_1_fail,
    "4.2.2 kubelet authorization.mode=AlwaysAllow": cis_4_2_2_fail,
    "4.2.4 kubelet readOnlyPort=26492": cis_4_2_4_fail,
    # bug streamingConnectionIdleTimeout=0 should fail
    # From benchmark: If using a Kubelet config file,
    # edit the file to set `streamingConnectionIdleTimeout` to a value other than 0.
    # "4.2.5 kubelet streamingConnectionIdleTimeout=0": cis_4_2_5_fail,
    "4.2.6 kubelet protectKernelDefaults=false": cis_4_2_6_fail,
    "4.2.7 kubelet makeIPTablesUtilChains=false": cis_4_2_7_fail,
    "4.2.9 kubelet eventRecordQPS=4": cis_4_2_9_fail,
    "4.2.10 kubelet tlsCertFile does not exist": cis_4_2_9_fail,
    # kind-test-proc-conf1.yml: although configured, on the node this option still true
    # "4.2.11 kubelet rotateCertificates=false": cis_4_2_11_fail,
    # kind-test-proc-conf1.yml: although configured, in the kubelet config file this property does not appear
    # "4.2.12 kubelet featureGates.RotateKubeletServerCertificate=false and serverTLSBootstrap=false": cis_4_2_12_fail,
}

k8s_process_config_2 = {
    "1.2.2 kube-apiserver --token-auth-file is not set": cis_1_2_2_pass,
    # TODO: rule 1.2.3 is not implemented yet
    # "1.2.3 kube-apiserver --DenyServiceExternalIPs is not set": cis_1_2_3_pass,
    "1.2.4 kube-apiserver --kubelet-https is not set": cis_1_2_4_pass,
    "1.2.5 kube-apiserver --kubelet-client-certificate and --kubelet-client-key arguments are set": cis_1_2_5_pass,
    "1.2.7 kube-apiserver --authorization-mode=Node,RBAC": cis_1_2_7_pass,
    "1.2.8 kube-apiserver --authorization-mode Node is set": cis_1_2_8_pass,
    "1.2.9 kube-apiserver --authorization-mode RBAC is set": cis_1_2_9_pass,
    "1.2.11 kube-apiserver --enable-admission-plugins=AlwaysAdmit is not set": cis_1_2_11_pass,
    # "1.2.12 kube-apiserver --enable-admission-plugins=AlwaysPullImages": cis_1_2_12_pass,
    # "1.2.13 kube-apiserver --enable-admission-plugins=SecurityContextDeny": cis_1_2_13_pass,
    # "1.2.15 kube-apiserver --disable-admission-plugins=NamespaceLifecycle is not set": cis_1_2_15_pass,
    "1.2.16 kube-apiserver --disable-admission-plugins=NodeRestriction is set": cis_1_2_16_pass,
    "1.2.17 kube-apiserver --secure-port=6443": cis_1_2_17_pass,
    "1.2.18 kube-controller --profiling=false": cis_1_2_18_pass,
    "1.2.20 kube-controller --audit-log-maxage=30": cis_1_2_20_pass,
    "1.2.21 kube-controller --audit-log-maxbackup=10": cis_1_2_21_pass,
    "1.2.22 kube-controller --audit-log-maxsize=100": cis_1_2_22_pass,
    "1.2.23 kube-controller --request-timeout=default": cis_1_2_23_pass,
    "1.2.24 kube-controller --service-account-lookup is not set": cis_1_2_24_pass,
    "1.2.25 kube-controller --service-account-key-file exists": cis_1_2_25_pass,
    "1.2.26 kube-controller --etcd-certfile and --etcd-keyfile exist": cis_1_2_26_pass,
    "1.2.27 kube-controller --tls-cert-file and --tls-private-key-file exist": cis_1_2_27_pass,
    "1.2.28 kube-controller --client-ca-file exists": cis_1_2_28_pass,
    "1.2.29 kube-controller --etcd-cafile exists": cis_1_2_29_pass,
    "1.2.32 kube-controller --tls-cipher-suites=TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256": cis_1_2_32_pass,
    "1.3.2 kube-controller --profiling=false": cis_1_3_2_pass,
    "1.3.3 kube-controller --use-service-account-credentials=true": cis_1_3_3_pass,
    "1.3.4 kube-controller --service-account-private-key-file=<file>": cis_1_3_4_pass,
    "1.3.5 kube-controller --root-ca-file=<path/to/file>": cis_1_3_5_pass,
    "1.3.6 kube-controller --feature-gates=RotateKubeletServerCertificate=true": cis_1_3_6_pass,
    "1.3.7 kube-controller --bind-address=127.0.0.1": cis_1_3_7_pass,
    "1.4.1 kube-scheduler --profiling=false": cis_1_4_1_pass,
    "1.4.2 kube-scheduler --bind-address=127.0.0.1": cis_1_4_2_pass,
    "2.1 etcd --cert-file and --key-file are set": cis_2_1_pass,
    "2.2 etcd --client-cert-auth=true": cis_2_2_pass,
    "2.3 etcd --auto-tls=false": cis_2_3_pass,
    "2.4 etcd --peer-cert-file and --peer-key-file are set": cis_2_4_pass,
    "2.5 etcd --peer-client-cert-auth=true": cis_2_5_pass,
    "2.6 etcd --peer-auto-tls=false": cis_2_6_pass,
    "4.2.1 kubelet authentication.anonymous.enabled=false": cis_4_2_1_pass,
    "4.2.2 kubelet authorization.mode=Webhook": cis_4_2_2_pass,
    "4.2.3 kubelet authentication.x509.clientCAFile exists": cis_4_2_3_pass,
    # "4.2.4 kubelet readOnlyPort=0": cis_4_2_4_pass,
    "4.2.5 kubelet streamingConnectionIdleTimeout=5m": cis_4_2_5_pass,
    "4.2.6 kubelet protectKernelDefaults=true": cis_4_2_6_pass,
    "4.2.7 kubelet makeIPTablesUtilChains=true": cis_4_2_7_pass,
    "4.2.9 kubelet eventRecordQPS=0": cis_4_2_9_pass,
    "4.2.11 kubelet rotateCertificates=true": cis_4_2_11_pass,
    "4.2.12 kubelet featureGates.RotateKubeletServerCertificate=true": cis_4_2_12_pass,
    "4.2.13 kubelet tlsCipherSuites=TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256": cis_4_2_13_pass,
}

cis_k8s_process_all = {
    "test-k8s-config-1": k8s_process_config_1,
    "test-k8s-config-2": k8s_process_config_2,
}

# The name of this variable needs to be `tests_cases` in order to CIS Rules coverage stats to be generated
test_cases = {**k8s_process_config_1, **k8s_process_config_2}

# Get the test cases for the provided configuration
test_cases_by_config = cis_k8s_process_all.get(kubernetes.current_config, {})
