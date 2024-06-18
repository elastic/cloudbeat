"""
This module provides k8s test case definition
"""

from dataclasses import astuple, dataclass


@dataclass
class FileTestCase:
    """
    Represents k8s nodes test case
    """

    rule_tag: str
    node_hostname: str
    resource_name: str
    expected: str

    def __iter__(self):
        return iter(astuple(self))

    def __len__(self):
        return len(astuple(self))


@dataclass
class K8sTestCase:
    """
    Represents k8s object and process test case
    """

    rule_tag: str
    resource_name: str
    expected: str

    def __iter__(self):
        return iter(astuple(self))

    def __len__(self):
        return len(astuple(self))
