"""
This module provides GCP test cases for Asset Inventory.
"""

from ..asset_inventory_test_case import AssetInventoryCase

test_cases = {
    "[Asset Inventory][GCP][Service Account] assets found": AssetInventoryCase(
        category="Access Management",
        type_="GCP Service Account",
    ),
    "[Asset Inventory][GCP][Service Account Key] assets found": AssetInventoryCase(
        category="Access Management",
        type_="GCP Service Account Key",
    ),
    "[Asset Inventory][GCP][Instance] assets found": AssetInventoryCase(
        category="Host",
        type_="GCP Compute Instance",
    ),
    "[Asset Inventory][GCP][Subnet] assets found": AssetInventoryCase(
        category="Subnet",
        type_="GCP Subnet",
    ),
    "[Asset Inventory][GCP][Project] assets found": AssetInventoryCase(
        category="Account",
        type_="GCP Project",
    ),
    # "[Asset Inventory][GCP][Organization] assets found": AssetInventoryCase(
    #     category="Infrastructure",
    #     type_="cloud-account",
    # ),
    # "[Asset Inventory][GCP][Folder] assets found": AssetInventoryCase(
    #     category="Infrastructure",
    #     type_="resource-hierarchy",
    # ),
    "[Asset Inventory][GCP][Bucket] assets found": AssetInventoryCase(
        category="Storage Bucket",
        type_="GCP Bucket",
    ),
    "[Asset Inventory][GCP][Firewall] assets found": AssetInventoryCase(
        category="Firewall",
        type_="GCP Firewall",
    ),
    # "[Asset Inventory][GCP][GKE Cluster] assets found": AssetInventoryCase(
    #     category="Infrastructure",
    #     type_="orchestration",
    # ),
    # "[Asset Inventory][GCP][Forwarding Rule] assets found": AssetInventoryCase(
    #     category="Infrastructure",
    #     type_="load-balancing",
    # ),
    "[Asset Inventory][GCP][IAM Role] assets found": AssetInventoryCase(
        category="Service Usage Technology",
        type_="GCP IAM Role",
    ),
    # "[Asset Inventory][GCP][Cloud Function] assets found": AssetInventoryCase(
    #     category="Infrastructure",
    #     type_="function",
    # ),
    "[Asset Inventory][GCP][Cloud Run Service] assets found": AssetInventoryCase(
        category="Container Service",
        type_="GCP Cloud Run Service",
    ),
}
