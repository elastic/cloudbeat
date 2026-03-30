#!/usr/bin/env python
"""
Fleet-only Okta package policy on the CDR Wiz agent policy (same EC2 host).
Requires install_wiz_integration.py to have run first (cdr_wiz_agent_policy.json).
"""
import sys
from pathlib import Path

import configuration_fleet as cnfg
from cdr_wiz_fleet_helpers import cdr_wiz_agent_policy_id, fleet_epm_package_version
from fleet_api.package_policy_api import create_integration
from fleet_api.utils import read_json, update_key_value
from loguru import logger
from package_policy import generate_random_name


def _skip_okta() -> bool:
    """True when OKTA_LOGS_URL is unset or left at the workflow default."""
    url = (cnfg.okta_config.url or "").strip()
    return url in ("", "default")


def main() -> None:
    """Create Okta package policy on the Wiz CDR agent policy, or skip/no-op if not configured."""
    if _skip_okta():
        logger.info("OKTA_LOGS_URL unset or default; skipping Okta Fleet integration")
        return

    api_key = (cnfg.okta_config.api_key or "").strip()
    if not api_key or api_key == "default":
        logger.error("OKTA_API_KEY is required when OKTA_LOGS_URL is set")
        sys.exit(1)

    base = Path(__file__).parent
    agent_policy_id = cdr_wiz_agent_policy_id(base)
    package_version = fleet_epm_package_version(
        "okta",
        "Could not resolve okta package version from Fleet",
    )
    logger.info(f"Okta package version: {package_version}")

    package_data = read_json(base / "data/okta-pkg.json")
    package_data["name"] = generate_random_name("pkg-okta-cdr")
    package_data["package"]["version"] = package_version

    for key, value in (("url", cnfg.okta_config.url), ("api_key", cnfg.okta_config.api_key)):
        update_key_value(
            data=package_data["inputs"]["okta-httpjson"],
            search_key=key,
            value_to_apply=value,
        )

    logger.info("Create Okta integration on Wiz agent policy")
    create_integration(
        cfg=cnfg.elk_config,
        pkg_policy=package_data,
        agent_policy_id=agent_policy_id,
        data={},
    )
    logger.info("Okta Fleet integration finished")


if __name__ == "__main__":
    main()
