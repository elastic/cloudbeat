"""
This module provides common configuration of the test project,
and also mapping environment variables

"""

import os

from loguru import logger
from munch import Munch

FINDINGS_INDEX_PATTERN = "*cloud_security_posture.findings*"
VULNERABILITIES_INDEX_PATTERN = "*cloud_security_posture.vulnerabilities*"
ASSET_INVENTORY_INDEX_PATTERN = "*cloud_asset_inventory.asset_inventory*"

# --- Cloudbeat agent environment definition ----------------
agent = Munch()
agent.name = os.getenv("AGENT_NAME", "cloudbeat")
agent.namespace = os.getenv("AGENT_NAMESPACE", "kube-system")
agent.findings_timeout = 120
agent.eks_findings_timeout = 120
agent.aws_findings_timeout = 10
agent.azure_findings_timeout = 10
agent.cluster_type = os.getenv("CLUSTER_TYPE", "eks")  # options: vanilla / eks / vanilla_aws
agent.agentless = os.getenv("AGENTLESS", "false") == "false"

# The K8S Node on which the test Pod is running.
agent.node_name = os.getenv("NODE_NAME")

# --- Kubernetes environment definition --------------------
kubernetes = Munch()
kubernetes.is_in_cluster_config = bool(
    os.getenv("KUBERNETES_IN_CLUSTER", "false") == "true",
)
kubernetes.use_kubernetes = bool(
    os.getenv("USE_K8S", "true") == "true",
)
kubernetes.current_config = os.getenv("CLUSTER_CONFIG", "test-k8s-config-1")
kubernetes.config_1 = os.getenv("K8S_CONFIG_1", "test-k8s-config-1")
kubernetes.config_2 = os.getenv("K8S_CONFIG_2", "test-k8s-config-2")

# --- AWS EKS ---------------------------------------------
eks = Munch()
eks.current_config = os.getenv("CLUSTER_CONFIG", "test-eks-config-1")
eks.config_1 = os.getenv("EKS_CONFIG_1", "test-eks-config-1")
eks.config_1_node_1 = os.getenv(
    "EKS_CONFIG_1_NODE_1",
    "ip-192-168-15-75.eu-west-2.compute.internal",
)
eks.config_1_node_2 = os.getenv(
    "EKS_CONFIG_1_NODE_2",
    "ip-192-168-38-87.eu-west-2.compute.internal",
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
elasticsearch.use_ssl = os.getenv("ES_USE_SSL", "true") == "true"
elasticsearch.url = os.getenv("ES_URL", f"{elasticsearch.protocol}://{elasticsearch.hosts}:{elasticsearch.port}")
elasticsearch.kibana_url = os.getenv("KIBANA_URL", "")
elasticsearch.kspm_index = os.getenv("KSPM_INDEX", FINDINGS_INDEX_PATTERN)
elasticsearch.cspm_index = os.getenv("CSPM_INDEX", FINDINGS_INDEX_PATTERN)
elasticsearch.cnvm_index = os.getenv("CNVM_INDEX", VULNERABILITIES_INDEX_PATTERN)
elasticsearch.asset_inventory_index = os.getenv("ASSET_INVENTORY_INDEX", ASSET_INVENTORY_INDEX_PATTERN)
elasticsearch.stack_version = os.getenv("STACK_VERSION", "")
elasticsearch.agent_version = os.getenv("AGENT_VERSION", "")

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
