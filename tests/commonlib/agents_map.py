"""
Generate agent parameterization for pytest.
"""

from configuration import agent, elasticsearch
from fleet_api.agent_policy_api import get_agents
from loguru import logger
from munch import Munch

CIS_AWS_COMPONENT = "cloudbeat/cis_aws"
CIS_GCP_COMPONENT = "cloudbeat/cis_gcp"
CIS_AZURE_COMPONENT = "cloudbeat/cis_azure"


class AgentExpectedMapping:
    """
    This class is used to map expected number of agents that are running each component.
    """

    def __init__(self):
        self.expected_map = {
            CIS_AWS_COMPONENT: 1,
            CIS_GCP_COMPONENT: 1,
            CIS_AZURE_COMPONENT: 1,
        }
        if agent.agentless:
            self.expected_map[CIS_AWS_COMPONENT] += 1
            self.expected_map[CIS_AZURE_COMPONENT] += 1


class AgentComponentMapping:
    """
    This class is used to map agent IDs that are running each component.
    """

    def __init__(self):
        self.reset_map()

    def reset_map(self):
        """
        Reset the components map.
        """
        self.component_map = {
            CIS_AWS_COMPONENT: [],
            CIS_GCP_COMPONENT: [],
            CIS_AZURE_COMPONENT: [],
        }

    def load_map(self):
        """
        Load the components map with the agent IDs.
        """
        self.reset_map()
        cfg = Munch()
        cfg.auth = elasticsearch.basic_auth
        cfg.kibana_url = elasticsearch.kibana_url

        active_agents = get_agents(cfg)
        logger.info(f"found {len(active_agents)} agents")
        for integration in self.component_map.copy():
            for active in active_agents:
                for component in active.components:
                    if integration in component.id:
                        self.component_map[integration].append(active.id)
