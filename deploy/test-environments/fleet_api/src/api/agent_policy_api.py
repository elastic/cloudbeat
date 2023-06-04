"""
This module contains API calls related to the agent policy API.
"""

from munch import Munch, munchify
from loguru import logger
from api.headers import base_headers as headers
from api.base_call_api import APICallException, perform_api_call


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
    # pylint: disable=duplicate-code
    url = f"{cfg.kibana_url}/api/fleet/agent_policies"

    try:
        response = perform_api_call(
            method="POST",
            url=url,
            headers=headers,
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
        return ""


def delete_agent_policy(cfg: Munch, agent_policy_id: str):
    """This function deletes an agent policy

    Args:
        cfg (Munch): Config object containing authentication data.
        agent_policy_id (str): The ID of the agent policy to be deleted
    Raises:
        APICallException: If the API call fails or returns a non-200 status code.
    """
    # pylint: disable=duplicate-code
    url = f"{cfg.kibana_url}/api/fleet/agent_policies/delete"
    json_data = {
        "agentPolicyId": agent_policy_id,
    }

    try:
        perform_api_call(
            method="POST",
            url=url,
            headers=headers,
            auth=cfg.auth,
            params={"json": json_data},
        )
        logger.info(f"Agent policy '{agent_policy_id}' deleted successfully")
    except APICallException as api_ex:
        logger.error(
            f"API call failed, status code {api_ex.status_code}. Response: {api_ex.response_text}",
        )
        return


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
    # pylint: disable=duplicate-code
    url = f"{cfg.kibana_url}/api/fleet/agents"

    try:
        response = perform_api_call(
            method="GET",
            url=url,
            headers=headers,
            auth=cfg.auth,
            params={},
        )
        response_obj = munchify(response)
        return response_obj.list
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
            headers=headers,
            auth=cfg.auth,
            params={"json": json_data},
        )
        logger.info(f"Agents '{agents}' deleted successfully")
    except APICallException as api_ex:
        logger.error(
            f"API call failed, status code {api_ex.status_code}. Response: {api_ex.response_text}",
        )
        return
