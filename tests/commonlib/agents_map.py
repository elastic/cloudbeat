"""
Generate agent parameterization for pytest.
"""
import os
import sys
from loguru import logger
from munch import Munch
from configuration import elasticsearch

sys.path.append(os.path.relpath("../deploy/test-environments/fleet_api/src"))
from api.agent_policy_api import get_agents  # pylint: disable=wrong-import-position # noqa: E402

CIS_AWS_COMPONENT = "cloudbeat/cis_aws"
CIS_GCP_COMPONENT = "cloudbeat/cis_gcp"
CIS_AZURE_COMPONENT = "cloudbeat/cis_azure"


class AgentExpectedMapping:
    """
    This class is used to map expected number of agents that are running each component.
    """

    def __init__(self):
        agentless = os.getenv("TEST_AGENTLESS", "false") == "true"
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
        """
        Load the components map with the agent IDs.
        """
        cfg = Munch()
        cfg.auth = elasticsearch.basic_auth
        cfg.kibana_url = elasticsearch.kibana_url

        agents = get_agents(cfg)
        logger.info(f"found {len(agents)} agents")
        for integration in self.component_map.copy():
            for agent in agents:
                for component in agent.components:
                    if integration in component.id:
                        self.component_map[integration].append(agent.id)
