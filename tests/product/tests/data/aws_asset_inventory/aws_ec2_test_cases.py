"""
This module provides AWS Elastic Compute Cloud EC2 service rule test cases for Asset
Inventory.
"""

from ..asset_inventory_test_case import AssetInventoryCase

aws_ec2_asset_inventory_test_case_1 = AssetInventoryCase()

test_cases = {
    "[Asset Inventory][AWS][EC2] assets found": aws_ec2_asset_inventory_test_case_1,
}
