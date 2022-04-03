import pytest
import configuration
from commonlib.kubernetes import KubernetesHelper
from commonlib.elastic_wrapper import ElasticWrapper


@pytest.fixture(scope="session", autouse=True)
def k8s():
    return KubernetesHelper(is_in_cluster_config=configuration.kubernetes.is_in_cluster_config)


@pytest.fixture(scope="session", autouse=True)
def cloudbeat_agent():
    return configuration.agent


@pytest.fixture(scope="session", autouse=True)
def elastic_client():
    elastic_config = configuration.elasticsearch
    es_client = ElasticWrapper(elastic_params=elastic_config)
    return es_client
