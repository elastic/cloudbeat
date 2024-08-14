#!/usr/bin/env python
"""
This script installs CSPM AWS integration

The following steps are performed:
1. Create an agent policy.
2. Create a CSPM AWS integration.
3. Create a CSPM bash script to be deployed on a host.
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
from fleet_api.package_policy_api import create_cspm_integration
from fleet_api.utils import read_json, render_template, update_key_value
from loguru import logger
from munch import Munch
from package_policy import SIMPLIFIED_AGENT_POLICY, generate_random_name
from state_file_manager import HostType, PolicyState, state_manager

CSPM_EXPECTED_AGENTS = 0
INTEGRATION_NAME = "CLOUDTRAIL AWS"
aws_config = cnfg.aws_config

cloudtrail_template = Path(__file__).parent / "data/cspm-linux.j2"

if __name__ == "__main__":
    # pylint: disable=duplicate-code
    package_version = get_package_version(cfg=cnfg.elk_config, package_name="aws", prerelease=False)
    logger.info(f"Package version: {package_version}")

    logger.info(f"Starting installation of {INTEGRATION_NAME} integration.")
    agent_data = SIMPLIFIED_AGENT_POLICY
    agent_data["name"] = generate_random_name("cloudtrail-aws")

    package_data = read_json(Path(__file__).parent / "data/cloudtrail-pkg.json")
    package_data["name"] = generate_random_name("pkg-cloudtrail-aws")
    package_data["package"]["version"] = package_version
    package_data["vars"]["access_key_id"] = cnfg.aws_config.access_key_id
    package_data["vars"]["secret_access_key"] = cnfg.aws_config.secret_access_key

    update_key_value(package_data, "bucket_arn", cnfg.aws_config.cloudtrail_s3)

    logger.info("Create agent policy")
    agent_policy_id = create_agent_policy(cfg=cnfg.elk_config, json_policy=agent_data)

    logger.info(f"Create {INTEGRATION_NAME} integration")
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
            CSPM_EXPECTED_AGENTS,
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
    manifest_params.file_path = Path(__file__).parent / "cloudtrail.sh"
    manifest_params.agent_version = cnfg.elk_config.stack_version
    manifest_params.artifacts_url = get_artifact_server(cnfg.elk_config.stack_version)

    # Render the template and get the replaced content
    rendered_content = render_template(cloudtrail_template, manifest_params.toDict())

    logger.info(f"Creating {INTEGRATION_NAME} linux manifest")
    # Write the rendered content to a file
    with open(Path(__file__).parent / "cloudtrail-linux.sh", "w", encoding="utf-8") as cloudtrail_file:
        cloudtrail_file.write(rendered_content)

    logger.info(f"Installation of {INTEGRATION_NAME} integration is done")
