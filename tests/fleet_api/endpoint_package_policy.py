"""
Helpers for Elastic Defend (endpoint) Fleet package policies: malware/ransomware detect mode updates.
"""

import copy
import time
from typing import Any, Dict

from fleet_api.base_call_api import APICallException, perform_api_call
from fleet_api.package_policy_api import get_package_policy_by_id
from loguru import logger
from munch import Munch

READONLY_PACKAGE_POLICY_KEYS = frozenset(
    {"id", "revision", "created_at", "created_by", "updated_at", "updated_by"},
)


def _package_policy_body_for_put(item: Dict[str, Any]) -> Dict[str, Any]:
    body = copy.deepcopy(item)
    for k in READONLY_PACKAGE_POLICY_KEYS:
        body.pop(k, None)
    return body


def _apply_detect_on_os_policy(os_policy: Dict[str, Any]) -> None:
    for feature in ("malware", "ransomware"):
        block = os_policy.get(feature)
        if isinstance(block, dict) and "mode" in block and block.get("mode") != "detect":
            block["mode"] = "detect"


def apply_endpoint_malware_ransomware_detect_modes(package_policy: Dict[str, Any]) -> None:
    """
    Set malware (and ransomware where present) to mode 'detect' for windows, linux, and mac sections
    under Fleet endpoint integration inputs.
    """
    inputs = package_policy.get("inputs")
    if not isinstance(inputs, list):
        return
    for inp in inputs:
        cfg_block = inp.get("config") if isinstance(inp, dict) else None
        if not isinstance(cfg_block, dict):
            continue
        pol_wrapped = cfg_block.get("policy")
        if not isinstance(pol_wrapped, dict):
            continue
        policy_val = pol_wrapped.get("value")
        if not isinstance(policy_val, dict):
            continue
        for os_key in ("windows", "linux", "mac"):
            os_pol = policy_val.get(os_key)
            if isinstance(os_pol, dict):
                _apply_detect_on_os_policy(os_pol)


def update_package_policy(cfg: Munch, package_policy_id: str, body: Dict[str, Any]) -> None:
    url = f"{cfg.kibana_url}/api/fleet/package_policies/{package_policy_id}"
    try:
        perform_api_call(
            method="PUT",
            url=url,
            auth=cfg.auth,
            params={"json": body},
        )
        logger.info(f"Package policy '{package_policy_id}' updated successfully")
    except APICallException as api_ex:
        logger.error(
            f"API call failed, status code {api_ex.status_code}. Response: {api_ex.response_text}",
        )
        raise


def enable_endpoint_malware_ransomware_detect(cfg: Munch, package_policy_id: str) -> None:
    item = get_package_policy_by_id(cfg=cfg, policy_id=package_policy_id)
    if not item:
        raise ValueError(f"No package policy returned for id {package_policy_id}")
    body = _package_policy_body_for_put(item)
    apply_endpoint_malware_ransomware_detect_modes(body)
    update_package_policy(cfg, package_policy_id, body)
    time.sleep(5)
