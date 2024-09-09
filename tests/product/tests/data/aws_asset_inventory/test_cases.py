"""
This module provides AWS Elastic Compute Cloud EC2 service rule test cases for Asset
Inventory.
"""

from ..asset_inventory_test_case import AssetInventoryCase

test_cases = {
    "[Asset Inventory][AWS][EC2] assets found": AssetInventoryCase(
        category="infrastructure",
        sub_category="compute",
        type_="virtual-machine",
        sub_type="ec2-instance",
    ),
    "[Asset Inventory][AWS][IAM Role] assets found": AssetInventoryCase(
        category="identity",
        sub_category="digital-identity",
        type_="role",
        sub_type="iam-role",
    ),
}
