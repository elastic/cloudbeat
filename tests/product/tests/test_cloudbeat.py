import pytest


@pytest.fixture
def data(k8s, cloudbeat_agent):
    pods = k8s.get_agent_pod_instances(agent_name=cloudbeat_agent.name,
                                       namespace=cloudbeat_agent.namespace)
    nodes = k8s.get_cluster_nodes()
    return pods, nodes


@pytest.mark.sanity
@pytest.mark.product
@pytest.mark.ci_cloudbeat
def test_cloudbeat_pod_exist(data):
    """
    This test verifies that pods count is equal to nodes count
    :param data: (Pods list, Nodes list)
    :return:
    """
    pods, nodes = data
    pods_count = len(pods)
    nodes_count = len(nodes)
    assert pods_count == nodes_count, f"Pods count is {pods_count}, and nodes count is {nodes_count}"


@pytest.mark.sanity
@pytest.mark.product
@pytest.mark.ci_cloudbeat
def test_cloudbeat_pods_running(data):
    """
    This test verifies that all pods are in status "Running"
    :param data: (Pods list, Nodes list)
    :return:
    """
    # Verify that at least 1 pod is running the cluster
    assert len(data[0]) > 0, "There are no cloudbeat pod instances running in the cluster"
    # Verify that each pod is in running state
    assert all(pod.status.phase == "Running" for pod in data[0]), "Not all pods are running"
