"""
This module provides functionality for performing API calls and handling API call exceptions.
It utilizes the 'requests' library for making HTTP requests.

Module contents:
    - APICallException: Exception class raised for API call failures.
    - perform_api_call: Function for making API calls and handling the response.

Dependencies:
    - requests: Library for making HTTP requests
"""
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


def perform_api_call(method, url, headers=None, auth=None, params=None):
    """
    Perform an API call using the provided parameters.

    Args:
        method (str): The HTTP method for the API call (e.g., 'GET', 'POST', 'PUT', 'DELETE').
        url (str): The URL of the API endpoint.
        headers (dict, optional): The headers to be included in the API request.
                                  If not provided, default headers will be used.
        auth (tuple or None, optional): The authentication tuple (username, password)
                                        for basic authentication. Set to None for no authentication.
                                        Defaults to None.
        params (dict, optional): The parameters to be included in the API request.
                                 Defaults to None.

    Returns:
        dict: The JSON response from the API call.

    Raises:
        APICallException: If the API call returns a non-200 status code.
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

    response = requests.request(method=method, url=url, headers=headers, auth=auth, **params)
    if response.status_code != 200:
        raise APICallException(response.status_code, response.text)

    return response.json()


def download_file(url, destination, timeout=30):
    """
    Download a file from a URL and save it to the specified destination.

    Args:
        url (str): The URL of the file to download.
        destination (str): The path where the downloaded file will be saved.
        timeout (int, optional): The maximum time (in seconds) to wait for the server's response.
                                 Defaults to 30 seconds.

    Raises:
        APICallException: If there's an issue with the HTTP request.
        IOError: If there's an issue with saving the downloaded file.
    """
    try:
        response = requests.get(url, stream=True, timeout=timeout)
        response.raise_for_status()

        with open(destination, "wb") as file:
            for chunk in response.iter_content(chunk_size=8192):
                file.write(chunk)

        logger.info(f"File downloaded to {destination}")
    except requests.exceptions.RequestException as ex:
        raise APICallException(500, f"HTTP Request Error: {ex}") from ex
    except IOError as io_ex:
        raise IOError(f"IO Error: {io_ex}") from io_ex
