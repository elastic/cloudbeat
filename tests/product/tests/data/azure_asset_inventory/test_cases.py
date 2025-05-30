"""
This module provides Azure test cases for Asset Inventory.
"""

from ..asset_inventory_test_case import AssetInventoryCase

test_cases = {
    "[Asset Inventory][Azure][Azure App Service] assets found": AssetInventoryCase(
        type_="Web Service",
        sub_type="Azure App Service",
    ),
    "[Asset Inventory][Azure][Azure Virtual Machine] assets found": AssetInventoryCase(
        type_="Host",
        sub_type="Azure Virtual Machine",
    ),
    "[Asset Inventory][Azure][Azure Container Registry] assets found": AssetInventoryCase(
        type_="Container Registry",
        sub_type="Azure Container Registry",
    ),
    "[Asset Inventory][Azure][Azure Cosmos DB Account] assets found": AssetInventoryCase(
        type_="Infrastructure",
        sub_type="Azure Cosmos DB Account",
    ),
    # "[Asset Inventory][Azure][Azure Cosmos DB SQL Database] assets found": AssetInventoryCase(
    #     type_="Infrastructure",
    #     sub_type="Azure Cosmos DB SQL Database",
    # ),
    "[Asset Inventory][Azure][Azure SQL Database] assets found": AssetInventoryCase(
        type_="Database",
        sub_type="Azure SQL Database",
    ),
    "[Asset Inventory][Azure][Azure SQL Server] assets found": AssetInventoryCase(
        type_="Database",
        sub_type="Azure SQL Server",
    ),
    # "[Asset Inventory][Azure][Azure Principal] assets found": AssetInventoryCase(
    #     type_="Identity",
    #     sub_type="Azure Principal",
    # ),
    "[Asset Inventory][Azure][Azure Elastic Pool] assets found": AssetInventoryCase(
        type_="Database",
        sub_type="Azure Elastic Pool",
    ),
    "[Asset Inventory][Azure][Azure Subscription] assets found": AssetInventoryCase(
        type_="Access Management",
        sub_type="Azure Subscription",
    ),
    "[Asset Inventory][Azure][Azure Tenant] assets found": AssetInventoryCase(
        type_="Access Management",
        sub_type="Azure Tenant",
    ),
    "[Asset Inventory][Azure][Azure Resource Group] assets found": AssetInventoryCase(
        type_="Access Management",
        sub_type="Azure Resource Group",
    ),
    "[Asset Inventory][Azure][Azure Disk] assets found": AssetInventoryCase(
        type_="Volume",
        sub_type="Azure Disk",
    ),
    "[Asset Inventory][Azure][Azure Snapshot] assets found": AssetInventoryCase(
        type_="Snapshot",
        sub_type="Azure Snapshot",
    ),
    "[Asset Inventory][Azure][Azure Storage Account] assets found": AssetInventoryCase(
        type_="Private Endpoint",
        sub_type="Azure Storage Account",
    ),
    "[Asset Inventory][Azure][Azure Storage Queue] assets found": AssetInventoryCase(
        type_="Messaging Service",
        sub_type="Azure Storage Queue",
    ),
    "[Asset Inventory][Azure][Azure Storage Queue Service] assets found": AssetInventoryCase(
        type_="Messaging Service",
        sub_type="Azure Storage Queue Service",
    ),
    "[Asset Inventory][Azure][Azure Storage Blob Container] assets found": AssetInventoryCase(
        type_="Storage Bucket",
        sub_type="Azure Storage Blob Container",
    ),
    "[Asset Inventory][Azure][Azure Storage Blob Service] assets found": AssetInventoryCase(
        type_="Service Usage Technology",
        sub_type="Azure Storage Blob Service",
    ),
    "[Asset Inventory][Azure][Azure Storage File Service] assets found": AssetInventoryCase(
        type_="File System Service",
        sub_type="Azure Storage File Service",
    ),
    "[Asset Inventory][Azure][Azure Storage File Share] assets found": AssetInventoryCase(
        type_="File System Service",
        sub_type="Azure Storage File Share",
    ),
    "[Asset Inventory][Azure][Azure Table] assets found": AssetInventoryCase(
        type_="Database",
        sub_type="Azure Storage Table",
    ),
    "[Asset Inventory][Azure][Azure Table Service] assets found": AssetInventoryCase(
        type_="Service Usage Technology",
        sub_type="Azure Storage Table Service",
    ),
}
