"""
Kubernetes CIS rules verification.
This module verifies correctness of retrieved findings by manipulating audit and remediation actions
"""
import time

import pytest
from commonlib.io_utils import get_logs_from_stream, get_k8s_yaml_objects
from pathlib import Path


# DEPLOY_YAML = "../../deploy/cloudbeat-ds.yaml"
DEPLOY_YAML = "../../deploy/k8s-cloudbeat-tests/manifests/k8s-cloudbeat-tests/templates/cloudbeat-ds.yaml"


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


@pytest.mark.rules
@pytest.mark.parametrize(
    ("rule_tag", "command", "param_value", "resource", "expected"),
    [
        ('CIS 1.1.1', 'chmod', '700', '/etc/kubernetes/manifests/kube-apiserver.yaml', 'failed'),
        ('CIS 1.1.1', 'chmod', '644', '/etc/kubernetes/manifests/kube-apiserver.yaml', 'passed'),
        ('CIS 1.1.2', 'chown', 'daemon:daemon', '/etc/kubernetes/manifests/kube-apiserver.yaml', 'failed'),
        ('CIS 1.1.2', 'chown', 'root:root', '/etc/kubernetes/manifests/kube-apiserver.yaml', 'passed')
    ],
    ids=['CIS 1.1.1 mode 700',
         'CIS 1.1.1 mode 644',
         'CIS 1.1.2 daemon:daemon',
         'CIS 1.1.2 root:root'
         ]
)
def test_master_node_configuration(data,
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
    api_client.exec_command(container_name=node.metadata.name,
                            command=command,
                            param_value=param_value,
                            resource=resource)

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
    iteration = 0
    while time.time() - start_time < timeout:
        logs = get_logs_from_stream(k8s.get_pod_logs(pod_name=pod_name,
                                                     namespace=namespace,
                                                     since_seconds=1))
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
