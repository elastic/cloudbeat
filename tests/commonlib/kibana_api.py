"""
Minimal Kibana API client used by the test suite (Fleet agents and cloud security telemetry).
"""

import time

import requests
from loguru import logger
from munch import Munch, munchify

REQUEST_TIMEOUT_SEC = 60
MAX_ATTEMPTS = 3
RETRY_BACKOFF_SEC = 5.0


def request_json(method: str, cfg: Munch, path: str, **kwargs) -> dict:
    """
    Call a Kibana API and return the parsed JSON body, retrying on failures.

    Args:
        method (str): HTTP method.
        cfg (Munch): Configuration with ``kibana_url`` and ``auth`` (a ``(user, password)`` tuple).
        path (str): API path, starting with ``/``.
        **kwargs: Extra arguments for ``requests.request`` (``json``, ``params``, ``headers``).

    Raises:
        RuntimeError: If every attempt failed.
    """
    url = f"{cfg.kibana_url}{path}"
    headers = {"kbn-xsrf": "true", **kwargs.pop("headers", {})}
    last_error = ""
    for attempt in range(1, MAX_ATTEMPTS + 1):
        retryable = True
        try:
            response = requests.request(
                method,
                url,
                auth=cfg.auth,
                headers=headers,
                timeout=REQUEST_TIMEOUT_SEC,
                **kwargs,
            )
            if response.status_code == 200:
                return response.json()
            last_error = f"status code {response.status_code}"
            # Only rate limiting and server errors are worth retrying; 401, 403, 404 etc. fail fast.
            retryable = response.status_code == 429 or response.status_code >= 500
        except (requests.RequestException, ValueError) as ex:
            last_error = str(ex)
        logger.error(f"Kibana API {method} {path} failed: {last_error} (attempt {attempt}/{MAX_ATTEMPTS})")
        if not retryable:
            break
        if attempt < MAX_ATTEMPTS:
            time.sleep(RETRY_BACKOFF_SEC)
    raise RuntimeError(f"Kibana API {method} {path} failed: {last_error}")


def get_agents(cfg: Munch) -> list:
    """
    Retrieve the agents enrolled in Fleet.

    Returns:
        list: The enrolled agents, or an empty list if the API call keeps failing.
    """
    try:
        body = request_json("GET", cfg, "/api/fleet/agents")
    except RuntimeError:
        return []
    # Newer stacks return "items", older ones "list".
    return munchify(body.get("items", body.get("list", [])))


def get_telemetry(cfg: Munch) -> dict:
    """
    Fetch the Kibana snapshot telemetry payload.

    Raises:
        RuntimeError: If the API call keeps failing.
    """
    return request_json(
        "POST",
        cfg,
        "/internal/telemetry/clusters/_stats",
        headers={
            "Content-Type": "application/json",
            "elastic-api-version": "2",
            "x-elastic-internal-origin": "Kibana",
        },
        json={"unencrypted": True},
    )
