"""
Wait for agents to be enrolled for a given policies
If the expected number of agents is not enrolled within the timeout, the test will fail
"""
import sys
import time
import re
from api.agent_policy_api import get_agents
import configuration_fleet as cnfg
from state_file_manager import state_manager
from loguru import logger

TIMEOUT = 600


class AgentPolicyEnrolled:
    """
    Class to represent the details of an enrolled agent.
    """
    def __init__(self, count: int, tags: list[str]) -> None:
        """
        Args:
            count (int): Number of agents to be enrolled.
            tags (list[str]): Tags of the agent.
        """
        self.count = count
        self.tags = tags


def get_expected_agents() -> dict:
    """
    Returns:
        dict: The name of the policy and the number of agents expected to be enrolled
    """
    logger.info("Loading agent policies state file")
    policies_dict = {}
    for policy in state_manager.get_policies():
        policies_dict[policy.agnt_policy_id] = AgentPolicyEnrolled(policy.expected_agents, policy.expected_tags)
    return policies_dict


def get_actual_agents() -> dict:
    """
    Returns:
        dict: The name of the policy and the number of agents enrolled
    """
    agents = get_agents(cfg=cnfg.elk_config)
    policies_dict = {}
    for agent in agents:
        policies_dict[agent.policy_id] = policies_dict.get(agent.policy_id, 0) + 1
    return policies_dict


def verify_agent_count(expected: dict, actual: dict) -> bool:
    """
    Verify that the expected number of agents are enrolled
    """
    result = True
    for policy_id, expected_count in expected.items():
        if policy_id not in actual:
            result = False
            logger.info(f"Policy {policy_id} not found in the actual agents mapping")
        elif actual[policy_id] != expected_count:
            result = False
            logger.info(f"Policy {policy_id} expected {expected_count} agents, but got {actual[policy_id]}")
    return result


def verify_agent_tags(agent, expected_agents) -> bool:
    expected_tags = []
    if agent.policy_id in expected_agents:
        expected_tags = expected_agents[agent.policy_id].tags
    for pattern in expected_tags:
        pattern_exist = False
        for tag in agent.tags:
            if re.match(pattern, tag):
                pattern_exist = True
                break
        if not pattern_exist:
            logger.info(f"Agent {agent.id} does not have the expected tag {pattern}")
            return False
    return True


def verify_agents_enrolled() -> bool:
    """
    Construct a dictionary of the expected agents and the actual agents
    Returns:
        bool: True if the expected agents are enrolled, False otherwise
    """
    expected = get_expected_agents()
    agents = get_agents(cfg=cnfg.elk_config)
    actual = {}
    for agent in agents:
        if verify_agent_tags(agent, expected):
            actual[agent.policy_id] = actual.get(agent.policy_id, 0) + 1
    return verify_agent_count(expected, actual)


def wait_for_agents_enrolled(timeout) -> bool:
    """
    Wait for agents to be enrolled
    """
    start_time = time.time()
    while time.time() - start_time < timeout:
        if verify_agents_enrolled():
            return True
        time.sleep(10)

    return False


if __name__ == "__main__":
    logger.info("Waiting for agents to be enrolled...")
    if wait_for_agents_enrolled(TIMEOUT):
        logger.info("All agents enrolled successfully")
    else:
        logger.error("Not all agents were enrolled")
        sys.exit(1)
