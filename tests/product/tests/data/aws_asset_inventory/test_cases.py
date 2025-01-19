"""
This module provides AWS Elastic Compute Cloud EC2 service rule test cases for Asset
Inventory.
"""

from ..asset_inventory_test_case import AssetInventoryCase

test_cases = {
    "[Asset Inventory][AWS][EC2] assets found": AssetInventoryCase(
        category="Host",
        type_="AWS EC2 Instance",
    ),
    "[Asset Inventory][AWS][IAM Role] assets found": AssetInventoryCase(
        category="Service Account",
        type_="AWS IAM Role",
    ),
    "[Asset Inventory][AWS][ELBv1] assets found": AssetInventoryCase(
        category="Load Balancer",
        type_="AWS Elastic Load Balancer",
    ),
    "[Asset Inventory][AWS][ELBv2] assets found": AssetInventoryCase(
        category="Load Balancer",
        type_="AWS Elastic Load Balancer v2",
    ),
    "[Asset Inventory][AWS][IAM Policy] assets found": AssetInventoryCase(
        category="Access Management",
        type_="AWS IAM Policy",
    ),
    "[Asset Inventory][AWS][IAM User] assets found": AssetInventoryCase(
        category="Identity",
        type_="AWS IAM User",
    ),
    "[Asset Inventory][AWS][Lambda Event Source Mapping] assets found": AssetInventoryCase(
        category="FaaS",
        type_="AWS Lambda Event Source Mapping",
    ),
    "[Asset Inventory][AWS][Lambda Function] assets found": AssetInventoryCase(
        category="FaaS",
        type_="AWS Lambda Function",
    ),
    "[Asset Inventory][AWS][Lambda Layer] assets found": AssetInventoryCase(
        category="FaaS",
        type_="AWS Lambda Layer",
    ),
    "[Asset Inventory][AWS][Internet Gateway] assets found": AssetInventoryCase(
        category="Gateway",
        type_="AWS Internet Gateway",
    ),
    "[Asset Inventory][AWS][NAT Gateway] assets found": AssetInventoryCase(
        category="Gateway",
        type_="AWS NAT Gateway",
    ),
    "[Asset Inventory][AWS][EC2 Network ACL] assets found": AssetInventoryCase(
        category="Networking",
        type_="AWS EC2 Network ACL",
    ),
    "[Asset Inventory][AWS][EC2 Network Interface] assets found": AssetInventoryCase(
        category="Networking",
        type_="AWS EC2 Network Interface",
    ),
    "[Asset Inventory][AWS][Security Group] assets found": AssetInventoryCase(
        category="Firewall",
        type_="AWS EC2 Security Group",
    ),
    "[Asset Inventory][AWS][EC2 Subnet] assets found": AssetInventoryCase(
        category="Networking",
        type_="AWS EC2 Subnet",
    ),
    "[Asset Inventory][AWS][Transit Gateway] assets found": AssetInventoryCase(
        category="Gateway",
        type_="AWS Transit Gateway",
    ),
    "[Asset Inventory][AWS][Transit Gateway Attachment] assets found": AssetInventoryCase(
        category="Gateway",
        type_="AWS Transit Gateway Attachment",
    ),
    "[Asset Inventory][AWS][VPC Peering Connection] assets found": AssetInventoryCase(
        category="Networking",
        type_="AWS VPC Peering Connection",
    ),
    "[Asset Inventory][AWS][VPC] assets found": AssetInventoryCase(
        category="Networking",
        type_="AWS VPC",
    ),
    "[Asset Inventory][AWS][RDS] assets found": AssetInventoryCase(
        category="Database",
        type_="AWS RDS Instance",
    ),
    "[Asset Inventory][AWS][S3] assets found": AssetInventoryCase(
        category="Storage Bucket",
        type_="AWS S3 Bucket",
    ),
    "[Asset Inventory][AWS][SNS Topic] assets found": AssetInventoryCase(
        category="Messaging Service",
        type_="AWS SNS Topic",
    ),
}
