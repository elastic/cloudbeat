"""
This is a basic test suite to validate the end-to-end (e2e) process.
The suite assumes that the environment is already deployed and all integrations,
such as KSPM (Kubernetes Security Posture Management) and CSPM (Cloud Security Posture Management),
have been created.
The goal of this suite is to perform basic sanity checks by querying Elasticsearch (ES) and
verifying that there are findings of 'resource.type' for each feature.
"""

import time

import pytest
from commonlib.agents_map import (
    CIS_AWS_COMPONENT,
    CIS_AZURE_COMPONENT,
    CIS_GCP_COMPONENT,
    AgentComponentMapping,
    AgentExpectedMapping,
)
from commonlib.utils import get_findings
from configuration import elasticsearch
from loguru import logger

CONFIG_TIMEOUT = 120
GCP_CONFIG_TIMEOUT = 600
CNVM_CONFIG_TIMEOUT = 3600

# The timeout and backoff for waiting all agents are running the specified component.
COMPONENTS_TIMEOUT = 300
COMPONENTS_BACKOFF = 10

AGENT_VERSION = elasticsearch.agent_version
if AGENT_VERSION.endswith("SNAPSHOT"):
    AGENT_VERSION = AGENT_VERSION.split("-")[0]

# Check if AGENT_VERSION is provided
if not AGENT_VERSION:
    logger.warning("AGENT_VERSION is not provided. Please set the AGENT_VERSION in the configuration.")

tests_data = {
    "cis_aws": [
        "identity-management",
        "monitoring",
        "cloud-audit",
        "cloud-database",
        "cloud-config",
        "cloud-compute",
        "cloud-storage",
    ],
    "cis_gcp": [
        "cloud-compute",
        "cloud-database",
        "key-management",
        "identity-management",
        "monitoring",
        "cloud-storage",
        "cloud-dns",
        "project-management",
        "data-processing",
    ],
    "cis_azure": [
        "cloud-storage",
        "cloud-database",
        "key-management",
        "monitoring",
        "cloud-dns",
    ],  # Exclude "cloud-compute", Azure environment is not static, so we can't guarantee findings of all types.
    "cis_k8s": ["file", "process", "k8s_object"],
    "cis_eks": [
        "file",
        "process",
        "k8s_object",
    ],
    "cnvm": ["vulnerability"],
}


def build_query_list(benchmark_id: str = "", match_type: str = "", version: str = "", agent: str = "") -> list:
    """
    Build a list of terms for Elasticsearch query based on the provided parameters.

    Parameters:
    - benchmark_id (str, optional): The benchmark ID for filtering.
    - match_type (str, optional): The resource type for filtering.
    - version (str, optional): The agent version for filtering.

    Returns:
    list: A list of terms for Elasticsearch query.
    """
    query_list = []
    if benchmark_id:
        query_list.append({"term": {"rule.benchmark.id": benchmark_id}})

    if match_type:
        query_list.append({"term": {"resource.type": match_type}})

    if version:
        query_list.append({"term": {"agent.version": version}})

    if agent:
        query_list.append({"term": {"agent.id": agent}})

    return query_list


@pytest.mark.sanity
@pytest.mark.parametrize("match_type", tests_data["cis_k8s"])
def test_kspm_unmanaged_findings(kspm_client, match_type):
    """
    Test case to check for unmanaged findings in KSPM.

    Args:
        kspm_client: The kspm client object.
        match_type (str): The resource type to match.

    Returns:
        None

    Raises:
        AssertionError: If the resource type is missing.
    """
    query_list = build_query_list(
        benchmark_id="cis_k8s",
        match_type=match_type,
        version=AGENT_VERSION,
    )
    query, sort = kspm_client.build_es_must_match_query(must_query_list=query_list, time_range="now-4h")
    result = get_findings(kspm_client, CONFIG_TIMEOUT, query, sort, match_type)
    assert len(result) > 0, f"The resource type '{match_type}' is missing"


@pytest.mark.sanity
@pytest.mark.parametrize("match_type", tests_data["cis_eks"])
def test_kspm_e_k_s_findings(kspm_client, match_type):
    """
    Test case to check for EKS findings in KSPM.

    Args:
        kspm_client: The elastic client object.
        match_type (str): The resource type to match.

    Returns:
        None

    Raises:
        AssertionError: If the resource type is missing.
    """
    query_list = build_query_list(
        benchmark_id="cis_eks",
        match_type=match_type,
        version=AGENT_VERSION,
    )
    query, sort = kspm_client.build_es_must_match_query(must_query_list=query_list, time_range="now-4h")

    results = get_findings(kspm_client, CONFIG_TIMEOUT, query, sort, match_type)
    assert len(results) > 0, f"The resource type '{match_type}' is missing"


@pytest.mark.sanity
@pytest.mark.agentless
@pytest.mark.parametrize("match_type", tests_data["cis_aws"])
def test_cspm_aws_findings(
    cspm_client,
    match_type,
    agents_actual_components: AgentComponentMapping,
    agents_expected_components: AgentExpectedMapping,
):
    """
    Test case to check for AWS findings in CSPM.

    Args:
        cspm_client: The elastic client object.
        match_type (str): The resource type to match.

    Returns:
        None

    Raises:
        AssertionError: If the resource type is missing.
    """
    aws_agents = wait_components_list(agents_actual_components, agents_expected_components, CIS_AWS_COMPONENT)
    for agent in aws_agents:
        query_list = build_query_list(
            benchmark_id="cis_aws",
            match_type=match_type,
            agent=agent,
        )
        query, sort = cspm_client.build_es_must_match_query(must_query_list=query_list, time_range="now-24h")

        results = get_findings(cspm_client, CONFIG_TIMEOUT, query, sort, match_type)
        assert len(results) > 0, f"The resource type '{match_type}' is missing for agent {agent}"


@pytest.mark.sanity
@pytest.mark.parametrize("match_type", tests_data["cnvm"])
def test_cnvm_findings(cnvm_client, match_type):
    """
    Test case to check for vulnerabilities found by CNVM.

    Args:
        cnvm_client: The elastic client object.
        match_type (str): The resource type to match.

    Returns:
        None

    Raises:
        AssertionError: If the resource type is missing.
    """
    query_list = build_query_list(version=AGENT_VERSION)
    query, sort = cnvm_client.build_es_must_match_query(must_query_list=query_list, time_range="now-24h")
    results = get_findings(cnvm_client, CNVM_CONFIG_TIMEOUT, query, sort, match_type)
    assert len(results) > 0, f"The resource type '{match_type}' is missing"
    # Check every finding has host section
    for finding in results["hits"]["hits"]:
        resource = finding["_source"]
        assert "host" in resource, f"Resource '{match_type}' is missing 'host' section"
        assert "name" in resource["host"], f"Resource '{match_type}' is missing 'host.name'"


@pytest.mark.sanity
@pytest.mark.agentless
@pytest.mark.parametrize("match_type", tests_data["cis_gcp"])
def test_cspm_gcp_findings(
    cspm_client,
    match_type,
    agents_actual_components: AgentComponentMapping,
    agents_expected_components: AgentExpectedMapping,
):
    """
    Test case to check for GCP findings in CSPM.

    Args:
        cspm_client: The elastic client object.
        match_type (str): The resource type to match.

    Returns:
        None

    Raises:
        AssertionError: If the resource type is missing.
    """
    gcp_agents = wait_components_list(agents_actual_components, agents_expected_components, CIS_GCP_COMPONENT)
    for agent in gcp_agents:
        query_list = build_query_list(
            benchmark_id="cis_gcp",
            match_type=match_type,
            agent=agent,
        )
        query, sort = cspm_client.build_es_must_match_query(must_query_list=query_list, time_range="now-24h")

        results = get_findings(cspm_client, GCP_CONFIG_TIMEOUT, query, sort, match_type)
        assert len(results) > 0, f"The resource type '{match_type}' is missing for agent {agent}"


@pytest.mark.sanity
@pytest.mark.agentless
@pytest.mark.parametrize("match_type", tests_data["cis_azure"])
def test_cspm_azure_findings(
    cspm_client,
    match_type,
    agents_actual_components: AgentComponentMapping,
    agents_expected_components: AgentExpectedMapping,
):
    """
    Test case to check for Azure findings in CSPM.

    Args:
        cspm_client: The elastic client object.
        match_type (str): The resource type to match.

    Returns:
        None

    Raises:
        AssertionError: If the resource type is missing.
    """
    azure_agents = wait_components_list(agents_actual_components, agents_expected_components, CIS_AZURE_COMPONENT)
    for agent in azure_agents:
        query_list = build_query_list(
            benchmark_id="cis_azure",
            match_type=match_type,
            agent=agent,
        )
        query, sort = cspm_client.build_es_must_match_query(must_query_list=query_list, time_range="now-24h")

        results = get_findings(cspm_client, CONFIG_TIMEOUT, query, sort, match_type)
        assert len(results) > 0, f"The resource type '{match_type}' is missing for agent {agent}"


def wait_components_list(actual: AgentComponentMapping, expected: AgentExpectedMapping, component: str) -> list[str]:
    """
    Wait for the list of agents running the specified component.

    Args:
        component (str): The component to match.

    Returns:
        list: The list of agents running the specified component.
    """

    # Skip waiting for agents if fleet is not available.
    if not elasticsearch.kibana_url:
        return [""]

    actual.load_map()
    start_time = time.time()
    while time.time() - start_time < COMPONENTS_TIMEOUT:
        if len(actual.component_map[component]) == expected.expected_map[component]:
            break

        time.sleep(COMPONENTS_BACKOFF)
        actual.load_map()

    assert expected.expected_map[component] == len(
        actual.component_map[component],
    ), f"Expected {expected.expected_map[component]} agents running\
 {component}, but got {len(actual.component_map[component])}"

    return actual.component_map[component]
