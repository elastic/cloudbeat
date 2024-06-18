#!/usr/bin/env python
"""
Purge Integrations and Policies

This script is used to purge integrations and policies
based on the stored IDs in the state_data.json file.

The following steps are performed:
1. Read the package policy IDs and agent policy IDs from the state_data.json file.
2. Delete the package policies and agent policies based on the stored IDs.
3. Delete the state_data.json file.

Usage:
    python purge_integrations.py

"""
import configuration_fleet as cnfg
from fleet_api.agent_policy_api import (
    delete_agent_policy,
    get_agents,
    unenroll_agents_from_policy,
)
from fleet_api.package_policy_api import delete_package_policy
from loguru import logger
from state_file_manager import state_manager


def purge_integrations():
    """
    Purge integrations and policies based on stored IDs in the state_data.json file.
    """
    # Check if the state_data.json file exists

    agents = get_agents(cfg=cnfg.elk_config)
    # Delete policies based on the stored IDs
    for policy in state_manager.get_policies():
        logger.info("Deleting policy", policy.pkg_policy_id, policy.agnt_policy_id)
        delete_package_policy(cfg=cnfg.elk_config, policy_ids=[policy.pkg_policy_id])

        agents_list = [item.agent.id for item in agents if item.policy_id == policy.agnt_policy_id]
        if agents_list:
            unenroll_agents_from_policy(cfg=cnfg.elk_config, agents=agents_list)

        # Check if there is more than one package policy using the same agent policy
        agent_policy_id = policy.agnt_policy_id
        agent_policies = [p for p in state_manager.get_policies() if p.agnt_policy_id == agent_policy_id]
        if len(agent_policies) > 1:
            state_manager.delete_by_package_policy(pkg_policy_id=policy.pkg_policy_id)
            continue

        delete_agent_policy(cfg=cnfg.elk_config, agent_policy_id=policy.agnt_policy_id)

    state_manager.delete_all()


if __name__ == "__main__":
    logger.info("Start purging integrations and policies process...")
    purge_integrations()
    logger.info("Integration and policy purge process completed successfully.")
