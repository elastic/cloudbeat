#!/usr/bin/env python
"""
Fleet setup for Elastic Defend (endpoint) on CDR: one agent policy, endpoint integration,
malware/ransomware detect modes, and Linux/Windows install artifacts.
"""

import json
import os
from pathlib import Path
from typing import Dict, Tuple

import configuration_fleet as cnfg
from fleet_api.agent_policy_api import create_agent_policy
from fleet_api.endpoint_package_policy import enable_endpoint_malware_ransomware_detect
from fleet_api.common_api import (
    get_artifact_server,
    get_enrollment_token,
    get_fleet_server_host,
    get_package_version,
    update_package_version,
)
from fleet_api.package_policy_api import create_integration
from fleet_api.utils import get_install_servers_option, read_json, render_template
from loguru import logger
from munch import Munch
from state_file_manager import HostType, PolicyState, state_manager

AGENT_POLICY_JSON = "data/agent-policy-elastic-defend.json"
PACKAGE_POLICY_JSON = "data/package-policy-elastic-defend.json"
LINUX_TEMPLATE = "data/elastic-defend-linux.j2"
WINDOWS_TEMPLATE = "data/elastic-defend-windows.j2"
INTEGRATION_LABEL = "ELASTIC_DEFEND_CDR"


def _truthy_env(name: str) -> bool:
    return os.getenv(name, "false").lower() in ("1", "true", "yes")


def _expected_enrolled_agents() -> int:
    n = 0
    if _truthy_env("ELASTIC_DEFEND_ENROLL_LINUX"):
        n += 1
    if _truthy_env("ELASTIC_DEFEND_ENROLL_WINDOWS"):
        n += 1
    return n


def _agent_version() -> str:
    v = (cnfg.elk_config.agent_version or "").strip()
    if v:
        return v
    return (cnfg.elk_config.stack_version or "").strip()


def load_policies() -> Tuple[Dict, Dict]:
    base = Path(__file__).parent
    agent_policy = read_json(base / AGENT_POLICY_JSON)
    package_policy = read_json(base / PACKAGE_POLICY_JSON)
    return agent_policy, package_policy


def _write_hosts_metadata(
    agent_policy_id: str,
    package_policy_id: str,
) -> None:
    meta = {
        "agent_policy_id": agent_policy_id,
        "package_policy_id": package_policy_id,
        "integration": INTEGRATION_LABEL,
        "elastic_defend_linux_public_ip": os.getenv("ELASTIC_DEFEND_LINUX_PUBLIC_IP", ""),
        "elastic_defend_windows_public_ip": os.getenv("ELASTIC_DEFEND_WINDOWS_PUBLIC_IP", ""),
        "elastic_defend_windows_instance_id": os.getenv("ELASTIC_DEFEND_WINDOWS_INSTANCE_ID", ""),
    }
    out = Path(__file__).parent / "elastic_defend_hosts.json"
    out.write_text(json.dumps(meta, indent=2), encoding="utf-8")
    logger.info(f"Wrote {out}")


if __name__ == "__main__":
    package_version = get_package_version(cfg=cnfg.elk_config, package_name="endpoint", prerelease=True)
    if not package_version:
        logger.error("Could not resolve endpoint package version from Fleet")
        raise SystemExit(1)
    logger.info(f"Endpoint package version: {package_version}")

    update_package_version(
        cfg=cnfg.elk_config,
        package_name="endpoint",
        package_version=package_version,
    )

    agent_data, package_data = load_policies()
    package_data["package"]["version"] = package_version

    logger.info("Create Elastic Defend agent policy")
    agent_policy_id = create_agent_policy(cfg=cnfg.elk_config, json_policy=agent_data)

    logger.info("Create Elastic Defend integration")
    package_policy_id = create_integration(
        cfg=cnfg.elk_config,
        pkg_policy=package_data,
        agent_policy_id=agent_policy_id,
        data={},
    )

    logger.info("Set endpoint malware/ransomware modes to detect")
    enable_endpoint_malware_ransomware_detect(cfg=cnfg.elk_config, package_policy_id=package_policy_id)

    expected = _expected_enrolled_agents()
    if expected > 0:
        state_manager.add_policy(
            PolicyState(
                agent_policy_id,
                package_policy_id,
                expected,
                [],
                HostType.LINUX_TAR.value,
                agent_data["name"],
            ),
        )
    else:
        logger.info("Skipping state_manager entry (no enroll steps; expected agents = 0)")

    _write_hosts_metadata(agent_policy_id, package_policy_id)

    enrollment_token = get_enrollment_token(cfg=cnfg.elk_config, policy_id=agent_policy_id)
    fleet_url = get_fleet_server_host(cfg=cnfg.elk_config)
    agent_version = _agent_version()
    artifacts_url = get_artifact_server(agent_version)
    install_servers = get_install_servers_option(cnfg.elk_config.stack_version)

    manifest_params = Munch(
        enrollment_token=enrollment_token,
        fleet_url=fleet_url,
        agent_version=agent_version,
        artifacts_url=artifacts_url,
    )
    if install_servers:
        manifest_params.install_servers = install_servers

    base = Path(__file__).parent
    linux_rendered = render_template(base / LINUX_TEMPLATE, manifest_params.toDict())
    (base / "elastic-defend-linux.sh").write_text(linux_rendered, encoding="utf-8")
    logger.info("Wrote elastic-defend-linux.sh")

    win_rendered = render_template(base / WINDOWS_TEMPLATE, manifest_params.toDict())
    (base / "elastic-defend-windows.ps1").write_text(win_rendered, encoding="utf-8")
    logger.info("Wrote elastic-defend-windows.ps1")

    logger.info("Elastic Defend Fleet setup finished")
