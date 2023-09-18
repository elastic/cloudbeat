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

package fetchers

import (
	"context"

	"github.com/elastic/elastic-agent-libs/logp"
	"golang.org/x/exp/maps"

	"github.com/elastic/cloudbeat/resources/fetching"
	"github.com/elastic/cloudbeat/resources/providers/azurelib/inventory"
)

type AzureAssetsFetcher struct {
	log        *logp.Logger
	resourceCh chan fetching.ResourceInfo
	provider   inventory.ServiceAPI
}

type AzureAsset struct {
	Type    string
	SubType string
	Asset   AzureAssetInfo `json:"asset,omitempty"`
}

// TODO: Fill this struct with the required fields
type AzureAssetInfo struct {
	Id             string         `json:"id,omitempty"`
	Name           string         `json:"name,omitempty"`
	Location       string         `json:"location,omitempty"`
	Properties     map[string]any `json:"properties,omitempty"`
	ResourceGroup  string         `json:"resource_group,omitempty"`
	SubscriptionId string         `json:"subscription_id,omitempty"`
	TenantId       string         `json:"tenant_id,omitempty"`
}

// TODO: Implement other types
var AzureAssetTypes = map[string]string{
	"microsoft.compute/virtualmachines": fetching.AzureVMType,
	"microsoft.storage/storageaccounts": fetching.AzureStorageAccountType,
}

func NewAzureAssetsFetcher(log *logp.Logger, ch chan fetching.ResourceInfo, provider inventory.ServiceAPI) *AzureAssetsFetcher {
	return &AzureAssetsFetcher{
		log:        log,
		resourceCh: ch,
		provider:   provider,
	}
}

func (f *AzureAssetsFetcher) Fetch(ctx context.Context, cMetadata fetching.CycleMetadata) error {
	f.log.Info("Starting AzureAssetsFetcher.Fetch")
	// TODO: Maybe we should use a query per type instead of listing all assets in a single query
	assets, err := f.provider.ListAllAssetTypesByName(maps.Keys(AzureAssetTypes))
	if err != nil {
		return err
	}

	for _, asset := range assets {
		select {
		case <-ctx.Done():
			f.log.Infof("AzureAssetsFetcher.Fetch context err: %s", ctx.Err().Error())
			return nil
		case f.resourceCh <- fetching.ResourceInfo{
			CycleMetadata: cMetadata,
			// TODO: Safe guard this conversion
			Resource: getAssetFromData(asset.(map[string]any)),
		}:
		}
	}

	return nil
}

func getAssetFromData(data map[string]any) *AzureAsset {
	assetType := getString(data, "type")
	properties, _ := data["properties"].(map[string]any)

	return &AzureAsset{
		Type:    AzureAssetTypes[assetType],
		SubType: getAzureSubType(assetType),
		Asset: AzureAssetInfo{
			Id:             getString(data, "id"),
			Name:           getString(data, "name"),
			Location:       getString(data, "location"),
			Properties:     properties,
			ResourceGroup:  getString(data, "resourceGroup"),
			SubscriptionId: getString(data, "subscriptionId"),
			TenantId:       getString(data, "tenantId"),
		},
	}
}

func getString(data map[string]any, key string) string {
	value, _ := data[key].(string)
	return value
}

// TODO: Handle this function
func (f *AzureAssetsFetcher) Stop() {
	// f.provider.Close()
}

func (r *AzureAsset) GetData() any {
	return r.Asset
}

func (r *AzureAsset) GetMetadata() (fetching.ResourceMetadata, error) {
	return fetching.ResourceMetadata{
		ID:      r.Asset.Id,
		Type:    r.Type,
		SubType: r.SubType,
		Name:    r.Asset.Name,
		Region:  r.Asset.Location,
	}, nil
}

func (r *AzureAsset) GetElasticCommonData() (map[string]any, error) { return nil, nil }

// TODO: Implement this function
func getAzureSubType(assetType string) string {
	return ""
}
