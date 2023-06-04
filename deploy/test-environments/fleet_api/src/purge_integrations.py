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
import json
from munch import munchify
from loguru import logger
from api.agent_policy_api import delete_agent_policy
from api.package_policy_api import delete_package_policy
import configuration as cnfg
from utils import delete_file

state_data_file = cnfg.state_data_file

def purge_integrations():
    """
    Purge integrations and policies based on stored IDs in the state_data.json file.
    """
    # Check if the state_data.json file exists
    if not state_data_file.is_file():
        logger.error("state_data.json file does not exist.")
        return

    # Read the package policy IDs and agent policy IDs from the file
    try:
        with state_data_file.open("r") as state_file:
            policy_data = munchify(json.load(state_file))
    except FileNotFoundError:
        logger.error("state_data.json file not found.")
        return

    # Delete policies based on the stored IDs
    for policy in policy_data.policies:
        delete_package_policy(cfg=cnfg.elk_config, policy_ids=[policy.pkg_policy_id])
        delete_agent_policy(cfg=cnfg.elk_config, agent_policy_id=policy.agnt_policy_id)

    delete_file(state_data_file)

if __name__ == "__main__":
    logger.info("Start purging integrations and policies process...")
    purge_integrations()
    logger.info("Integration and policy purge process completed successfully.")
