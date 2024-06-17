"""
Define a class to manage the policies state using a file.
Exports state_manager object as a singleton.
"""

import json
from enum import Enum
from pathlib import Path

from fleet_api.utils import delete_file
from loguru import logger

__state_file = Path(__file__).parent / "state_data.json"


class HostType(Enum):
    """
    Enumeration representing different host types for deployment.

    The `HostType` enumeration defines constants for various host types,
    such as Kubernetes or Linux-based deployments.

    Attributes:
        KUBERNETES (str): Represents a Kubernetes-based deployment.
        LINUX_TAR (str): Represents a Linux-based deployment using TAR archives.
    """

    KUBERNETES = "kubernetes"
    LINUX_TAR = "linux"


class PolicyStateEncoder(json.JSONEncoder):
    """
    Custom JSON encoder for PolicyState objects.
    """

    def default(self, o):
        """
        Encode a PolicyState object.
        """
        return o.__dict__


class PolicyState:
    """
    Class to represent a policy state.
    """

    def __init__(
        self,
        agnt_policy_id: str,
        pkg_policy_id: str,
        expected_agents: int,
        expected_tags: list[str],
        host_type: HostType,
        integration_name: str,
    ):
        """
        Args:
            agnt_policy_id (str): ID of the agent policy.
            pkg_policy_id (str): ID of the package policy.
            expected_agents (int): Expected number of deployed agents.
            expected_tags: (list(int)): List of expected tags count.
            host_type (HostType): Deployment host type
            integration_name (str): Name of installed integration
        """
        self.agnt_policy_id = agnt_policy_id
        self.pkg_policy_id = pkg_policy_id
        self.expected_agents = expected_agents
        self.expected_tags = expected_tags
        self.host_type = host_type
        self.integration_name = integration_name


class StateFileManager:
    """
    Class to manage the policies state using a file.
    """

    def __init__(self, state_file: Path):
        """
        Args:
            state_file (Path): Path of a file to cache the state.
        """
        self.__state_file = state_file
        self.__policies = []
        self.__load()

    def __load(self) -> None:
        """
        Load the policies data from a file.
        """
        if not self.__state_file.exists():
            return
        with self.__state_file.open("r") as policies_file:
            policies_data = json.load(policies_file)
            for policy in policies_data["policies"]:
                self.__policies.append(PolicyState(**policy))
        logger.info(f" {len(self.__policies)} policies loaded to state from {self.__state_file}")

    def __save(self) -> None:
        """
        Save the policies data to a file.
        """
        policies_data = {"policies": self.__policies}
        with self.__state_file.open("w") as policies_file:
            json.dump(policies_data, policies_file, cls=PolicyStateEncoder)
        logger.info(f" {len(self.__policies)} policies saved to state in {self.__state_file}")

    def add_policy(self, data: PolicyState) -> None:
        """
        Add a policy to the current state.

        Args:
            data (PolicyState): Policy data to be added.
        """
        self.__policies.append(data)
        self.__save()

    def get_policies(self) -> list[PolicyState]:
        """
        Get the current state.

        Returns:
            list: List of policies.
        """
        return self.__policies

    def delete_by_package_policy(self, pkg_policy_id: str) -> None:
        """
        Delete policies with a given package policy ID.

        Args:
            pkg_policy_id (str): Package policy ID to match for deletion.
        """
        self.__policies = [policy for policy in self.__policies if policy.pkg_policy_id != pkg_policy_id]
        self.__save()
        logger.info(f"Policies with package policy ID {pkg_policy_id} deleted from state.")

    def delete_all(self) -> None:
        """
        Delete the current state.
        """
        self.__policies = []
        delete_file(self.__state_file)


state_manager = StateFileManager(__state_file)
