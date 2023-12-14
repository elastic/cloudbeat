"""
Aggregates API calls to the Kibana Fleet API.
"""

from commonlib.base_call_api import APICallException, perform_api_call
from munch import Munch, munchify
from loguru import logger


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
            auth=cfg.basic_auth,
        )
        response_obj = munchify(response)
        return response_obj.list
    except APICallException as api_ex:
        logger.error(
            f"API call failed, status code {api_ex.status_code}. Response: {api_ex.response_text}",
        )
        return []
