#!/usr/bin/env python
"""
This script installs CSPM AWS integration

The following steps are performed:
1. Create an agent policy.
2. Create a CSPM AWS integration.
3. Create a CSPM bash script to be deployed on a host.
"""

from pathlib import Path
from typing import Dict, Tuple
from munch import Munch
import configuration_fleet as cnfg
from api.agent_policy_api import create_agent_policy, get_agents
from api.package_policy_api import create_cspm_integration
from api.common_api import (
    get_enrollment_token,
    get_fleet_server_host,
    get_build_info,
)
from loguru import logger
from utils import (
    read_json,
    save_state,
    render_template,
)

CSPM_AGENT_POLICY = "../../../cloud/data/agent_policy_cspm_aws.json"
CSPM_PACKAGE_POLICY = "../../../cloud/data/package_policy_cspm_aws.json"

cspm_agent_policy_data = Path(__file__).parent / CSPM_AGENT_POLICY
cspm_pkg_policy_data = Path(__file__).parent / CSPM_PACKAGE_POLICY
cspm_template = Path(__file__).parent / "data/cspm-linux.j2"


def load_data() -> Tuple[Dict, Dict]:
    """Loads data.

    Returns:
        Tuple[Dict, Dict]: A tuple containing the loaded agent and package policies.
    """
    logger.info("Loading agent and package policies")
    agent_policy = read_json(json_path=cspm_agent_policy_data)
    package_policy = read_json(json_path=cspm_pkg_policy_data)
    return agent_policy, package_policy


if __name__ == "__main__":
    # pylint: disable=duplicate-code
    logger.info("Starting installation of CSPM AWS integration.")
    agent_data, package_data = load_data()

    logger.info("Create agent policy")
    agent_policy_id = create_agent_policy(cfg=cnfg.elk_config, json_policy=agent_data)

    aws_config = cnfg.aws_config
    cspm_data = {
        "access_key_id": aws_config.access_key_id,
        "secret_access_key": aws_config.secret_access_key,
        "aws.credentials.type": "direct_access_keys",
    }

    logger.info("Create CSPM integration")
    package_policy_id = create_cspm_integration(
        cfg=cnfg.elk_config,
        pkg_policy=package_data,
        agent_policy_id=agent_policy_id,
        cspm_data=cspm_data,
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
    manifest_params.file_path = Path(__file__).parent / "cspm.sh"
    manifest_params.agent_version = get_agents(cfg=cnfg.elk_config)[0].agent.version
    if "SNAPSHOT" in manifest_params.agent_version:
        manifest_params.artifacts_url = cnfg.artifactory_url["snapshot"] + get_build_info(
            version=manifest_params.agent_version,
            is_snapshot=True,
        )
    else:
        manifest_params.artifacts_url = cnfg.artifactory_url["staging"] + get_build_info(
            version=manifest_params.agent_version,
            is_snapshot=False,
        )

    # Render the template and get the replaced content
    rendered_content = render_template(cspm_template, manifest_params.toDict())

    logger.info("Creating CSPM linux manifest")
    # Write the rendered content to a file
    with open(Path(__file__).parent / "cspm-linux.sh", "w", encoding="utf-8") as cspm_file:
        cspm_file.write(rendered_content)

    logger.info("Installation of CSPM integration is done")
