"""
This module provides GCP test cases for Asset Inventory.
"""

from ..asset_inventory_test_case import AssetInventoryCase

test_cases = {
    "[Asset Inventory][GCP][Service Account] assets found": AssetInventoryCase(
        category="identity",
        # sub_category="service-identity",
        type_="service-account",
        # sub_type="gcp-service-account",
    ),
    "[Asset Inventory][GCP][Service Account Key] assets found": AssetInventoryCase(
        category="identity",
        # sub_category="service-identity",
        type_="service-account-key",
        # sub_type="gcp-service-account-key",
    ),
    "[Asset Inventory][GCP][Instance] assets found": AssetInventoryCase(
        category="infrastructure",
        # sub_category="compute",
        type_="virtual-machine",
        # sub_type="gcp-instance",
    ),
    "[Asset Inventory][GCP][Subnet] assets found": AssetInventoryCase(
        category="infrastructure",
        # sub_category="network",
        type_="subnet",
        # sub_type="gcp-subnet",
    ),
    "[Asset Inventory][GCP][Project] assets found": AssetInventoryCase(
        category="infrastructure",
        # sub_category="management",
        type_="cloud-account",
        # sub_type="gcp-project",
    ),
    # "[Asset Inventory][GCP][Organization] assets found": AssetInventoryCase(
    #     category="infrastructure",
    #     # sub_category="management",
    #     type_="cloud-account",
    #     # sub_type="gcp-organization",
    # ),
    # "[Asset Inventory][GCP][Folder] assets found": AssetInventoryCase(
    #     category="infrastructure",
    #     # sub_category="management",
    #     type_="resource-hierarchy",
    #     # sub_type="gcp-folder",
    # ),
    "[Asset Inventory][GCP][Bucket] assets found": AssetInventoryCase(
        category="infrastructure",
        # sub_category="storage",
        type_="object-storage",
        # sub_type="gcp-bucket",
    ),
    "[Asset Inventory][GCP][Firewall] assets found": AssetInventoryCase(
        category="infrastructure",
        # sub_category="network",
        type_="firewall",
        # sub_type="gcp-firewall",
    ),
    # "[Asset Inventory][GCP][GKE Cluster] assets found": AssetInventoryCase(
    #     category="infrastructure",
    #     # sub_category="container",
    #     type_="orchestration",
    #     # sub_type="gcp-gke-cluster",
    # ),
    # "[Asset Inventory][GCP][Forwarding Rule] assets found": AssetInventoryCase(
    #     category="infrastructure",
    #     # sub_category="network",
    #     type_="load-balancing",
    #     # sub_type="gcp-forwarding-rule",
    # ),
    "[Asset Inventory][GCP][IAM Role] assets found": AssetInventoryCase(
        category="identity",
        # sub_category="access-management",
        type_="iam-role",
        # sub_type="gcp-iam-role",
    ),
    # "[Asset Inventory][GCP][Cloud Function] assets found": AssetInventoryCase(
    #     category="infrastructure",
    #     # sub_category="serverless",
    #     type_="function",
    #     # sub_type="gcp-cloud-function",
    # ),
    "[Asset Inventory][GCP][Cloud Run Service] assets found": AssetInventoryCase(
        category="infrastructure",
        # sub_category="container",
        type_="serverless",
        # sub_type="gcp-cloud-run-service",
    ),
}
