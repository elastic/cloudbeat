#!/usr/bin/env python
"""
This script installs WIZ integration

The following steps are performed:
1. Create an agent policy.
2. Create a WIZ integration.
3. Create an WIZ bash script to be deployed on a host.
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

WIZ_EXPECTED_AGENTS = 1
INTEGRATION_NAME = "Wiz"

agent_launcher_template = Path(__file__).parent / "data/cspm-linux.j2"

if __name__ == "__main__":
    # pylint: disable=duplicate-code
    package_version = get_package_version(cfg=cnfg.elk_config, package_name="wiz", prerelease=False)
    logger.info(f"Package version: {package_version}")

    logger.info(f"Starting installation of {INTEGRATION_NAME} integration.")
    agent_data = SIMPLIFIED_AGENT_POLICY
    agent_data["name"] = generate_random_name("wiz-3rd-party")

    package_data = read_json(Path(__file__).parent / "data/wiz-pkg.json")
    package_data["name"] = generate_random_name("pkg-wiz-3rd-party")
    package_data["package"]["version"] = package_version

    wiz_data = {
        "client_id": cnfg.wiz_config.client_id,
        "client_secret": cnfg.wiz_config.client_secret,
        "url": cnfg.wiz_config.url,
        "token_url": cnfg.wiz_config.token_url,
    }

    for key, value in wiz_data.items():
        update_key_value(
            data=package_data["inputs"]["wiz-cel"],
            search_key=key,
            value_to_apply=value,
        )

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
            WIZ_EXPECTED_AGENTS,
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
    manifest_params.file_path = Path(__file__).parent / "wiz.sh"
    manifest_params.agent_version = cnfg.elk_config.agent_version
    manifest_params.artifacts_url = get_artifact_server(cnfg.elk_config.agent_version)
    install_servers = get_install_servers_option(cnfg.elk_config.agent_version)
    if install_servers:
        manifest_params.install_servers = install_servers
    # Render the template and get the replaced content
    rendered_content = render_template(agent_launcher_template, manifest_params.toDict())

    logger.info(f"Creating {INTEGRATION_NAME} linux manifest")
    # Write the rendered content to a file
    with open(Path(__file__).parent / "wiz.sh", "w", encoding="utf-8") as agent_launcher_file:
        agent_launcher_file.write(rendered_content)

    logger.info(f"Installation of {INTEGRATION_NAME} integration is done")
