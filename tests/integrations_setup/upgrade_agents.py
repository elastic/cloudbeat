#!/usr/bin/env python
"""
This script upgrades Linux-based agents.

The following steps are performed:
1. Generate a custom agent binary download URL.
2. Update all Linux-based agent policies with the custom download URL.
3. Execute a bulk upgrade process for all agents.
4. Wait until all agent upgrades are complete.

Note: This script requires a 'state_data.json' file to identify all Linux agents to be updated.

For execution, create a configuration file 'cnvm_config.json' in the same directory.

Example 'state_data.json':
{
    "policies": [
        {
            "agnt_policy_id": "c3a6d9d0-6b58-11ee-8fd8-b709d88b5892",
            "pkg_policy_id": "226965a4-e07a-4ddd-a64d-765ddd9946e5",
            "expected_agents": 1,
            "expected_tags": [
                "cft_version:cft_version",
                "cft_arn:arn:aws:cloudformation:.*"
            ],
            "type": "linux",
            "integration_name": "cnvm-int"
        }
    ]
}
"""

import sys
import time
from pathlib import Path

import configuration_fleet as cnfg
from fleet_api.agent_policy_api import (
    create_agent_download_source,
    get_agents,
    update_agent_policy,
)
from fleet_api.common_api import (
    bulk_upgrade_agents,
    get_artifact_server,
    get_package_version,
    update_package_version,
    wait_for_action_status,
)
from fleet_api.package_policy_api import get_package_policy_by_id
from loguru import logger
from state_file_manager import HostType, state_manager

STATE_DATA_PATH = Path(__file__).parent / "state_data.json"


def create_custom_agent_download_source() -> str:
    """Create a custom agent download source and return its ID."""
    host_url = get_artifact_server(version=cnfg.elk_config.stack_version, is_short_url=True)
    download_source_id = create_agent_download_source(
        cfg=cnfg.elk_config,
        name="custom_source",
        host=host_url,
    )
    logger.info(f"Download source id '{download_source_id}' is created")
    return download_source_id


def update_linux_policies(download_source_id: str):
    """Update all Linux-based agent policies with the custom download source."""
    state_policies = state_manager.get_policies()
    linux_policies_list = []

    for policy in state_policies:
        if policy.host_type == HostType.LINUX_TAR.value:
            linux_policies_list.append(policy.agnt_policy_id)
            update_agent_policy(
                cfg=cnfg.elk_config,
                policy_id=policy.agnt_policy_id,
                json_policy={
                    "name": policy.integration_name,
                    "namespace": "default",
                    "download_source_id": download_source_id,
                },
            )

    return linux_policies_list


def wait_for_packages_upgrade():
    """
    This function waits until all packages version is upgraded.
    """
    desired_version = get_package_version(cfg=cnfg.elk_config)
    policies = state_manager.get_policies()
    for policy in policies:
        if policy.integration_name == "tf-ap-d4c":
            continue
        if not wait_for_package_policy_version(
            cfg=cnfg.elk_config,
            policy_id=policy.pkg_policy_id,
            desired_version=desired_version,
        ):
            logger.error(f"Integration {policy.integration_name} failed to upgrade.")
            sys.exit(1)


def wait_for_package_policy_version(
    cfg,
    policy_id,
    desired_version,
    timeout_secs=300,
    poll_interval_secs=10,
):
    """
    Wait for a package policy to reach the desired version with a timeout.

    Args:
        cfg (Munch): A configuration object containing Kibana URL, authentication details, etc.
        policy_id (str): The package policy ID to monitor.
        desired_version (str): The desired version to wait for.
        timeout_secs (int, optional): Maximum time to wait in seconds. Default is 300 seconds.
        poll_interval_secs (int, optional): Time to wait between polling for the package version.
                                            Default is 10 seconds.

    Returns:
        bool: True if the package policy reaches the desired version within the timeout,
              False otherwise.
    """
    start_time = time.time()

    while time.time() - start_time < timeout_secs:
        policy_info = get_package_policy_by_id(cfg, policy_id)
        policy_name = policy_info.get("name", "")
        policy_version = policy_info.get("package", {}).get("version", "")
        logger.info(
            f"Integration: {policy_name}, current version: {policy_version}, desired version: {desired_version}",
        )
        if policy_version == desired_version:
            return True  # Desired version reached

        time.sleep(poll_interval_secs)  # Wait and poll again

    return False  # Desired version not reached within the timeout


def main():
    """
    Main linux agents upgrade flow
    """
    # If the version is not released, the package version should be updated manually
    update_package_version(
        cfg=cnfg.elk_config,
        package_name="cloud_security_posture",
        package_version=get_package_version(cfg=cnfg.elk_config),
    )
    # Ensure that all packages are on the latest version
    wait_for_packages_upgrade()

    download_source_id = create_custom_agent_download_source()

    if not download_source_id:
        logger.error("Failed to create the agent download source.")
        sys.exit(1)

    linux_policies_list = update_linux_policies(download_source_id)
    time.sleep(180)  # To ensure that policies updated
    agents = get_agents(cfg=cnfg.elk_config)
    linux_agent_ids = [agent.id for agent in agents if agent.policy_id in linux_policies_list]
    for agent_id in linux_agent_ids:
        action_id = bulk_upgrade_agents(
            cfg=cnfg.elk_config,
            agent_ids=agent_id,
            version=cnfg.elk_config.stack_version,
            source_uri=get_artifact_server(version=cnfg.elk_config.stack_version),
        )

        if not wait_for_action_status(
            cfg=cnfg.elk_config,
            target_action_id=action_id,
            target_type="UPGRADE",
            target_status="COMPLETE",
        ):
            logger.error("Failed to complete the upgrade action within the expected timeframe.")
            sys.exit(1)
        logger.info(f"Agent {agent_id} upgrade is finished")


if __name__ == "__main__":
    main()
