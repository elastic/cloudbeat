"""
This module contains API calls related to internal Kibana settings.
"""

from fleet_api.base_call_api import (
    APICallException,
    perform_api_call,
)
from loguru import logger
from munch import Munch


def get_kibana_status(cfg: Munch) -> dict:
    """Gets Kibana status information including deployment type.

    Args:
        cfg (Munch): Config object containing authentication data.

    Returns:
        dict: The status response from Kibana API.
    """
    url = f"{cfg.kibana_url}/api/status"
    try:
        response = perform_api_call(
            method="GET",
            url=url,
            auth=cfg.auth,
        )
        logger.info("Kibana status retrieved successfully.")
        return response
    except APICallException as api_ex:
        logger.error(f"Failed to get Kibana status, status code {api_ex.status_code}. Response: {api_ex.response_text}")
        raise api_ex


def is_serverless_deployment(cfg: Munch) -> bool:
    """Checks if the Kibana deployment is serverless.

    Args:
        cfg (Munch): Config object containing authentication data.

    Returns:
        bool: True if serverless deployment, False otherwise.
    """
    status = get_kibana_status(cfg)
    build_flavor = status.get("version", {}).get("build_flavor", "")
    return build_flavor == "serverless"


def update_kibana_settings(cfg: Munch, settings: dict) -> dict:
    """Updates internal Kibana settings.

    Args:
        cfg (Munch): Config object containing authentication data.
        settings (dict): Dictionary of settings to update in Kibana.

    Returns:
        dict: The response from the Kibana settings API.
    """
    # Use standard API endpoint by default
    url = f"{cfg.kibana_url}/api/kibana/settings"
    logger.info("Using standard Kibana settings API endpoint.")

    # Update to serverless endpoint if needed
    if is_serverless_deployment(cfg):
        url = f"{cfg.kibana_url}/internal/kibana/settings"
        logger.info("Detected serverless deployment, switching to internal API endpoint.")

    headers = {
        "Content-Type": "application/json",
        "kbn-xsrf": "true",
        # Needed for internal Kibana API compatibility
        "x-elastic-internal-origin": "kibana",
    }
    payload = {
        "changes": settings,
    }
    try:
        response = perform_api_call(
            method="POST",
            url=url,
            headers=headers,
            auth=cfg.auth,
            params={"json": payload},
        )
        logger.info("Kibana settings updated successfully.")
        return response
    except APICallException as api_ex:
        logger.error(f"API call failed, status code {api_ex.status_code}. Response: {api_ex.response_text}")
        raise api_ex
