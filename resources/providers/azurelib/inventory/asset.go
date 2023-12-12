// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package inventory

import (
	"github.com/elastic/cloudbeat/resources/utils/strings"
)

const (
	// Resources group
	ActivityLogAlertAssetType          = "microsoft.insights/activitylogalerts"
	ApplicationInsights                = "microsoft.insights/components"
	BastionAssetType                   = "microsoft.network/bastionhosts"
	ClassicStorageAccountAssetType     = "microsoft.classicstorage/storageaccounts"
	ClassicVirtualMachineAssetType     = "microsoft.classiccompute/virtualmachines"
	DiagnosticSettingsAssetType        = "microsoft.insights/diagnosticSettings"
	DiskAssetType                      = "microsoft.compute/disks"
	DocumentDBDatabaseAccountAssetType = "microsoft.documentdb/databaseaccounts"
	MySQLDBAssetType                   = "microsoft.dbformysql/servers"
	NetworkWatchersAssetType           = "microsoft.network/networkwatchers"
	NetworkWatchersFlowLogAssetType    = "microsoft.network/networkwatchers/flowlogs"
	PostgreSQLDBAssetType              = "microsoft.dbforpostgresql/servers"
	SQLServersAssetType                = "microsoft.sql/servers"
	StorageAccountAssetType            = "microsoft.storage/storageaccounts"
	VaultAssetType                     = "microsoft.keyvault/vaults"
	VirtualMachineAssetType            = "microsoft.compute/virtualmachines"
	WebsitesAssetType                  = "microsoft.web/sites"

	// Authorizationresources group
	RoleDefinitionsType = "microsoft.authorization/roledefinitions"

	// Azure Resource Graph table groups
	AssetGroupResources              = "resources"
	AssetGroupAuthorizationResources = "authorizationresources"
)

type AzureAsset struct {
	Id             string         `json:"id,omitempty"`
	Name           string         `json:"name,omitempty"`
	DisplayName    string         `json:"display_name,omitempty"`
	Location       string         `json:"location,omitempty"`
	Properties     map[string]any `json:"properties,omitempty"`
	Extension      map[string]any `json:"extension,omitempty"`
	ResourceGroup  string         `json:"resource_group,omitempty"`
	SubscriptionId string         `json:"subscription_id,omitempty"`
	TenantId       string         `json:"tenant_id,omitempty"`
	Type           string         `json:"type,omitempty"`
	Sku            string         `json:"sku,omitempty"`
}

func getAssetFromData(data map[string]any) AzureAsset {
	subId := strings.FromMap(data, "subscriptionId")
	properties, _ := data["properties"].(map[string]any)

	return AzureAsset{
		Id:             strings.FromMap(data, "id"),
		Name:           strings.FromMap(data, "name"),
		DisplayName:    strings.FromMap(data, "displayName"),
		Location:       strings.FromMap(data, "location"),
		Properties:     properties,
		ResourceGroup:  strings.FromMap(data, "resourceGroup"),
		SubscriptionId: subId,
		TenantId:       strings.FromMap(data, "tenantId"),
		Sku:            strings.FromMap(data, "sku"),
		Type:           strings.FromMap(data, "type"),
	}
}
