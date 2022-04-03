"""
Kubernetes CIS rules verification.
This module verifies correctness of retrieved findings by manipulating audit and remediation actions
"""
import time

import pytest
from commonlib.io_utils import get_logs_from_stream, get_k8s_yaml_objects
from pathlib import Path


DEPLOY_YAML = "../../deploy/k8s-cloudbeat-tests/templates/cloudbeat-ds.yaml"


@pytest.fixture(scope='module')
def data(k8s, docker_client, cloudbeat_agent):

    file_path = Path(__file__).parent / DEPLOY_YAML
    if k8s.get_agent_pod_instances(agent_name=cloudbeat_agent.name, namespace=cloudbeat_agent.namespace):
        k8s.stop_agent(get_k8s_yaml_objects(file_path=file_path))
    k8s.start_agent(yaml_file=file_path, namespace=cloudbeat_agent.namespace)
    time.sleep(5)
    yield k8s, docker_client, cloudbeat_agent
    k8s_yaml_list = get_k8s_yaml_objects(file_path=file_path)
    k8s.stop_agent(yaml_objects_list=k8s_yaml_list)


@pytest.mark.parametrize(
    ("rule_tag", "command", "param_value", "resource", "expected"),
    [
        ('CIS 1.1.1', 'chmod', '700', '/etc/kubernetes/manifests/kube-apiserver.yaml', 'failed'),
        ('CIS 1.1.1', 'chmod', '644', '/etc/kubernetes/manifests/kube-apiserver.yaml', 'passed'),
        # ('CIS 1.1.2', 'chown', 'daemon:daemon', '/etc/kubernetes/manifests/kube-apiserver.yaml', 'failed'),
        # ('CIS 1.1.2', 'chown', 'root:root', '/etc/kubernetes/manifests/kube-apiserver.yaml', 'passed')
    ],
    ids=['CIS 1.1.1 mode 700',
         'CIS 1.1.1 mode 644',
         # 'CIS 1.1.2 daemon:daemon',
         # 'CIS 1.1.2 root:root'
         ]
)
def test_master_node_configuration(data,
                                   rule_tag,
                                   command,
                                   param_value,
                                   resource,
                                   expected):
    k8s_client, d_client, agent_config = data
    node = k8s_client.get_cluster_nodes()[0]
    pods = k8s_client.get_agent_pod_instances(agent_name=agent_config.name, namespace=agent_config.namespace)
    command_f = f"{command} {param_value} {resource}"
    d_client.exec_command(container_name=node.metadata.name,
                          command=command_f)

    verification_result = check_logs(k8s=k8s_client,
                                     pod_name=pods[0].metadata.name,
                                     namespace=agent_config.namespace,
                                     rule_tag=rule_tag,
                                     expected=expected,
                                     timeout=agent_config.findings_timeout)

    assert verification_result, f"Rule {rule_tag} verification failed."


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
    while time.time() - start_time < timeout:
        logs = get_logs_from_stream(k8s.get_pod_logs(pod_name=pod_name,
                                                     namespace=namespace,
                                                     since_seconds=3))
        for log in logs:
            # print(log)
            if not log.get('result'):
                continue
            findings = log.get('result').get('findings')
            if findings:
                for finding in findings:
                    print(finding.get('rule').get('tags'))
                    if rule_tag in finding.get('rule').get('tags'):
                        agent_evaluation = finding.get('result').get('evaluation')
                        print(f"current state - expected: {expected}, evaluation: {agent_evaluation}")
                        if agent_evaluation == expected:
                            print(finding.get('rule').get('tags'))
                            return True
        time.sleep(1)
    return False
