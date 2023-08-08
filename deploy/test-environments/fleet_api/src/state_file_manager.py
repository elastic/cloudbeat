"""
Define a class to manage the policies state using a file.
Exports state_manager object as a singleton.
"""
import json
from pathlib import Path
from utils import delete_file
from loguru import logger

__state_file = Path(__file__).parent / "state_data.json"


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

    def __init__(self, agnt_policy_id: str, pkg_policy_id: str, expected_agents: int, expected_tags: list[str]):
        """
        Args:
            agnt_policy_id (str): ID of the agent policy.
            pkg_policy_id (str): ID of the package policy.
            expected_agents (int): Expected number of deployed agents.
        """
        self.agnt_policy_id = agnt_policy_id
        self.pkg_policy_id = pkg_policy_id
        self.expected_agents = expected_agents
        self.expected_tags = expected_tags


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

    def delete_all(self) -> None:
        """
        Delete the current state.
        """
        self.__policies = []
        delete_file(self.__state_file)


state_manager = StateFileManager(__state_file)
