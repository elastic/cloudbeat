"""
This is a basic test suite to validate the end-to-end (e2e) process.
The suite assumes that the environment is already deployed and all integrations,
such as KSPM (Kubernetes Security Posture Management) and CSPM (Cloud Security Posture Management),
have been created.
The goal of this suite is to perform basic sanity checks by querying Elasticsearch (ES) and
verifying that there are findings of 'resource.type' for each feature.
"""
import pytest
from commonlib.utils import get_findings
from configuration import elasticsearch

CONFIG_TIMEOUT = 120
GCP_CONFIG_TIMEOUT = 600
CNVM_CONFIG_TIMEOUT = 3600

STACK_VERSION = elasticsearch.stack_version
# Check if STACK_VERSION is provided
if not STACK_VERSION:
    raise ValueError("STACK_VERSION is not provided. Please set the STACK_VERSION in the configuration.")

tests_data = {
    "cis_aws": [
        "cloud-compute",
        "identity-management",
        "monitoring",
        "cloud-audit",
        "cloud-database",
        "cloud-config",
    ],  # Exclude "cloud-storage" due to lack of fetcher control and potential delays.
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
    "cis_k8s": ["file", "process", "k8s_object"],
    "cis_eks": ["process", "k8s_object"],  # Optimize search findings by excluding 'file'.
    "cnvm": ["vulnerability"],
}


@pytest.mark.sanity
@pytest.mark.upgrade
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
    query_list = [
        {"term": {"rule.benchmark.id": "cis_k8s"}},
        {"term": {"resource.type": match_type}},
        {"term": {"agent.version": STACK_VERSION}},
    ]
    query, sort = kspm_client.build_es_must_match_query(must_query_list=query_list, time_range="now-4h")

    result = get_findings(kspm_client, CONFIG_TIMEOUT, query, sort, match_type)
    assert len(result) > 0, f"The resource type '{match_type}' is missing"


@pytest.mark.sanity
@pytest.mark.upgrade
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
    query_list = [
        {"term": {"rule.benchmark.id": "cis_eks"}},
        {"term": {"resource.type": match_type}},
        {"term": {"agent.version": STACK_VERSION}},
    ]
    query, sort = kspm_client.build_es_must_match_query(must_query_list=query_list, time_range="now-4h")

    results = get_findings(kspm_client, CONFIG_TIMEOUT, query, sort, match_type)
    assert len(results) > 0, f"The resource type '{match_type}' is missing"


@pytest.mark.sanity
@pytest.mark.parametrize("match_type", tests_data["cis_aws"])
def test_cspm_findings(cspm_client, match_type):
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
    query_list = [{"term": {"rule.benchmark.id": "cis_aws"}}, {"term": {"resource.type": match_type}}]
    query, sort = cspm_client.build_es_must_match_query(must_query_list=query_list, time_range="now-24h")

    results = get_findings(cspm_client, CONFIG_TIMEOUT, query, sort, match_type)
    assert len(results) > 0, f"The resource type '{match_type}' is missing"


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
    query_list = []
    query, sort = cnvm_client.build_es_must_match_query(must_query_list=query_list, time_range="now-24h")
    results = get_findings(cnvm_client, CNVM_CONFIG_TIMEOUT, query, sort, match_type)
    assert len(results) > 0, f"The resource type '{match_type}' is missing"


@pytest.mark.sanity
@pytest.mark.parametrize("match_type", tests_data["cis_gcp"])
def test_cspm_gcp_findings(cspm_client, match_type):
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
    query_list = [{"term": {"rule.benchmark.id": "cis_gcp"}}, {"term": {"resource.type": match_type}}]
    query, sort = cspm_client.build_es_must_match_query(must_query_list=query_list, time_range="now-24h")

    results = get_findings(cspm_client, GCP_CONFIG_TIMEOUT, query, sort, match_type)
    assert len(results) > 0, f"The resource type '{match_type}' is missing"
