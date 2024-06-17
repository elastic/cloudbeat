#!/usr/bin/env python
"""
This script fetches the available versions of Elastic Agent from the Elastic website and
performs various operations on the versions.

Functions:
- get_available_versions:
    Fetches the available versions of Elastic Agent.
- parse_version:
    Parses a version string into a tuple of integers.
- filter_versions:
    Filters a list of versions based on the given prefix or after a specific version.
- get_package_version:
    Retrieves the package version of cloud_security_posture for a given Kibana version.
- generate_job_matrix:
    Generates a job matrix based on the given versions.
- main:
    Retrieves available versions of Elastic Agent and performs operations on them.
"""
import argparse
import json
import os

import requests


def get_available_versions():
    """
    Fetches the available versions of Elastic Agent from the Elastic API.

    Returns:
        list: A list of the latest versions of Elastic Agent.

    Raises:
        Exception: If there is an error while fetching the versions list.
    """
    url = "https://www.elastic.co/api/product_versions"
    headers = {"Content-Type": "application/json"}

    try:
        response = requests.get(url, headers=headers, timeout=60)
        response.raise_for_status()
        json_body = response.json()

        versions_dict = {}
        for item in json_body[0]:
            title = item.get("title", "")
            version_number = item.get("version_number", "")
            # Filter out versions with no alphabetic characters in the version number
            if "Elastic Agent" in title and not any(char.isalpha() for char in version_number):
                major_minor = ".".join(version_number.split(".")[:2])
                versions_dict[major_minor] = max(versions_dict.get(major_minor, ""), version_number)

        # Sort versions and select the latest num_versions
        versions = sorted(
            list(versions_dict.values()),
            key=lambda x: tuple(map(int, x.split("."))),
            reverse=True,
        )

        return versions
    except requests.exceptions.RequestException as e:
        raise requests.exceptions.RequestException("Failed to fetch versions list") from e


def parse_version(version):
    """
    Parses a version string into a tuple of integers.

    Args:
        version (str): The version string to parse.

    Returns:
        tuple: A tuple of integers representing the parsed version.

    Example:
        >>> parse_version('1.2.3')
        (1, 2, 3)
    """
    return tuple(map(int, version.split(".")))


def filter_versions(versions, prefix=None, min_version=None):
    """
    Filter a list of versions based on the given prefix or after a specific version.

    Args:
        versions (list): A list of versions to filter.
        prefix (str, optional): Only include versions that start with this prefix. Defaults to None.
        min_version (str, optional): Only include versions that are greater than this version. Defaults to None.

    Returns:
        list: A filtered list of versions.

    """
    if prefix:
        return [version for version in versions if version.startswith(prefix)]

    if min_version:
        min_version_tuple = parse_version(min_version)
        return [version for version in versions if parse_version(version) > min_version_tuple]

    return []


def get_package_version(kibana_version: str):
    """
    Retrieves the package version of cloud_security_posture for a given Kibana version.

    Args:
        kibana_version (str): The Kibana version.

    Returns:
        str: The package version of cloud_security_posture.

    Raises:
        Exception: If there is an error while retrieving the package version.
    """
    url = f"https://epr.elastic.co/search?package=cloud_security_posture&kibana.version={kibana_version}"
    try:
        response = requests.get(url, timeout=10)
        response.raise_for_status()
        data = response.json()
        return data[0]["version"]
    except requests.exceptions.RequestException as e:
        raise requests.exceptions.RequestException("Failed to retrieve package version") from e


def generate_job_matrix(versions):
    """
    Generate a job matrix based on the given versions.

    Args:
        versions (list): A list of versions.

    Returns:
        str: A JSON string representing the job matrix.

    Raises:
        None

    """
    job_matrix = []
    for version in versions:
        package_version = get_package_version(version)
        if not package_version:
            raise ValueError(f"Package version not found for Kibana version {version}")
        job_matrix.append({"agent-version": version, "package-version": package_version})

    return json.dumps({"include": job_matrix})


def main():
    """
    Retrieve available versions of Elastic Agent.

    This function retrieves a specified number of available versions of Elastic Agent
    and generates a job matrix based on the versions.

    Returns:
        None
    """
    parser = argparse.ArgumentParser(description="Retrieve available versions of Elastic Agent.")
    parser.add_argument(
        "--after",
        type=str,
        default="8.11",
        help="Only include versions that are greater than this version.",
    )
    args = parser.parse_args()

    available_versions = get_available_versions()
    filtered_versions = filter_versions(available_versions, min_version=args.after)
    with open(os.environ["GITHUB_OUTPUT"], "a", encoding="utf-8") as fh:
        print(f"matrix={generate_job_matrix(filtered_versions)}", file=fh)


if __name__ == "__main__":
    main()
