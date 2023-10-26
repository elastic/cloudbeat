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
from pathlib import Path
from loguru import logger
import configuration_fleet as cnfg
from api.agent_policy_api import (
    create_agent_download_source,
    get_agents,
    update_agent_policy,
)
from api.common_api import (
    get_artifact_server,
    bulk_upgrade_agents,
    wait_for_action_status,
)
from state_file_manager import state_manager, HostType

STATE_DATA_PATH = Path(__file__).parent / "state_data.json"


def create_custom_agent_download_source() -> str:
    """Create a custom agent download source and return its ID."""
    host_url = get_artifact_server(version=cnfg.elk_config.stack_version, is_short_url=True)
    download_source_id = create_agent_download_source(
        cfg=cnfg.elk_config,
        name="custom_source",
        host=host_url,
    )
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


def main():
    """
    Main linux agents upgrade flow
    """
    download_source_id = create_custom_agent_download_source()

    if not download_source_id:
        logger.error("Failed to create the agent download source.")
        sys.exit(1)

    linux_policies_list = update_linux_policies(download_source_id)

    agents = get_agents(cfg=cnfg.elk_config)
    linux_agent_ids = [agent.id for agent in agents if agent.policy_id in linux_policies_list]

    action_id = bulk_upgrade_agents(
        cfg=cnfg.elk_config,
        agent_ids=linux_agent_ids,
        version=cnfg.elk_config.stack_version,
        source_uri=get_artifact_server(version=cnfg.elk_config.stack_version),
    )

    wait_for_action_status(
        cfg=cnfg.elk_config,
        target_action_id=action_id,
        target_type="UPGRADE",
        target_status="COMPLETE",
    )


if __name__ == "__main__":
    main()
