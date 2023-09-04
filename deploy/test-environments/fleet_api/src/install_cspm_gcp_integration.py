#!/usr/bin/env python
"""
This script installs CSPM GCP integration

The following steps are performed:
1. Create an agent policy.
2. Create a CSPM GCP integration.
3. Create a deploy/deployment-manager/config.json file to be used by the just deploy-dm command.
"""
import json
from pathlib import Path
from typing import Dict, Tuple
from munch import Munch
import configuration_fleet as cnfg
from api.agent_policy_api import create_agent_policy
from api.package_policy_api import create_cspm_integration
from api.common_api import (
    get_enrollment_token,
    get_fleet_server_host,
    get_artifact_server,
    get_package_version,
    update_package_version,
)
from loguru import logger
from utils import read_json
from state_file_manager import state_manager, PolicyState

CSPM_GCP_AGENT_POLICY = "../../../cloud/data/agent_policy_cspm_gcp.json"
CSPM_GCP_PACKAGE_POLICY = "../../../cloud/data/package_policy_cspm_gcp.json"
CSPM_GCP_EXPECTED_AGENTS = 1
DEPLOYMENT_MANAGER_CONFIG = "../../../deployment-manager/config.json"

cspm_gcp_agent_policy_data = Path(__file__).parent / CSPM_GCP_AGENT_POLICY
cspm_gcp_pkg_policy_data = Path(__file__).parent / CSPM_GCP_PACKAGE_POLICY
cspm_gcp_deployment_manager_config = Path(__file__).parent / DEPLOYMENT_MANAGER_CONFIG
INTEGRATION_NAME = "CSPM GCP"


def load_data() -> Tuple[Dict, Dict]:
    """Loads data.

    Returns:
        Tuple[Dict, Dict]: A tuple containing the loaded agent and package policies.
    """
    logger.info("Loading agent and package policies")
    agent_policy = read_json(json_path=cspm_gcp_agent_policy_data)
    package_policy = read_json(json_path=cspm_gcp_pkg_policy_data)
    return agent_policy, package_policy


if __name__ == "__main__":
    # pylint: disable=duplicate-code
    package_version = get_package_version(cfg=cnfg.elk_config)
    logger.info(f"Package version: {package_version}")
    update_package_version(
        cfg=cnfg.elk_config,
        package_name="cloud_security_posture",
        package_version=package_version,
    )
    logger.info(f"Starting installation of {INTEGRATION_NAME} integration.")
    agent_data, package_data = load_data()

    logger.info("Create agent policy")
    agent_policy_id = create_agent_policy(cfg=cnfg.elk_config, json_policy=agent_data)

    logger.info(f"Create {INTEGRATION_NAME} integration for policy {agent_policy_id}")
    package_policy_id = create_cspm_integration(
        cfg=cnfg.elk_config,
        pkg_policy=package_data,
        agent_policy_id=agent_policy_id,
        cspm_data={},
    )

    state_manager.add_policy(
        PolicyState(
            agent_policy_id,
            package_policy_id,
            CSPM_GCP_EXPECTED_AGENTS,
            [],
        ),
    )

    deploy_manager_params = Munch()
    deploy_manager_params.ENROLLMENT_TOKEN = get_enrollment_token(
        cfg=cnfg.elk_config,
        policy_id=agent_policy_id,
    )

    deploy_manager_params.FLEET_URL = get_fleet_server_host(cfg=cnfg.elk_config)
    deploy_manager_params.ELASTIC_ARTIFACT_SERVER = get_artifact_server(cnfg.elk_config.stack_version)
    deploy_manager_params.DEPLOYMENT_NAME = cnfg.gcp_dm_config.deployment_name
    deploy_manager_params.ZONE = cnfg.gcp_dm_config.zone
    deploy_manager_params.ALLOW_SSH = cnfg.gcp_dm_config.allow_ssh
    deploy_manager_params.STACK_VERSION = cnfg.elk_config.stack_version

    with open(cspm_gcp_deployment_manager_config, "w") as file:
        json.dump(deploy_manager_params, file)

    logger.info(f"Installation of {INTEGRATION_NAME} integration is done")
