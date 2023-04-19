"""
This module provides common configuration of the test project,
and also mapping environment variables

"""
import os
from munch import Munch
from loguru import logger

# --- Cloudbeat agent environment definition ----------------
agent = Munch()
agent.name = os.getenv("AGENT_NAME", "cloudbeat")
agent.namespace = os.getenv("AGENT_NAMESPACE", "kube-system")
agent.findings_timeout = 500
agent.eks_findings_timeout = 120
agent.aws_findings_timeout = 10
agent.cluster_type = os.getenv("CLUSTER_TYPE", "eks")  # options: vanilla / eks / vanilla_aws

# The K8S Node on which the test Pod is running.
agent.node_name = os.getenv("NODE_NAME")

# --- Kubernetes environment definition --------------------
kubernetes = Munch()
kubernetes.is_in_cluster_config = bool(
    os.getenv("KUBERNETES_IN_CLUSTER", "false") == "true",
)

# --- AWS EKS ---------------------------------------------
eks = Munch()
eks.current_config = os.getenv("EKS_CONFIG", "test-eks-config-1")
eks.config_1 = os.getenv("EKS_CONFIG_1", "test-eks-config-1")
eks.config_1_node_1 = os.getenv(
    "EKS_CONFIG_1_NODE_1",
    "ip-192-168-57-173.eu-west-2.compute.internal",
)
eks.config_1_node_2 = os.getenv(
    "EKS_CONFIG_1_NODE_2",
    "ip-192-168-83-229.eu-west-2.compute.internal",
)
eks.config_2 = os.getenv("EKS_CONFIG_2", "test-eks-config-2")
eks.config_2_node_1 = os.getenv(
    "EKS_CONFIG_2_NODE_1",
    "ip-192-168-14-74.eu-west-2.compute.internal",
)
eks.config_2_node_2 = os.getenv(
    "EKS_CONFIG_2_NODE_2",
    "ip-192-168-89-216.eu-west-2.compute.internal",
)

# --- Elasticsearch environment definition --------------------------------
elasticsearch = Munch()
elasticsearch.hosts = os.getenv("ES_HOST", "localhost")
elasticsearch.user = os.getenv("ES_USER", "kube_system")
elasticsearch.password = os.getenv("ES_PASSWORD", "changeme")
elasticsearch.basic_auth = (elasticsearch.user, elasticsearch.password)
elasticsearch.port = os.getenv("ES_PORT", "9200")
elasticsearch.protocol = os.getenv("ES_PROTOCOL", "http")
elasticsearch.url = f"{elasticsearch.protocol}://{elasticsearch.hosts}:{elasticsearch.port}"
elasticsearch.cis_index = os.getenv("CIS_INDEX", "*cloud_security_posture.findings*")

# --- Docker environment definition
docker = Munch()
docker.base_url = os.getenv("DOCKER_URL", "")
docker.use_docker = bool(os.getenv("USE_DOCKER", "true") == "true")


def print_environment(name: str, config_object: Munch):
    """
    C-ATF configuration environment printer
    @param name: Config object name
    @param config_object: Object to be printed
    @return:
    """
    logger.info(f"==== Test Config Environment: {name} ====")
    for item in config_object:
        logger.info(f"{item}='{config_object[item]}'")
