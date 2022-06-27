from pathlib import Path
import pytest
from kubernetes.client import ApiException
from kubernetes.utils import FailToCreateError
import json
from commonlib.io_utils import get_k8s_yaml_objects


KUBE_RULES_ENV_YML = "../../deploy/mock-pod.yml"
POD_RESOURCE_TYPE = "Pod"


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
def clean_test_env(data):
    """
    Sets up a testing env with needed kube resources
    """
    k8s_client, api_client, cloudbeat_agent = data

    file_path = Path(__file__).parent / KUBE_RULES_ENV_YML
    k8s_resources = get_k8s_yaml_objects(file_path=file_path)

    for yml_resource in k8s_resources:
        # check if we already have one - delete if so
        resource_type, metadata = yml_resource['kind'], yml_resource['metadata']
        relevant_metadata = {k: metadata[k] for k in ('name', 'namespace') if k in metadata}
        try:
            # try getting the resource before deleting it - will raise exception if not found
            k8s_client.get_resource(resource_type=resource_type, **relevant_metadata)
            k8s_client.delete_resources(resource_type=resource_type, **relevant_metadata)
            k8s_client.wait_for_resource(resource_type=resource_type, status_list=["DELETED"], **relevant_metadata)
        except ApiException as notFound:
            print(f"no {relevant_metadata['name']} online - setting up a new one: {notFound}")
            # create resource
            k8s_client.create_from_dict(data=yml_resource, **relevant_metadata)

    yield k8s_client, api_client, cloudbeat_agent
    # teardown
    k8s_client.delete_from_yaml(yaml_objects_list=k8s_resources)


@pytest.fixture(scope='module')
def test_env(data):
    """
    Sets up a testing env with needed kube resources
    """
    k8s, api_client, cloudbeat_agent = data

    file_path = Path(__file__).parent / KUBE_RULES_ENV_YML
    k8s_resources = get_k8s_yaml_objects(file_path=file_path)

    try:
        k8s.create_from_yaml(yaml_file=file_path, namespace=cloudbeat_agent.namespace)
    except FailToCreateError as conflict:
        print([json.loads(c.body)['message'] for c in conflict.api_exceptions])

    for yml_resource in k8s_resources:
        resource_type, metadata = yml_resource['kind'], yml_resource['metadata']
        relevant_metadata = {k: metadata[k] for k in ('name', 'namespace') if k in metadata}
        k8s.wait_for_resource(resource_type=resource_type, status_list=["RUNNING", "ADDED"], **relevant_metadata)

    yield k8s, api_client, cloudbeat_agent
    # teardown
    k8s.delete_from_yaml(yaml_objects_list=k8s_resources)  # stop agent
