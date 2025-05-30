"""
This module provides AWS Elastic Compute Cloud EC2 service rule test cases for Asset
Inventory.
"""

from ..asset_inventory_test_case import AssetInventoryCase

test_cases = {
    "[Asset Inventory][AWS][EC2] assets found": AssetInventoryCase(
        type_="Host",
        sub_type="AWS EC2 Instance",
    ),
    "[Asset Inventory][AWS][IAM Role] assets found": AssetInventoryCase(
        type_="Service Account",
        sub_type="AWS IAM Role",
    ),
    "[Asset Inventory][AWS][ELBv1] assets found": AssetInventoryCase(
        type_="Load Balancer",
        sub_type="AWS Elastic Load Balancer",
    ),
    "[Asset Inventory][AWS][ELBv2] assets found": AssetInventoryCase(
        type_="Load Balancer",
        sub_type="AWS Elastic Load Balancer v2",
    ),
    "[Asset Inventory][AWS][IAM Policy] assets found": AssetInventoryCase(
        type_="Access Management",
        sub_type="AWS IAM Policy",
    ),
    "[Asset Inventory][AWS][IAM User] assets found": AssetInventoryCase(
        type_="Identity",
        sub_type="AWS IAM User",
    ),
    "[Asset Inventory][AWS][Lambda Event Source Mapping] assets found": AssetInventoryCase(
        type_="FaaS",
        sub_type="AWS Lambda Event Source Mapping",
    ),
    "[Asset Inventory][AWS][Lambda Function] assets found": AssetInventoryCase(
        type_="FaaS",
        sub_type="AWS Lambda Function",
    ),
    "[Asset Inventory][AWS][Lambda Layer] assets found": AssetInventoryCase(
        type_="FaaS",
        sub_type="AWS Lambda Layer",
    ),
    "[Asset Inventory][AWS][Internet Gateway] assets found": AssetInventoryCase(
        type_="Gateway",
        sub_type="AWS Internet Gateway",
    ),
    "[Asset Inventory][AWS][NAT Gateway] assets found": AssetInventoryCase(
        type_="Gateway",
        sub_type="AWS NAT Gateway",
    ),
    "[Asset Inventory][AWS][EC2 Network ACL] assets found": AssetInventoryCase(
        type_="Networking",
        sub_type="AWS EC2 Network ACL",
    ),
    "[Asset Inventory][AWS][EC2 Network Interface] assets found": AssetInventoryCase(
        type_="Networking",
        sub_type="AWS EC2 Network Interface",
    ),
    "[Asset Inventory][AWS][Security Group] assets found": AssetInventoryCase(
        type_="Firewall",
        sub_type="AWS EC2 Security Group",
    ),
    "[Asset Inventory][AWS][EC2 Subnet] assets found": AssetInventoryCase(
        type_="Networking",
        sub_type="AWS EC2 Subnet",
    ),
    "[Asset Inventory][AWS][Transit Gateway] assets found": AssetInventoryCase(
        type_="Gateway",
        sub_type="AWS Transit Gateway",
    ),
    "[Asset Inventory][AWS][Transit Gateway Attachment] assets found": AssetInventoryCase(
        type_="Gateway",
        sub_type="AWS Transit Gateway Attachment",
    ),
    "[Asset Inventory][AWS][VPC Peering Connection] assets found": AssetInventoryCase(
        type_="Networking",
        sub_type="AWS VPC Peering Connection",
    ),
    "[Asset Inventory][AWS][VPC] assets found": AssetInventoryCase(
        type_="Networking",
        sub_type="AWS VPC",
    ),
    "[Asset Inventory][AWS][RDS] assets found": AssetInventoryCase(
        type_="Database",
        sub_type="AWS RDS Instance",
    ),
    "[Asset Inventory][AWS][S3] assets found": AssetInventoryCase(
        type_="Storage Bucket",
        sub_type="AWS S3 Bucket",
    ),
    "[Asset Inventory][AWS][SNS Topic] assets found": AssetInventoryCase(
        type_="Messaging Service",
        sub_type="AWS SNS Topic",
    ),
}
