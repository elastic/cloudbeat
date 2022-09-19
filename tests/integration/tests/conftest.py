"""
Integration tests setup configurations and fixtures
"""
# pylint: disable=E0401
from pathlib import Path
import pytest
from commonlib.io_utils import get_k8s_yaml_objects
from commonlib.kubernetes import ApiException

DEPLOY_YML_DICT = {
    'cloudbeat_vanilla': "../../deploy/cloudbeat-pytest.yml",
    'elastic-agent_vanilla': "../../deploy/sa-agent-pytest.yml",
    'cloudbeat_eks': "../../deploy/cloudbeat-eks-pytest.yaml"
}


@pytest.fixture(scope='module', name='start_stop_cloudbeat')
def fixture_start_stop_cloudbeat(k8s, api_client, cloudbeat_agent):
    """
    This fixture starts cloudbeat on test module setup and
    stops on teardown of the test module
    @param k8s: Kubernetes client object
    @param api_client: Docker api / FileSystem api
    @param cloudbeat_agent: Cloudbeat configuration
    @return: Kubernetes object, Api client, Cloudbeat config
    """

    try:
        file_path = Path(__file__).parent / \
            DEPLOY_YML_DICT[f"{cloudbeat_agent.name}_{cloudbeat_agent.cluster_type}"]
    except KeyError:
        raise Exception(
            f"configuration {cloudbeat_agent.name}_{cloudbeat_agent.cluster_type} is unknown") from KeyError

    if k8s.get_agent_pod_instances(agent_name=cloudbeat_agent.name,
                                   namespace=cloudbeat_agent.namespace):
        k8s.delete_from_yaml(get_k8s_yaml_objects(file_path=file_path))
        k8s.wait_for_resource(resource_type='Pod',
                              name=cloudbeat_agent.name,
                              status_list=['DELETED'],
                              namespace=cloudbeat_agent.namespace)
    k8s.start_agent(yaml_file=file_path, namespace=cloudbeat_agent.namespace)
    k8s.wait_for_resource(resource_type='Pod',
                          name=cloudbeat_agent.name,
                          status_list=['ADDED', 'MODIFIED'],
                          namespace=cloudbeat_agent.namespace)
    yield k8s, api_client, cloudbeat_agent
    k8s_yaml_list = get_k8s_yaml_objects(file_path=file_path)
    k8s.delete_from_yaml(yaml_objects_list=k8s_yaml_list)  # stop agent

    lease_resources = [{
        'name': 'cloudbeat-cluster-leader',
        'namespace': cloudbeat_agent.namespace
    },
        {
            'name': 'elastic-agent-cluster-leader',
            'namespace': cloudbeat_agent.namespace
    }
    ]
    # Delete lease resources
    for resource in lease_resources:
        try:
            k8s.delete_resources(resource_type='Lease', **resource)
        except ApiException:
            continue


@pytest.fixture(name='fixture_data')
def fixture_data(start_stop_cloudbeat, k8s, cloudbeat_agent):
    """
    This fixture is used for preparing data for integration test
    @param start_stop_cloudbeat: fixture to start and stop agent / cloudbeat
    @param k8s: Kubernetes wrapper object
    @param cloudbeat_agent: config object
    @return: pods, nodes in cluster
    """
    # pylint: disable=W0612,W0613
    pods = k8s.get_agent_pod_instances(agent_name=cloudbeat_agent.name,
                                       namespace=cloudbeat_agent.namespace)
    nodes = k8s.get_cluster_nodes()

    return pods, nodes


@pytest.fixture(name='fixture_sa_data')
def fixture_sa_data(k8s, cloudbeat_agent):
    """
    This fixture is used for preparing data for integration test
    @param k8s: Kubernetes wrapper object
    @param cloudbeat_agent: config object
    @return: pods, nodes in cluster
    """
    # pylint: disable=W0612,W0613
    pods = k8s.get_agent_pod_instances(agent_name=cloudbeat_agent.name,
                                       namespace=cloudbeat_agent.namespace)
    nodes = k8s.get_cluster_nodes()
    return pods, nodes
