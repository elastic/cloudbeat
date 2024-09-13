#!/usr/bin/env python
"""
This script installs CSPM GCP integration

The following steps are performed:
1. Create an agent policy.
2. Create a CSPM GCP integration.
3. Create a deploy/deployment-manager/config.json file to be used by the just deploy-dm command.
"""
import json
import sys
from pathlib import Path

import configuration_fleet as cnfg
from fleet_api.agent_policy_api import create_agent_policy
from fleet_api.common_api import (
    get_artifact_server,
    get_enrollment_token,
    get_fleet_server_host,
    get_package_version,
    update_package_version,
)
from fleet_api.package_policy_api import create_cspm_integration
from fleet_api.utils import read_json
from loguru import logger
from munch import Munch
from package_policy import (
    VERSION_MAP,
    generate_random_name,
    load_data,
    version_compatible,
)
from packaging import version
from state_file_manager import HostType, PolicyState, state_manager

CSPM_GCP_EXPECTED_AGENTS = 1
DEPLOYMENT_MANAGER_CONFIG = "../../deploy/deployment-manager/config.json"

cspm_gcp_deployment_manager_config = Path(__file__).parent / DEPLOYMENT_MANAGER_CONFIG
INTEGRATION_NAME = "CSPM GCP"
PKG_DEFAULT_VERSION = VERSION_MAP.get("cis_gcp", "")
INTEGRATION_INPUT = {
    "name": generate_random_name("pkg-cspm-gcp"),
    "input_name": "cis_gcp",
    "posture": "cspm",
    "deployment": "gcp",
}
AGENT_INPUT = {
    "name": generate_random_name("cspm-gcp"),
}

if __name__ == "__main__":
    # pylint: disable=duplicate-code
    package_version = get_package_version(cfg=cnfg.elk_config)
    if not version_compatible(
        current_version=package_version,
        required_version=PKG_DEFAULT_VERSION,
    ):
        logger.warning(f"{INTEGRATION_NAME} is not supported in version {package_version}")
        sys.exit(0)

    logger.info(f"Package version: {package_version}")
    update_package_version(
        cfg=cnfg.elk_config,
        package_name="cloud_security_posture",
        package_version=package_version,
    )
    if version.parse(package_version) >= version.parse("1.6"):
        INTEGRATION_INPUT["vars"] = {
            "gcp.account_type": "single-account",
        }
        if cnfg.gcp_dm_config.service_account_json_path:
            logger.info("Using service account credentials json")
            json_path = Path(__file__).parent / cnfg.gcp_dm_config.service_account_json_path
            service_account_json = read_json(json_path)
            INTEGRATION_INPUT["vars"]["gcp.credentials.json"] = json.dumps(service_account_json)

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
            HostType.LINUX_TAR.value,
            INTEGRATION_INPUT["name"],
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
