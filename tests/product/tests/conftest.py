"""
This module provides fixtures and configurations for
product tests.
"""
from pathlib import Path
import time
import json
import pytest
from loguru import logger
from kubernetes.client import ApiException
from kubernetes.utils import FailToCreateError
from commonlib.io_utils import get_k8s_yaml_objects

from product.tests.parameters import TEST_PARAMETERS


DEPLOY_YML = "../../deploy/cloudbeat-pytest.yml"
KUBE_RULES_ENV_YML = "../../deploy/mock-pod.yml"
POD_RESOURCE_TYPE = "Pod"


@pytest.fixture(scope="module", name="cloudbeat_start_stop")
def data(k8s, api_client, cloudbeat_agent):
    """
    This fixture starts cloudbeat, in case cloudbeat exists
    restart will be performed
    @param k8s: Kubernetes wrapper object
    @param api_client: Docker or FileSystem client
    @param cloudbeat_agent: Cloudbeat configuration
    @return:
    """
    file_path = Path(__file__).parent / DEPLOY_YML
    if k8s.get_agent_pod_instances(
        agent_name=cloudbeat_agent.name,
        namespace=cloudbeat_agent.namespace,
    ):
        k8s.delete_from_yaml(get_k8s_yaml_objects(file_path=file_path))
    k8s.start_agent(yaml_file=file_path, namespace=cloudbeat_agent.namespace)
    time.sleep(5)
    yield k8s, api_client, cloudbeat_agent
    k8s_yaml_list = get_k8s_yaml_objects(file_path=file_path)
    k8s.delete_from_yaml(yaml_objects_list=k8s_yaml_list)  # stop agent


@pytest.fixture(scope="module", name="config_node_pre_test")
def config_node_pre_test(cloudbeat_start_stop):
    """
    This fixture performs extra operations required in
    file system rules tests.
    Before test execution creates temporary files
    After test execution delete files created in Before section
    @param cloudbeat_start_stop: Cloudbeat fixture execution
    @return: Kubernetes object, Api client, Cloudbeat configuration
    """
    k8s_client, api_client, cloudbeat_agent = cloudbeat_start_stop

    nodes = k8s_client.get_cluster_nodes()

    temp_file_list = [
        "/var/lib/etcd/some_file.txt",
        "/etc/kubernetes/pki/some_file.txt",
    ]

    config_files = {
        "/etc/kubernetes/pki/admission_config.yaml": """apiVersion: apiserver.config.k8s.io/v1
kind: AdmissionConfiguration
plugins:
  - name: EventRateLimit
    path: /etc/kubernetes/pki/event_config.yaml""",
        "/etc/kubernetes/pki/event_config.yaml": """apiVersion: eventratelimit.admission.k8s.io/v1alpha1
kind: Configuration
limits:
  - type: Namespace
    qps: 50
    burst: 100
    cacheSize: 2000
  - type: User
    qps: 10
    burst: 50""",
    }

    # create temporary files:
    for node in nodes:
        if node.metadata.name != cloudbeat_agent.node_name:
            continue
        for temp_file in temp_file_list:
            api_client.exec_command(
                container_name=node.metadata.name,
                command="touch",
                param_value=temp_file,
                resource="",
            )

        # create config files:
        for config_file, contents in config_files.items():
            api_client.exec_command(
                container_name=node.metadata.name,
                command="cat",
                param_value=contents,
                resource=config_file,
            )

    yield k8s_client, api_client, cloudbeat_agent

    # delete temporary files:
    for node in nodes:
        if node.metadata.name != cloudbeat_agent.node_name:
            continue
        for temp_file in temp_file_list:
            api_client.exec_command(
                container_name=node.metadata.name,
                command="unlink",
                param_value=temp_file,
                resource="",
            )


@pytest.fixture(scope="module", name="clean_test_env")
def clean_test_env(cloudbeat_start_stop):
    """
    Sets up a testing env with needed kube resources
    """
    k8s_client, api_client, cloudbeat_agent = cloudbeat_start_stop

    file_path = Path(__file__).parent / KUBE_RULES_ENV_YML
    k8s_resources = get_k8s_yaml_objects(file_path=file_path)

    for yml_resource in k8s_resources:
        # check if we already have one - delete if so
        resource_type, metadata = yml_resource["kind"], yml_resource["metadata"]
        relevant_metadata = {k: metadata[k] for k in ("name", "namespace") if k in metadata}
        try:
            # try getting the resource before deleting it - will raise exception if not found
            k8s_client.get_resource(resource_type=resource_type, **relevant_metadata)
            k8s_client.delete_resources(resource_type=resource_type, **relevant_metadata)
            k8s_client.wait_for_resource(
                resource_type=resource_type,
                status_list=["DELETED"],
                **relevant_metadata,
            )
        except ApiException as not_found:
            logger.error(
                f"no {relevant_metadata['name']} online - setting up a new one: {not_found}",
            )
            # create resource

        k8s_client.create_from_dict(data=yml_resource, **relevant_metadata)

    yield k8s_client, api_client, cloudbeat_agent
    # teardown
    k8s_client.delete_from_yaml(yaml_objects_list=k8s_resources)


@pytest.fixture(scope="module", name="test_env")
def test_env(cloudbeat_start_stop):
    """
    Sets up a testing env with needed kube resources
    """
    k8s, api_client, cloudbeat_agent = cloudbeat_start_stop

    file_path = Path(__file__).parent / KUBE_RULES_ENV_YML
    k8s_resources = get_k8s_yaml_objects(file_path=file_path)

    try:
        k8s.create_from_yaml(yaml_file=file_path, namespace=cloudbeat_agent.namespace)
    except FailToCreateError as conflict:
        logger.error([json.loads(c.body)["message"] for c in conflict.api_exceptions])

    for yml_resource in k8s_resources:
        resource_type, metadata = yml_resource["kind"], yml_resource["metadata"]
        relevant_metadata = {k: metadata[k] for k in ("name", "namespace") if k in metadata}
        k8s.wait_for_resource(
            resource_type=resource_type,
            status_list=["RUNNING", "ADDED"],
            **relevant_metadata,
        )

    yield k8s, api_client, cloudbeat_agent
    # teardown
    k8s.delete_from_yaml(yaml_objects_list=k8s_resources)  # stop agent


def pytest_generate_tests(metafunc):
    """
    This function generates the test cases to run using the set of
    test cases registered in TEST_PARAMETERS and the values passed to
    relevant custom cmdline parameters such as --range.
    """
    if (
        metafunc.definition.get_closest_marker(
            metafunc.config.getoption("markexpr", default=None),
        )
        is None
        and metafunc.config.getoption("keyword", default=None) is None
    ):
        return
    params = TEST_PARAMETERS.get(metafunc.function)
    if params is None:
        raise ValueError(f"Params for function {metafunc.function} are not registered.")

    test_range = metafunc.config.getoption("range")
    test_range_start, test_range_end = test_range.split("..")

    if test_range_end != "" and int(test_range_end) < len(params.argvalues):
        params.argvalues = params.argvalues[: int(test_range_end)]

        if params.ids is not None:
            params.ids = params.ids[: int(test_range_end)]

    if test_range_start != "":
        if int(test_range_start) >= len(params.argvalues):
            raise ValueError(f"Invalid range for test function {metafunc.function}")

        params.argvalues = params.argvalues[int(test_range_start) :]

        if params.ids is not None:
            params.ids = params.ids[int(test_range_start) :]

    metafunc.parametrize(params.argnames, params.argvalues, ids=params.ids)
