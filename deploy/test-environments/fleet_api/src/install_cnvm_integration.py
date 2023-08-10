#!/usr/bin/env python
"""
This script installs CNVM AWS integration

The following steps are performed:
1. Create an agent policy.
2. Create a CNVM AWS integration.
3. Create a deploy/cloudformation/config.json file to be used by the just deploy-cloudformation command.
"""
import json
from pathlib import Path
from typing import Dict, Tuple
from munch import Munch
import configuration_fleet as cnfg
from api.agent_policy_api import create_agent_policy
from api.package_policy_api import create_cnvm_integration
from api.common_api import (
    get_enrollment_token,
    get_fleet_server_host,
    get_artifact_server,
)
from loguru import logger
from utils import read_json
from state_file_manager import state_manager, PolicyState

CNVM_AGENT_POLICY = "../../../cloud/data/agent_policy_cnvm_aws.json"
CNVM_PACKAGE_POLICY = "../../../cloud/data/package_policy_cnvm_aws.json"
CNVM_EXPECTED_AGENTS = 1
CNVM_CLOUDFORMATION_CONFIG = "../../../cloudformation/config.json"
CNVM_AGENT_TAGS = ["cft_version:CFT_VERSION", "cft_arn:arn:aws:cloudformation:.*"]

cnvm_agent_policy_data = Path(__file__).parent / CNVM_AGENT_POLICY
cnvm_pkg_policy_data = Path(__file__).parent / CNVM_PACKAGE_POLICY
cnvm_cloudformation_config = Path(__file__).parent / CNVM_CLOUDFORMATION_CONFIG


def load_data() -> Tuple[Dict, Dict]:
    """Loads data.

    Returns:
        Tuple[Dict, Dict]: A tuple containing the loaded agent and package policies.
    """
    logger.info("Loading agent and package policies")
    agent_policy = read_json(json_path=cnvm_agent_policy_data)
    package_policy = read_json(json_path=cnvm_pkg_policy_data)
    return agent_policy, package_policy


if __name__ == "__main__":
    # pylint: disable=duplicate-code
    logger.info("Starting installation of CNVM AWS integration.")
    agent_data, package_data = load_data()

    logger.info("Create agent policy")
    agent_policy_id = create_agent_policy(cfg=cnfg.elk_config, json_policy=agent_data)

    logger.info("Create CNVM integration for policy", agent_policy_id)
    package_policy_id = create_cnvm_integration(
        cfg=cnfg.elk_config,
        pkg_policy=package_data,
        agent_policy_id=agent_policy_id,
    )

    state_manager.add_policy(PolicyState(agent_policy_id, package_policy_id, CNVM_EXPECTED_AGENTS, CNVM_AGENT_TAGS))

    cloudformation_params = Munch()
    cloudformation_params.ENROLLMENT_TOKEN = get_enrollment_token(
        cfg=cnfg.elk_config,
        policy_id=agent_policy_id,
    )

    cloudformation_params.FLEET_URL = get_fleet_server_host(cfg=cnfg.elk_config)
    cloudformation_params.ELASTIC_AGENT_VERSION = cnfg.elk_config.stack_version
    cloudformation_params.ELASTIC_ARTIFACT_SERVER = get_artifact_server(cnfg.elk_config.stack_version)

    with open(cnvm_cloudformation_config, "w") as file:
        json.dump(cloudformation_params, file)

    logger.info("Installation of CNVM integration is done")
