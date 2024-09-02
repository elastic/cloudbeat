#!/usr/bin/env python
"""
TODO(kuba): UPDATE THIS DOCSTRING!

This script installs Asset Inventory AWS integration

The following steps are performed:
1. Create an agent policy.
2. Create a Asset Inventory AWS integration.
3. Create a deploy/cloudformation/config.json file to be used by the just deploy-cloudformation command.
"""
import json
import sys
from pathlib import Path

import configuration_fleet as cnfg
from fleet_api.agent_policy_api import create_agent_policy
from fleet_api.common_api import (
    get_artifact_server,
    get_cnvm_template,
    get_enrollment_token,
    get_fleet_server_host,
    get_package_version,
)
from fleet_api.package_policy_api import create_integration
from fleet_api.utils import rename_file_by_suffix
from loguru import logger
from munch import Munch
from package_policy import (
    VERSION_MAP,
    extract_template_url,
    generate_random_name,
    get_package_default_url,
    load_data,
    version_compatible,
)
from state_file_manager import HostType, PolicyState, state_manager

EXPECTED_AGENTS = 1
# AI_CLOUDFORMATION_CONFIG = "../../deploy/cloudformation/config.json"
# TODO(kuba): The CF template does not exist yet. Create it
# AI_TEMPLATE = "../../deploy/cloudformation/elastic-agent-ec2-asset-inventory.yml"
# AI_TEMP_FILE = "elastic-agent-ec2-asset-inventory-temp.yml"
# TODO(kuba): Not sure what the tags do yet...
# AI_AGENT_TAGS = ["cft_version:*", "cft_arn:arn:aws:cloudformation:.*"]
# TODO(kuba): Add the integration to version map. But how do I enable betas?
PKG_DEFAULT_VERSION = VERSION_MAP.get("asset_inventory_aws", "")
INTEGRATION_NAME = "Asset Inventory AWS"
INTEGRATION_INPUT = {
    "name": generate_random_name("pkg-asset-inventory-aws"),
    "input_name": "asset_inventory_aws",
    "vars": {
        "access_key_id": aws_config.access_key_id,
        "secret_access_key": aws_config.secret_access_key,
        "aws.credentials.type": "direct_access_keys",
    },
}
AGENT_INPUT = {
    "name": generate_random_name("asset-inventory-aws"),
}
aws_config = cnfg.aws_config

# ai_cloudformation_config = Path(__file__).parent / AI_CLOUDFORMATION_CONFIG
# ai_cloudformation_template = Path(__file__).parent / AI_TEMPLATE


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

    # KUBA: CloudFormation approach
    # cloudformation_params = Munch()
    # cloudformation_params.ENROLLMENT_TOKEN = get_enrollment_token(
    #     cfg=cnfg.elk_config,
    #     policy_id=agent_policy_id,
    # )

    # cloudformation_params.FLEET_URL = get_fleet_server_host(cfg=cnfg.elk_config)
    # cloudformation_params.ELASTIC_AGENT_VERSION = cnfg.elk_config.stack_version
    # cloudformation_params.ELASTIC_ARTIFACT_SERVER = get_artifact_server(cnfg.elk_config.stack_version)

    # with open(ai_cloudformation_config, "w") as file:
    #     json.dump(cloudformation_params, file)

    # logger.info(f"Get {INTEGRATION_NAME} template")
    # default_url = get_package_default_url(
    #     cfg=cnfg.elk_config,
    #     # TODO(kuba): Can policy name just be anything?
    #     policy_name=INTEGRATION_INPUT["posture"],
    #     # TODO(kuba): Policy type needs to be investigated for sure.
    #     policy_type="cloudbeat/vuln_mgmt_aws",
    # )
    # template_url = extract_template_url(url_string=default_url)

    # logger.info(f"Using {template_url} for stack creation")
    # if template_url:
    #     rename_file_by_suffix(
    #         file_path=ai_cloudformation_template,
    #         suffix="-orig",
    #     )
    # # TODO(kuba): I think we could do with a new function for a generic/AI template.
    # get_cnvm_template(
    #     url=template_url,
    #     template_path=ai_cloudformation_template,
    #     # TODO(kuba): let's define our own tags to use, but either way, seem optional
    #     cnvm_tags=cnfg.aws_config.cnvm_tags,
    # )

    # CSPM install approach
    manifest_params = Munch()
    manifest_params.enrollment_token = get_enrollment_token(
        cfg=cnfg.elk_config,
        policy_id=agent_policy_id,
    )
    manifest_params.fleet_url = get_fleet_server_host(cfg=cnfg.elk_config)
    manifest_params.file_path = Path(__file__).parent / "cspm.sh"
    manifest_params.agent_version = cnfg.elk_config.stack_version
    manifest_params.artifacts_url = get_artifact_server(cnfg.elk_config.stack_version)

    # Render the template and get the replaced content
    rendered_content = render_template(cspm_template, manifest_params.toDict())

    logger.info(f"Creating {INTEGRATION_NAME} linux manifest")
    # Write the rendered content to a file
    with open(Path(__file__).parent / "cspm-linux.sh", "w", encoding="utf-8") as cspm_file:
        cspm_file.write(rendered_content)

    logger.info(f"Installation of {INTEGRATION_NAME} integration is done")
