"""
Kubernetes CIS rules verification.
This module verifies correctness of retrieved findings by manipulating audit and remediation actions
"""
import time

import pytest
from commonlib.io_utils import get_logs_from_stream, get_k8s_yaml_objects
from pathlib import Path


DEPLOY_YAML = "../../deploy/cloudbeat-pytest.yml"


@pytest.fixture(scope='module')
def data(k8s, api_client, cloudbeat_agent):

    file_path = Path(__file__).parent / DEPLOY_YAML
    if k8s.get_agent_pod_instances(agent_name=cloudbeat_agent.name, namespace=cloudbeat_agent.namespace):
        k8s.stop_agent(get_k8s_yaml_objects(file_path=file_path))
    k8s.start_agent(yaml_file=file_path, namespace=cloudbeat_agent.namespace)
    time.sleep(5)
    yield k8s, api_client, cloudbeat_agent
    k8s_yaml_list = get_k8s_yaml_objects(file_path=file_path)
    k8s.stop_agent(yaml_objects_list=k8s_yaml_list)


def check_logs(k8s, timeout, pod_name, namespace, rule_tag, expected) -> bool:
    """
    This function retrieves pod logs and verifies if evaluation result is equal to expected result.
    @param k8s: Kubernetes wrapper instance
    @param timeout: Exit timeout
    @param pod_name: Name of pod the logs shall be retrieved from
    @param namespace: Kubernetes namespace
    @param rule_tag: Log rule tag
    @param expected: Expected result
    @return: bool True / False
    """
    start_time = time.time()
    iteration = 0
    while time.time() - start_time < timeout:
        try:
            logs = get_logs_from_stream(k8s.get_pod_logs(pod_name=pod_name,
                                                         namespace=namespace,
                                                         since_seconds=2))
        except:
            continue
        iteration += 1
        for log in logs:
            if not log.get('result'):
                print(f"{iteration}: no result")
                continue
            findings = log.get('result').get('findings')
            if findings:
                for finding in findings:
                    if rule_tag in finding.get('rule').get('tags'):
                        print(f"{iteration}: expected:"
                              f"{expected} tags:"
                              f"{finding.get('rule').get('tags')}, "
                              f"evaluation: {finding.get('result').get('evaluation')}")
                        agent_evaluation = finding.get('result').get('evaluation')
                        if agent_evaluation == expected:
                            return True
    return False


@pytest.mark.rules
@pytest.mark.parametrize(
    ("rule_tag", "command", "param_value", "resource", "expected"),
    [
        ('CIS 1.1.1', 'chmod', '0700', '/etc/kubernetes/manifests/kube-apiserver.yaml', 'failed'),
        ('CIS 1.1.1', 'chmod', '0644', '/etc/kubernetes/manifests/kube-apiserver.yaml', 'passed'),
        ('CIS 1.1.2', 'chown', 'daemon:daemon', '/etc/kubernetes/manifests/kube-apiserver.yaml', 'failed'),
        ('CIS 1.1.2', 'chown', 'root:root', '/etc/kubernetes/manifests/kube-apiserver.yaml', 'passed'),
        ('CIS 1.1.3', 'chmod', '0700', '/etc/kubernetes/manifests/kube-controller-manager.yaml', 'failed'),
        ('CIS 1.1.3', 'chmod', '0644', '/etc/kubernetes/manifests/kube-controller-manager.yaml', 'passed'),
        ('CIS 1.1.4', 'chown', 'root:daemon', '/etc/kubernetes/manifests/kube-controller-manager.yaml', 'failed'),
        ('CIS 1.1.4', 'chown', 'daemon:root', '/etc/kubernetes/manifests/kube-controller-manager.yaml', 'failed'),
        ('CIS 1.1.4', 'chown', 'daemon:daemon', '/etc/kubernetes/manifests/kube-controller-manager.yaml', 'failed'),
        ('CIS 1.1.4', 'chown', 'root:root', '/etc/kubernetes/manifests/kube-controller-manager.yaml', 'passed'),
        ('CIS 1.1.5', 'chmod', '0700', '/etc/kubernetes/manifests/kube-scheduler.yaml', 'failed'),
        ('CIS 1.1.5', 'chmod', '0644', '/etc/kubernetes/manifests/kube-scheduler.yaml', 'passed'),
        ('CIS 1.1.6', 'chown', 'root:daemon', '/etc/kubernetes/manifests/kube-scheduler.yaml', 'failed'),
        ('CIS 1.1.6', 'chown', 'root:root', '/etc/kubernetes/manifests/kube-scheduler.yaml', 'passed'),
        ('CIS 1.1.7', 'chmod', '0700', '/etc/kubernetes/manifests/etcd.yaml', 'failed'),
        ('CIS 1.1.7', 'chmod', '0644', '/etc/kubernetes/manifests/etcd.yaml', 'passed'),
        ('CIS 1.1.8', 'chown', 'root:daemon', '/etc/kubernetes/manifests/etcd.yaml', 'failed'),
        ('CIS 1.1.8', 'chown', 'daemon:root', '/etc/kubernetes/manifests/etcd.yaml', 'failed'),
        ('CIS 1.1.8', 'chown', 'daemon:daemon', '/etc/kubernetes/manifests/etcd.yaml', 'failed'),
        ('CIS 1.1.8', 'chown', 'root:root', '/etc/kubernetes/manifests/etcd.yaml', 'passed'),
        ('CIS 1.1.11', 'chmod', '0710', '/var/lib/etcd', 'failed'),
        ('CIS 1.1.11', 'chmod', '0710', '/var/lib/etcd/somefile', 'failed'),
        ('CIS 1.1.11', 'chmod', '0600', '/var/lib/etcd', 'passed'),
        ('CIS 1.1.11', 'chmod', '0600', '/var/lib/etcd/somefile', 'passed'),
        ('CIS 1.1.12', 'chown', 'root:root', '/var/lib/etcd', 'failed'),
        ('CIS 1.1.12', 'chown', 'daemon:root', '/var/lib/etcd', 'failed'),
        ('CIS 1.1.12', 'chown', 'root:daemon', '/var/lib/etcd', 'failed'),
        ('CIS 1.1.12', 'chown', 'root:daemon', '/var/lib/etcd/some_file.txt', 'failed'),
        ('CIS 1.1.12', 'chown', 'daemon:daemon', '/var/lib/etcd/', 'passed'),
        ('CIS 1.1.12', 'chown', 'daemon:daemon', '/var/lib/etcd/some_file.txt', 'passed'),
        ('CIS 1.1.13', 'chmod', '0700', '/etc/kubernetes/admin.conf', 'failed'),
        ('CIS 1.1.13', 'chmod', '0644', '/etc/kubernetes/admin.conf', 'failed'),
        ('CIS 1.1.13', 'chmod', '0600', '/etc/kubernetes/admin.conf', 'passed'),
        ('CIS 1.1.14', 'chown', 'root:daemon', '/etc/kubernetes/admin.conf', 'failed'),
        ('CIS 1.1.14', 'chown', 'daemon:root', '/etc/kubernetes/admin.conf', 'failed'),
        ('CIS 1.1.14', 'chown', 'daemon:daemon', '/etc/kubernetes/admin.conf', 'failed'),
        ('CIS 1.1.14', 'chown', 'root:root', '/etc/kubernetes/admin.conf', 'passed'),
        ('CIS 1.1.15', 'chmod', '0700', '/etc/kubernetes/scheduler.conf', 'failed'),
        ('CIS 1.1.15', 'chmod', '0644', '/etc/kubernetes/scheduler.conf', 'passed'),
        ('CIS 1.1.16', 'chown', 'root:daemon', '/etc/kubernetes/scheduler.conf', 'failed'),
        ('CIS 1.1.16', 'chown', 'daemon:root', '/etc/kubernetes/scheduler.conf', 'failed'),
        ('CIS 1.1.16', 'chown', 'daemon:daemon', '/etc/kubernetes/scheduler.conf', 'failed'),
        ('CIS 1.1.16', 'chown', 'root:root', '/etc/kubernetes/scheduler.conf', 'passed'),
        ('CIS 1.1.17', 'chmod', '0700', '/etc/kubernetes/controller-manager.conf', 'failed'),
        ('CIS 1.1.17', 'chmod', '0644', '/etc/kubernetes/controller-manager.conf', 'passed'),
        ('CIS 1.1.18', 'chown', 'root:daemon', '/etc/kubernetes/controller-manager.conf', 'failed'),
        ('CIS 1.1.18', 'chown', 'daemon:root', '/etc/kubernetes/controller-manager.conf', 'failed'),
        ('CIS 1.1.18', 'chown', 'daemon:daemon', '/etc/kubernetes/controller-manager.conf', 'failed'),
        ('CIS 1.1.18', 'chown', 'root:root', '/etc/kubernetes/controller-manager.conf', 'passed'),
        ('CIS 1.1.19', 'chown', 'root:daemon', '/etc/kubernetes/pki/', 'failed'),
        ('CIS 1.1.19', 'chown', 'daemon:root', '/etc/kubernetes/pki/', 'failed'),
        ('CIS 1.1.19', 'chown', 'daemon:daemon', '/etc/kubernetes/pki/', 'failed'),
        ('CIS 1.1.19', 'chown', 'root:daemon', '/etc/kubernetes/pki/some_file.txt', 'failed'),
        ('CIS 1.1.19', 'chown', 'root:root', '/etc/kubernetes/pki/', 'passed'),
        ('CIS 1.1.19', 'chown', 'root:root', '/etc/kubernetes/pki/some_file.txt', 'passed'),
        ('CIS 4.1.1', 'chmod', '0700', '/etc/systemd/system/kubelet.service.d/10-kubeadm.conf', 'failed'),
        ('CIS 4.1.1', 'chmod', '0644', '/etc/systemd/system/kubelet.service.d/10-kubeadm.conf', 'passed'),
        ('CIS 4.1.2', 'chown', 'root:daemon', '/etc/systemd/system/kubelet.service.d/10-kubeadm.conf', 'failed'),
        ('CIS 4.1.2', 'chown', 'daemon:root', '/etc/systemd/system/kubelet.service.d/10-kubeadm.conf', 'failed'),
        ('CIS 4.1.2', 'chown', 'daemon:daemon', '/etc/systemd/system/kubelet.service.d/10-kubeadm.conf', 'failed'),
        ('CIS 4.1.2', 'chown', 'root:root', '/etc/systemd/system/kubelet.service.d/10-kubeadm.conf', 'passed'),
        ('CIS 4.1.5', 'chmod', '0700', '/etc/kubernetes/kubelet.conf', 'failed'),
        ('CIS 4.1.5', 'chmod', '0644', '/etc/kubernetes/kubelet.conf', 'passed'),
        ('CIS 4.1.9', 'chmod', '0700', '/var/lib/kubelet/config.yaml', 'failed'),
        ('CIS 4.1.9', 'chmod', '0644', '/var/lib/kubelet/config.yaml', 'passed'),
        ('CIS 4.1.10', 'chown', 'root:daemon', '/etc/kubernetes/kubelet.conf', 'failed'),
        ('CIS 4.1.10', 'chown', 'daemon:root', '/etc/kubernetes/kubelet.conf', 'failed'),
        ('CIS 4.1.10', 'chown', 'daemon:daemon', '/etc/kubernetes/kubelet.conf', 'failed'),
        ('CIS 4.1.10', 'chown', 'root:root', '/etc/kubernetes/kubelet.conf', 'passed'),
    ],
    ids=[
         'CIS 1.1.1 mode 700',
         'CIS 1.1.1 mode 644',
         'CIS 1.1.2 daemon:daemon',
         'CIS 1.1.2 root:root',
         'CIS 1.1.3 mode 744',
         'CIS 1.1.3 mode 644',
         'CIS 1.1.4 root:daemon',
         'CIS 1.1.4 daemon:root',
         'CIS 1.1.4 daemon:daemon',
         'CIS 1.1.4 root:root',
         'CIS 1.1.5 mode 0700',
         'CIS 1.1.5 mode 0644',
         'CIS 1.1.6 root:daemon',
         'CIS 1.1.6 root:root',
         'CIS 1.1.7 mode 0700',
         'CIS 1.1.7 mode 0644',
         'CIS 1.1.8 root:daemon',
         'CIS 1.1.8 daemon:root',
         'CIS 1.1.8 daemon:daemon',
         'CIS 1.1.8 root:root',
         'CIS 1.1.11 mode 0710',
         'CIS 1.1.11 mode 0710',
         'CIS 1.1.11 mode 0600',
         'CIS 1.1.11 mode 0600',
         'CIS 1.1.12 root:root',
         'CIS 1.1.12 daemon:root',
         'CIS 1.1.12 root:daemon',
         'CIS 1.1.12 root:daemon',
         'CIS 1.1.12 daemon:daemon',
         'CIS 1.1.12 daemon:daemon',
         'CIS 1.1.13 0700',
         'CIS 1.1.13 0644',
         'CIS 1.1.13 0600',
         'CIS 1.1.14 root:daemon',
         'CIS 1.1.14 daemon:root',
         'CIS 1.1.14 daemon:daemon',
         'CIS 1.1.14 root:root',
         'CIS 1.1.15 0700',
         'CIS 1.1.15 0644',
         'CIS 1.1.16 root:daemon',
         'CIS 1.1.16 daemon:root',
         'CIS 1.1.16 daemon:daemon',
         'CIS 1.1.16 root:root',
         'CIS 1.1.17 0700',
         'CIS 1.1.17 0644',
         'CIS 1.1.18 root:daemon',
         'CIS 1.1.18 daemon:root',
         'CIS 1.1.18 daemon:daemon',
         'CIS 1.1.18 root:root',
         'CIS 1.1.19 root:root',
         'CIS 1.1.19 daemon:root',
         'CIS 1.1.19 root:daemon',
         'CIS 1.1.19 root:daemon',
         'CIS 1.1.19 daemon:daemon',
         'CIS 1.1.19 daemon:daemon',
         'CIS 4.1.1 0700',
         'CIS 4.1.1 0644',
         'CIS 4.1.2 root:daemon',
         'CIS 4.1.2 daemon:root',
         'CIS 4.1.2 daemon:daemon',
         'CIS 4.1.2 root:root',
         'CIS 4.1.5 0700',
         'CIS 4.1.5 0644',
         'CIS 4.1.9 0700',
         'CIS 4.1.9 0644',
         'CIS 4.1.10 root:daemon',
         'CIS 4.1.10 daemon:root',
         'CIS 4.1.10 daemon:daemon',
         'CIS 4.1.10 root:root',
         ]
)
def test_file_system_configuration(data,
                                   rule_tag,
                                   command,
                                   param_value,
                                   resource,
                                   expected):
    """
    This data driven test verifies rules and findings return by cloudbeat agent.
    In order to add new cases @pytest.mark.parameterize section shall be updated.
    Setup and teardown actions are defined in data method.
    This test creates cloudbeat agent instance, changes node resources (modes, users, groups) and verifies,
    that cloudbeat returns correct finding.
    @param data: Fixture that returns object instances and configurations.
    @param rule_tag: Name of rule to be verified.
    @param command: Command to be executed, for example chmod / chown
    @param param_value: Value to be used when executing command.
    @param resource: Full path to resource / file
    @param expected: Result to be found in finding evaluation field.
    @return: None - Test Pass / Fail result is generated.
    """
    k8s_client, api_client, agent_config = data
    # Currently, single node is used, in the future may be extended for all nodes.
    node = k8s_client.get_cluster_nodes()[0]
    pods = k8s_client.get_agent_pod_instances(agent_name=agent_config.name, namespace=agent_config.namespace)
    res = api_client.exec_command(container_name=node.metadata.name,
                            command=command,
                            param_value=param_value,
                            resource=resource)
    time.sleep(10)
    print(f'exec command output: {res}')
    verification_result = check_logs(k8s=k8s_client,
                                     pod_name=pods[0].metadata.name,
                                     namespace=agent_config.namespace,
                                     rule_tag=rule_tag,
                                     expected=expected,
                                     timeout=agent_config.findings_timeout)

    assert verification_result, f"Rule {rule_tag} verification failed."
