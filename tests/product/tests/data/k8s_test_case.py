"""
This module provides k8s test case definition
"""

from dataclasses import dataclass, astuple


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
