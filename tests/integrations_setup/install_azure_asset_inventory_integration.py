#!/usr/bin/env python
"""
This script installs Asset Inventory Azure integration

The following steps are performed:
1. Create an agent policy.
2. Create a CSPM Azure integration.
"""
import json
import sys
from pathlib import Path

import configuration_fleet as cnfg
from fleet_api.agent_policy_api import create_agent_policy
from fleet_api.common_api import (
    get_arm_template,
    get_artifact_server,
    get_enrollment_token,
    get_fleet_server_host,
    get_package_version,
    update_package_version,
)
from fleet_api.package_policy_api import create_integration
from fleet_api.utils import rename_file_by_suffix
from loguru import logger
from munch import Munch
from package_policy import (
    VERSION_MAP,
    extract_arm_template_url,
    generate_random_name,
    get_package_default_url,
    load_data,
    version_compatible,
)
from state_file_manager import HostType, PolicyState, state_manager

EXPECTED_AGENTS = 1
AGENT_POLICY = "../../deploy/cloud/data/agent_policy_asset_inventory_azure.json"
PACKAGE_POLICY = "../../deploy/cloud/data/package_policy_asset_inventory_azure.json"
ARM_PARAMETERS = "../../deploy/azure/arm_parameters.json"
ARM_TEMPLATE = "../../deploy/azure/ARM-for-single-account.json"
PACKAGE_NAME = "cloud_asset_inventory"

path_agent_policy_data = Path(__file__).parent / AGENT_POLICY
path_pkg_policy_data = Path(__file__).parent / PACKAGE_POLICY
path_arm_parameters = Path(__file__).parent / ARM_PARAMETERS
path_arm_template = Path(__file__).parent / ARM_TEMPLATE
INTEGRATION_NAME = "Asset Inventory Azure"

PKG_DEFAULT_VERSION = VERSION_MAP.get("asset_inventory_azure", "")
INTEGRATION_INPUT = {
    "name": generate_random_name("pkg-asset-inventory-azure"),
    "input_name": "asset_inventory_azure",
    "vars": {
        "azure.account_type": "single-account",
        "azure.credentials.type": "arm_template",
    },
}
AGENT_INPUT = {
    "name": generate_random_name("asset-inventory-azure"),
}


if __name__ == "__main__":
    # pylint: disable=duplicate-code
    package_version = get_package_version(cfg=cnfg.elk_config, package_name=PACKAGE_NAME)
    logger.info(f"Package version: {package_version}")
    if not version_compatible(
        current_version=package_version,
        required_version=PKG_DEFAULT_VERSION,
    ):
        logger.warning(f"{INTEGRATION_NAME} is not supported in version {package_version}")
        sys.exit(0)
    logger.info(f"Package version: {package_version}")
    update_package_version(
        cfg=cnfg.elk_config,
        package_name=PACKAGE_NAME,
        package_version=package_version,
    )

    logger.info(f"Starting installation of {INTEGRATION_NAME} integration.")
    agent_data, package_data = load_data(
        cfg=cnfg.elk_config,
        agent_input=AGENT_INPUT,
        package_input=INTEGRATION_INPUT,
        stream_name="cloud_asset_inventory.asset_inventory",
        package_name=PACKAGE_NAME,
    )

    logger.info("Create agent policy")
    agent_policy_id = create_agent_policy(cfg=cnfg.elk_config, json_policy=agent_data)

    logger.info(f"Create {INTEGRATION_NAME} integration for policy {agent_policy_id}")
    package_policy_id = create_integration(
        cfg=cnfg.elk_config,
        pkg_policy=package_data,
        agent_policy_id=agent_policy_id,
        data={},
    )

    state_manager.add_policy(
        PolicyState(
            agent_policy_id,
            package_policy_id,
            EXPECTED_AGENTS,
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

    with open(path_arm_parameters, "w") as file:
        json.dump(azure_arm_parameters, file)

    logger.info(f"Get {INTEGRATION_NAME} template")
    default_url = get_package_default_url(
        cfg=cnfg.elk_config,
        policy_name="asset_inventory",
        policy_type="cloudbeat/asset_inventory_azure",
        package_name=PACKAGE_NAME,
    )
    template_url = extract_arm_template_url(url_string=default_url)

    logger.info(f"Using {template_url} for stack creation")
    # If file exists, rename it
    rename_file_by_suffix(
        file_path=path_arm_template,
        suffix="-orig",
    )
    get_arm_template(
        url=template_url,
        template_path=path_arm_template,
    )

    logger.info(f"Installation of {INTEGRATION_NAME} integration is done")
