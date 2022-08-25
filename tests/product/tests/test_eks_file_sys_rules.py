"""
CIS EKS rules verification.
This module verifies correctness of retrieved findings by manipulating audit and remediation actions
"""
from datetime import datetime
import pytest
from commonlib.utils import get_ES_evaluation
from .data.file_system.eks_test_cases import cis_eks_kubeconfig, cis_eks_kubelet_config


@pytest.mark.eks_file_system_rules
@pytest.mark.parametrize(
    ("rule_tag", "command", "param_value", "resource", "expected"),
    cis_eks_kubeconfig,
)
def test_eks_kubeconfig(elastic_client,
                        config_node_pre_test,
                        rule_tag,
                        command,
                        param_value,
                        resource,
                        expected):
    """
    This data driven test verifies rules and findings return by cloudbeat agent.
    In order to add new cases @pytest.mark.parameterize section shall be updated.
    Setup and teardown actions are defined in data method.
    This test creates cloudbeat agent instance,
    changes node resources (modes, users, groups) and verifies,
    that cloudbeat returns correct finding.
    @param rule_tag: Name of rule to be verified.
    @param command: Command to be executed, for example chmod / chown
    @param param_value: Value to be used when executing command.
    @param resource: Full path to resource / file
    @param expected: Result to be found in finding evaluation field.
    @return: None - Test Pass / Fail result is generated.
    """
    k8s_client, api_client, cloudbeat_agent = config_node_pre_test
    change_data(k8s_client=k8s_client,
                api_client=api_client,
                command=command,
                param_value=param_value,
                resource=resource)
    verify_results(elastic_client=elastic_client,
                   cloudbeat_agent=cloudbeat_agent,
                   rule_tag=rule_tag,
                   resource=resource,
                   expected=expected)


@pytest.mark.eks_file_system_rules
@pytest.mark.parametrize(
    ("rule_tag", "command", "param_value", "resource", "expected"),
    cis_eks_kubelet_config,
)
def test_eks_kubelet_config(elastic_client,
                            config_node_pre_test,
                            rule_tag,
                            command,
                            param_value,
                            resource,
                            expected):
    """
    This data driven test verifies rules and findings return by cloudbeat agent.
    In order to add new cases @pytest.mark.parameterize section shall be updated.
    Setup and teardown actions are defined in data method.
    This test creates cloudbeat agent instance,
    changes node resources (modes, users, groups) and verifies,
    that cloudbeat returns correct finding.
    @param rule_tag: Name of rule to be verified.
    @param command: Command to be executed, for example chmod / chown
    @param param_value: Value to be used when executing command.
    @param resource: Full path to resource / file
    @param expected: Result to be found in finding evaluation field.
    @return: None - Test Pass / Fail result is generated.
    """
    k8s_client, api_client, cloudbeat_agent = config_node_pre_test
    change_data(k8s_client=k8s_client,
                api_client=api_client,
                command=command,
                param_value=param_value,
                resource=resource)
    verify_results(elastic_client=elastic_client,
                   cloudbeat_agent=cloudbeat_agent,
                   rule_tag=rule_tag,
                   resource=resource,
                   expected=expected)


def change_data(k8s_client, api_client, command, param_value, resource):
    # Currently, single node is used, in the future may be extended for all nodes.
    node = k8s_client.get_cluster_nodes()[0]
    api_client.exec_command(container_name=node.metadata.name,
                            command=command,
                            param_value=param_value,
                            resource=resource)


def verify_results(elastic_client, cloudbeat_agent, rule_tag, resource, expected):
    def identifier(res):
        return res.name in resource

    evaluation = get_ES_evaluation(
        elastic_client=elastic_client,
        timeout=cloudbeat_agent.findings_timeout,
        rule_tag=rule_tag,
        exec_timestamp=datetime.utcnow(),
        resource_identifier=identifier,
    )

    assert evaluation is not None, f"No evaluation for rule {rule_tag} could be found"
    assert evaluation == expected, f"Rule {rule_tag} verification failed," \
                                   f"expected: {expected}, got: {evaluation}"
