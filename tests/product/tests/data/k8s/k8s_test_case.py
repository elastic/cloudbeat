"""
This module provides k8s test case definition
"""

from dataclasses import dataclass


@dataclass
class FileTestCase:
    """
    Represents k8s nodes test case
    """

    rule_tag: str
    node_hostname: str
    resource_name: str
    expected: str


@dataclass
class K8sTestCase:
    """
    Represents k8s object and process test case
    """

    rule_tag: str
    resource_name: str
    expected: str
