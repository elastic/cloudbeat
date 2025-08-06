"""
This module contains API calls related to internal Kibana settings.
"""

from fleet_api.base_call_api import (
    APICallException,
    perform_api_call,
)
from loguru import logger
from munch import Munch


def update_kibana_settings(cfg: Munch, settings: dict) -> None:
    """Updates internal Kibana settings.

    Args:
        cfg (Munch): Config object containing authentication data.
        settings (dict): Dictionary of settings to update in Kibana.
    """
    url = f"{cfg.kibana_url}/api/kibana/settings"
    headers = {
        "Content-Type": "application/json",
        "kbn-xsrf": "true",
        "x-elastic-internal-origin": "kibana",
    }
    payload = {
        "changes": settings
    }
    try:
        response = perform_api_call(
            method="POST",
            url=url,
            headers=headers,
            auth=cfg.auth,
            params= {"json": payload}
        )
        logger.info("Kibana settings updated successfully.")
        return response
    except APICallException as api_ex:
        logger.error(f"API call failed, status code {api_ex.status_code}. Response: {api_ex.response_text}")
        raise api_ex
