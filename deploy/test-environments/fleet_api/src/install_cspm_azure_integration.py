#!/usr/bin/env python
"""
This script installs CSPM Azure integration

The following steps are performed:
1. Create an agent policy.
2. Create a CSPM Azure integration.
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
from state_file_manager import state_manager, PolicyState, HostType

CSPM_AZURE_AGENT_POLICY = "../../../cloud/data/agent_policy_cspm_azure.json"
CSPM_AZURE_PACKAGE_POLICY = "../../../cloud/data/package_policy_cspm_azure.json"
CSPM_AZURE_EXPECTED_AGENTS = 1
AZURE_ARM_PARAMETERS = "../../../azure/arm_parameters.json"

cspm_azure_agent_policy_data = Path(__file__).parent / CSPM_AZURE_AGENT_POLICY
cspm_azure_pkg_policy_data = Path(__file__).parent / CSPM_AZURE_PACKAGE_POLICY
cspm_azure_arm_parameters = Path(__file__).parent / AZURE_ARM_PARAMETERS
INTEGRATION_NAME = "CSPM Azure"


def load_data() -> Tuple[Dict, Dict]:
    """Loads data.

    Returns:
        Tuple[Dict, Dict]: A tuple containing the loaded agent and package policies.
    """
    logger.info("Loading agent and package policies")
    agent_policy = read_json(json_path=cspm_azure_agent_policy_data)
    package_policy = read_json(json_path=cspm_azure_pkg_policy_data)
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
            CSPM_AZURE_EXPECTED_AGENTS,
            [],
            HostType.LINUX_TAR.value,
            package_data.get("name", ""),
        ),
    )

    azure_arm_parameters = Munch(
        {
            "$schema": "https://schema.management.azure.com/schemas/2019-04-01/deploymentParameters.json#",
            "contentVersion": "1.0.0.0",
            "parameters": {},
        },
    )
    azure_arm_parameters["parameters"]["EnrollmentToken"] = {
        "value": get_enrollment_token(
            cfg=cnfg.elk_config,
            policy_id=agent_policy_id,
        ),
    }

    azure_arm_parameters["parameters"]["FleetUrl"] = {
        "value": get_fleet_server_host(cfg=cnfg.elk_config),
    }

    azure_arm_parameters["parameters"]["ElasticArtifactServer"] = {
        "value": get_artifact_server(cnfg.elk_config.stack_version),
    }

    azure_arm_parameters["parameters"]["ElasticAgentVersion"] = {
        "value": cnfg.elk_config.stack_version,
    }

    azure_arm_parameters["parameters"]["ResourceGroupName"] = {
        "value": cnfg.azure_arm_parameters.deployment_name,
    }

    azure_arm_parameters["parameters"]["Location"] = {
        "value": cnfg.azure_arm_parameters.location,
    }

    with open(cspm_azure_arm_parameters, "w") as file:
        json.dump(azure_arm_parameters, file)

    logger.info(f"Installation of {INTEGRATION_NAME} integration is done")
