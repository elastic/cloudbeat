"""
This is a basic test suite to validate the end-to-end (e2e) process.
The suite assumes that the environment is already deployed and all integrations,
such as KSPM (Kubernetes Security Posture Management) and CSPM (Cloud Security Posture Management),
have been created.
The goal of this suite is to perform basic sanity checks by querying Elasticsearch (ES) and
verifying that there are findings of 'resource.type' for each feature.
"""
import pytest
from commonlib.utils import get_findings, get_telemetry
from configuration import elasticsearch
from loguru import logger
from munch import munchify

CONFIG_TIMEOUT = 120
GCP_CONFIG_TIMEOUT = 600
CNVM_CONFIG_TIMEOUT = 3600

STACK_VERSION = elasticsearch.stack_version

# Check if STACK_VERSION is provided
if not STACK_VERSION:
    logger.warning("STACK_VERSION is not provided. Please set the STACK_VERSION in the configuration.")

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
        # Exclude "data-processing" due to lack of Dataproc assets in the test account.
    ],
    "cis_azure": [
        "configuration",
    ],  # Azure environment is not static, so we can't guarantee findings of all types.
    "cis_k8s": ["file", "process", "k8s_object"],
    "cis_eks": [
        "process",
        "k8s_object",
    ],  # Optimize search findings by excluding 'file'.
    "cnvm": ["vulnerability"],
}


def build_query_list(benchmark_id: str = "", match_type: str = "", version: str = "") -> list:
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
        version=STACK_VERSION,
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
        version=STACK_VERSION,
    )
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
    query_list = build_query_list(
        benchmark_id="cis_aws",
        match_type=match_type,
        version=STACK_VERSION,
    )
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
    query_list = build_query_list(version=STACK_VERSION)
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
    query_list = build_query_list(
        benchmark_id="cis_gcp",
        match_type=match_type,
        version=STACK_VERSION,
    )
    query, sort = cspm_client.build_es_must_match_query(must_query_list=query_list, time_range="now-24h")

    results = get_findings(cspm_client, GCP_CONFIG_TIMEOUT, query, sort, match_type)
    assert len(results) > 0, f"The resource type '{match_type}' is missing"


@pytest.mark.sanity
@pytest.mark.parametrize("match_type", tests_data["cis_azure"])
def test_cspm_azure_findings(cspm_client, match_type):
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
    query_list = build_query_list(benchmark_id="cis_azure", version=STACK_VERSION)
    query, sort = cspm_client.build_es_must_match_query(
        must_query_list=query_list,
        time_range="now-24h",
    )

    results = get_findings(cspm_client, CONFIG_TIMEOUT, query, sort, match_type)
    assert len(results) > 0, f"The resource type '{match_type}' is missing"


# @pytest.mark.parametrize("match_type", tests_data["cis_azure"])
@pytest.mark.sanity
def test_telemetry():
    """
    Test case to check telemetry is fetched as expected

    Raises:
        AssertionError: If one of the payload keys is missing.
    """
    telemetry_payload = get_telemetry(elasticsearch)

    telemetry_object = munchify(telemetry_payload[0])
    assert len(telemetry_object.stats.stack_stats.kibana.plugins.cloud_security_posture.keys()) > 0,\
        f"The telemetry payload is missing the cloud_security_posture key"

    cloud_security_telemetry_data = telemetry_object.stats.stack_stats.kibana.plugins.cloud_security_posture

    indices_stats = cloud_security_telemetry_data.indices
    indices = ["findings", "latest_findings", "vulnerabilities", "latest_vulnerabilities", "score"]

    ## indices stats
    for index in indices:
        assert indices_stats[index].doc_count > 0, \
            f"Expected {index} index to contain data"
    assert len(telemetry_payload[0]["stats"]["stack_stats"]["kibana"]["plugins"]["cloud_security_posture"]["indices"]["latestPackageVersion"]) > 0, \
            f"Package version is missing from telemetry payload"

    ## account stats
    cloud_account_stats = cloud_security_telemetry_data.cloud_account_stats
    for account in cloud_account_stats:
        if account.product == 'cspm':
            assert len(account.account_id) > 0 , f"Telemetry data missing account_id for cloud_account_stats {account}"
            assert len(account.cloud_provider) > 0 , f"Telemetry data missing cloud_provider for cloud_account_stats {account}"
            assert len(account.package_policy_id) > 0 , f"Telemetry data missing package_policy_id for cloud_account_stats {account}"
            assert len(account.product) > 0 , f"Telemetry data missing product for cloud_account_stats {account}"
            assert len(account.posture_management_stats.benchmark_name) > 0 , f"Telemetry data missing benchmark_name for cloud_account_stats {account}"
            assert len(account.posture_management_stats.benchmark_version) > 0 , f"Telemetry data missing benchmark_version for cloud_account_stats {account}"

    ## installation stats
    installation_stats = cloud_security_telemetry_data.installation_stats
    for installation in installation_stats:
        assert len(installation.package_policy_id) > 0, f"Telemetry data missing package_policy_id in installation stats"
        assert len(installation.feature) > 0, f"Telemetry data missing feature in installation_stats for {installation}"
        assert installation.agent_count > 0, f"Telemetry data missing agent in installation_stats for {installation}"
        assert len(installation.deployment_mode) > 0, f"Telemetry data missing account_type in installation_stats for {installation}"
        assert len(installation.package_version) > 0, f"Telemetry data missing package_version in installation_stats for {installation}"
        if installation.feature == 'cspm':
            assert isinstance(installation.is_setup_automatic, bool), f"Telemetry data missing is_setup_automatic in installation_stats for {installation}"
            # assert len(installation.setup_access_option) > 0, f"Telemetry data missing setup_access_option in installation_stats for {installation}"

