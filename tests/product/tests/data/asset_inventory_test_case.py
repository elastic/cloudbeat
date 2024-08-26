"""
This module provides Azure test case definition
"""

from dataclasses import astuple, dataclass


@dataclass
class AssetInventoryCase:
    """
    Represents Asset Inventory test case
    """

    # rule_tag: str
    # case_identifier: str
    # expected: str

    def __iter__(self):
        return iter(astuple(self))

    def __len__(self):
        return len(astuple(self))
