"""
This module contains API calls related to the package policy API.
"""

from fleet_api.base_call_api import (
    APICallException,
    perform_api_call,
    uses_new_fleet_api_response,
)
from fleet_api.utils import delete_key, update_key
from loguru import logger
from munch import Munch, munchify


def create_kspm_unmanaged_integration(cfg: Munch, pkg_policy: dict, agent_policy_id: str) -> str:
    """Creates an unmanaged integration for KSPM

    Args:
        cfg (Munch): Config object containing authentication data.
        pkg_policy (dict): The package policy to be associated with the integration.
        agent_policy_id (str): The ID of the agent policy to be used.

    Returns:
        str: The ID of the created unmanaged integration.
    """
    return create_integration(
        cfg=cfg,
        pkg_policy=pkg_policy,
        agent_policy_id=agent_policy_id,
        data={},
    )


def create_kspm_eks_integration(
    cfg: Munch,
    pkg_policy: dict,
    agent_policy_id: str,
    eks_data: dict,
) -> str:
    """Creates an eks integration for KSPM

    Args:
        cfg (Munch): Config object containing authentication data.
        pkg_policy (dict): The package policy to be associated with the integration.
        agent_policy_id (str): The ID of the agent policy to be used.
        eks_data (dict): The EKS data to be modified in the package policy.

    Returns:
        str: The ID of the created unmanaged integration.
    """
    package_policy = munchify(pkg_policy)
    delete_key(package_policy, search_key="role_arn", key_to_delete="value")

    return create_integration(
        cfg=cfg,
        pkg_policy=package_policy,
        agent_policy_id=agent_policy_id,
        data=eks_data,
    )


def create_cspm_integration(
    cfg: Munch,
    pkg_policy: dict,
    agent_policy_id: str,
    cspm_data: dict,
) -> str:
    """Creates an CSPM AWS integration

    Args:
        cfg (Munch): Config object containing authentication data.
        pkg_policy (dict): The package policy to be associated with the integration.
        agent_policy_id (str): The ID of the agent policy to be used.
        cspm_data (dict): The CSPM data to be modified in the package policy.

    Returns:
        str: The ID of the created unmanaged integration.
    """
    return create_integration(
        cfg=cfg,
        pkg_policy=pkg_policy,
        agent_policy_id=agent_policy_id,
        data=cspm_data,
    )


def create_cnvm_integration(cfg: Munch, pkg_policy: dict, agent_policy_id: str) -> str:
    """Creates an integration for CNVM

    Args:
        cfg (Munch): Config object containing authentication data.
        pkg_policy (dict): The package policy to be associated with the integration.
        agent_policy_id (str): The ID of the agent policy to be used.

    Returns:
        str: The ID of the created unmanaged integration.
    """
    return create_integration(
        cfg=cfg,
        pkg_policy=pkg_policy,
        agent_policy_id=agent_policy_id,
        data={},
    )


def create_integration(cfg: Munch, pkg_policy: dict, agent_policy_id: str, data: dict) -> str:
    """Creates an elastic integration

    Args:
        cfg (Munch): Config object containing authentication data.
        pkg_policy (dict): The package policy to be associated with the integration.
        agent_policy_id (str): The ID of the agent policy to be used.
        data (dict): The integration data to be modified in the package policy.

    Returns:
        str: The ID of the created unmanaged integration.
    """
    url = f"{cfg.kibana_url}/api/fleet/package_policies"

    if pkg_policy.get("policy_id") is not None:
        pkg_policy["policy_id"] = agent_policy_id
    else:
        pkg_policy["policy_ids"] = [agent_policy_id]

    for key, value in data.items():
        update_key(pkg_policy, key, value)

    try:
        response = perform_api_call(
            method="POST",
            url=url,
            auth=cfg.auth,
            params={"json": pkg_policy},
        )
        policy_data = response.get("response", {}).get("item", {})
        if uses_new_fleet_api_response(cfg.stack_version):
            policy_data = response.get("item", {})
        package_policy_id = policy_data.get("id", "")
        logger.info(f"Package policy '{package_policy_id}' created successfully")
        return package_policy_id
    except APICallException as api_ex:
        logger.error(
            f"API call failed, status code {api_ex.status_code}. Response: {api_ex.response_text}",
        )
        raise api_ex


def delete_package_policy(cfg: Munch, policy_ids: list):
    """Delete package policy

    Args:
        cfg (Munch): Config object containing authentication data.
        policy_ids (list): A list of policy IDs to be deleted.
    """
    data_json = {
        "packagePolicyIds": policy_ids,
        "force": "true",
    }

    url = f"{cfg.kibana_url}/api/fleet/package_policies/delete"

    try:
        perform_api_call(
            method="POST",
            url=url,
            auth=cfg.auth,
            params={"json": data_json},
        )
        logger.info("Package policy deleted successfully")
    except APICallException as api_ex:
        logger.error(
            f"API call failed, status code {api_ex.status_code}. Response: {api_ex.response_text}",
        )


def get_package_policy_by_id(cfg: Munch, policy_id: str) -> dict:
    """
    Retrieve package policy information by its ID.

    Args:
        cfg (Munch): A configuration object containing Kibana URL, authentication details, etc.
        policy_id (str): The package policy ID to retrieve.

    Returns:
        dict: A dictionary containing the package policy information,
              or an empty dictionary if not found.

    Raises:
        APICallException: If the API call to retrieve the package policy fails.
    """
    url = f"{cfg.kibana_url}/api/fleet/package_policies/{policy_id}"

    try:
        response = perform_api_call(
            method="GET",
            url=url,
            auth=cfg.auth,
        )

        return response.get("item", {})
    except APICallException as api_ex:
        logger.error(
            f"API call failed, status code {api_ex.status_code}. Response: {api_ex.response_text}",
        )
        raise
