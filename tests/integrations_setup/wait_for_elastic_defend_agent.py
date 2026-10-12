"""
Poll Fleet for the Elastic Defend agent enrolled onto the Endpoint Replay VM
and emit the fields endpoint_replay.py needs (--host-name, --agent-id,
--host-id), so nobody has to copy them out of the Fleet UI by hand.

Reads agent_policy_id from elastic_defend_hosts.json, written alongside this
script by install_elastic_defend_integration.py.

Required env: KIBANA_URL, ES_USER, ES_PASSWORD
Optional env: STACK_VERSION, WAIT_TIMEOUT_SEC (default 300), WAIT_POLL_INTERVAL_SEC (default 10)

NOTE: this assumes the standard Fleet agent metadata shape
(local_metadata.host.hostname / local_metadata.host.id). That has not been
verified against a live enrolled agent from this script yet -- if it
mismatches, the "no field" error below prints the full raw agent JSON so the
field names are easy to fix.
"""

import json
import os
import sys
import time
from pathlib import Path
from typing import Optional

import configuration_fleet as cnfg
from loguru import logger

from fleet_api.agent_policy_api import get_agents

HOSTS_METADATA_FILE = Path(__file__).parent / "elastic_defend_hosts.json"
GITHUB_OUTPUT_ENV = "GITHUB_OUTPUT"

TIMEOUT_SEC = int(os.getenv("WAIT_TIMEOUT_SEC", "300"))
POLL_INTERVAL_SEC = int(os.getenv("WAIT_POLL_INTERVAL_SEC", "10"))


def load_agent_policy_id() -> str:
    """Read agent_policy_id from the metadata file written by install_elastic_defend_integration.py."""
    if not HOSTS_METADATA_FILE.exists():
        logger.error(f"{HOSTS_METADATA_FILE} not found; run install_elastic_defend_integration.py first")
        sys.exit(1)
    meta = json.loads(HOSTS_METADATA_FILE.read_text(encoding="utf-8"))
    agent_policy_id = meta.get("agent_policy_id")
    if not agent_policy_id:
        logger.error(f"agent_policy_id missing from {HOSTS_METADATA_FILE}")
        sys.exit(1)
    return agent_policy_id


def find_enrolled_agent(agent_policy_id: str) -> Optional[dict]:
    """Return the first active agent enrolled in agent_policy_id, or None."""
    for agent in get_agents(cfg=cnfg.elk_config):
        if agent.get("policy_id") != agent_policy_id:
            continue
        if not agent.get("active", True):
            continue
        return agent
    return None


def extract_fields(agent: dict) -> tuple[str, str, str]:
    """Pull agent id, Fleet host name, and host id out of a Fleet agent document."""
    local_metadata = agent.get("local_metadata") or {}
    host = local_metadata.get("host") or {}
    agent_id = agent.get("id")
    host_name = host.get("hostname") or host.get("name")
    host_id = host.get("id")

    fields = (
        ("id", agent_id),
        ("local_metadata.host.hostname", host_name),
        ("local_metadata.host.id", host_id),
    )
    missing = [label for label, value in fields if not value]
    if missing:
        logger.error(f"Enrolled agent is missing expected field(s): {missing}")
        logger.error(f"Full agent document for debugging:\n{json.dumps(agent, indent=2, default=str)}")
        sys.exit(1)

    return agent_id, host_name, host_id


def write_outputs(agent_id: str, host_name: str, host_id: str) -> None:
    """Append agent-id/host-name/host-id to $GITHUB_OUTPUT and log them."""
    output_path = os.getenv(GITHUB_OUTPUT_ENV)
    lines = [f"agent-id={agent_id}", f"host-name={host_name}", f"host-id={host_id}"]
    if output_path:
        with open(output_path, "a", encoding="utf-8") as fh:
            fh.write("\n".join(lines) + "\n")
    logger.info("Discovered enrolled Elastic Defend agent:")
    for line in lines:
        logger.info(f"  {line}")


def main() -> None:
    """Poll Fleet until the Endpoint Replay agent enrolls, then emit its details."""
    agent_policy_id = load_agent_policy_id()
    logger.info(f"Waiting up to {TIMEOUT_SEC}s for an agent enrolled in policy {agent_policy_id}")

    deadline = time.time() + TIMEOUT_SEC
    agent = None
    while time.time() < deadline:
        agent = find_enrolled_agent(agent_policy_id)
        if agent is not None:
            break
        logger.info(f"No enrolled agent yet; sleeping {POLL_INTERVAL_SEC}s")
        time.sleep(POLL_INTERVAL_SEC)

    if agent is None:
        logger.error(f"Timed out after {TIMEOUT_SEC}s waiting for an agent in policy {agent_policy_id}")
        sys.exit(1)

    agent_id, host_name, host_id = extract_fields(agent)
    write_outputs(agent_id, host_name, host_id)


if __name__ == "__main__":
    main()
