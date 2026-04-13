"""
This module provides functionality for performing API calls and handling API call exceptions.
It utilizes the 'requests' library for making HTTP requests.

Module contents:
    - APICallException: Exception class raised for API call failures.
    - perform_api_call: Function for making API calls and handling the response.

Dependencies:
    - requests: Library for making HTTP requests
"""

import random
import time

import requests
from loguru import logger

TRANSIENT_STATUSES = {502, 503, 504}
TRANSIENT_EXCEPTIONS = (
    requests.exceptions.ConnectionError,
    requests.exceptions.Timeout,
    requests.exceptions.ChunkedEncodingError,
)


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
    retries=8,
    retry_backoff_sec=2.0,
    retry_max_sleep_sec=60.0,
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
        retries (int, optional): Number of retry attempts on transient failures. Defaults to 8.
        retry_backoff_sec (float, optional): Base backoff seconds for exponential delay. Defaults to 2.0.
        retry_max_sleep_sec (float, optional): Maximum sleep seconds between retries. Defaults to 60.0.

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

    attempt = 0
    while True:
        attempt += 1
        try:
            response = requests.request(
                method=method,
                url=url,
                headers=headers,
                auth=auth,
                timeout=60,
                **params,
            )
        except TRANSIENT_EXCEPTIONS as exc:
            if attempt <= retries:
                sleep = _retry_sleep(attempt, retry_backoff_sec, retry_max_sleep_sec)
                logger.warning(
                    f"Transient request error on attempt {attempt}/{retries} ({exc}). "
                    f"Retrying in {sleep:.1f}s..."
                )
                time.sleep(sleep)
                continue
            raise

        if response.status_code not in ok_statuses:
            if response.status_code in TRANSIENT_STATUSES and attempt <= retries:
                sleep = _retry_sleep(attempt, retry_backoff_sec, retry_max_sleep_sec)
                logger.warning(
                    f"Transient HTTP {response.status_code} on attempt {attempt}/{retries}. "
                    f"Retrying in {sleep:.1f}s... Response: {response.text[:200]}"
                )
                time.sleep(sleep)
                continue
            raise APICallException(response.status_code, response.text)

        if not return_json:
            return response.content
        if response.status_code == 204 or not (response.content or b"").strip():
            return {}
        return response.json()


def _retry_sleep(attempt: int, base: float, maximum: float) -> float:
    """Return a jittered exponential backoff sleep duration, capped at maximum."""
    sleep = base * (2 ** (attempt - 1))
    sleep = sleep * (0.75 + random.random() * 0.5)
    return min(maximum, sleep)


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
