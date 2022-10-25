"""
This module provides EKS test case definition
"""

from dataclasses import dataclass, astuple


@dataclass
class EksTestCase:
    """
    Represents common EKS test case
    """
    rule_tag: str
    node_hostname: str
    expected: str

    def __iter__(self):
        return iter(astuple(self))

    def __len__(self):
        return len(astuple(self))
