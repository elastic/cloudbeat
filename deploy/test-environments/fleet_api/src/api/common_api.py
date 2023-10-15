"""
This module contains API calls related to Fleet settings
"""
import codecs
from typing import Dict, Any
from munch import Munch, munchify
from loguru import logger
from api.base_call_api import APICallException, perform_api_call
from utils import (
    replace_image_field,
    add_capabilities,
    add_tags,
)

AGENT_ARTIFACT_SUFFIX = "/downloads/beats/elastic-agent"

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
        return


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
        return


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


def get_artifact_server(version: str) -> str:
    """
    Retrieve the artifact server for a specific version of Elastic.

    Args:
        version (str): The version of Elastic.

    Returns:
        str: The artifact server of the specified version.

    Raises:
        APICallException: If the API call to retrieve the artifact server fails.
    """

    if is_snapshot(version):
        url = SNAPSHOT_ARTIFACTORY_URL
    else:
        url = STAGING_ARTIFACTORY_URL

    return url + get_build_info(version) + AGENT_ARTIFACT_SUFFIX


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
        return {}


def update_package_version(cfg: Munch, package_name: str, package_version: str):
    """
    Updates the version of a package.

    Args:
        cfg (Munch): Configuration object containing Kibana URL, authentication details, etc.
        package_version (str): The version to update the 'cloud_security_posture' package to.

    Raises:
        APICallException: If the API call fails with an error.

    """
    # pylint: disable=duplicate-code
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
