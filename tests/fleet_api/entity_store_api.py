"""
This module contains API calls related to Entity Store interactions.
"""

from fleet_api.base_call_api import (
    APICallException,
    perform_api_call,
)
from loguru import logger
from munch import Munch


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
