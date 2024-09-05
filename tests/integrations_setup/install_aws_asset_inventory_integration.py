#!/usr/bin/env python
"""
TODO(kuba): UPDATE THIS DOCSTRING!

This script installs Asset Inventory AWS integration

The following steps are performed:
1. Create an agent policy.
2. Create a Asset Inventory AWS integration.
3. Create a deploy/cloudformation/config.json file to be used by the just deploy-cloudformation command.
"""
import sys
from pathlib import Path

import configuration_fleet as cnfg
from fleet_api.agent_policy_api import create_agent_policy
from fleet_api.common_api import (
    get_artifact_server,
    get_enrollment_token,
    get_fleet_server_host,
    get_package_version,
)
from fleet_api.package_policy_api import create_integration
from fleet_api.utils import render_template
from loguru import logger
from munch import Munch
from package_policy import (
    VERSION_MAP,
    generate_random_name,
    load_data,
    version_compatible,
)
from state_file_manager import HostType, PolicyState, state_manager

EXPECTED_AGENTS = 1
PKG_DEFAULT_VERSION = VERSION_MAP.get("asset_inventory_aws", "")
INTEGRATION_NAME = "Asset Inventory AWS"
INTEGRATION_INPUT = {
    "name": generate_random_name("pkg-asset-inventory-aws"),
    "input_name": "asset_inventory_aws",
    "vars": {
        "access_key_id": cnfg.aws_config.access_key_id,
        "secret_access_key": cnfg.aws_config.secret_access_key,
        "aws.credentials.type": "direct_access_keys",
    },
}
AGENT_INPUT = {
    "name": generate_random_name("asset-inventory-aws"),
}
script_template = Path(__file__).parent / "data/cspm-linux.j2"


if __name__ == "__main__":
    package_version = get_package_version(cfg=cnfg.elk_config)
    logger.info(f"Package version: {package_version}")
    if not version_compatible(
        current_version=package_version,
        required_version=PKG_DEFAULT_VERSION,
    ):
        logger.warning(f"{INTEGRATION_NAME} is not supported in version {package_version}")
        sys.exit(0)
    logger.info(f"Starting installation of {INTEGRATION_NAME} integration.")
    agent_data, package_data = load_data(
        cfg=cnfg.elk_config,
        agent_input=AGENT_INPUT,
        package_input=INTEGRATION_INPUT,
        stream_name="cloud_asset_inventory.asset_inventory",
        package_name="cloud_asset_inventory",
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
            INTEGRATION_INPUT["name"],
        ),
    )

    manifest_params = Munch()
    manifest_params.enrollment_token = get_enrollment_token(
        cfg=cnfg.elk_config,
        policy_id=agent_policy_id,
    )
    manifest_params.fleet_url = get_fleet_server_host(cfg=cnfg.elk_config)
    # TODO(kuba): Rename cspm.sh files maybe?
    manifest_params.file_path = Path(__file__).parent / "aws-asset-inventory.sh"
    manifest_params.agent_version = cnfg.elk_config.stack_version
    manifest_params.artifacts_url = get_artifact_server(cnfg.elk_config.stack_version)

    # Render the template and get the replaced content
    rendered_content = render_template(script_template, manifest_params.toDict())

    logger.info(f"Creating {INTEGRATION_NAME} linux manifest")
    # Write the rendered content to a file
    # TODO(kuba): Rename cspm.sh files maybe?
    with open(Path(__file__).parent / "aws-asset-inventory-linux.sh", "w", encoding="utf-8") as cspm_file:
        cspm_file.write(rendered_content)

    logger.info(f"Installation of {INTEGRATION_NAME} integration is done")
