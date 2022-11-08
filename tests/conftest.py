"""
Global pytest file for fixtures and test configs
"""
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
    if docker_config.use_docker:
        client = DockerWrapper(config=docker_config)
    else:
        client = FsClient
    return client


def pytest_addoption(parser):
    parser.addoption(
        '--range',
        default=['..'],
        help='range to run tests on',
    )


def get_fixtures():
    return cloudbeat_agent, k8s


def pytest_sessionfinish(session, exitstatus):
    """
    Called after whole test run finished, right before returning the exit status to the system.
    @param session: The pytest session object.
    @param exitstatus: (Union[int, ExitCode]) – The status which pytest will return to the system
    @return:
    """

    report_dir = session.config.option.allure_report_dir
    cloudbeat = configuration.agent
    kube_client = KubernetesHelper()
    app_list = [cloudbeat.name, 'kibana', 'elasticsearch']
    apps_dict = {}
    for app in app_list:
        apps_dict.update(kube_client.get_pod_image_version(pod_name=app, namespace=cloudbeat.namespace))
    kubernetes_data = kube_client.get_nodes_versions()
    report_data = {**apps_dict, **kubernetes_data}
    try:
        if report_dir:
            with open('{}/{}'.format(report_dir, 'environment.properties'), 'w') as allure_env:
                allure_env.writelines([f"{key}:{value}\n" for key, value in report_data.items()])
    except ValueError:
        print("Warning fail to create allure environment report")
