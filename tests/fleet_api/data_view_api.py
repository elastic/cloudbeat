"""
This module contains API calls related to Data View interactions in Kibana.
"""

from fleet_api.base_call_api import (
    APICallException,
    perform_api_call,
)
from loguru import logger
from munch import Munch


def create_security_default_data_view(cfg: Munch, name: str, namespace: str = "default") -> dict:
    """Creates a security default data view in Kibana if it doesn't exist.

    This function creates a data view with security-related indices including alerts,
    beats data, logs, and traces commonly used by security solutions.

    Args:
        cfg (Munch): Config object containing authentication data.
        name (str): The name of the data view to create.
        namespace (str, optional): The Kibana space namespace. Defaults to "default".

    Returns:
        dict: The data view object (either existing or newly created).

    Raises:
        APICallException: If the API call fails.
    """
    data_view_id = f"{name}-{namespace}"

    # Check if data view already exists
    if data_view_exists(cfg, name, namespace):
        logger.info(f"Data view '{data_view_id}' already exists.")
        return get_data_view(cfg, name, namespace)

    # Data view doesn't exist, create it
    logger.info(f"Data view '{data_view_id}' not found. Creating new data view.")
    create_url = f"{cfg.kibana_url}/s/{namespace}/api/data_views/data_view"
    payload = {
        "data_view": {
            "title": (
                ".alerts-security.alerts-default,apm-*-transaction*,auditbeat-*,endgame-*,"
                "filebeat-*,logs-*,packetbeat-*,traces-apm*,winlogbeat-*,-*elastic-cloud-logs-*"
            ),
            "timeFieldName": "@timestamp",
            "name": data_view_id,
            "id": data_view_id,
        },
    }

    try:
        response = perform_api_call(
            method="POST",
            url=create_url,
            auth=cfg.auth,
            params={"json": payload},
        )
        logger.info(f"Data view '{data_view_id}' created successfully.")
        return response
    except APICallException as e:
        logger.error(f"Failed to create data view '{data_view_id}': {e}")
        raise


def get_data_view(cfg: Munch, name: str, namespace: str = "default") -> dict:
    """Gets a data view from Kibana.

    Args:
        cfg (Munch): Config object containing authentication data.
        name (str): The name of the data view to retrieve.
        namespace (str, optional): The Kibana space namespace. Defaults to "default".

    Returns:
        dict: The data view object.

    Raises:
        APICallException: If the API call fails or data view doesn't exist.
    """
    data_view_id = f"{name}-{namespace}"
    url = f"{cfg.kibana_url}/s/{namespace}/api/data_views/data_view/{data_view_id}"

    try:
        response = perform_api_call(
            method="GET",
            url=url,
            auth=cfg.auth,
        )
        logger.info(f"Retrieved data view '{data_view_id}' successfully.")
        return response
    except APICallException as e:
        logger.error(f"Failed to get data view '{data_view_id}': {e}")
        raise


def data_view_exists(cfg: Munch, name: str, namespace: str = "default") -> bool:
    """Checks if a data view exists in Kibana.

    Args:
        cfg (Munch): Config object containing authentication data.
        name (str): The name of the data view to check.
        namespace (str, optional): The Kibana space namespace. Defaults to "default".

    Returns:
        bool: True if the data view exists, False otherwise.
    """
    try:
        get_data_view(cfg, name, namespace)
        return True
    except APICallException as e:
        if e.status_code == 404:
            return False
        # Re-raise for other errors
        raise
