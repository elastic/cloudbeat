"""
This test is a basic integration test for cloudbeat.
The test executed on pre-merge events as required test.
The following flow is tested:
Cloudbeat -> ElasticSearch
"""

import configuration
import pytest
from commonlib.utils import get_findings, wait_for_cycle_completion
from loguru import logger

CONFIG_TIMEOUT = 45

cluster_data_dict = {
    "vanilla": ["file", "process", "k8s_object"],
    "vanilla_aws": ["aws-iam", "aws-s3", "aws-ec2-network", "aws-trail", "aws-monitoring", "aws-rds"],
    "eks": ["file", "process", "k8s_object", "load-balancer", "container-registry"],
}


def get_test_data() -> list:
    """
    This function retrieves test data that depends on cluster environment
    @return: test data list
    """
    try:
        return cluster_data_dict[configuration.agent.cluster_type]
    except KeyError as key:
        logger.error(f"Key not found in cluster_data_dict: {key}")
        return []


testdata = get_test_data()


@pytest.mark.pre_merge
@pytest.mark.order(1)
@pytest.mark.dependency()
def test_cloudbeat_pod_exist(fixture_data):
    """
    This test verifies that pods count is equal to nodes count
    :param fixture_data: (Pods list, Nodes list)
    :return:
    """
    # pylint: disable=duplicate-code

    pods, nodes = fixture_data
    pods_count = len(pods)
    nodes_count = len(nodes)
    assert pods_count == nodes_count, f"Pods count is {pods_count}, and nodes count is {nodes_count}"


@pytest.mark.pre_merge
@pytest.mark.order(3)
@pytest.mark.dependency(depends=["test_cloudbeat_pod_exist"])
def test_cloudbeat_pods_running(k8s, cloudbeat_agent):
    """
    This test verifies that all pods are in status "Running"
    :param k8s: Kubernetes client
    :param cloudbeat_agent: cloudbeat config
    :return:
    """
    pods = k8s.get_agent_pod_instances(
        agent_name=cloudbeat_agent.name,
        namespace=cloudbeat_agent.namespace,
    )
    # Verify that at least 1 pod is running the cluster
    assert len(pods) > 0, "There are no cloudbeat pod instances running in the cluster"
    # Verify that each pod is in running state
    for pod in pods:
        assert pod.status.phase == "Running", f"The pod '{pod.metadata.name}' status is: '{pod.status.phase}'"


@pytest.mark.pre_merge
@pytest.mark.order(2)
@pytest.mark.dependency(depends=["test_cloudbeat_pod_exist"])
@pytest.mark.parametrize("match_type", testdata)
def test_elastic_index_exists(kspm_client, match_type):
    """
    This test verifies that findings of all types are sending to elasticsearch
    :param kspm_client: Elastic API client
    :param match_type: Findings type for matching
    :return:
    """
    query, sort = kspm_client.build_es_query(term={"resource.type": match_type})
    result = get_findings(kspm_client, CONFIG_TIMEOUT, query, sort, match_type)

    assert len(result) > 0, f"The findings of type {match_type} not found"


@pytest.mark.pre_merge
@pytest.mark.skip(reason="https://github.com/elastic/cloudbeat/issues/2383")
@pytest.mark.order(4)
@pytest.mark.dependency(depends=["test_cloudbeat_pod_exist"])
def test_leader_election(fixture_data, kspm_client, cloudbeat_agent, k8s):
    """
    This test verifies that k8s related findings are sent by a single agent
    :param fixture_data: (Pods list, Nodes list)
    :param kspm_client: Elastic API client
    :param cloudbeat_agent: Cloudbeat configuration
    :param k8s: Kubernetes client object
    :return:
    """

    query, sort = kspm_client.build_es_query(term={"type": "k8s_object"})
    pods, nodes = fixture_data
    leader_node = k8s.get_cluster_leader(namespace=cloudbeat_agent.namespace, pods=pods)
    assert leader_node != "", "The Leader node could not be found"

    # Wait for all agents to send resources to ES
    res = wait_for_cycle_completion(elastic_client=kspm_client, nodes=nodes)
    assert res, "Not all nodes have completed a cycle within the configured threshold"

    result = kspm_client.get_index_data(
        query=query,
        size=1000,
        sort=sort,
    )
    # checking that k8s_objects are being sent only by the leader node.
    for resource in result["hits"]["hits"]:
        assert (
            leader_node == resource["_source"]["agent"]["name"]
        ), f"Multiple agents send k8s_objects, leader: {leader_node}, resource: {resource['_source']}"
