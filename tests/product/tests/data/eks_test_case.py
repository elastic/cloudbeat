"""
This module provides EKS test case definition
"""

from dataclasses import astuple, dataclass


@dataclass
class EksTestCase:
    """
    Represents EKS nodes test case
    """

    rule_tag: str
    node_hostname: str
    expected: str

    def __iter__(self):
        return iter(astuple(self))

    def __len__(self):
        return len(astuple(self))


@dataclass
class EksKubeObjectCase:
    """
    Represents Kube Object test case
    """

    rule_tag: str
    test_resource_id: str
    expected: str

    def __iter__(self):
        return iter(astuple(self))

    def __len__(self):
        return len(astuple(self))


@dataclass
class EksAwsServiceCase:
    """
    Represents EKS AWS service test case
    """

    rule_tag: str
    case_identifier: str
    expected: str

    def __iter__(self):
        return iter(astuple(self))

    def __len__(self):
        return len(astuple(self))
