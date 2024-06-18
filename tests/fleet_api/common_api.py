"""
This module contains API calls related to Fleet settings
"""

import codecs
import json
import time
from typing import Any, Dict, List

from fleet_api.base_call_api import APICallException, perform_api_call
from fleet_api.utils import add_capabilities, add_tags, replace_image_field
from loguru import logger
from munch import Munch, munchify

AGENT_ARTIFACT_SUFFIX = "/downloads/beats/elastic-agent"
AGENT_ARTIFACT_SUFFIX_SHORT = "/downloads/"

STAGING_ARTIFACTORY_URL = "https://staging.elastic.co/"
SNAPSHOT_ARTIFACTORY_URL = "https://snapshots.elastic.co/"


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

    url = f"{cfg.kibana_url}/api/fleet/fleet_server_hosts"

    try:
        response = perform_api_call(
            method="GET",
            url=url,
            auth=cfg.auth,
        )
        response_obj = munchify(response)
        for fleet_server in response_obj["items"]:
            if fleet_server.is_default:
                return fleet_server.host_urls[0]
        raise KeyError("No default fleet server found")
    except APICallException as api_ex:
        logger.error(
            f"API call failed, status code {api_ex.status_code}. Response: {api_ex.response_text}",
        )
        raise


def create_kubernetes_manifest(cfg: Munch, params: Munch):
    """Create a Kubernetes manifest based on the provided configuration and parameters.

    Args:
        cfg (Munch): Config object containing authentication data.
        params (Munch): The parameters object containing additional information.
    """
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
        manifest_yaml = response_obj.item
        if params.docker_image_override:
            manifest_yaml = replace_image_field(
                response_obj.item,
                new_image=params.docker_image_override,
            )
        if hasattr(params, "capabilities") and params.capabilities:
            manifest_yaml = add_capabilities(yaml_content=manifest_yaml)
        with codecs.open(params.yaml_path, "w", encoding="utf-8-sig") as k8s_yaml:
            k8s_yaml.write(manifest_yaml)
        logger.info(f"KSPM manifest is available at: '{params.yaml_path}'")
    except APICallException as api_ex:
        logger.error(
            f"API call failed, status code {api_ex.status_code}. Response: {api_ex.response_text}",
        )


def get_cnvm_template(url: str, template_path: str, cnvm_tags: str):
    """
    Download a CloudFormation template from a specified URL,
    add custom tags to it, and save it to a file.

    Args:
        url (str): The URL to download the CloudFormation template.
        template_path (str): The file path where the modified template will be saved.
        cnvm_tags (str): Custom tags to be added to the template in the format "key1=value1 key2=value2 ...".

    Returns:
        None

    Raises:
        APICallException: If there's an issue with the API call.
    """
    try:
        template_yaml = perform_api_call(
            method="GET",
            url=url,
            return_json=False,
        )
        template_yaml = add_tags(tags=cnvm_tags, yaml_content=template_yaml)

        with codecs.open(template_path, "w", encoding="utf-8") as cnvm_yaml:
            cnvm_yaml.write(template_yaml)
        logger.info(f"CNVM template is available at: '{template_path}'")
    except APICallException as api_ex:
        logger.error(
            f"API call failed, status code {api_ex.status_code}. Response: {api_ex.response_text}",
        )


def get_arm_template(url: str, template_path: str):
    """
    Download an ARM template from a specified URL and save it to a file.

    Args:
        url (str): The URL to download the ARM template.
        template_path (str): The file path where the modified template will be saved.

    Returns:
        None

    Raises:
        APICallException: If there's an issue with the API call.
    """
    try:
        template_json = perform_api_call(
            method="GET",
            url=url,
            return_json=True,
        )

        with open(template_path, "w") as arm_json:
            json.dump(template_json, arm_json)
        logger.info(f"ARM template is available at: '{template_path}'")
    except APICallException as api_ex:
        logger.error(
            f"API call failed, status code {api_ex.status_code}. Response: {api_ex.response_text}",
        )


def get_build_info(version: str) -> str:
    """
    Retrieve the build ID for a specific version of Elastic.

    Args:
        version (str): The version of Elastic.

    Returns:
        str: The build ID of the specified version.

    Raises:
        APICallException: If the API call to retrieve the build ID fails.
    """
    if is_snapshot(version):
        url = f"https://snapshots.elastic.co/latest/{version}.json"
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


def get_artifact_server(version: str, is_short_url: bool = False) -> str:
    """
    Retrieve the artifact server URL for a specific version of Elastic.

    Args:
        elastic_version (str): The version of Elastic.
        is_short_url (bool, optional): Indicates whether to use the short artifact URL.
                                          Defaults to False.

    Returns:
        str: The artifact server URL for the specified Elastic version.

    Raises:
        APICallException: If the API call to retrieve the artifact server fails.
    """
    if is_snapshot(version):
        url = SNAPSHOT_ARTIFACTORY_URL
    else:
        url = STAGING_ARTIFACTORY_URL

    artifacts_suffix = AGENT_ARTIFACT_SUFFIX
    if is_short_url:
        artifacts_suffix = AGENT_ARTIFACT_SUFFIX_SHORT

    return url + get_build_info(version) + artifacts_suffix


def is_snapshot(version: str) -> bool:
    """
    Determine if the specified version is a snapshot version.

    Args:
        version (str): The version of Elastic.

    Returns:
        bool: True if the version is a snapshot version, False otherwise.
    """
    return "SNAPSHOT" in version


def get_stack_latest_version() -> str:
    """
    Retrieve the latest version of the stack from the Elastic snapshots API.

    Returns:
        str: The latest version of the stack.

    Raises:
        APICallException: If the API call to retrieve the version fails.

    """
    url = "https://snapshots.elastic.co/latest/master.json"
    try:
        response = perform_api_call(
            method="GET",
            url=url,
        )
        return response.get("version", "")

    except APICallException as api_ex:
        logger.error(
            f"API call failed, status code {api_ex.status_code}. Response: {api_ex.response_text}",
        )
        return ""


def get_package_version(
    cfg: Munch,
    package_name: str = "cloud_security_posture",
    prerelease: bool = True,
) -> str:
    """
    Retrieve the version of a specified package.

    Args:
        cfg (Munch): A configuration object containing Kibana URL, authentication details, etc.
        package_name (str, optional): The name of the package to retrieve the version for.
                                      Default is "cloud_security_posture".
        prerelease (bool, optional): A flag indicating whether to include prerelease versions.
                                     Default is True.

    Returns:
        str: The version of the specified package, or None if the API call fails or the package is not found.
    """
    url = f"{cfg.kibana_url}/api/fleet/epm/packages"

    request_params = {
        "prerelease": prerelease,
    }

    try:
        response = perform_api_call(
            method="GET",
            url=url,
            auth=cfg.auth,
            params={"params": request_params},
        )

        cloud_security_posture_version = None
        for package in response["response"]:
            if package.get("name", "") == package_name:
                cloud_security_posture_version = package.get("version", "")
                break

        return cloud_security_posture_version
    except APICallException as api_ex:
        logger.error(
            f"API call failed, status code {api_ex.status_code}. Response: {api_ex.response_text}",
        )
        return None


def get_package(
    cfg: Munch,
    package_name: str = "cloud_security_posture",
    is_full: bool = True,
    prerelease: bool = False,
) -> Dict[str, Any]:
    """
    Retrieve package information from the Elastic Fleet Server API.

    Args:
        cfg (Munch): Configuration data.
        package_name (str, optional): The name of the package to retrieve.
                                      Default is "cloud_security_posture".
        is_full (bool, optional): Whether to retrieve full package information. Default is True.
        prerelease (bool, optional): Whether to include prerelease versions. Default is False.

    Returns:
        Dict[str, Any]: A dictionary containing the package information
                        or an empty dictionary if the API call fails.
    """
    url = f"{cfg.kibana_url}/api/fleet/epm/packages/{package_name}"

    request_params = {
        "full": is_full,
        "prerelease": prerelease,
    }

    try:
        response = perform_api_call(
            method="GET",
            url=url,
            auth=cfg.auth,
            params={"params": request_params},
        )
        return response.get("response", {})
    except APICallException as api_ex:
        logger.error(
            f"API call failed, status code {api_ex.status_code}. Response: {api_ex.response_text}",
        )
        raise


def update_package_version(cfg: Munch, package_name: str, package_version: str):
    """
    Updates the version of a package.

    Args:
        cfg (Munch): Configuration object containing Kibana URL, authentication details, etc.
        package_version (str): The version to update the 'cloud_security_posture' package to.

    Raises:
        APICallException: If the API call fails with an error.

    """
    url = f"{cfg.kibana_url}/api/fleet/epm/packages/{package_name}/{package_version}"
    try:
        perform_api_call(
            method="POST",
            url=url,
            auth=cfg.auth,
            params={
                "json": {
                    "force": True,
                    "ignore_constraints": True,
                },
            },
        )

    except APICallException as api_ex:
        logger.error(
            f"API call failed, status code {api_ex.status_code}. Response: {api_ex.response_text}",
        )


def bulk_upgrade_agents(cfg: Munch, agent_ids: List[str], version: str, source_uri: str) -> str:
    """
    Upgrade a list of agents to a specified version using the Kibana API.

    Args:
        cfg (Munch): Configuration object containing Kibana URL and authentication details.
        agent_ids (List[str]): List of agent IDs to upgrade.
        version (str): The version to upgrade to.
        source_uri (str): The source URI for the agent package.

    Returns:
        str: The action ID of the upgrade.

    Raises:
        APICallException: If the API call fails with a non-200 status code.
    """
    url = f"{cfg.kibana_url}/api/fleet/agents/bulk_upgrade"
    json_data = {
        "agents": agent_ids,
        "version": version,
        "source_uri": source_uri,
    }
    logger.info(f"Source URI: {source_uri}")
    try:
        response = perform_api_call(
            method="POST",
            url=url,
            auth=cfg.auth,
            params={"json": json_data},
        )
        action_id = response.get("actionId")
        if not action_id:
            raise APICallException(
                response.status_code,
                "API response did not include an actionId",
            )
        logger.info(f"Agents '{agent_ids}' upgrade to version '{version}' is started")
        logger.info(f"Action status id: {action_id}")
        return action_id
    except APICallException as api_ex:
        logger.error(
            f"API call failed, status code {api_ex.status_code}. Response: {api_ex.response_text}",
        )
        raise APICallException(api_ex.status_code, api_ex.response_text) from api_ex


def get_action_status(cfg: Munch) -> List[dict]:
    """
    Retrieve action status for agents using the Kibana API.

    Args:
        cfg (Munch): Configuration object containing Kibana URL and authentication details.

    Returns:
        List[dict]: A list of action status items.

    Raises:
        APICallException: If the API call fails with a non-200 status code.
    """
    url = f"{cfg.kibana_url}/api/fleet/agents/action_status"

    try:
        response = perform_api_call(
            method="GET",
            url=url,
            auth=cfg.auth,
        )
        return response.get("items", [])
    except APICallException as api_ex:
        logger.error(
            f"API call failed, status code {api_ex.status_code}. Response: {api_ex.response_text}",
        )
        raise APICallException(api_ex.status_code, api_ex.response_text) from api_ex


def wait_for_action_status(
    cfg: Munch,
    target_action_id: str,
    target_type: str,
    target_status: str,
    timeout_secs: int = 600,
):
    """
    Wait for a specific action status to match the target criteria.

    Args:
        cfg (Munch): Configuration object containing Kibana URL and authentication details.
        target_action_id (str): The action ID to match.
        target_type (str): The target action type to match.
        target_status (str): The target status to match.
        timeout_secs (int): Maximum time to wait in seconds (default is 600 seconds).

    Returns:
        bool: True if the target criteria is met, False if the timeout is reached.

    Raises:
        APICallException: If the API call fails with a non-200 status code.
    """
    start_time = time.time()
    while True:
        action_status = get_action_status(cfg)
        for item in action_status:
            if item.get("actionId") == target_action_id:
                logger.info(f"Type: {item.get('type')}, Status: {item.get('status')}")
                if item.get("type") == target_type and item.get("status") == target_status:
                    return True  # Found the target criteria

        if time.time() - start_time >= timeout_secs:
            logger.error(f"Agent upgrade process reached a timeout of {timeout_secs} seconds.")
            return False  # Timeout reached

        time.sleep(2)  # Fixed sleep interval of 1 second


def get_telemetry(cfg: Munch) -> dict:
    """
    This function create API call to Kibana snapshot telemetry api and return is payload.

    Args:
        cfg: configuration object contains kibana host and auth

    Returns:
        dict: Telemetry payload
    """
    url = f"{cfg.kibana_url}/internal/telemetry/clusters/_stats"
    headers = {
        "Content-Type": "application/json",
        "kbn-xsrf": "true",
        "elastic-api-version": "2",
        "x-elastic-internal-origin": "Kibana",
    }
    try:
        response = perform_api_call(
            method="POST",
            url=url,
            headers=headers,
            auth=cfg.auth,
            params={
                "json": {
                    "unencrypted": True,
                },
            },
        )
        return response
    except APICallException as api_ex:
        logger.error(
            f"API call failed, status code {api_ex.status_code}. Response: {api_ex.response_text}",
        )
        raise
