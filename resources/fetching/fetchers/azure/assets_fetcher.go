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
	Id             string                 `json:"id,omitempty"`
	Name           string                 `json:"name,omitempty"`
	Location       string                 `json:"location,omitempty"`
	Properties     map[string]interface{} `json:"properties,omitempty"`
	ResourceGroup  string                 `json:"resource_group,omitempty"`
	SubscriptionId string                 `json:"subscription_id,omitempty"`
	TenantId       string                 `json:"tenant_id,omitempty"`
}

// TODO: Implement other types
var AzureAssetTypes = map[string]string{
	"microsoft.compute/virtualmachines": fetching.AzureVMType,
	"microsoft.storage/storageaccounts": fetching.AzureStorageAccountType,
}

func NewAzureAssetsFetcher(_ context.Context, log *logp.Logger, ch chan fetching.ResourceInfo, provider inventory.ServiceAPI) *AzureAssetsFetcher {
	return &AzureAssetsFetcher{
		log:        log,
		resourceCh: ch,
		provider:   provider,
	}
}

func getAssetTypes() []string {
	types := make([]string, len(AzureAssetTypes))
	i := 0

	for k := range AzureAssetTypes {
		types[i] = k
		i++
	}

	return types
}

func (f *AzureAssetsFetcher) Fetch(ctx context.Context, cMetadata fetching.CycleMetadata) error {
	f.log.Info("Starting AzureAssetsFetcher.Fetch")
	// TODO: Maybe we should use a query per type instead of listing all assets in a single query
	assets, err := f.provider.ListAllAssetTypesByName(getAssetTypes())
	if err != nil {
		f.log.Errorf("Failed to list assets: %s", err.Error())
		return err
	}

	for _, asset := range assets {
		select {
		case <-ctx.Done():
			f.log.Infof("AzureAssetsFetcher.Fetch context err: %s", ctx.Err().Error())
			return nil
		case f.resourceCh <- fetching.ResourceInfo{
			CycleMetadata: cMetadata,
			// TODO: Safe guard this convertion
			Resource: getAssetFromData(asset.(map[string]interface{})),
		}:
		}
	}

	return nil
}

// TODO: Safe guard this function
func getAssetFromData(data map[string]interface{}) *AzureAsset {
	assetType := data["type"].(string)

	asset := &AzureAsset{
		Type:    AzureAssetTypes[assetType],
		SubType: getAzureSubType(assetType),
		Asset: AzureAssetInfo{
			Id:             data["id"].(string),
			Name:           data["name"].(string),
			Location:       data["location"].(string),
			Properties:     data["properties"].(map[string]interface{}),
			ResourceGroup:  data["resourceGroup"].(string),
			SubscriptionId: data["subscriptionId"].(string),
			TenantId:       data["tenantId"].(string),
		},
	}

	return asset
}

// TODO: Handle this function
func (f *AzureAssetsFetcher) Stop() {
	// f.provider.Close()
}

func (r *AzureAsset) GetData() interface{} {
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

func (r *AzureAsset) GetElasticCommonData() any { return nil }

// TODO: Implement this function
func getAzureSubType(assetType string) string {
	return ""
}
