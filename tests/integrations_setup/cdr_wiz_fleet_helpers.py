"""Shared helpers for Fleet package policies on the CDR Wiz agent policy."""

import re
import sys
from pathlib import Path

import configuration_fleet as cnfg
from fleet_api.common_api import get_package_version
from fleet_api.utils import read_json
from loguru import logger

# GA stacks look like 8.16.0; SNAPSHOT and BC builds add a hyphen suffix (e.g. 9.4.0-SNAPSHOT, 9.4.0-sdfjh).
_EPM_PRERELEASE_STACK_VERSION = re.compile(r"^\d+\.\d+\.\d+-.+")


def stack_version_uses_epm_prerelease(stack_version: str) -> bool:
    """Whether get_package_version should pass prerelease=True (SNAPSHOT or BC hash suffix)."""
    v = (stack_version or "").strip()
    if not v:
        return False
    if "SNAPSHOT" in v.upper():
        return True
    return bool(_EPM_PRERELEASE_STACK_VERSION.match(v))


def cdr_wiz_agent_policy_id(integrations_setup_dir: Path) -> str:
    """Load agent_policy_id from cdr_wiz_agent_policy.json; exit if missing."""
    wiz_context = read_json(integrations_setup_dir / "cdr_wiz_agent_policy.json")
    agent_policy_id = (wiz_context.get("agent_policy_id") or "").strip()
    if not agent_policy_id:
        logger.error("cdr_wiz_agent_policy.json has no agent_policy_id")
        sys.exit(1)
    return agent_policy_id


def fleet_epm_package_version(package_name: str, missing_log_message: str) -> str:
    """Resolve package version from Fleet / EPM; exit if not found."""
    prerelease = stack_version_uses_epm_prerelease(cnfg.elk_config.stack_version or "")
    package_version = get_package_version(
        cfg=cnfg.elk_config,
        package_name=package_name,
        prerelease=prerelease,
    )
    if not package_version:
        logger.error(missing_log_message)
        sys.exit(1)
    return package_version
