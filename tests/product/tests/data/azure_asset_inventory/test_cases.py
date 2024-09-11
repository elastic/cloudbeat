"""
This module provides Azure test cases for Asset Inventory.
"""

from ..asset_inventory_test_case import AssetInventoryCase

test_cases = {
    "[Asset Inventory][Azure][Azure App Service] assets found": AssetInventoryCase(
        category="infrastructure",
        sub_category="application",
        type_="web-application",
        sub_type="azure-app-service",
    ),
    "[Asset Inventory][Azure][Azure Virtual Machine] assets found": AssetInventoryCase(
        category="infrastructure",
        sub_category="compute",
        type_="virtual-machine",
        sub_type="azure-virtual-machine",
    ),
    "[Asset Inventory][Azure][Azure Container Registry] assets found": AssetInventoryCase(
        category="infrastructure",
        sub_category="container",
        type_="registry",
        sub_type="azure-container-registry",
    ),
    "[Asset Inventory][Azure][Azure SQL Database] assets found": AssetInventoryCase(
        category="infrastructure",
        sub_category="database",
        type_="relational",
        sub_type="azure-sql-database",
    ),
    "[Asset Inventory][Azure][Azure SQL Server] assets found": AssetInventoryCase(
        category="infrastructure",
        sub_category="database",
        type_="relational",
        sub_type="azure-sql-server",
    ),
    "[Asset Inventory][Azure][Azure Elastic Pool] assets found": AssetInventoryCase(
        category="infrastructure",
        sub_category="database",
        type_="scalability",
        sub_type="azure-elastic-pool",
    ),
    "[Asset Inventory][Azure][Azure Subscription] assets found": AssetInventoryCase(
        category="infrastructure",
        sub_category="management",
        type_="cloud-account",
        sub_type="azure-subscription",
    ),
    "[Asset Inventory][Azure][Azure Tenant] assets found": AssetInventoryCase(
        category="infrastructure",
        sub_category="management",
        type_="cloud-account",
        sub_type="azure-tenant",
    ),
    "[Asset Inventory][Azure][Azure Resource Group] assets found": AssetInventoryCase(
        category="infrastructure",
        sub_category="management",
        type_="resource-group",
        sub_type="azure-resource-group",
    ),
    "[Asset Inventory][Azure][Azure Disk] assets found": AssetInventoryCase(
        category="infrastructure",
        sub_category="storage",
        type_="disk",
        sub_type="azure-disk",
    ),
    "[Asset Inventory][Azure][Azure Snapshot] assets found": AssetInventoryCase(
        category="infrastructure",
        sub_category="storage",
        type_="snapshot",
        sub_type="azure-snapshot",
    ),
    "[Asset Inventory][Azure][Azure Storage Account] assets found": AssetInventoryCase(
        category="infrastructure",
        sub_category="storage",
        type_="storage",
        sub_type="azure-storage-account",
    ),
}
