#!/usr/bin/env python
"""
This script installs Activity Logs Azure integration

The following steps are performed:
1. Create an agent policy.
2. Create an Activity Logs integration.
3. Create an Activity Logs bash script to be deployed on a host.
"""
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
from fleet_api.utils import (
    get_install_servers_option,
    read_json,
    render_template,
    update_key_value,
)
from loguru import logger
from munch import Munch
from package_policy import SIMPLIFIED_AGENT_POLICY, generate_random_name
from state_file_manager import HostType, PolicyState, state_manager

LOGS_EXPECTED_AGENTS = 1
INTEGRATION_NAME = "ACTIVITY LOGS AZURE"

agent_launcher_template = Path(__file__).parent / "data/cspm-linux.j2"

if __name__ == "__main__":
    # pylint: disable=duplicate-code
    package_version = get_package_version(cfg=cnfg.elk_config, package_name="azure", prerelease=False)
    logger.info(f"Package version: {package_version}")

    logger.info(f"Starting installation of {INTEGRATION_NAME} integration.")
    agent_data = SIMPLIFIED_AGENT_POLICY
    agent_data["name"] = generate_random_name("activity-logs-azure")

    package_data = read_json(Path(__file__).parent / "data/az-activity-logs-pkg.json")
    package_data["name"] = generate_random_name("pkg-activity-logs-azure")
    package_data["package"]["version"] = package_version
    package_data["vars"]["eventhub"] = cnfg.azure_config.eventhub
    package_data["vars"]["consumer_group"] = cnfg.azure_config.consumer_group
    package_data["vars"]["connection_string"] = cnfg.azure_config.connection_string
    package_data["vars"]["storage_account"] = cnfg.azure_config.storage_account
    package_data["vars"]["storage_account_key"] = cnfg.azure_config.storage_account_key

    update_key_value(package_data, "storage_account_container", generate_random_name("activity-logs"))

    logger.info("Create agent policy")
    agent_policy_id = create_agent_policy(cfg=cnfg.elk_config, json_policy=agent_data)

    logger.info(f"Create {INTEGRATION_NAME} integration")
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
            LOGS_EXPECTED_AGENTS,
            [],
            HostType.LINUX_TAR.value,
            package_data["name"],
        ),
    )

    manifest_params = Munch()
    manifest_params.enrollment_token = get_enrollment_token(
        cfg=cnfg.elk_config,
        policy_id=agent_policy_id,
    )

    manifest_params.fleet_url = get_fleet_server_host(cfg=cnfg.elk_config)
    manifest_params.file_path = Path(__file__).parent / "az_activity_logs.sh"
    manifest_params.agent_version = cnfg.elk_config.agent_version
    manifest_params.artifacts_url = get_artifact_server(cnfg.elk_config.agent_version)
    install_servers = get_install_servers_option(cnfg.elk_config.agent_version)
    if install_servers:
        manifest_params.install_servers = install_servers

    # Render the template and get the replaced content
    rendered_content = render_template(agent_launcher_template, manifest_params.toDict())

    logger.info(f"Creating {INTEGRATION_NAME} linux manifest")
    # Write the rendered content to a file
    with open(Path(__file__).parent / "az_activity_logs.sh", "w", encoding="utf-8") as agent_launcher_file:
        agent_launcher_file.write(rendered_content)

    logger.info(f"Installation of {INTEGRATION_NAME} integration is done")
