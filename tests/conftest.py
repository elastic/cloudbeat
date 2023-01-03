"""
Global pytest file for fixtures and test configs
"""
import pytest
import configuration
from commonlib.kubernetes import KubernetesHelper
from commonlib.elastic_wrapper import ElasticWrapper
from commonlib.docker_wrapper import DockerWrapper
from commonlib.io_utils import FsClient
from _pytest.logging import LogCaptureFixture
from loguru import logger


@pytest.fixture(autouse=True)
def caplog(caplog: LogCaptureFixture) -> None:
    """Emitting logs from loguru's logger.log means that they will not show up in
    caplog which only works with Python's standard logging. This adds the same
    LogCaptureHandler being used by caplog to hook into loguru.
    Args:
        caplog_arg (LogCaptureFixture): caplog fixture
    Returns:
        None
    """

    def filter_(record):
        return record["level"].no >= caplog.handler.level

    handler_id = logger.add(
        caplog.handler,
        level=0,
        format="{message}",
        filter=filter_,
    )
    yield caplog
    logger.remove(handler_id)


@pytest.fixture(scope="session", autouse=True)
def k8s():
    """
    This function (fixture) instantiates KubernetesHelper depends on configuration.
    When executing tests code local, kubeconfig file is used for connecting to K8s cluster.
    When code executed as container (pod / job) in K8s cluster in cluster configuration is used.
    @return: Kubernetes Helper instance.
    """
    return KubernetesHelper(
        is_in_cluster_config=configuration.kubernetes.is_in_cluster_config,
    )


@pytest.fixture(scope="session", autouse=True)
def cloudbeat_agent():
    """
    This function (fixture) retrieves agent configuration, defined in configuration.py file.
    @return: Agent config
    """
    return configuration.agent


@pytest.fixture(scope="session", autouse=True)
def eks_cluster():
    """
    This function (fixture) retrieves eks_cluster configuration, defined in configuration.py file.
    @return: EKS config
    """
    return configuration.eks


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
        logger.info("docker client")
    else:
        client = FsClient
        logger.info("fs client")
    return client


def pytest_addoption(parser):
    """
    Add custom options to the pytest commandline utility.
    """
    parser.addoption(
        "--range",
        default="..",
        help="range to run tests on",
    )


def get_fixtures():
    """
    This function returns all fixtures in the current file.
    @return: List of fixtures.
    """
    return cloudbeat_agent, k8s


def pytest_sessionfinish(session):
    """
    Called after whole test run finished, right before returning the exit status to the system.
    @param session: The pytest session object.
    @return:
    """

    report_dir = session.config.option.allure_report_dir
    cloudbeat = configuration.agent
    kube_client = KubernetesHelper(
        is_in_cluster_config=configuration.kubernetes.is_in_cluster_config,
    )
    app_list = [cloudbeat.name, "kibana", "elasticsearch"]
    apps_dict = {}
    for app in app_list:
        apps_dict.update(
            kube_client.get_pod_image_version(
                pod_name=app,
                namespace=cloudbeat.namespace,
            ),
        )
    kubernetes_data = kube_client.get_nodes_versions()
    report_data = {**apps_dict, **kubernetes_data}
    try:
        if report_dir:
            with open(
                f"{report_dir}/{'environment.properties'}",
                "w",
                encoding="utf8",
            ) as allure_env:
                allure_env.writelines(
                    [f"{key}:{value}\n" for key, value in report_data.items()],
                )
    except ValueError:
        logger.exception("Warning fail to create allure environment report")
