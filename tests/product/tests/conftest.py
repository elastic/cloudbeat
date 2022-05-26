from pathlib import Path
import time
import pytest
from kubernetes.client import ApiException

from commonlib.io_utils import get_k8s_yaml_objects

DEPLOY_YML = "../../deploy/cloudbeat-pytest.yml"
BUSYBOX_POD_YML = "../../deploy/mock-pod.yml"
POD_RESOURCE_TYPE = "Pod"


@pytest.fixture(scope='module')
def data(k8s, api_client, cloudbeat_agent):
    file_path = Path(__file__).parent / DEPLOY_YML
    if k8s.get_agent_pod_instances(agent_name=cloudbeat_agent.name, namespace=cloudbeat_agent.namespace):
        k8s.delete_from_yaml(get_k8s_yaml_objects(file_path=file_path))
    k8s.start_agent(yaml_file=file_path, namespace=cloudbeat_agent.namespace)
    time.sleep(5)
    yield k8s, api_client, cloudbeat_agent
    k8s_yaml_list = get_k8s_yaml_objects(file_path=file_path)
    k8s.delete_from_yaml(yaml_objects_list=k8s_yaml_list)  # stop agent


@pytest.fixture(scope='module')
def config_node_pre_test(data):
    k8s_client, api_client, cloudbeat_agent = data

    node = k8s_client.get_cluster_nodes()[0]

    api_client.exec_command(container_name=node.metadata.name,
                            command='touch',
                            param_value='/var/lib/etcd/some_file.txt',
                            resource='')

    api_client.exec_command(container_name=node.metadata.name,
                            command='touch',
                            param_value='/etc/kubernetes/pki/some_file.txt',
                            resource='')
    yield k8s_client, api_client, cloudbeat_agent


@pytest.fixture(scope='module')
def setup_busybox_pod(data):
    """
    Creates busybox pod to play with
    """
    file_path = Path(__file__).parent / BUSYBOX_POD_YML
    k8s_client, api_client, cloudbeat_agent = data
    k8s_yaml_list = get_k8s_yaml_objects(file_path=file_path)
    pod_name = k8s_yaml_list[0]['Pod']['name']

    # check if we already have one - delete if so
    try:
        # try getting the pod before deleting it - will raise exception if not found
        k8s_client.get_resource(resource_type=POD_RESOURCE_TYPE, name=pod_name, namespace=cloudbeat_agent.namespace)
        k8s_client.delete_from_yaml(get_k8s_yaml_objects(file_path=file_path))
        deleted = k8s_client.wait_for_resource(
            resource_type=POD_RESOURCE_TYPE,
            name=pod_name,
            namespace=cloudbeat_agent.namespace,
            status="DELETED")
        print(f"busybox deleted: {deleted}")
    except ApiException as notFound:
        print(f"no busybox online - setting up a new one: {notFound}")

    # create busybox
    k8s_client.create_from_yaml(
        yaml_file=file_path,
        namespace=cloudbeat_agent.namespace,
        verbose=True
    )
    yield k8s_client, api_client, cloudbeat_agent
    # teardown
    k8s_client.delete_from_yaml(yaml_objects_list=k8s_yaml_list)
