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
from commonlib.io_utils import FsClient

testdata = ['file', 'process', 'k8s_object']
CONFIG_TIMEOUT = 60


@pytest.mark.post_merge_agent
@pytest.mark.order(1)
@pytest.mark.dependency()
def test_agent_pod_exist(fixture_data):
    """
    This test verifies that pods count is equal to nodes count
    :param fixture_data: (Pods list, Nodes list)
    :return:
    """
    pods, nodes = fixture_data
    pods_count = len(pods)
    nodes_count = len(nodes)
    assert pods_count == nodes_count,\
        f"Pods count is {pods_count}, and nodes count is {nodes_count}"


@pytest.mark.post_merge_agent
@pytest.mark.order(2)
@pytest.mark.dependency(depends=["test_agent_pod_exist"])
def test_agent_pods_running(fixture_data):
    """
    This test verifies that all pods are in status "Running"
    :param fixture_sa_data: (Pods list, Nodes list)
    :return:
    """
    # Verify that at least 1 pod is running the cluster
    assert len(
        fixture_data[0]) > 0, "There are no elastic-agent pod instances running in the cluster"
    # Verify that each pod is in running state
    assert all(pod.status.phase ==
               "Running" for pod in fixture_data[0]), "Not all pods are running"


@pytest.mark.post_merge_agent
@pytest.mark.order(3)
@pytest.mark.dependency(depends=["test_agent_pod_exist"])
@pytest.mark.parametrize("match_type", testdata)
def test_elastic_index_exists(elastic_client, match_type):
    """
    This test verifies that findings of all types are sending to elasticsearch
    :param elastic_client: Elastic API client
    :param match_type: Findings type for matching
    :return:
    """
    query, sort = elastic_client.build_es_query(
        term={"resource.ResourceMetadata.type": match_type})
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

    assert len(result) > 0,\
        f"The findings of type {match_type} not found"


@pytest.mark.post_merge_agent
@pytest.mark.order(4)
@pytest.mark.dependency(depends=["test_agent_pods_running"])
def test_cloudbeat_status(k8s, cloudbeat_agent):
    """
    This test connects to all elastic agents, executes command to
    retrieve beats status, verifies that cloud beat status in state "Running"
    @param k8s: Kubernetes wrapper client
    @param cloudbeat_agent: Cloudbeat configuration
    @return: Pass / Fail
    """

    pods = k8s.get_agent_pod_instances(agent_name=cloudbeat_agent.name,
                                       namespace=cloudbeat_agent.namespace)
    results = []
    exec_command = ["/usr/share/elastic-agent/elastic-agent",
                    "status", "--output", "json"]
    for pod in pods:
        response = k8s.pod_exec(name=pod.metadata.name,
                                namespace=cloudbeat_agent.namespace,
                                command=exec_command)
        status = FsClient.get_beat_status_from_json(response=response,
                                                    beat_name='cloudbeat')
        if status != 'Running':
            results.append(f"Pod: {pod.metadata.name} status: {status}")

    assert len(results) == 0, '\n'.join(results)
