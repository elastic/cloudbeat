"""
This module contains API calls related to the agent policy API.
"""

from typing import Optional

from fleet_api.base_call_api import (
    APICallException,
    perform_api_call,
    uses_new_fleet_api_response,
)
from loguru import logger
from munch import Munch, munchify


def create_agent_policy(cfg: Munch, json_policy: dict) -> str:
    """This function creates an agent policy

    Args:
        cfg (Munch): Config object containing authentication data.
        json_policy (dict): Data for the agent policy to be created.

    Returns:
        str: The ID of the created agent policy

    Raises:
        APICallException: If the API call fails or returns a non-200 status code.
    """
    url = f"{cfg.kibana_url}/api/fleet/agent_policies"
    logger.info(f"Creating agent policy at {url}")
    logger.info(f"Agent policy data: {json_policy}")
    try:
        response = perform_api_call(
            method="POST",
            url=url,
            auth=cfg.auth,
            params={"json": json_policy},
        )
        agent_policy_id = munchify(response).item.id
        logger.info(f"Agent policy '{agent_policy_id}' created successfully")
        return agent_policy_id
    except APICallException as api_ex:
        logger.error(
            f"API call failed, status code {api_ex.status_code}. Response: {api_ex.response_text}",
        )
        raise api_ex


def update_agent_policy(cfg: Munch, policy_id, json_policy: dict):
    """This function updates an agent policy

    Args:
        cfg (Munch): Config object containing authentication data.
        policy_id (str): Policy id to be updated.
        json_policy (dict): Data for the agent policy to be updated.

    Raises:
        APICallException: If the API call fails or returns a non-200 status code.
    """
    url = f"{cfg.kibana_url}/api/fleet/agent_policies/{policy_id}"

    try:
        perform_api_call(
            method="PUT",
            url=url,
            auth=cfg.auth,
            params={"json": json_policy},
        )
        logger.info(
            f"Agent policy '{policy_id}' for integration '{json_policy.get('name', '')}' has been updated",
        )
    except APICallException as api_ex:
        logger.error(
            f"API call failed, status code {api_ex.status_code}. Response: {api_ex.response_text}",
        )
        raise api_ex


def delete_agent_policy(cfg: Munch, agent_policy_id: str):
    """This function deletes an agent policy

    Args:
        cfg (Munch): Config object containing authentication data.
        agent_policy_id (str): The ID of the agent policy to be deleted
    Raises:
        APICallException: If the API call fails or returns a non-200 status code.
    """
    url = f"{cfg.kibana_url}/api/fleet/agent_policies/delete"
    json_data = {
        "agentPolicyId": agent_policy_id,
    }

    try:
        perform_api_call(
            method="POST",
            url=url,
            auth=cfg.auth,
            params={"json": json_data},
        )
        logger.info(f"Agent policy '{agent_policy_id}' deleted successfully")
    except APICallException as api_ex:
        logger.error(
            f"API call failed, status code {api_ex.status_code}. Response: {api_ex.response_text}",
        )


def get_agent_policy_id_by_name(cfg: Munch, policy_name: str) -> str:
    """
    Check if an agent policy with the specified name exists and return its ID.

    Args:
        cfg (Munch): Config object containing authentication data and endpoint URLs.
        policy_name (str): The name of the agent policy to check.

    Returns:
        str: The ID of the agent policy if it exists, otherwise None.
    """
    url = f"{cfg.kibana_url}/api/fleet/agent_policies"

    try:
        response = perform_api_call(
            method="GET",
            url=url,
            auth=cfg.auth,
        )
        agent_policies = munchify(response).get("items", [])
        for policy in agent_policies:
            if policy.name == policy_name:
                return policy.id
    except APICallException as api_ex:
        logger.error(
            f"API call failed, status code {api_ex.status_code}. Response: {api_ex.response_text}",
        )
        raise api_ex

    return None


def get_agents(cfg: Munch) -> list:
    """
    Retrieves a list of agents from the specified Kibana URL.

    Args:
        cfg (Munch): Configuration object containing Kibana URL and authentication details.

    Returns:
        list: A list of agents retrieved from the API.

    Raises:
        APICallException: If the API call fails with a non-200 status code.
    """
    url = f"{cfg.kibana_url}/api/fleet/agents"

    try:
        response = perform_api_call(
            method="GET",
            url=url,
            auth=cfg.auth,
        )
        if uses_new_fleet_api_response(cfg.stack_version):
            return munchify(response.get("items", []))
        return munchify(response.get("list", []))
    except APICallException as api_ex:
        logger.error(
            f"API call failed, status code {api_ex.status_code}. Response: {api_ex.response_text}",
        )
        return []


def unenroll_agents_from_policy(cfg: Munch, agents: list):
    """
    Unenrolls a list of agents from a policy using the specified Kibana URL.

    Args:
        cfg (Munch): Configuration object containing Kibana URL and authentication details.
        agents (list): A list of agent IDs to unenroll from the policy.

    Raises:
        APICallException: If the API call fails with a non-200 status code.
    """
    url = f"{cfg.kibana_url}/api/fleet/agents/bulk_unenroll"
    json_data = {
        "agents": agents,
        "revoke": "true",
    }

    try:
        perform_api_call(
            method="POST",
            url=url,
            auth=cfg.auth,
            params={"json": json_data},
        )
        logger.info(f"Agents '{agents}' deleted successfully")
    except APICallException as api_ex:
        logger.error(
            f"API call failed, status code {api_ex.status_code}. Response: {api_ex.response_text}",
        )


def create_agent_download_source(
    cfg: Munch,
    name: str,
    host: str,
    is_default: bool = False,
) -> Optional[str]:
    """
    Create a new agent download source using the Kibana Fleet API.

    Args:
        cfg (Munch): Configuration object containing Kibana URL and authentication details.
        name (str): The name of the agent download source.
        host (str): The host URL where agents will download packages from.
        is_default (bool, optional): Whether this source should be the default. Default is False.

    Returns:
        str: The ID of the newly created agent download source,
             or None if the ID cannot be retrieved.
    """
    url = f"{cfg.kibana_url}/api/fleet/agent_download_sources"
    json_data = {
        "name": name,
        "host": host,
        "is_default": is_default,
    }

    try:
        response = perform_api_call(
            method="POST",
            url=url,
            auth=cfg.auth,
            params={"json": json_data},
        )
        source_id = response.get("item", {}).get("id")
        return source_id
    except APICallException as api_ex:
        logger.error(
            f"API call failed, status code {api_ex.status_code}. Response: {api_ex.response_text}",
        )
        return None
