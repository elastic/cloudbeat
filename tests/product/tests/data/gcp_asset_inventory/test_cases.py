"""
This module provides GCP test cases for Asset Inventory.
"""

from ..asset_inventory_test_case import AssetInventoryCase

test_cases = {
    "[Asset Inventory][GCP][Service Account] assets found": AssetInventoryCase(
        type_="Access Management",
        sub_type="GCP Service Account",
    ),
    "[Asset Inventory][GCP][Service Account Key] assets found": AssetInventoryCase(
        type_="Access Management",
        sub_type="GCP Service Account Key",
    ),
    "[Asset Inventory][GCP][Instance] assets found": AssetInventoryCase(
        type_="Host",
        sub_type="GCP Compute Instance",
    ),
    "[Asset Inventory][GCP][Subnet] assets found": AssetInventoryCase(
        type_="Subnet",
        sub_type="GCP Subnet",
    ),
    "[Asset Inventory][GCP][Project] assets found": AssetInventoryCase(
        type_="Account",
        sub_type="GCP Project",
    ),
    # "[Asset Inventory][GCP][Organization] assets found": AssetInventoryCase(
    #     type_="Infrastructure",
    #     sub_type="cloud-account",
    # ),
    # "[Asset Inventory][GCP][Folder] assets found": AssetInventoryCase(
    #     type_="Infrastructure",
    #     sub_type="resource-hierarchy",
    # ),
    "[Asset Inventory][GCP][Bucket] assets found": AssetInventoryCase(
        type_="Storage Bucket",
        sub_type="GCP Bucket",
    ),
    "[Asset Inventory][GCP][Firewall] assets found": AssetInventoryCase(
        type_="Firewall",
        sub_type="GCP Firewall",
    ),
    # "[Asset Inventory][GCP][GKE Cluster] assets found": AssetInventoryCase(
    #     type_="Infrastructure",
    #     sub_type="orchestration",
    # ),
    # "[Asset Inventory][GCP][Forwarding Rule] assets found": AssetInventoryCase(
    #     type_="Infrastructure",
    #     sub_type="load-balancing",
    # ),
    "[Asset Inventory][GCP][IAM Role] assets found": AssetInventoryCase(
        type_="Service Usage Technology",
        sub_type="GCP IAM Role",
    ),
    # "[Asset Inventory][GCP][Cloud Function] assets found": AssetInventoryCase(
    #     type_="Infrastructure",
    #     sub_type="function",
    # ),
    "[Asset Inventory][GCP][Cloud Run Service] assets found": AssetInventoryCase(
        type_="Container Service",
        sub_type="GCP Cloud Run Service",
    ),
}
