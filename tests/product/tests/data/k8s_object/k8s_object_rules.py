"""
This module defines k8s object test cases
"""

from .k8s_object_test_cases import KubeTestCase

DEFAULT = 'default'
RULE_FAIL_STATUS = 'failed'
RULE_PASS_STATUS = 'passed'
TEST_POD_NAME = 'busybox-pod'
TEST_CONTAINER_NAME = 'busybox'
TEST_ROLE_NAME = 'test-role'
TEST_CLUSTER_ROLE_NAME = 'test-cluster-role'
TEST_SERVICE_ACCOUNT_NAME = 'test-service-account'
TEST_CLUSTER_ROLE_BINDING = 'test-cluster-role-binding'
TEST_POD_SECURITY_POLICY = 'test-psp'
KUBE_SYSTEM_NAMESPACE = 'kube-system'

# CIS 5.1.3
cis_5_1_3_role_fail = KubeTestCase(
    rule_tag='CIS 5.1.3',
    resource_type='Role',
    resource_body={
        'metadata': {'name': TEST_ROLE_NAME, 'namespace': KUBE_SYSTEM_NAMESPACE},
        'rules': [
            {
                "apiGroups": ["*"],
                "resources": ["*"],
                "verbs": ["*"],
            },
        ],
    },
    expected=RULE_FAIL_STATUS,
)

cis_5_1_3_role_pass = KubeTestCase(
    rule_tag='CIS 5.1.3',
    resource_type='Role',
    resource_body={
        'metadata': {'name': TEST_ROLE_NAME, 'namespace': KUBE_SYSTEM_NAMESPACE},
        'rules': [
            {
                'apiGroups': [''],
                'resources': ['pods'],
                'verbs': ['get', 'watch', 'list'],
            }
        ]
    },
    expected=RULE_PASS_STATUS,
)

cis_5_1_3_cluster_role_fail = KubeTestCase(
    rule_tag='CIS 5.1.3',
    resource_type='ClusterRole',
    resource_body={
        'metadata': {'name': TEST_CLUSTER_ROLE_NAME},
        'rules': [
            {
                "apiGroups": ["*"],
                "resources": ["*"],
                "verbs": ["*"],
            },
        ],
    },
    expected=RULE_FAIL_STATUS,
)

cis_5_1_3_cluster_role_pass = KubeTestCase(
    rule_tag='CIS 5.1.3',
    resource_type='ClusterRole',
    resource_body={
        'metadata': {'name': TEST_CLUSTER_ROLE_NAME},
        'rules': [
            {
                'apiGroups': [''],
                'resources': ['pods'],
                'verbs': ['get', 'watch', 'list'],
            }
        ]
    },
    expected=RULE_PASS_STATUS,
)

cis_5_1_3 = {
    '5.1.3 Role with wildcards': cis_5_1_3_role_fail,
    '5.1.3 Role with no wildcards': cis_5_1_3_role_pass,
    '5.1.3 ClusterRole with wildcards': cis_5_1_3_cluster_role_fail,
    '5.1.3 ClusterRole with no wildcards': cis_5_1_3_cluster_role_pass,
}

# CIS 5.1.5
cis_5_1_5_pod_serviceAccount = KubeTestCase(
    rule_tag='CIS 5.1.5',
    resource_type='Pod',
    resource_body={
        'metadata': {'name': TEST_POD_NAME, 'namespace': KUBE_SYSTEM_NAMESPACE},
        'spec': {'serviceAccount': DEFAULT},
    },
    expected=RULE_FAIL_STATUS,
)

cis_5_1_5_pod_serviceAccountName = KubeTestCase(
    rule_tag='CIS 5.1.5',
    resource_type='Pod',
    resource_body={
        'metadata': {'name': TEST_POD_NAME, 'namespace': KUBE_SYSTEM_NAMESPACE},
        'spec': {'serviceAccountName': DEFAULT},
    },
    expected=RULE_FAIL_STATUS,
)

cis_5_1_5_service_account = KubeTestCase(
    rule_tag='CIS 5.1.5',
    resource_type='ServiceAccount',
    resource_body={
        'metadata': {'name': DEFAULT, 'namespace': DEFAULT},
        'automountServiceAccountToken': True,
    },
    expected=RULE_FAIL_STATUS,
)

cis_5_1_5 = {
    "5.1.5 ServiceAccount.Name == default and automountServiceAccountToken == true":
        cis_5_1_5_service_account,
    '5.1.5 Pod.serviceAccount == default': cis_5_1_5_pod_serviceAccount,
    '5.1.5 Pod.serviceAccountName == default': cis_5_1_5_pod_serviceAccountName,
}

# CIS 5.1.6
cis_5_1_6_pod_fail = KubeTestCase(
    rule_tag='CIS 5.1.6',
    resource_type='Pod',
    resource_body={
        'metadata': {'name': TEST_POD_NAME, 'namespace': KUBE_SYSTEM_NAMESPACE},
        'spec': {'automountServiceAccountToken': True},
    },
    expected=RULE_FAIL_STATUS,
)

cis_5_1_6_pod_pass = KubeTestCase(
    rule_tag='CIS 5.1.6',
    resource_type='Pod',
    resource_body={
        'metadata': {'name': TEST_POD_NAME, 'namespace': KUBE_SYSTEM_NAMESPACE},
        'spec': {'automountServiceAccountToken': False},
    },
    expected=RULE_PASS_STATUS,
)

cis_5_1_6_service_account_fail = KubeTestCase(
    rule_tag='CIS 5.1.6',
    resource_type='ServiceAccount',
    resource_body={
        'metadata': {'name': TEST_SERVICE_ACCOUNT_NAME, 'namespace': KUBE_SYSTEM_NAMESPACE},
        'automountServiceAccountToken': True,
    },
    expected=RULE_FAIL_STATUS,
)

cis_5_1_6_service_account_pass = KubeTestCase(
    rule_tag='CIS 5.1.6',
    resource_type='ServiceAccount',
    resource_body={
        'metadata': {'name': TEST_SERVICE_ACCOUNT_NAME, 'namespace': KUBE_SYSTEM_NAMESPACE},
        'automountServiceAccountToken': False,
    },
    expected=RULE_PASS_STATUS,
)

cis_5_1_6 = {
    '5.1.6 Pod.spec.automountServiceAccountToken == true': cis_5_1_6_pod_fail,
    '5.1.6 Pod.spec.automountServiceAccountToken == false': cis_5_1_6_pod_pass,
    '5.1.6 ServiceAccount.automountServiceAccountToken == true': cis_5_1_6_service_account_pass,
    '5.1.6 ServiceAccount.automountServiceAccountToken == false': cis_5_1_6_service_account_fail,
}

# CIS 5.2.2
cis_5_2_2_pod_fail = KubeTestCase(
    rule_tag='CIS 5.2.2',
    resource_type='Pod',
    resource_body={
        'metadata': {'name': TEST_POD_NAME, 'namespace': KUBE_SYSTEM_NAMESPACE},
        'spec': {
            'containers': [{
                'name': TEST_CONTAINER_NAME,
                'securityContext': {
                    'privileged': True
                }
            }]
        },
    },
    expected=RULE_FAIL_STATUS,
)

cis_5_2_2_pod_pass = KubeTestCase(
    rule_tag='CIS 5.2.2',
    resource_type='Pod',
    resource_body={
        'metadata': {'name': TEST_POD_NAME, 'namespace': KUBE_SYSTEM_NAMESPACE},
        'spec': {
            'containers': [{
                'name': TEST_CONTAINER_NAME,
                'securityContext': {
                    'privileged': False
                }
            }]
        },
    },
    expected=RULE_PASS_STATUS,
)

cis_5_2_2 = {
    '5.2.2 Pod.spec.containers.securityContext.privileged == true': cis_5_2_2_pod_fail,
    '5.2.2 Pod.spec.containers.securityContext.privileged == false': cis_5_2_2_pod_pass,
}

# CIS 5.2.3
cis_5_2_3_pod_fail = KubeTestCase(
    rule_tag='CIS 5.2.3',
    resource_type='Pod',
    resource_body={
        'metadata': {'name': TEST_POD_NAME, 'namespace': KUBE_SYSTEM_NAMESPACE},
        'spec': {'hostPID': True},
    },
    expected=RULE_FAIL_STATUS,
)

cis_5_2_3_pod_pass = KubeTestCase(
    rule_tag='CIS 5.2.3',
    resource_type='Pod',
    resource_body={
        'metadata': {'name': TEST_POD_NAME, 'namespace': KUBE_SYSTEM_NAMESPACE},
        'spec': {'hostPID': False}
    },
    expected=RULE_PASS_STATUS,
)

cis_5_2_3 = {
    '5.2.3 Pod.spec.hostPID == true': cis_5_2_3_pod_fail,
    '5.2.3 Pod.spec.hostPID == false': cis_5_2_3_pod_pass,
}

# CIS 5.2.4
cis_5_2_4_pod_fail = KubeTestCase(
    rule_tag='CIS 5.2.4',
    resource_type='Pod',
    resource_body={
        'metadata': {'name': TEST_POD_NAME, 'namespace': KUBE_SYSTEM_NAMESPACE},
        'spec': {'hostIPC': True},
    },
    expected=RULE_FAIL_STATUS,
)

cis_5_2_4_pod_pass = KubeTestCase(
    rule_tag='CIS 5.2.4',
    resource_type='Pod',
    resource_body={
        'metadata': {'name': TEST_POD_NAME, 'namespace': KUBE_SYSTEM_NAMESPACE},
        'spec': {'hostIPC': False},
    },
    expected=RULE_PASS_STATUS,
)

cis_5_2_4 = {
    '5.2.4 Pod.spec.hostIPC == true': cis_5_2_4_pod_fail,
    '5.2.4 Pod.spec.hostIPC == false': cis_5_2_4_pod_pass
}

# CIS 5.2.5
cis_5_2_5_pod_fail = KubeTestCase(
    rule_tag='CIS 5.2.5',
    resource_type='Pod',
    resource_body={
        'metadata': {'name': TEST_POD_NAME, 'namespace': KUBE_SYSTEM_NAMESPACE},
        'spec': {'hostNetwork': True},
    },
    expected=RULE_FAIL_STATUS,
)

cis_5_2_5_pod_pass = KubeTestCase(
    rule_tag='CIS 5.2.5',
    resource_type='Pod',
    resource_body={
        'metadata': {'name': TEST_POD_NAME, 'namespace': KUBE_SYSTEM_NAMESPACE},
        'spec': {'hostNetwork': False},
    },
    expected=RULE_PASS_STATUS,
)

cis_5_2_5 = {
    '5.2.5 Pod.spec.hostNetwork == true': cis_5_2_5_pod_fail,
    '5.2.5 Pod.spec.hostNetwork == false': cis_5_2_5_pod_pass,
}

# CIS 5.2.6
cis_5_2_6_pod_fail = KubeTestCase(
    rule_tag='CIS 5.2.6',
    resource_type='Pod',
    resource_body={
        'metadata': {'name': TEST_POD_NAME, 'namespace': KUBE_SYSTEM_NAMESPACE},
        'spec': {
            'containers': [{
                'name': TEST_CONTAINER_NAME,
                'securityContext': {
                    'allowPrivilegeEscalation': True
                }
            }]
        },
    },
    expected=RULE_FAIL_STATUS,
)

cis_5_2_6_pod_pass = KubeTestCase(
    rule_tag='CIS 5.2.6',
    resource_type='Pod',
    resource_body={
        'metadata': {'name': TEST_POD_NAME, 'namespace': KUBE_SYSTEM_NAMESPACE},
        'spec': {
            'containers': [{
                'name': TEST_CONTAINER_NAME,
                'securityContext': {
                    'allowPrivilegeEscalation': False
                }
            }]
        },
    },
    expected=RULE_PASS_STATUS,
)

cis_5_2_6 = {
    '5.2.6 Pod.spec.containers.securityContext.allowPrivilegeEscalation == true':
        cis_5_2_6_pod_fail,
    '5.2.6 Pod.spec.containers.securityContext.allowPrivilegeEscalation == false':
        cis_5_2_6_pod_pass,
}

# CIS 5.2.7
cis_5_2_7_pod_fail = KubeTestCase(
    rule_tag='CIS 5.2.7',
    resource_type='Pod',
    resource_body={
        'metadata': {'name': TEST_POD_NAME, 'namespace': KUBE_SYSTEM_NAMESPACE},
        'spec': {'runAsUser': {'rule': 'MustRunAs', 'ranges': [{'min': 0, 'max': 65535}]}},
    },
    expected=RULE_FAIL_STATUS,
)

cis_5_2_7_pod_pass = KubeTestCase(
    rule_tag='CIS 5.2.7',
    resource_type='Pod',
    resource_body={
        'metadata': {'name': TEST_POD_NAME, 'namespace': KUBE_SYSTEM_NAMESPACE},
        'spec': {'runAsUser': {'rule': 'MustRunAs', 'ranges': [{'min': 1, 'max': 65535}]}},
    },
    expected=RULE_PASS_STATUS,
)

cis_5_2_7_pod_container_fail = KubeTestCase(
    rule_tag='CIS 5.2.7',
    resource_type='Pod',
    resource_body={
        'metadata': {'name': TEST_POD_NAME, 'namespace': KUBE_SYSTEM_NAMESPACE},
        'spec': {
            'containers': [{
                'name': TEST_CONTAINER_NAME,
                'securityContext': {
                    'runAsUser': 0
                }
            }]
        },
    },
    expected=RULE_FAIL_STATUS,
)

cis_5_2_7 = {
    '5.2.7 Pod.spec.runAsUser allows root': cis_5_2_7_pod_fail,
    '5.2.7 Pod.spec.runAsUser forbids root': cis_5_2_7_pod_pass,
    '5.2.7 Pod.container.spec.securityContext.runAsUser == root': cis_5_2_7_pod_container_fail,
}

# CIS 5.2.8
cis_5_2_8_pod_container_fail = KubeTestCase(
    rule_tag='CIS 5.2.8',
    resource_type='Pod',
    resource_body={
        'metadata': {'name': TEST_POD_NAME, 'namespace': KUBE_SYSTEM_NAMESPACE},
        'spec': {
            'containers': [{
                'name': TEST_CONTAINER_NAME,
                'securityContext': {
                    'runAsUser': 0
                }
            }]
        },
    },
    expected=RULE_FAIL_STATUS,
)

cis_5_2_8 = {
    '5.2.8 Pod.container.spec.securityContext.runAsUser == root': cis_5_2_8_pod_container_fail,
}
