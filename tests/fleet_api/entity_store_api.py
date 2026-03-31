"""
This module contains API calls related to Entity Store interactions.
"""

import time

from fleet_api.base_call_api import (
    APICallException,
    perform_api_call,
)
from loguru import logger
from munch import Munch

# Matches ecp-synthetics-monitors/projects/entity-store/lib/common/kibana-api.ts
_ENTITY_STORE_V2_INTERNAL_HEADERS = {
    "Content-Type": "application/json",
    "kbn-xsrf": "true",
    "x-elastic-internal-origin": "kibana",
}

_ENTITY_STORE_V2_SETTING_KEY = "securitySolution:entityStoreEnableV2"
_ENTITY_STORE_V2_POLL_TIMEOUT_SEC = 60
_ENTITY_STORE_V2_POLL_INTERVAL_SEC = 5


def _entity_store_v2_setting_user_value(cfg: Munch):
    """Read userValue for the v2 feature flag from GET /internal/kibana/settings."""
    url = f"{cfg.kibana_url}/internal/kibana/settings"
    data = perform_api_call(
        "GET",
        url,
        auth=cfg.auth,
        headers=_ENTITY_STORE_V2_INTERNAL_HEADERS.copy(),
        params={"params": {"query": _ENTITY_STORE_V2_SETTING_KEY}},
    )
    return data.get("settings", {}).get(_ENTITY_STORE_V2_SETTING_KEY, {}).get("userValue")


def enable_entity_store(cfg: Munch) -> dict:
    """Enables the entity store in Kibana.

    Args:
        cfg (Munch): Config object containing authentication data.

    Returns:
        dict: The response from the entity store enable API.
    """
    url = f"{cfg.kibana_url}/api/entity_store/enable"
    payload = {
        "securitySolution": "enableAssetInventory",
    }
    try:
        response = perform_api_call(
            method="POST",
            url=url,
            auth=cfg.auth,
            params={"json": payload},
        )
        logger.info("Entity store enabled successfully.")
        return response
    except APICallException as api_ex:
        logger.error(f"API call failed, status code {api_ex.status_code}. Response: {api_ex.response_text}")
        raise api_ex


def entity_store_status(cfg: Munch) -> dict:
    """Checks the status of the entity store in Kibana.

    Args:
        cfg (Munch): Config object containing authentication data.

    Returns:
        dict: The status response from the entity store API.
    """
    url = f"{cfg.kibana_url}/api/entity_store/status"
    try:
        response = perform_api_call(
            method="GET",
            url=url,
            auth=cfg.auth,
        )
        logger.info("Entity store status retrieved successfully.")
        return response
    except APICallException as api_ex:
        logger.error(f"API call failed, status code {api_ex.status_code}. Response: {api_ex.response_text}")
        raise api_ex


def entity_store_status_v2(cfg: Munch) -> dict:
    """Checks the status of Entity Store v2 using the internal API (apiVersion=2)."""
    url = f"{cfg.kibana_url}/internal/security/entity_store/status"
    try:
        response = perform_api_call(
            method="GET",
            url=url,
            auth=cfg.auth,
            headers=_ENTITY_STORE_V2_INTERNAL_HEADERS.copy(),
            params={"params": {"apiVersion": "2"}},
            ok_statuses=(200, 201, 204),
        )
        logger.info("Entity Store v2 status retrieved successfully.")
        return response
    except APICallException as api_ex:
        logger.error(
            "Entity Store v2 status API call failed, status {}. Response: {}",
            api_ex.status_code,
            api_ex.response_text,
        )
        raise api_ex


def enable_entity_store_v2(cfg: Munch) -> None:
    """Turn on Entity Store v2 via internal settings and poll until active.

    Same sequence as enableEntityStoreV2 in kibana-api.ts (POST then GET until userValue is true).
    """
    url = f"{cfg.kibana_url}/internal/kibana/settings"
    try:
        perform_api_call(
            "POST",
            url,
            auth=cfg.auth,
            headers=_ENTITY_STORE_V2_INTERNAL_HEADERS.copy(),
            params={"json": {"changes": {_ENTITY_STORE_V2_SETTING_KEY: True}}},
        )
        logger.info("Entity Store v2 setting posted; waiting until active...")
        deadline = time.time() + _ENTITY_STORE_V2_POLL_TIMEOUT_SEC
        while time.time() < deadline:
            if _entity_store_v2_setting_user_value(cfg) is True:
                logger.info("Entity Store v2 feature flag is active.")
                return
            time.sleep(_ENTITY_STORE_V2_POLL_INTERVAL_SEC)
    except APICallException as api_ex:
        logger.error(
            "enable_entity_store_v2 failed, status {}. Response: {}",
            api_ex.status_code,
            api_ex.response_text,
        )
        raise
    raise TimeoutError(
        f"Entity Store v2 setting not active within {_ENTITY_STORE_V2_POLL_TIMEOUT_SEC}s",
    )


def install_entity_store_v2(cfg: Munch) -> dict:
    """Install Entity Store v2 (POST /internal/security/entity_store/install?apiVersion=2, empty body).

    Same as installEntityStoreV2 in kibana-api.ts.
    """
    url = f"{cfg.kibana_url}/internal/security/entity_store/install"
    try:
        result = perform_api_call(
            "POST",
            url,
            auth=cfg.auth,
            headers=_ENTITY_STORE_V2_INTERNAL_HEADERS.copy(),
            params={"json": {}, "params": {"apiVersion": "2"}},
            ok_statuses=(200, 201, 204),
        )
        logger.info("Entity Store v2 install completed.")
        return result
    except APICallException as api_ex:
        logger.error(
            "install_entity_store_v2 failed, status {}. Response: {}",
            api_ex.status_code,
            api_ex.response_text,
        )
        raise api_ex


def init_entity_store_v2_maintainers(cfg: Munch) -> dict:
    """Initialize Entity Store v2 maintainers (internal API, apiVersion=2 query param)."""
    url = f"{cfg.kibana_url}/internal/security/entity_store/entity_maintainers/init"
    try:
        result = perform_api_call(
            "POST",
            url,
            auth=cfg.auth,
            headers=_ENTITY_STORE_V2_INTERNAL_HEADERS.copy(),
            params={"json": {}, "params": {"apiVersion": "2"}},
            ok_statuses=(200, 201, 204),
        )
        logger.info("Entity Store v2 maintainers init completed.")
        return result
    except APICallException as api_ex:
        logger.error(
            "init_entity_store_v2_maintainers failed, status {}. Response: {}",
            api_ex.status_code,
            api_ex.response_text,
        )
        raise api_ex


def is_entity_store_fully_started(cfg: Munch) -> bool:
    """Checks if the entity store is fully started (status is running and all engines are started)."""
    status_response = entity_store_status(cfg)
    global_status = status_response.get("status")
    engines = status_response.get("engines", [])
    if global_status != "running":
        logger.info(f"Entity store global status is: '{global_status}'")
        return False
    logger.info("====== Entity Store Engines Status ====")
    for engine in engines:
        if engine.get("status") != "started":
            logger.error(f"Engine {engine.get('type')} status is not started: {engine.get('status')}")
            return False
        logger.info(f"Engine {engine.get('type')} is started.")
    logger.info("Entity store is fully started.")
    return True


def is_entity_store_v2_fully_started(cfg: Munch) -> bool:
    """Checks if Entity Store v2 is fully started (internal v2 status is running and all engines started)."""
    status_response = entity_store_status_v2(cfg)
    global_status = status_response.get("status")
    engines = status_response.get("engines", [])
    if global_status != "running":
        logger.info("Entity Store v2 global status is: '{}'", global_status)
        return False
    logger.info("====== Entity Store v2 Engines Status ====")
    for engine in engines:
        if engine.get("status") != "started":
            logger.error(
                "Entity Store v2 engine {} status is not started: {}",
                engine.get("type"),
                engine.get("status"),
            )
            return False
        logger.info("Entity Store v2 engine {} is started.", engine.get("type"))
    logger.info("Entity Store v2 is fully started.")
    return True
