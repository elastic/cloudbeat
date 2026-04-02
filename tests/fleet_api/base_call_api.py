"""
This module provides functionality for performing API calls and handling API call exceptions.
It utilizes the 'requests' library for making HTTP requests.

Module contents:
    - APICallException: Exception class raised for API call failures.
    - perform_api_call: Function for making API calls and handling the response.

Dependencies:
    - requests: Library for making HTTP requests
"""

import time

import requests


class APICallException(Exception):
    """
    Exception raised for API call failures.

    Attributes:
        status_code (int): The HTTP status code of the failed API call.
        response_text (str): The response text of the failed API call.
    """

    def __init__(self, status_code, response_text):
        """
        Initialize the APICallException.

        Args:
            status_code (int): The HTTP status code of the failed API call.
            response_text (str): The response text of the failed API call.
        """
        self.status_code = status_code
        self.response_text = response_text


def perform_api_call(
    method,
    url,
    return_json=True,
    headers=None,
    auth=None,
    params=None,
    ok_statuses=None,
    max_retries: int = 8,
    retry_backoff_sec: float = 5.0,
    retry_backoff_max_sec: float = 30.0,
):
    """
    Perform an API call using the provided parameters.

    Args:
        method (str): The HTTP method for the API call (e.g., 'GET', 'POST', 'PUT', 'DELETE').
        url (str): The URL of the API endpoint.
        return_json (bool, optional): Indicates whether the function should return
                                      JSON data (default is True).
        headers (dict, optional): The headers to be included in the API request.
                                  If not provided, default headers will be used.
        auth (tuple or None, optional): The authentication tuple (username, password)
                                        for basic authentication. Set to None for no authentication.
                                        Defaults to None.
        params (dict, optional): The parameters to be included in the API request.
                                 Defaults to None.
        ok_statuses (tuple, optional): HTTP status codes treated as success. Defaults to (200,).

    Returns:
        dict or bytes: Parsed JSON (empty dict for 204 or empty body), or raw content.

    Raises:
        APICallException: If the API call returns a non-success status code.
    """
    if headers is None:
        headers = {
            "Content-Type": "application/json",
            "kbn-xsrf": "true",
        }
    if auth is None:
        auth = ()
    if params is None:
        params = {}
    if ok_statuses is None:
        ok_statuses = (200,)

    for attempt in range(max_retries):
        response = requests.request(method=method, url=url, headers=headers, auth=auth, **params)
        if response.status_code in ok_statuses:
            break
        _fleet_not_ready = (
            response.status_code == 400 and "not available with the current configuration" in response.text
        )
        if (response.status_code >= 500 or _fleet_not_ready) and attempt < max_retries - 1:
            delay = min(retry_backoff_sec * (2**attempt), retry_backoff_max_sec)
            print(
                f"perform_api_call: {method} {url} returned {response.status_code} "
                f"(attempt {attempt + 1}/{max_retries}), retrying in {delay:.0f}s",
            )
            time.sleep(delay)
            continue
        raise APICallException(response.status_code, response.text)

    if response.status_code not in ok_statuses:
        raise APICallException(response.status_code, response.text)

    if not return_json:
        return response.content
    if response.status_code == 204 or not (response.content or b"").strip():
        return {}
    return response.json()


def uses_new_fleet_api_response(version: str) -> bool:
    """
    Determine if the specified version uses the new Fleet API response format.

    Args:
        version (str): Elastic stack version.

    Returns:
        bool: True if the version uses the new Fleet API response format, False otherwise.
    """
    if not version:
        return ValueError("Stack version must be provided.")
    return version.startswith("9.") or version.startswith("8.17")
