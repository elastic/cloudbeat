"""
This is a basic test suite to validate the end-to-end (e2e) process.
The suite assumes that the environment is already deployed and all integrations,
such as KSPM (Kubernetes Security Posture Management) and CSPM (Cloud Security Posture Management),
have been created.
The goal of this suite is to perform basic sanity checks by querying Elasticsearch (ES) and
verifying that there are findings of 'resource.type' for each feature.
"""
import pytest

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
    "cis_eks": ["file", "process", "k8s_object"],
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
    query, sort = build_query_string(benchmark_id="cis_k8s", resource_type=match_type)

    result = elastic_client.get_index_data(
        index_name=elastic_client.index,
        query=query,
        size=1,
        sort=sort,
    )
    total_results = result["hits"]["total"]["value"]
    assert total_results > 0, f"The resource type '{match_type}' is missing"


@pytest.mark.sanity
@pytest.mark.parametrize("match_type", tests_data["cis_eks"])
def test_kspm_eks_findings(elastic_client, match_type):
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
    query, sort = build_query_string(benchmark_id="cis_eks", resource_type=match_type)

    result = elastic_client.get_index_data(
        index_name=elastic_client.index,
        query=query,
        size=1,
        sort=sort,
    )
    total_results = result["hits"]["total"]["value"]
    assert total_results > 0, f"The resource type '{match_type}' is missing"


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
    query, sort = build_query_string(benchmark_id="cis_aws", resource_type=match_type)

    result = elastic_client.get_index_data(
        index_name=elastic_client.index,
        query=query,
        size=1,
        sort=sort,
    )
    total_results = result["hits"]["total"]["value"]
    assert total_results > 0, f"The resource type '{match_type}' is missing"


def build_query_string(benchmark_id, resource_type):
    """
    Build the query string and sort parameters for querying Elasticsearch.

    Args:
        benchmark_id (str): The benchmark ID.
        resource_type (str): The resource type.

    Returns:
        tuple: A tuple containing the query and sort parameters.

    Example:
        query, sort = build_query_string("cis_aws", "monitoring")
    """
    query = {
        "bool": {
            "filter": [
                {"term": {"rule.benchmark.id": benchmark_id}},
                {"term": {"resource.type": resource_type}},
                {"range": {"@timestamp": {"gte": "now-24h"}}},
            ],
        },
    }

    sort = [{"@timestamp": {"order": "desc"}}]
    return query, sort
