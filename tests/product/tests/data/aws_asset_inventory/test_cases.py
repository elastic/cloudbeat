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
    "[Asset Inventory][AWS][ELBv1] assets found": AssetInventoryCase(
        category="infrastructure",
        sub_category="network",
        type_="load-balancer",
        sub_type="elastic-load-balancer",
    ),
    "[Asset Inventory][AWS][ELBv2] assets found": AssetInventoryCase(
        category="infrastructure",
        sub_category="network",
        type_="load-balancer",
        sub_type="elastic-load-balancer-v2",
    ),
    "[Asset Inventory][AWS][IAM Policy] assets found": AssetInventoryCase(
        category="identity",
        sub_category="digital-identity",
        type_="policy",
        sub_type="iam-policy",
    ),
    "[Asset Inventory][AWS][IAM User] assets found": AssetInventoryCase(
        category="identity",
        sub_category="digital-identity",
        type_="user",
        sub_type="iam-user",
    ),
    "[Asset Inventory][AWS][Lambda Event Source Mapping] assets found": AssetInventoryCase(
        category="infrastructure",
        sub_category="integration",
        type_="event-source",
        sub_type="lambda-event-source-mapping",
    ),
    "[Asset Inventory][AWS][Lambda Function] assets found": AssetInventoryCase(
        category="infrastructure",
        sub_category="compute",
        type_="serverless",
        sub_type="lambda-function",
    ),
    "[Asset Inventory][AWS][Lambda Layer] assets found": AssetInventoryCase(
        category="infrastructure",
        sub_category="compute",
        type_="serverless",
        sub_type="lambda-layer",
    ),
    "[Asset Inventory][AWS][Internet Gateway] assets found": AssetInventoryCase(
        category="infrastructure",
        sub_category="network",
        type_="gateway",
        sub_type="internet-gateway",
    ),
    "[Asset Inventory][AWS][NAT Gateway] assets found": AssetInventoryCase(
        category="infrastructure",
        sub_category="network",
        type_="gateway",
        sub_type="nat-gateway",
    ),
    "[Asset Inventory][AWS][VPC ACL] assets found": AssetInventoryCase(
        category="identity",
        sub_category="authorization",
        type_="acl",
        sub_type="s3-access-control-list",
    ),
    "[Asset Inventory][AWS][EC2 Network Interface] assets found": AssetInventoryCase(
        category="infrastructure",
        sub_category="network",
        type_="interface",
        sub_type="ec2-network-interface",
    ),
    "[Asset Inventory][AWS][Security Group] assets found": AssetInventoryCase(
        category="infrastructure",
        sub_category="network",
        type_="firewall",
        sub_type="ec2-security-group",
    ),
    "[Asset Inventory][AWS][EC2 Subnet] assets found": AssetInventoryCase(
        category="infrastructure",
        sub_category="network",
        type_="subnet",
        sub_type="ec2-subnet",
    ),
    "[Asset Inventory][AWS][Transit Gateway] assets found": AssetInventoryCase(
        category="infrastructure",
        sub_category="network",
        type_="virtual-network",
        sub_type="transit-gateway",
    ),
    "[Asset Inventory][AWS][Transit Gateway Attachment] assets found": AssetInventoryCase(
        category="infrastructure",
        sub_category="network",
        type_="virtual-network",
        sub_type="transit-gateway-attachment",
    ),
    "[Asset Inventory][AWS][VPC Peering Connection] assets found": AssetInventoryCase(
        category="infrastructure",
        sub_category="network",
        type_="peering",
        sub_type="vpc-peering-connection",
    ),
    "[Asset Inventory][AWS][VPC] assets found": AssetInventoryCase(
        category="infrastructure",
        sub_category="network",
        type_="virtual-network",
        sub_type="vpc",
    ),
    "[Asset Inventory][AWS][RDS] assets found": AssetInventoryCase(
        category="infrastructure",
        sub_category="database",
        type_="relational",
        sub_type="rds-instance",
    ),
    "[Asset Inventory][AWS][S3] assets found": AssetInventoryCase(
        category="infrastructure",
        sub_category="storage",
        type_="object-storage",
        sub_type="s3-bucket",
    ),
    "[Asset Inventory][AWS][SNS Topic] assets found": AssetInventoryCase(
        category="infrastructure",
        sub_category="messaging",
        type_="notification-service",
        sub_type="sns-topic",
    ),
}
