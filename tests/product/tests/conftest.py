import pytest
from commonlib.io_utils import *
from commonlib.utils import *
from pathlib import Path

DEPLOY_YML = "../../deploy/cloudbeat-pytest.yml"
BUSYBOX_POD_YML = "../../deploy/mock-pod.yml"


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

    # add etcd group if not exists
    # groups = api_client.exec_command(container_name=node.metadata.name, command='getent', param_value='group',
    #                                  resource='')
    # if 'etcd' not in groups:
    #     api_client.exec_command(container_name=node.metadata.name, command='groupadd',
    #                             param_value='etcd',
    #                             resource='')
    #
    # # add etcd user if not exists
    # users = api_client.exec_command(container_name=node.metadata.name, command='getent', param_value='passwd',
    #                                 resource='')
    # if 'etcd' not in users:
    #     api_client.exec_command(container_name=node.metadata.name,
    #                             command='useradd',
    #                             param_value='-g etcd etcd',
    #                             resource='')

    # create stub file
    # etcd_content = api_client.exec_command(container_name=node.metadata.name, command='ls',
    #                                        param_value='/var/lib/etcd/', resource='')
    # if 'some_file.txt' not in etcd_content:

    api_client.exec_command(container_name=node.metadata.name,
                            command='touch',
                            param_value='/var/lib/etcd/some_file.txt',
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
    if k8s_client.get_resource(resource_type="Pod", name=pod_name, namespace=cloudbeat_agent.namespace):
        k8s_client.delete_from_yaml(get_k8s_yaml_objects(file_path=file_path))
        time.sleep(20) # todo: wait for status change

    # create busybox
    k8s_client.create_from_yaml(
        yaml_file=file_path,
        namespace=cloudbeat_agent.namespace,
        verbose=True
    )
    yield k8s_client, api_client, cloudbeat_agent
    # teardown
    k8s_client.delete_from_yaml(yaml_objects_list=k8s_yaml_list)
