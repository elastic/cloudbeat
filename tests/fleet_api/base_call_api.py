"""
This module provides functionality for performing API calls and handling API call exceptions.
It utilizes the 'requests' library for making HTTP requests.

Module contents:
    - APICallException: Exception class raised for API call failures.
    - perform_api_call: Function for making API calls and handling the response.

Dependencies:
    - requests: Library for making HTTP requests
    - loguru: Retry logging in perform_api_call
"""

import time

import requests
from loguru import logger


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


def _retry_delay_sec(attempt: int, retry_backoff_sec: float, retry_backoff_max_sec: float) -> float:
    return min(retry_backoff_sec * (2**attempt), retry_backoff_max_sec)


def _http_status_should_retry(response) -> bool:
    if response.status_code == 429 or response.status_code >= 500:
        return True
    return response.status_code == 400 and "not available with the current configuration" in response.text


def _sleep_transport_retry(
    method: str,
    url: str,
    attempt: int,
    max_retries: int,
    exc: BaseException,
    retry_backoff_sec: float,
    retry_backoff_max_sec: float,
) -> None:
    delay = _retry_delay_sec(attempt, retry_backoff_sec, retry_backoff_max_sec)
    logger.warning(
        "perform_api_call: {} {} raised {} (attempt {}/{}), retrying in {:.0f}s",
        method,
        url,
        type(exc).__name__,
        attempt + 1,
        max_retries,
        delay,
    )
    time.sleep(delay)


def _sleep_transient_http_retry(
    method: str,
    url: str,
    attempt: int,
    max_retries: int,
    status_code: int,
    retry_backoff_sec: float,
    retry_backoff_max_sec: float,
) -> None:
    delay = _retry_delay_sec(attempt, retry_backoff_sec, retry_backoff_max_sec)
    logger.warning(
        "perform_api_call: {} {} returned {} (attempt {}/{}), retrying in {:.0f}s",
        method,
        url,
        status_code,
        attempt + 1,
        max_retries,
        delay,
    )
    time.sleep(delay)


def _normalize_perform_api_call_inputs(headers, auth, params, ok_statuses):
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
    return headers, auth, params, ok_statuses


def _request_until_success_or_raise(
    method,
    url,
    headers,
    auth,
    params,
    ok_statuses,
    max_retries,
    retry_backoff_sec,
    retry_backoff_max_sec,
) -> requests.Response:
    """Perform HTTP request with retries; return ``response`` on success (status in ok_statuses)."""
    for attempt in range(max_retries):
        try:
            response = requests.request(method=method, url=url, headers=headers, auth=auth, **params)
        except requests.exceptions.RequestException as exc:
            if attempt < max_retries - 1:
                _sleep_transport_retry(
                    method,
                    url,
                    attempt,
                    max_retries,
                    exc,
                    retry_backoff_sec,
                    retry_backoff_max_sec,
                )
                continue
            raise APICallException(0, str(exc)) from exc
        if response.status_code in ok_statuses:
            return response
        if _http_status_should_retry(response) and attempt < max_retries - 1:
            _sleep_transient_http_retry(
                method,
                url,
                attempt,
                max_retries,
                response.status_code,
                retry_backoff_sec,
                retry_backoff_max_sec,
            )
            continue
        raise APICallException(response.status_code, response.text)
    raise AssertionError("unreachable: retry loop must return or raise")


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
        ValueError: If ``max_retries`` is less than 1.
        APICallException: If the API call returns a non-success status code after retries,
            or if transport errors (e.g. connection, DNS, timeout) persist after retries
            (in the latter case ``status_code`` is 0).
    """
    if max_retries < 1:
        raise ValueError("max_retries must be at least 1")
    headers, auth, params, ok_statuses = _normalize_perform_api_call_inputs(headers, auth, params, ok_statuses)
    response = _request_until_success_or_raise(
        method,
        url,
        headers,
        auth,
        params,
        ok_statuses,
        max_retries,
        retry_backoff_sec,
        retry_backoff_max_sec,
    )

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
        raise ValueError("Stack version must be provided.")
    return version.startswith("9.") or version.startswith("8.17")
