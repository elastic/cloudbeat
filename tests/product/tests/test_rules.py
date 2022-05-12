"""
Kubernetes CIS rules verification.
This module verifies correctness of retrieved findings by manipulating audit and remediation actions
"""
import datetime
import time

import pytest
from commonlib.io_utils import get_logs_from_stream
from product.tests.tests.file_system.file_system_test_cases import *


def check_logs(k8s, timeout, pod_name, namespace, rule_tag, expected, exec_timestamp) -> bool:
    """
    This function retrieves pod logs and verifies if evaluation result is equal to expected result.
    @param k8s: Kubernetes wrapper instance
    @param timeout: Exit timeout
    @param pod_name: Name of pod the logs shall be retrieved from
    @param namespace: Kubernetes namespace
    @param rule_tag: Log rule tag
    @param expected: Expected result
    @:param exec_timestamp: the timestamp the command executed
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
        for log in logs:
            if not log.get('result'):
                continue
            findings = log.get('result').get('findings')
            findings_timestamp = datetime.datetime.strptime(log["time"], '%Y-%m-%dT%H:%M:%Sz')
            if (findings_timestamp - exec_timestamp).total_seconds() < 0:
                continue

            if findings:
                for finding in findings:
                    if rule_tag in finding.get('rule').get('tags'):
                        iteration += 1
                        agent_evaluation = finding.get('result').get('evaluation')
                        if agent_evaluation == expected:
                            print(f"{iteration}: expected:"
                                  f"{expected} tags:"
                                  f"{finding.get('rule').get('tags')}, "
                                  f"evidence: {finding.get('result').get('evidence')} ",
                                  f"evaluation: {finding.get('result').get('evaluation')}")
                            return True
    if iteration == 0:
        raise EnvironmentError("no logs found")
    return False


@pytest.mark.rules
@pytest.mark.parametrize(
    ("rule_tag", "command", "param_value", "resource", "expected"),
    [*cis_1_1_1,
     *cis_1_1_2,
     *cis_1_1_3,
     *cis_1_1_4,
     *cis_1_1_5,
     *cis_1_1_6,
     *cis_1_1_7,
     *cis_1_1_8,
     *cis_1_1_11,
     # *cis_1_1_12, uncomment after fix https://github.com/elastic/cloudbeat/issues/118
     *cis_1_1_13,
     *cis_1_1_14,
     *cis_1_1_15,
     *cis_1_1_16,
     *cis_1_1_17,
     *cis_1_1_18,
     *cis_4_1_1,
     *cis_4_1_2,
     *cis_4_1_5,
     *cis_4_1_9,
     *cis_4_1_10
     ],
)
def test_file_system_configuration(config_node_pre_test,
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
    k8s_client, api_client, cloudbeat_agent = config_node_pre_test
    # Currently, single node is used, in the future may be extended for all nodes.
    node = k8s_client.get_cluster_nodes()[0]
    pods = k8s_client.get_agent_pod_instances(agent_name=cloudbeat_agent.name, namespace=cloudbeat_agent.namespace)

    api_client.exec_command(container_name=node.metadata.name,
                            command=command,
                            param_value=param_value,
                            resource=resource)

    exec_ts = datetime.datetime.utcnow()

    verification_result = check_logs(k8s=k8s_client,
                                     pod_name=pods[0].metadata.name,
                                     namespace=cloudbeat_agent.namespace,
                                     rule_tag=rule_tag,
                                     expected=expected,
                                     timeout=cloudbeat_agent.findings_timeout,
                                     exec_timestamp=exec_ts)

    assert verification_result, f"Rule {rule_tag} verification failed."
