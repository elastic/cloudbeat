"""
Kubernetes CIS rules verification.
This module verifies correctness of retrieved findings by manipulating audit and remediation actions
"""

from datetime import datetime
import pytest
from commonlib.utils import get_ES_evaluation

from product.tests.data.file_system import file_system_test_cases as fs_tc
from product.tests.parameters import register_params, Parameters


@pytest.mark.file_system_rules
def test_file_system_configuration(
    kspm_client,
    config_node_pre_test,
    rule_tag,
    command,
    param_value,
    resource,
    expected,
):
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
    # pylint: disable=duplicate-code

    k8s_client, api_client, cloudbeat_agent = config_node_pre_test
    # Currently, single node is used, in the future may be extended for all nodes.
    node = k8s_client.get_cluster_nodes()[0]
    api_client.exec_command(
        container_name=node.metadata.name,
        command=command,
        param_value=param_value,
        resource=resource,
    )

    def identifier(ident_resource):
        eval_resource = ident_resource.resource.raw
        if not eval_resource.path.endswith(resource):
            return False

        if command == "chmod":
            try:
                return int(eval_resource.mode) == int(param_value)
            except AttributeError:
                return False

        elif command == "chown":
            owner, group = param_value.split(":")
            try:
                return (eval_resource.owner == owner) and (eval_resource.group == group)
            except AttributeError:
                return False

        return False

    evaluation = get_ES_evaluation(
        elastic_client=kspm_client,
        timeout=cloudbeat_agent.findings_timeout,
        rule_tag=rule_tag,
        exec_timestamp=datetime.utcnow(),
        resource_identifier=identifier,
    )

    assert evaluation is not None, f"No evaluation for rule {rule_tag} could be found"
    assert evaluation == expected, f"Rule {rule_tag} verification failed," f"expected: {expected}, got: {evaluation}"


register_params(
    test_file_system_configuration,
    Parameters(
        ("rule_tag", "command", "param_value", "resource", "expected"),
        [
            *fs_tc.cis_1_1_1,
            *fs_tc.cis_1_1_2,
            *fs_tc.cis_1_1_3,
            *fs_tc.cis_1_1_4,
            *fs_tc.cis_1_1_5,
            *fs_tc.cis_1_1_6,
            *fs_tc.cis_1_1_7,
            *fs_tc.cis_1_1_8,
            *fs_tc.cis_1_1_11,
            *fs_tc.cis_1_1_12,
            *fs_tc.cis_1_1_13,
            *fs_tc.cis_1_1_14,
            *fs_tc.cis_1_1_15,
            *fs_tc.cis_1_1_16,
            *fs_tc.cis_1_1_17,
            *fs_tc.cis_1_1_18,
            *fs_tc.cis_1_1_19,
            *fs_tc.cis_1_1_20,
            *fs_tc.cis_1_1_21,
            *fs_tc.cis_4_1_1,
            *fs_tc.cis_4_1_2,
            *fs_tc.cis_4_1_5,
            *fs_tc.cis_4_1_6,
            *fs_tc.cis_4_1_9,
            *fs_tc.cis_4_1_10,
        ],
    ),
)
