#!/usr/bin/env python
"""
Fleet-only entityanalytics_okta package policy on the CDR Wiz agent policy (same EC2 host).
Requires install_wiz_integration.py to have run first (cdr_wiz_agent_policy.json).
Reuses OKTA_API_KEY as okta_token; OKTA_ENTITY_ANALYTICS_DOMAIN is the Okta org hostname only.
"""
import sys
from pathlib import Path

import configuration_fleet as cnfg
from cdr_wiz_fleet_helpers import cdr_wiz_agent_policy_id, fleet_epm_package_version
from fleet_api.package_policy_api import create_integration
from fleet_api.utils import read_json, update_key_value
from loguru import logger
from package_policy import generate_random_name


def _skip_entityanalytics_okta() -> bool:
    """True when OKTA_ENTITY_ANALYTICS_DOMAIN is unset or left at the workflow default."""
    domain = (cnfg.okta_config.entity_analytics_domain or "").strip()
    return domain in ("", "default")


def main() -> None:
    """Create entityanalytics_okta package policy on the Wiz CDR agent policy, or skip if not configured."""
    if _skip_entityanalytics_okta():
        logger.info(
            "OKTA_ENTITY_ANALYTICS_DOMAIN unset or default; skipping Okta Entity Analytics Fleet integration",
        )
        return

    api_key = (cnfg.okta_config.api_key or "").strip()
    if not api_key or api_key == "default":
        logger.error("OKTA_API_KEY is required when OKTA_ENTITY_ANALYTICS_DOMAIN is set")
        sys.exit(1)

    base = Path(__file__).parent
    agent_policy_id = cdr_wiz_agent_policy_id(base)
    package_version = fleet_epm_package_version(
        "entityanalytics_okta",
        "Could not resolve entityanalytics_okta package version from Fleet",
    )
    logger.info(f"entityanalytics_okta package version: {package_version}")

    package_data = read_json(base / "data/entityanalytics_okta-pkg.json")
    package_data["name"] = generate_random_name("pkg-entityanalytics-okta-cdr")
    package_data["package"]["version"] = package_version

    entity_input = package_data["inputs"]["entity-entity-analytics"]
    for key, value in (
        ("okta_domain", cnfg.okta_config.entity_analytics_domain.strip()),
        ("okta_token", cnfg.okta_config.api_key),
    ):
        update_key_value(data=entity_input, search_key=key, value_to_apply=value)

    logger.info("Create Okta Entity Analytics integration on Wiz agent policy")
    create_integration(
        cfg=cnfg.elk_config,
        pkg_policy=package_data,
        agent_policy_id=agent_policy_id,
        data={},
    )
    logger.info("Okta Entity Analytics Fleet integration finished")


if __name__ == "__main__":
    main()
