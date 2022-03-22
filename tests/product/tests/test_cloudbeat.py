import pytest


@pytest.fixture
def data(k8s, cloudbeat_agent):
    pods = k8s.get_agent_pod_instances(agent_name=cloudbeat_agent.name,
                                       namespace=cloudbeat_agent.namespace)
    nodes = k8s.get_cluster_nodes()
    return pods, nodes


@pytest.mark.sanity
@pytest.mark.product
def test_cloudbeat_pod_exist(data):
    """
    This test verifies that pods count is equal to nodes count
    :param data: (Pods list, Nodes list)
    :return:
    """

    pods, nodes = data
    assert len(pods) == len(nodes), f"Pods count is {len(pods)}, and nodes count is {len(nodes)}"


@pytest.mark.sanity
@pytest.mark.product
def test_cloudbeat_pods_running(data):
    """
    This test verifies that all pods are in status "Running"
    :param data: (Pods list, Nodes list)
    :return:
    """
    assert all(pod.status.phase == "Running" for pod in data[0]), "Not all pods are running"

