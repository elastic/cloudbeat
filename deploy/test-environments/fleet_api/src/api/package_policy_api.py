""" 
This module contains API calls related to the package policy API.
"""

from munch import Munch, munchify
from loguru import logger
from api.headers import base_headers as headers
from api.base_call_api import APICallException, perform_api_call
from utils import update_key, delete_key

def create_kspm_unmanaged_integration(cfg: Munch, pkg_policy: dict, agent_policy_id: str) -> str:
    """ Creates an unmanaged integration for KSPM

    Args:
        cfg (Munch): Config object containing authentication data.
        pkg_policy (dict): The package policy to be associated with the integration.
        agent_policy_id (str): The ID of the agent policy to be used.

    Returns:
        str: The ID of the created unmanaged integration.
    """
    package_policy = munchify(pkg_policy)
    package_policy.policy_id = agent_policy_id

    url = f"{cfg.kibana_url}/api/fleet/package_policies"

    try:
        response = perform_api_call(method="POST",
                                    url=url,
                                    headers=headers,
                                    auth=cfg.auth,
                                    params={"json": package_policy})
        package_policy_id = munchify(response).item.id
        logger.info(f"Package policy '{package_policy_id}' created successfully")
        return package_policy_id
    except APICallException as api_ex:
        logger.error(
            f"API call failed with status code {api_ex.status_code}. "
            f"Response: {api_ex.response_text}"
            )
        return


def create_kspm_eks_integration(cfg: Munch,
                                pkg_policy: dict,
                                agent_policy_id: str,
                                eks_data: dict) -> str:
    """ Creates an eks integration for KSPM

    Args:
        cfg (Munch): Config object containing authentication data.
        pkg_policy (dict): The package policy to be associated with the integration.
        agent_policy_id (str): The ID of the agent policy to be used.
        eks_data (dict): The EKS data to be modified in the package policy.

    Returns:
        str: The ID of the created unmanaged integration.
    """

    url = f"{cfg.kibana_url}/api/fleet/package_policies"

    try:
        package_policy = munchify(pkg_policy)
        package_policy.policy_id = agent_policy_id

        delete_key(package_policy, search_key="role_arn", key_to_delete="value")

        for key, value in eks_data.items():
            update_key(package_policy, key, value)

        response = perform_api_call(method="POST",
                                    url=url,
                                    headers=headers,
                                    auth=cfg.auth,
                                    params={"json": package_policy})
        package_policy_id = munchify(response).item.id
        logger.info(f"Package policy '{package_policy_id}' created successfully")
        return package_policy_id
    except APICallException as api_ex:
        logger.error(
            f"API call failed with status code {api_ex.status_code}. "
            f"Response: {api_ex.response_text}"
            )
        return



def delete_package_policy(cfg: Munch, policy_ids: list):
    """ Delete package policy

    Args:
        cfg (Munch): Config object containing authentication data.
        policy_ids (list): A list of policy IDs to be deleted.
    """

    data_json = {
        "packagePolicyIds": policy_ids,
        "force": "true"
    }

    url = f"{cfg.kibana_url}/api/fleet/package_policies/delete"

    try:
        perform_api_call(method="POST",
                         url=url,
                         headers=headers,
                         auth=cfg.auth,
                         params={"json": data_json})
        logger.info("Package policy deleted successfully")
    except APICallException as api_ex:
        logger.error(
            f"API call failed with status code {api_ex.status_code}. "
            f"Response: {api_ex.response_text}"
            )
