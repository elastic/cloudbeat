import pytest
import configuration
from commonlib.kubernetes import KubernetesHelper
from commonlib.elastic_wrapper import ElasticWrapper
from commonlib.docker_wrapper import DockerWrapper
from commonlib.io_utils import FsClient


@pytest.fixture(scope="session", autouse=True)
def k8s():
    """
    This function (fixture) instantiates KubernetesHelper depends on configuration.
    When executing tests code local, kubeconfig file is used for connecting to K8s cluster.
    When code executed as container (pod / job) in K8s cluster in cluster configuration is used.
    @return: Kubernetes Helper instance.
    """
    return KubernetesHelper(is_in_cluster_config=configuration.kubernetes.is_in_cluster_config)


@pytest.fixture(scope="session", autouse=True)
def cloudbeat_agent():
    """
    This function (fixture) retrieves agent configuration, defined in configuration.py file.
    @return: Agent config
    """
    return configuration.agent


@pytest.fixture(scope="session", autouse=True)
def elastic_client():
    """
    This function (fixture) instantiate ElasticWrapper.
    @return: ElasticWrapper client
    """
    elastic_config = configuration.elasticsearch
    es_client = ElasticWrapper(elastic_params=elastic_config)
    return es_client


@pytest.fixture(scope="session", autouse=True)
def api_client():
    """
    This function (fixture) instantiates client depends on configuration.
    For local development mode, the docker api may be used.
    For production mode (deployment to k8s cluster), FsClient shall be used.
    @return: Client (docker / FsClient).
    """
    docker_config = configuration.docker
    print(f"Config use_docker value: {docker_config.use_docker}")
    if docker_config.use_docker:
        client = DockerWrapper(config=docker_config)
    else:
        client = FsClient
    return client
