"""
Generate agent parameterization for pytest.
"""
import os
from commonlib.fleet_api import get_agents
from configuration import elasticsearch


CIS_AWS_COMPONENT = "cloudbeat/cis_aws"
CIS_GCP_COMPONENT = "cloudbeat/cis_gcp"
CIS_AZURE_COMPONENT = "cloudbeat/cis_azure"

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
