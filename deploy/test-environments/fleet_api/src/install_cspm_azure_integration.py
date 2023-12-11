#!/usr/bin/env python
"""
This script installs CSPM Azure integration

The following steps are performed:
1. Create an agent policy.
2. Create a CSPM Azure integration.
"""
import sys
import json
from pathlib import Path
from munch import Munch
import configuration_fleet as cnfg
from api.agent_policy_api import create_agent_policy
from api.package_policy_api import create_cspm_integration
from api.common_api import (
    get_enrollment_token,
    get_fleet_server_host,
    get_artifact_server,
    get_package_version,
    get_arm_template,
)
from package_policy import (
    get_package_default_url,
    extract_arm_template_url,
)
from package_policy import (
    version_compatible,
    VERSION_MAP,
    load_data,
    generate_random_name,
)


from loguru import logger
from utils import (
    rename_file_by_suffix,
)
from state_file_manager import state_manager, PolicyState, HostType

CSPM_AZURE_AGENT_POLICY = "../../../cloud/data/agent_policy_cspm_azure.json"
CSPM_AZURE_PACKAGE_POLICY = "../../../cloud/data/package_policy_cspm_azure.json"
CSPM_AZURE_EXPECTED_AGENTS = 1
AZURE_ARM_PARAMETERS = "../../../azure/arm_parameters.json"
AZURE_ARM_TEMPLATE = "../../../azure/ARM-for-single-account.json"

cspm_azure_agent_policy_data = Path(__file__).parent / CSPM_AZURE_AGENT_POLICY
cspm_azure_pkg_policy_data = Path(__file__).parent / CSPM_AZURE_PACKAGE_POLICY
cspm_azure_arm_parameters = Path(__file__).parent / AZURE_ARM_PARAMETERS
cspm_azure_arm_template = Path(__file__).parent / AZURE_ARM_TEMPLATE
INTEGRATION_NAME = "CSPM Azure"

PKG_DEFAULT_VERSION = VERSION_MAP.get("cis_azure", "")
INTEGRATION_INPUT = {
    "name": generate_random_name("pkg-cspm-azure"),
    "input_name": "cis_azure",
    "posture": "cspm",
    "deployment": "azure",
}
AGENT_INPUT = {
    "name": generate_random_name("cspm-azure"),
}


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

    logger.info(f"Get {INTEGRATION_NAME} template")
    default_url = get_package_default_url(
        cfg=cnfg.elk_config,
        policy_name=INTEGRATION_INPUT["posture"],
        policy_type="cloudbeat/cis_azure",
    )
    template_url = extract_arm_template_url(url_string=default_url)

    logger.info(f"Using {template_url} for stack creation")
    # If file exists, rename it
    rename_file_by_suffix(
        file_path=cspm_azure_arm_template,
        suffix="-orig",
    )
    get_arm_template(
        url=template_url,
        template_path=cspm_azure_arm_template,
    )

    logger.info(f"Installation of {INTEGRATION_NAME} integration is done")
