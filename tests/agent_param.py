"""
Generate agent parameterization for pytest.
"""
import os
from loguru import logger
from commonlib.fleet_api import get_agents
from configuration import elasticsearch


CIS_AWS_COMPONENT = "cloudbeat/cis_aws"
AWS_AGENT_PARAM = "cis_aws_agent"

CIS_GCP_COMPONENT = "cloudbeat/cis_gcp"
GCP_AGENT_PARAM = "cis_gcp_agent"

CIS_AZURE_COMPONENT = "cloudbeat/cis_azure"
AZURE_AGENT_PARAM = "cis_azure_agent"

PARAM_COMPONENT_MAP = {
    AWS_AGENT_PARAM: CIS_AWS_COMPONENT,
    GCP_AGENT_PARAM: CIS_GCP_COMPONENT,
    AZURE_AGENT_PARAM: CIS_AZURE_COMPONENT,
}


class AgentExpectedMapping:
    def __init__(self):
        agentless = os.getenv("TEST_AGENTLESS", False)
        self.expected_map = {
            CIS_AWS_COMPONENT: 1,
            CIS_GCP_COMPONENT: 1,
            CIS_AZURE_COMPONENT: 1,
        }
        if agentless:
            self.expected_map[CIS_AWS_COMPONENT] += 1


class AgentComponentMapping:
    """
    This class is used to map agent IDs that are running each component.
    """
    def __init__(self):
        self.component_map = {
            CIS_AWS_COMPONENT: [],
            CIS_GCP_COMPONENT: [],
            CIS_AZURE_COMPONENT: [],
        }

    def load_map(self):
        agents = get_agents(elasticsearch)
        logger.info(f"found {len(agents)} agents")
        for integration in self.component_map:
            for agent in agents:
                for component in agent.components:
                    if integration in component.id:
                        self.component_map[integration].append(agent.id)


class AgentComponentHelper:
    """
    This class is used to assert that the expected number of agents are running each component.
    """
    def __init__(self):
        self.expected_map = AgentExpectedMapping()
        self.component_map = AgentComponentMapping()

    def load_map(self) -> None:
        """
        Load the expected and actual agent component mapping.
        """
        self.component_map.load_map()

    def assert_agents(self) -> None:
        """
        Assert that the expected number of agents are running each component.
        """
        for component in self.expected_map.expected_map:
            expected_count = self.expected_map.expected_map[component]
            actual_count = len(self.component_map.component_map[component])
            assert actual_count == expected_count, f"Expected {expected_count} agents running component {component}, got {actual_count}"
    
    def parameterize_agent(self, param_name: str) -> list[str]:
        """
        Parameterize the agent ID for a component.
        """
        component = PARAM_COMPONENT_MAP[param_name]
        return self.component_map.component_map[component]
