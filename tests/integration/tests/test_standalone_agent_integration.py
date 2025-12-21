"""
This test is a basic integration test for cloudbeat.
The test executed on pre-merge events as required test.
The following flow is tested:
Cloudbeat -> ElasticSearch
"""

import pytest
from commonlib.io_utils import FsClient
from commonlib.utils import get_findings
from loguru import logger

testdata = ["file", "process", "k8s_object"]
CONFIG_TIMEOUT = 120


@pytest.mark.pre_merge_agent
@pytest.mark.order(1)
@pytest.mark.dependency()
def test_agent_pod_exist(fixture_data):
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


@pytest.mark.pre_merge_agent
@pytest.mark.order(2)
@pytest.mark.dependency(depends=["test_agent_pod_exist"])
def test_agent_pods_running(fixture_data):
    """
    This test verifies that all pods are in status "Running"
    :param fixture_data: (Pods list, Nodes list)
    :return:
    """
    # Verify that at least 1 pod is running the cluster
    assert len(fixture_data[0]) > 0, "There are no elastic-agent pod instances running in the cluster"
    # Verify that each pod is in running state
    assert all(pod.status.phase == "Running" for pod in fixture_data[0]), "Not all pods are running"


@pytest.mark.pre_merge_agent
@pytest.mark.order(3)
@pytest.mark.dependency(depends=["test_agent_pod_exist"])
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


@pytest.mark.pre_merge_agent
@pytest.mark.skip(reason="https://github.com/elastic/cloudbeat/issues/3777")
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

    pods = k8s.get_agent_pod_instances(
        agent_name=cloudbeat_agent.name,
        namespace=cloudbeat_agent.namespace,
    )
    results = []
    exec_command = [
        "/usr/share/elastic-agent/elastic-agent",
        "status",
        "--output",
        "json",
    ]
    for pod in pods:
        response = k8s.pod_exec(
            name=pod.metadata.name,
            namespace=cloudbeat_agent.namespace,
            command=exec_command,
        )
        logger.debug(response)
        status = FsClient.get_beat_status_from_json(
            response=response,
            beat_name="cloudbeat",
        )
        if not status.startswith("Healthy"):
            results.append(f"Pod: {pod.metadata.name} status: {status}")

    assert len(results) == 0, "\n".join(results)
