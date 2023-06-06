"""
This module contains API calls related to Fleet settings
"""
import codecs
from munch import Munch, munchify
from loguru import logger
from api.base_call_api import APICallException, perform_api_call


def get_enrollment_token(cfg: Munch, policy_id: str) -> str:
    """Retrieves the enrollment token for a specified policy ID.

    Args:
        cfg (Munch): Config object containing authentication data.
        policy_id (str): The ID of the policy for which to retrieve the enrollment token.

    Returns:
        str: The enrollment token associated with the specified policy ID.
    """

    url = f"{cfg.kibana_url}/api/fleet/enrollment_api_keys"

    try:
        response = perform_api_call(
            method="GET",
            url=url,
            auth=cfg.auth,
        )
        response_obj = munchify(response)

        api_key = ""
        for item in response_obj.list:
            if item.policy_id == policy_id:
                api_key = item.api_key
                break
    except APICallException as api_ex:
        logger.error(
            f"API call failed, status code {api_ex.status_code}. Response: {api_ex.response_text}",
        )
        return ""

    return api_key


def get_fleet_server_host(cfg: Munch) -> str:
    """Retrieves the Fleet server host URL.

    Args:
        cfg (Munch): Config object containing authentication data.

    Returns:
        str: The Fleet server host URL.
    """

    url = f"{cfg.kibana_url}/api/fleet/settings"

    try:
        response = perform_api_call(
            method="GET",
            url=url,
            auth=cfg.auth,
        )
        response_obj = munchify(response)
        return response_obj.item.fleet_server_hosts[0]
    except APICallException as api_ex:
        logger.error(
            f"API call failed, status code {api_ex.status_code}. Response: {api_ex.response_text}",
        )
        return ""


def create_kubernetes_manifest(cfg: Munch, params: Munch):
    """Create a Kubernetes manifest based on the provided configuration and parameters.

    Args:
        cfg (Munch): Config object containing authentication data.
        params (Munch): The parameters object containing additional information.
    """
    # pylint: disable=duplicate-code
    url = f"{cfg.kibana_url}/api/fleet/kubernetes"

    request_params = {
        "fleetServer": params.fleet_url,
        "enrolToken": params.enrollment_token,
    }

    try:
        response = perform_api_call(
            method="GET",
            url=url,
            auth=cfg.auth,
            params={"params": request_params},
        )
        response_obj = munchify(response)
        with codecs.open(params.yaml_path, "w", encoding="utf-8-sig") as k8s_yaml:
            k8s_yaml.write(response_obj.item)
        logger.info(f"KSPM manifest is available at: '{params.yaml_path}'")
    except APICallException as api_ex:
        logger.error(
            f"API call failed, status code {api_ex.status_code}. Response: {api_ex.response_text}",
        )
        return


def get_build_info(version: str, is_snapshot: bool) -> str:
    """
    Retrieve the build ID for a specific version of Elastic.

    Args:
        version (str): The version of Elastic.
        is_snapshot (bool): Flag indicating whether it is a snapshot build.

    Returns:
        str: The build ID of the specified version.

    Raises:
        APICallException: If the API call to retrieve the build ID fails.
    """
    # pylint: disable=duplicate-code
    if is_snapshot:
        url = "https://snapshots.elastic.co/latest/master.json"
    else:
        url = f"https://staging.elastic.co/latest/{version}.json"

    try:
        response = perform_api_call(
            method="GET",
            url=url,
        )
        response_obj = munchify(response)
        return response_obj.build_id

    except APICallException as api_ex:
        logger.error(
            f"API call failed, status code {api_ex.status_code}. Response: {api_ex.response_text}",
        )
        return ""
