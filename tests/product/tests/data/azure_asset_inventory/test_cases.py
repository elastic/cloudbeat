"""
This module provides Azure test cases for Asset Inventory.
"""

from ..asset_inventory_test_case import AssetInventoryCase

test_cases = {
    # TODO(kuba): add cases for all Azure resoruces
    "[Asset Inventory][Azure][Resource Group] assets found": AssetInventoryCase(
        category="infrastructure",
        sub_category="management",
        type_="resource-group",
        sub_type="azure-resource-group",
    ),
}
