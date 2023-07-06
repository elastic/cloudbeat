"""
Define a class to manage the policies state using a file.
Exports state_manager object as a singleton.
"""
import json
from pathlib import Path
from munch import munchify
from utils import delete_file
from loguru import logger

__state_file = Path(__file__).parent / "state_data.json"


class PolicyStateEncoder(json.JSONEncoder):
    """
    Custom JSON encoder for PolicyState objects.
    """
    def default(self, o):
        return o.__dict__


class PolicyState:
    """
    Class to represent a policy state.
    """
    def __init__(self, agnt_policy_id, pkg_policy_id, expected_agents):
        self.agnt_policy_id = agnt_policy_id
        self.pkg_policy_id = pkg_policy_id
        self.expected_agents = expected_agents


class StateFileManager:
    """
    Class to manage the policies state using a file.
    """
    def __init__(self, state_file: str):
        self.state_file = state_file
        self.policies = []
        self.__load()


    def __load(self) -> None:
        """
        Load the policies data from a file.
        """
        if not self.state_file.exists():
            return
        with self.state_file.open("r") as policies_file:
            policies_data = json.load(policies_file)
            for policy in policies_data["policies"]:
                self.policies.append(PolicyState(**policy))
        logger.info(f" {len(self.policies)} policies loaded to state from {self.state_file}")


    def __save(self) -> None:
        """
        Save the policies data to a file.
        """
        policies_data = munchify({"policies": self.policies})
        with self.state_file.open("w") as policies_file:
            json.dump(policies_data, policies_file, cls=PolicyStateEncoder)
        logger.info(f" {len(self.policies)} policies saved to state in {self.state_file}")


    def add_policy(self, data: PolicyState):
        """
        Add a policy to the current state.

        Args:
            data (PolicyState): Policy data to be added.
        """
        self.policies.append(data)
        self.__save()


    def delete_all(self):
        """
        Delete the current state.
        """
        self.policies = []
        delete_file(self.state_file)

state_manager = StateFileManager(__state_file)
