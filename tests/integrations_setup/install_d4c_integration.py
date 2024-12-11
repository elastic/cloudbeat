#!/usr/bin/env python
"""
This script installs Defend for Containers (D4C) integration

The following steps are performed:
1. Create an agent policy.
2. Create a D4C integration.
3. Create a D4C Kubernetes manifest to be deployed on a host.
"""

from pathlib import Path
from typing import Dict, Tuple

import configuration_fleet as cnfg
from fleet_api.agent_policy_api import create_agent_policy
from fleet_api.common_api import (
    create_kubernetes_manifest,
    get_enrollment_token,
    get_fleet_server_host,
    get_package_version,
    update_package_version,
)
from fleet_api.package_policy_api import create_integration
from fleet_api.utils import read_json
from loguru import logger
from munch import Munch
from state_file_manager import HostType, PolicyState, state_manager

D4C_AGENT_POLICY = "data/agent-policy-d4c.json"
D4C_PACKAGE_POLICY = "data/package-policy-d4c.json"
D4C_AGENT_POLICY_NAME = "tf-ap-d4c"
D4C_EXPECTED_AGENTS = 2
INTEGRATION_NAME = "D4C"

d4c_agent_policy_data = Path(__file__).parent / D4C_AGENT_POLICY
d4c_pkg_policy_data = Path(__file__).parent / D4C_PACKAGE_POLICY


def load_data() -> Tuple[Dict, Dict]:
    """
    Loads agent and package policies from JSON files.
    This function reads JSON data from specific paths for agent and package policies.

    Returns:
        Tuple[Dict, Dict]: A tuple containing the loaded agent and package policies.
    """
    logger.info("Loading agent and package policies")
    agent_policy = read_json(json_path=d4c_agent_policy_data)
    package_policy = read_json(json_path=d4c_pkg_policy_data)
    return agent_policy, package_policy


if __name__ == "__main__":
    package_version = get_package_version(cfg=cnfg.elk_config, package_name="cloud_defend")
    logger.info(f"Package version: {package_version}")
    update_package_version(
        cfg=cnfg.elk_config,
        package_name="cloud_defend",
        package_version=package_version,
    )

    logger.info(f"Starting installation of {INTEGRATION_NAME} integration.")
    agent_data, package_data = load_data()

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
            D4C_EXPECTED_AGENTS,
            [],
            HostType.KUBERNETES.value,
            D4C_AGENT_POLICY_NAME,
        ),
    )

    manifest_params = Munch()
    manifest_params.enrollment_token = get_enrollment_token(
        cfg=cnfg.elk_config,
        policy_id=agent_policy_id,
    )

    manifest_params.fleet_url = get_fleet_server_host(cfg=cnfg.elk_config)
    manifest_params.yaml_path = Path(__file__).parent / "kspm_d4c.yaml"
    manifest_params.docker_image_override = cnfg.kspm_config.docker_image_override
    manifest_params.capabilities = True
    logger.info(f"Creating {INTEGRATION_NAME} manifest")
    create_kubernetes_manifest(cfg=cnfg.elk_config, params=manifest_params)
    logger.info(f"Installation of {INTEGRATION_NAME} integration is done")
