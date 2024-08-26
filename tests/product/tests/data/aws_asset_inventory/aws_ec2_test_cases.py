"""
TODO(kuba)
"""

from ..constants import RULE_FAIL_STATUS, RULE_PASS_STATUS
from ..asset_inventory_test_case import AssetInventoryCase

aws_ec2_asset_inventory_test_case_1 = AssetInventoryCase()

test_cases = {
    "[Asset Inventory][AWS][EC2] assets found": aws_ec2_asset_inventory_test_case_1,
}
