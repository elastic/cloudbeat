#!/usr/bin/env python
"""
This script installs KSPM EKS integration

The following steps are performed:
1. Create an agent policy.
2. Create a KSPM EKS integration.
3. Create a KSPM manifest to be deployed on a host.
"""

from pathlib import Path
from typing import Dict, Tuple
from munch import Munch
import configuration_fleet as cnfg
from api.agent_policy_api import create_agent_policy
from api.package_policy_api import create_kspm_eks_integration
from api.common_api import (
    get_enrollment_token,
    get_fleet_server_host,
    create_kubernetes_manifest,
)
from loguru import logger
from utils import (
    read_json,
    save_state,
)

KSPM_EKS_AGENT_POLICY = "../../../cloud/data/agent_policy.json"
KSPM_EKS_PACKAGE_POLICY = "../../../cloud/data/package_policy.json"


kspm_agent_policy_data = Path(__file__).parent / KSPM_EKS_AGENT_POLICY
kspm_eks_pkg_policy_data = Path(__file__).parent / KSPM_EKS_PACKAGE_POLICY


def load_data() -> Tuple[Dict, Dict]:
    """Loads data.

    Returns:
        Tuple[Dict, Dict]: A tuple containing the loaded agent and package policies.
    """
    logger.info("Loading agent and package policies")
    agent_policy = read_json(json_path=kspm_agent_policy_data)
    package_policy = read_json(json_path=kspm_eks_pkg_policy_data)
    return agent_policy, package_policy


if __name__ == "__main__":
    # pylint: disable=duplicate-code
    logger.info("Starting installation of KSPM EKS integration.")
    agent_data, package_data = load_data()

    logger.info("Create agent policy")
    agent_policy_id = create_agent_policy(cfg=cnfg.elk_config, json_policy=agent_data)

    aws_config = cnfg.aws_config
    eks_data = {
        "access_key_id": aws_config.access_key_id,
        "secret_access_key": aws_config.secret_access_key,
    }

    logger.info("Create KSPM EKS integration")
    package_policy_id = create_kspm_eks_integration(
        cfg=cnfg.elk_config,
        pkg_policy=package_data,
        agent_policy_id=agent_policy_id,
        eks_data=eks_data,
    )

    save_state(
        cnfg.state_data_file,
        [
            {
                "pkg_policy_id": package_policy_id,
                "agnt_policy_id": agent_policy_id,
            },
        ],
    )
    manifest_params = Munch()
    manifest_params.enrollment_token = get_enrollment_token(
        cfg=cnfg.elk_config,
        policy_id=agent_policy_id,
    )

    manifest_params.fleet_url = get_fleet_server_host(cfg=cnfg.elk_config)
    manifest_params.yaml_path = Path(__file__).parent / "kspm_eks.yaml"
    logger.info("Creating KSPM EKS manifest")
    create_kubernetes_manifest(cfg=cnfg.elk_config, params=manifest_params)
    logger.info("Installation of KSPM EKS integration is done")
