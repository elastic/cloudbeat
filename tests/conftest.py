"""
Global pytest file for fixtures and test configs
"""

import functools
import logging
import sys
import time

import configuration
import pytest
from _pytest.logging import LogCaptureFixture
from commonlib.agents_map import AgentComponentMapping, AgentExpectedMapping
from commonlib.docker_wrapper import DockerWrapper
from commonlib.elastic_wrapper import ElasticWrapper
from commonlib.io_utils import FsClient
from commonlib.kubernetes import KubernetesHelper
from loguru import logger


class InterceptHandler(logging.Handler):
    """
    This class intercepts standard logging messages toward Loguru sinks
    """

    def emit(self, record):
        # Get corresponding Loguru level if it exists
        try:
            level = logger.level(record.levelname).name
        except ValueError:
            level = record.levelno

        # Find caller from where originated the logged message
        frame, depth = logging.currentframe(), 2
        while frame.f_code.co_filename == logging.__file__:
            frame = frame.f_back
            depth += 1

        logger.opt(depth=depth, exception=record.exc_info).log(level, record.getMessage())


def logger_wraps(*, entry=True, _exit=True, level="DEBUG"):
    """
    Adding logger wrapper for debugging functions
    @param entry: Boolean, Display entry message or not
    @param _exit: Boolean, Display exit message or not
    @param level: Logging level, default DEBUG
    @return:
    """

    def wrapper(func):
        name = func.__name__

        @functools.wraps(func)
        def wrapped(*args, **kwargs):
            logger_ = logger.opt(depth=1)
            if entry:
                logger_.log(level, f"Entering '{name}' (args={args}, kwargs={kwargs})")
            start = time.time()
            result = func(*args, **kwargs)
            end = time.time()
            logger_.log(level, "Function '{}' executed in {:f} s", func.__name__, end - start)
            if _exit:
                logger_.log(level, f"Exiting '{name}' (result={result})")
            return result

        return wrapped

    return wrapper


@pytest.fixture
def caplog(_caplog: LogCaptureFixture) -> None:
    """Emitting logs from loguru's logger.log means that they will not show up in
    caplog which only works with Python's standard logging. This adds the same
    LogCaptureHandler being used by caplog to hook into loguru.
    Args:
        _caplog (LogCaptureFixture): caplog fixture
    Returns:
        None
    """

    def filter_(record):
        return record["level"].no >= _caplog.handler.level

    handler_id = logger.add(
        _caplog.handler,
        level=0,
        format="{message}",
        filter=filter_,
    )
    yield _caplog
    logger.remove(handler_id)


def pytest_configure():
    """
    Update framework configuration
    Logger: set logger default format
    @return:
    """
    fmt = (
        "<green>{time:YYYY-MM-DD HH:mm:ss.SSS}</green> "
        "<level>[{level}]</level> | "
        "<cyan>{name}</cyan>:<cyan>{function}</cyan>:<cyan>{line}</cyan> | "
        "<level>{message}</level>"
    )
    config = {
        "handlers": [
            {"sink": sys.stderr, "format": fmt},
        ],
    }

    logger.configure(**config)


@pytest.fixture(scope="session", autouse=True)
def k8s():
    """
    This function (fixture) instantiates KubernetesHelper depends on configuration.
    When executing tests code local, kubeconfig file is used for connecting to K8s cluster.
    When code executed as container (pod / job) in K8s cluster in cluster configuration is used.
    @return: Kubernetes Helper instance.
    """
    logger.debug(f"Kubernetes 'in_cluster_config': {configuration.kubernetes.is_in_cluster_config}")
    logger.debug(f"Kubernetes 'use_k8s': {configuration.kubernetes.use_kubernetes}")
    if configuration.kubernetes.use_kubernetes:
        return KubernetesHelper(
            is_in_cluster_config=configuration.kubernetes.is_in_cluster_config,
        )

    return None


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


@pytest.fixture
def cspm_client():
    """
    This function (fixture) instantiate ElasticWrapper.
    @return: ElasticWrapper client with cspm index.
    """
    return create_es_client(configuration.elasticsearch.cspm_index)


@pytest.fixture
def cnvm_client():
    """
    This function (fixture) instantiate ElasticWrapper.
    @return: ElasticWrapper client with cnvm index.
    """
    return create_es_client(configuration.elasticsearch.cnvm_index)


@pytest.fixture
def kspm_client():
    """
    This function (fixture) instantiate ElasticWrapper.
    @return: ElasticWrapper client with kspm index.
    """
    return create_es_client(configuration.elasticsearch.kspm_index)

@pytest.fixture
def asset_inventory_client():
    """
    This function (fixture) instantiate ElasticWrapper.
    @return: ElasticWrapper client with asset inventory index.
    """
    return create_es_client(configuration.elasticsearch.asset_inventory_index)


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


def pytest_sessionstart():
    """
    Called after the Session object has been created and before performing collection and entering the run test loop.
    @return:
    """
    configs = {
        "Cloudbeat": configuration.agent,
        "Kubernetes": configuration.kubernetes,
    }
    if configuration.agent.cluster_type == "eks":
        configs["EKS"] = configuration.eks

    for key, val in configs.items():
        configuration.print_environment(name=key, config_object=val)


def pytest_sessionfinish(session):
    """
    Called after whole test run finished, right before returning the exit status to the system.
    @param session: The pytest session object.
    @return:
    """
    report_dir = session.config.option.allure_report_dir
    cloudbeat = configuration.agent
    use_kubernetes = configuration.kubernetes.use_kubernetes
    if use_kubernetes:
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
            logger.warning("Warning fail to create allure environment report")


def create_es_client(index: str) -> ElasticWrapper:
    """
    This function (fixture) instantiate ElasticWrapper.
    @return: ElasticWrapper client with cspm index.
    """
    es_client = ElasticWrapper(
        configuration.elasticsearch.url,
        configuration.elasticsearch.basic_auth,
        index,
        configuration.elasticsearch.use_ssl,
    )
    logger.info(f"client with ElasticSearch url: {configuration.elasticsearch.url}")
    return es_client


@pytest.fixture(scope="session")
def agents_expected_components() -> AgentExpectedMapping:
    """
    This function (fixture) instantiate an AgentExpectedMapping.
    @return: an mapping of the agent components.
    """
    return AgentExpectedMapping()


@pytest.fixture(scope="session")
def agents_actual_components() -> AgentComponentMapping:
    """
    This function (fixture) instantiate an AgentComponentMapping.
    @return: an mapping of the agent components.
    """
    return AgentComponentMapping()
