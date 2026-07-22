"""
This module contains API calls related to the managed integrations API.
(/api/fleet/managed_integrations — replaces the deprecated agentless_policies endpoint)
"""

from fleet_api.base_call_api import APICallException, perform_api_call
from loguru import logger
from munch import Munch, munchify

_FIELDS_NOT_ACCEPTED = frozenset({"policy_id", "policy_ids", "supports_agentless", "output_id"})


def create_managed_integration(cfg: Munch, json_policy: dict) -> str:
    """Create a managed (agentless) integration via the unified managed_integrations endpoint.

    Args:
        cfg (Munch): Config object containing authentication data.
        json_policy (dict): Simplified package policy body. Fields policy_id, policy_ids,
            supports_agentless, and output_id are stripped automatically.

    Returns:
        str: The ID of the created managed integration.

    Raises:
        APICallException: If the API call fails or returns a non-200 status code.
    """
    url = f"{cfg.kibana_url}/api/fleet/managed_integrations"
    body = {k: v for k, v in json_policy.items() if k not in _FIELDS_NOT_ACCEPTED}

    try:
        response = perform_api_call(
            method="POST",
            url=url,
            auth=cfg.auth,
            params={"json": body},
        )
        managed_id = munchify(response).item.id
        logger.info(f"Managed integration '{managed_id}' created successfully")
        return managed_id
    except APICallException as api_ex:
        logger.error(
            f"API call failed, status code {api_ex.status_code}. Response: {api_ex.response_text}",
        )
        raise api_ex


def delete_managed_integration(cfg: Munch, policy_id: str):
    """Delete a managed integration.

    Args:
        cfg (Munch): Config object containing authentication data.
        policy_id (str): The ID of the managed integration to delete.

    Raises:
        APICallException: If the API call fails or returns a non-200 status code.
    """
    url = f"{cfg.kibana_url}/api/fleet/managed_integrations/{policy_id}"

    try:
        perform_api_call(
            method="DELETE",
            url=url,
            auth=cfg.auth,
        )
        logger.info(f"Managed integration '{policy_id}' deleted successfully")
    except APICallException as api_ex:
        logger.error(
            f"API call failed, status code {api_ex.status_code}. Response: {api_ex.response_text}",
        )
        raise api_ex
