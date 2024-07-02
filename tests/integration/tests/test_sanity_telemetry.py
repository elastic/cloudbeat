"""
This test suite goal is to validate end-to-end cloud security telemetry payloads.
The suite assumes that the environment is already deployed and all integrations,
such as KSPM (Kubernetes Security Posture Management) and CSPM (Cloud Security Posture Management),
have been created.
The goal of this suite is to perform basic sanity checks to verify the telemetry fetchers are working
 as expected.
"""

import pytest
from fleet_api.common_api import get_telemetry
from integrations_setup.configuration_fleet import elk_config
from munch import munchify


@pytest.fixture(scope="module", name="cloud_security_telemetry_data")
def get_cloud_security_telemetry_data():
    """Fixture to fetch telemetry data"""
    telemetry_payload = get_telemetry(elk_config)
    telemetry_object = munchify(telemetry_payload[0])
    return telemetry_object.stats.stack_stats.kibana.plugins.cloud_security_posture


@pytest.mark.sanity
def test_telemetry_indices(cloud_security_telemetry_data):
    """
    Test case to check telemetry indices are fetched as expected

    Raises:
        AssertionError: If one of the payload keys is missing.
    """
    indices_stats = cloud_security_telemetry_data.indices
    indices = [
        "findings",
        "latest_findings",
        "vulnerabilities",
        # "latest_vulnerabilities",  # https://github.com/elastic/security-team/issues/8252
        "score",
    ]

    # indices stats
    for index in indices:
        assert indices_stats[index].doc_count > 0, f"Expected {index} index to contain data"


@pytest.mark.sanity
def test_telemetry_cloud_account_stats(cloud_security_telemetry_data):
    """
    Test case to check telemetry cloud account stats are fetched as expected

    Raises:
        AssertionError: If one of the payload keys is missing.
    """
    # account stats
    cloud_account_stats = cloud_security_telemetry_data.cloud_account_stats
    for account in cloud_account_stats:
        assert len(account.account_id) > 0, f"Telemetry data missing account_id for cloud_account_stats {account}"
        assert len(account.product) > 0, f"Telemetry data missing product for cloud_account_stats {account}"
        assert (
            len(account.package_policy_id) > 0
        ), f"Telemetry data missing package_policy_id for cloud_account_stats {account}"

        if not (account.product == "kspm" and "CIS Kubernetes" in account.posture_management_stats.benchmark_name):
            assert (
                len(account.cloud_provider) > 0
            ), f"Telemetry data missing cloud_provider for cloud_account_stats {account}"
        if account.product != "vuln_mgmt":
            assert (
                len(account.posture_management_stats.benchmark_name) > 0
            ), f"Telemetry data missing benchmark_name for cloud_account_stats {account}"
            assert (
                len(account.posture_management_stats.benchmark_version) > 0
            ), f"Telemetry data missing benchmark_version for cloud_account_stats {account}"


@pytest.mark.sanity
def test_telemetry_installation_stats(cloud_security_telemetry_data):
    """
    Test case to check telemetry installation stats are fetched as expected

    Raises:
        AssertionError: If one of the payload keys is missing.
    """
    # installation stats
    installation_stats = cloud_security_telemetry_data.installation_stats
    for installation in installation_stats:
        assert len(installation.package_policy_id) > 0, "Telemetry data missing package_policy_id in installation stats"
        assert len(installation.feature) > 0, f"Telemetry data missing feature in installation_stats for {installation}"
        assert installation.agent_count > 0, f"Telemetry data missing agent in installation_stats for {installation}"
        assert (
            len(installation.package_version) > 0
        ), f"Telemetry data missing package_version in installation_stats for {installation}"
        assert (
            len(installation.deployment_mode) > 0
        ), f"Telemetry data missing deployment_mode in installation_stats for {installation}"
        if installation.feature == "cspm":
            assert isinstance(
                installation.is_setup_automatic,
                bool,
            ), f"Telemetry data missing is_setup_automatic in installation_stats for {installation}"
            assert (
                len(installation.account_type) > 0
            ), f"Telemetry data missing account_type in installation_stats for {installation}"
