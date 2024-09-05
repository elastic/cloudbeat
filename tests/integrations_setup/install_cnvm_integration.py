#!/usr/bin/env python
"""
This script installs CNVM AWS integration

The following steps are performed:
1. Create an agent policy.
2. Create a CNVM AWS integration.
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
from fleet_api.package_policy_api import create_cnvm_integration
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

CNVM_EXPECTED_AGENTS = 1
CNVM_CLOUDFORMATION_CONFIG = "../../deploy/cloudformation/config.json"
CNMV_TEMPLATE = "../../deploy/cloudformation/elastic-agent-ec2-cnvm.yml"
CNMV_TEMP_FILE = "elastic-agent-ec2-cnvm-temp.yml"
CNVM_AGENT_TAGS = ["cft_version:*", "cft_arn:arn:aws:cloudformation:.*"]
PKG_DEFAULT_VERSION = VERSION_MAP.get("vuln_mgmt_aws", "")
INTEGRATION_NAME = "CNVM AWS"
INTEGRATION_INPUT = {
    "name": generate_random_name("pkg-cnvm-aws"),
    "input_name": "vuln_mgmt_aws",
    "posture": "vuln_mgmt",
    "deployment": "aws",
}
AGENT_INPUT = {
    "name": generate_random_name("cnvm-aws"),
}

cnvm_cloudformation_config = Path(__file__).parent / CNVM_CLOUDFORMATION_CONFIG
cnvm_cloudformation_template = Path(__file__).parent / CNMV_TEMPLATE


if __name__ == "__main__":
    # pylint: disable=duplicate-code
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
        stream_name="cloud_security_posture.findings",
    )

    logger.info("Create agent policy")
    agent_policy_id = create_agent_policy(cfg=cnfg.elk_config, json_policy=agent_data)

    logger.info(f"Create {INTEGRATION_NAME} integration for policy {agent_policy_id}")
    package_policy_id = create_cnvm_integration(
        cfg=cnfg.elk_config,
        pkg_policy=package_data,
        agent_policy_id=agent_policy_id,
    )

    state_manager.add_policy(
        PolicyState(
            agent_policy_id,
            package_policy_id,
            CNVM_EXPECTED_AGENTS,
            CNVM_AGENT_TAGS,
            HostType.LINUX_TAR.value,
            INTEGRATION_INPUT["name"],
        ),
    )

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

    logger.info(f"Get {INTEGRATION_NAME} template")
    default_url = get_package_default_url(
        cfg=cnfg.elk_config,
        policy_name=INTEGRATION_INPUT["posture"],
        policy_type="cloudbeat/vuln_mgmt_aws",
    )
    template_url = extract_template_url(url_string=default_url)

    logger.info(f"Using {template_url} for stack creation")
    if template_url:
        rename_file_by_suffix(
            file_path=cnvm_cloudformation_template,
            suffix="-orig",
        )
    get_cnvm_template(
        url=template_url,
        template_path=cnvm_cloudformation_template,
        cnvm_tags=cnfg.aws_config.cnvm_tags,
    )

    logger.info(f"Installation of {INTEGRATION_NAME} integration is done")
