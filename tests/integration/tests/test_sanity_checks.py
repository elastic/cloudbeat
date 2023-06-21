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

CONFIG_TIMEOUT = 120

tests_data = {
    "cis_aws": [
        "cloud-compute",
        "cloud-storage",
        "identity-management",
        "monitoring",
        "cloud-audit",
        "cloud-database",
        "cloud-config",
    ],
    "cis_k8s": ["file", "process", "k8s_object"],
    "cis_eks": ["process", "k8s_object"],  # Optimize search findings by excluding 'file'.
}


@pytest.mark.sanity
@pytest.mark.parametrize("match_type", tests_data["cis_k8s"])
def test_kspm_unmanaged_findings(elastic_client, match_type):
    """
    Test case to check for unmanaged findings in KSPM.

    Args:
        elastic_client: The elastic client object.
        match_type (str): The resource type to match.

    Returns:
        None

    Raises:
        AssertionError: If the resource type is missing.
    """
    query_list = [{"term": {"rule.benchmark.id": "cis_k8s"}}, {"term": {"resource.type": match_type}}]
    query, sort = elastic_client.build_es_must_match_query(must_query_list=query_list, time_range="now-4h")

    result = get_findings(elastic_client, CONFIG_TIMEOUT, query, sort, match_type)
    assert len(result) > 0, f"The resource type '{match_type}' is missing"


@pytest.mark.sanity
@pytest.mark.parametrize("match_type", tests_data["cis_eks"])
def test_kspm_e_k_s_findings(elastic_client, match_type):
    """
    Test case to check for EKS findings in KSPM.

    Args:
        elastic_client: The elastic client object.
        match_type (str): The resource type to match.

    Returns:
        None

    Raises:
        AssertionError: If the resource type is missing.
    """
    query_list = [{"term": {"rule.benchmark.id": "cis_eks"}}, {"term": {"resource.type": match_type}}]
    query, sort = elastic_client.build_es_must_match_query(must_query_list=query_list, time_range="now-4h")

    results = get_findings(elastic_client, CONFIG_TIMEOUT, query, sort, match_type)
    assert len(results) > 0, f"The resource type '{match_type}' is missing"


@pytest.mark.sanity
@pytest.mark.parametrize("match_type", tests_data["cis_aws"])
def test_cspm_findings(elastic_client, match_type):
    """
    Test case to check for AWS findings in CSPM.

    Args:
        elastic_client: The elastic client object.
        match_type (str): The resource type to match.

    Returns:
        None

    Raises:
        AssertionError: If the resource type is missing.
    """
    query_list = [{"term": {"rule.benchmark.id": "cis_aws"}}, {"term": {"resource.type": match_type}}]
    query, sort = elastic_client.build_es_must_match_query(must_query_list=query_list, time_range="now-24h")

    results = get_findings(elastic_client, CONFIG_TIMEOUT, query, sort, match_type)
    assert len(results) > 0, f"The resource type '{match_type}' is missing"
