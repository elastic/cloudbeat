from product.tests.kube_test_case import KubeTestCase

cis_5_1_5_pod_serviceAccount = KubeTestCase(
    rule_tag='CIS 5.1.5',
    resource_type='Pod',
    resource_body={
        "metadata": {"name": "busybox-pod"},
        "spec": {"serviceAccount": "default"}
    },
    expected='failed',
)

cis_5_1_5_pod_serviceAccountName = KubeTestCase(
    rule_tag='CIS 5.1.5',
    resource_type='Pod',
    resource_body={
        "metadata": {"name": "busybox-pod"},
        "spec": {"serviceAccountName": "default"}
    },
    expected='failed',
)

cis_5_1_5_service_account = KubeTestCase(
    rule_tag='CIS 5.1.5',
    resource_type='ServiceAccount',
    resource_body={
        "metadata": {"name": "default"},
        "automountServiceAccountToken": True
    },
    expected='failed',
)

cis_5_1_5 = {
    'CIS 5.1.5 Pod.serviceAccount == default': cis_5_1_5_pod_serviceAccount,
    'CIS 5.1.5 Pod.serviceAccountName == default': cis_5_1_5_pod_serviceAccountName,
    'CIS 5.1.5 ServiceAccount.Name == default and automountServiceAccountToken == true': cis_5_1_5_service_account
}
