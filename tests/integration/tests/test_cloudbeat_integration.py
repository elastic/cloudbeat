"""
This test is a basic integration test for cloudbeat.
The test executed on pre-merge events as required test.
The following flow is tested:
Cloudbeat -> ElasticSearch
"""
import time
import json
import pytest
import allure
from commonlib.utils import wait_for_cycle_completion


testdata = ['file', 'process', 'k8s_object']
CONFIG_TIMEOUT = 30


@pytest.mark.pre_merge
@pytest.mark.order(1)
@pytest.mark.dependency()
def test_cloudbeat_pod_exist(fixture_data):
    """
    This test verifies that pods count is equal to nodes count
    :param fixture_data: (Pods list, Nodes list)
    :return:
    """
    pods, nodes = fixture_data
    pods_count = len(pods)
    nodes_count = len(nodes)
    assert pods_count == nodes_count, \
        f"Pods count is {pods_count}, and nodes count is {nodes_count}"


@pytest.mark.pre_merge
@pytest.mark.order(2)
@pytest.mark.dependency(depends=["test_cloudbeat_pod_exist"])
def test_cloudbeat_pods_running(fixture_data):
    """
    This test verifies that all pods are in status "Running"
    :param fixture_data: (Pods list, Nodes list)
    :return:
    """
    # Verify that at least 1 pod is running the cluster
    assert len(fixture_data[0]) > 0, "There are no cloudbeat pod instances running in the cluster"
    # Verify that each pod is in running state
    assert all(pod.status.phase == "Running" for pod in fixture_data[0]), "Not all pods are running"


@pytest.mark.pre_merge
@pytest.mark.order(3)
@pytest.mark.dependency(depends=["test_cloudbeat_pod_exist"])
@pytest.mark.parametrize("match_type", testdata)
def test_elastic_index_exists(elastic_client, match_type):
    """
    This test verifies that findings of all types are sending to elasticsearch
    :param elastic_client: Elastic API client
    :param match_type: Findings type for matching
    :return:
    """
    query, sort = elastic_client.build_es_query(term={"type": match_type})
    start_time = time.time()
    result = {}
    while time.time() - start_time < CONFIG_TIMEOUT:
        current_result = elastic_client.get_index_data(index_name=elastic_client.index,
                                                       query=query,
                                                       sort=sort)
        if elastic_client.get_total_value(data=current_result) != 0:
            allure.attach(json.dumps(elastic_client.get_doc_source(data=current_result),
                                     indent=4,
                                     sort_keys=True),
                          match_type,
                          attachment_type=allure.attachment_type.JSON)
            result = current_result
            break
        time.sleep(1)

    assert len(result) > 0, \
        f"The findings of type {match_type} not found"


@pytest.mark.pre_merge
@pytest.mark.order(4)
@pytest.mark.dependency(depends=["test_cloudbeat_pod_exist"])
def test_leader_election(fixture_data, elastic_client, cloudbeat_agent, k8s):
    """
    This test verifies that k8s related findings are sent by a single agent
    :param fixture_data: (Pods list, Nodes list)
    :param elastic_client: Elastic API client
    :param cloudbeat_agent: Cloudbeat configuration
    :param k8s: Kubernetes client object
    :return:
    """

    query, sort = elastic_client.build_es_query(term={"type": "k8s_object"})
    pods, nodes = fixture_data
    leader_node = k8s.get_cluster_leader(namespace=cloudbeat_agent.namespace, pods=pods)
    assert leader_node != "", \
        "The Leader node could not be found"

    # Wait for all agents to send resources to ES
    res = wait_for_cycle_completion(elastic_client=elastic_client, nodes=nodes)
    assert res, 'Not all nodes have completed a cycle within the configured threshold'

    result = elastic_client.get_index_data(index_name=elastic_client.index,
                                           query=query,
                                           size=1000,
                                           sort=sort)
    # checking that k8s_objects are being sent only by the leader node.
    for resource in result['hits']['hits']:
        assert leader_node == resource['_source']['agent']['name'], \
            f"Multiple agents send k8s_objects, leader: {leader_node}, resource: {resource['_source']}"
