"""
This module provides Azure test cases for Asset Inventory.
"""

from ..asset_inventory_test_case import AssetInventoryCase

test_cases = {
    "[Asset Inventory][Azure][Azure App Service] assets found": AssetInventoryCase(
        category="Web Service",
        type_="Azure App Service",
    ),
    "[Asset Inventory][Azure][Azure Virtual Machine] assets found": AssetInventoryCase(
        category="Host",
        type_="Azure Virtual Machine",
    ),
    "[Asset Inventory][Azure][Azure Container Registry] assets found": AssetInventoryCase(
        category="Container Registry",
        type_="Azure Container Registry",
    ),
    "[Asset Inventory][Azure][Azure Cosmos DB Account] assets found": AssetInventoryCase(
        category="Infrastructure",
        type_="Azure Cosmos DB Account",
    ),
    # "[Asset Inventory][Azure][Azure Cosmos DB SQL Database] assets found": AssetInventoryCase(
    #     category="Infrastructure",
    #     type_="Azure Cosmos DB SQL Database",
    # ),
    "[Asset Inventory][Azure][Azure SQL Database] assets found": AssetInventoryCase(
        category="Database",
        type_="Azure SQL Database",
    ),
    "[Asset Inventory][Azure][Azure SQL Server] assets found": AssetInventoryCase(
        category="Database",
        type_="Azure SQL Server",
    ),
    "[Asset Inventory][Azure][Azure Principal] assets found": AssetInventoryCase(
        category="Identity",
        type_="Azure Principal",
    ),
    "[Asset Inventory][Azure][Azure Elastic Pool] assets found": AssetInventoryCase(
        category="Database",
        type_="Azure Elastic Pool",
    ),
    "[Asset Inventory][Azure][Azure Subscription] assets found": AssetInventoryCase(
        category="Access Management",
        type_="Azure Subscription",
    ),
    "[Asset Inventory][Azure][Azure Tenant] assets found": AssetInventoryCase(
        category="Access Management",
        type_="Azure Tenant",
    ),
    "[Asset Inventory][Azure][Azure Resource Group] assets found": AssetInventoryCase(
        category="Access Management",
        type_="Azure Resource Group",
    ),
    "[Asset Inventory][Azure][Azure Disk] assets found": AssetInventoryCase(
        category="Volume",
        type_="Azure Disk",
    ),
    "[Asset Inventory][Azure][Azure Snapshot] assets found": AssetInventoryCase(
        category="Snapshot",
        type_="Azure Snapshot",
    ),
    "[Asset Inventory][Azure][Azure Storage Account] assets found": AssetInventoryCase(
        category="Private Endpoint",
        type_="Azure Storage Account",
    ),
    "[Asset Inventory][Azure][Azure Storage Queue] assets found": AssetInventoryCase(
        category="Messaging Service",
        type_="Azure Storage Queue",
    ),
    "[Asset Inventory][Azure][Azure Storage Queue Service] assets found": AssetInventoryCase(
        category="Messaging Service",
        type_="Azure Storage Queue Service",
    ),
    "[Asset Inventory][Azure][Azure Storage Blob Service] assets found": AssetInventoryCase(
        category="Storage Bucket",
        type_="Azure Storage Blob Service",
    ),
}
