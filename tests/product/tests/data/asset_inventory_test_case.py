"""
This module provides Asset Inventory test case definition
"""

from dataclasses import astuple, dataclass


@dataclass
class AssetInventoryCase:
    """
    Represents Asset Inventory test case
    """

    category: str
    sub_category: str
    type_: str
    sub_type: str

    def __iter__(self):
        return iter(astuple(self))

    def __len__(self):
        return len(astuple(self))
