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
	"fmt"

	"github.com/elastic/elastic-agent-libs/logp"
	"github.com/samber/lo"
	"golang.org/x/exp/maps"

	"github.com/elastic/cloudbeat/resources/fetching"
	"github.com/elastic/cloudbeat/resources/providers/azurelib"
	"github.com/elastic/cloudbeat/resources/providers/azurelib/inventory"
)

type AzureBatchAssetFetcher struct {
	log        *logp.Logger
	resourceCh chan fetching.ResourceInfo
	provider   inventory.ServiceAPI
}

var AzureBatchAssets = map[string]typePair{
	inventory.ActivityLogAlertAssetType: newPair(fetching.AzureActivityLogAlertType, fetching.MonitoringIdentity),
	inventory.BastionAssetType:          newPair(fetching.AzureBastionType, fetching.CloudDns),
}

func NewAzureBatchAssetFetcher(log *logp.Logger, ch chan fetching.ResourceInfo, provider inventory.ServiceAPI) *AzureBatchAssetFetcher {
	return &AzureBatchAssetFetcher{
		log:        log,
		resourceCh: ch,
		provider:   provider,
	}
}

func (f *AzureBatchAssetFetcher) Fetch(ctx context.Context, cMetadata fetching.CycleMetadata) error {
	f.log.Info("Starting AzureBatchAssetFetcher.Fetch")

	allAssets, err := f.provider.ListAllAssetTypesByName(ctx, maps.Keys(AzureBatchAssets))
	if err != nil {
		return err
	}

	if len(allAssets) == 0 {
		return nil
	}

	assetGroups := lo.GroupBy(allAssets, func(item inventory.AzureAsset) string {
		return fmt.Sprintf("%s-%s", item.SubscriptionId, item.Type)
	})

	for _, assets := range assetGroups {
		pair := AzureBatchAssets[assets[0].Type]

		select {
		case <-ctx.Done():
			f.log.Infof("AzureBatchAssetFetcher.Fetch context err: %s", ctx.Err().Error())
			return nil
		case f.resourceCh <- fetching.ResourceInfo{
			CycleMetadata: cMetadata,
			Resource: &AzureBatchResource{
				// Every asset in the list has the same type and subtype
				Type:    pair.Type,
				SubType: pair.SubType,
				Assets:  assets,
			},
		}:
		}
	}

	return nil
}

func (f *AzureBatchAssetFetcher) Stop() {}

type AzureBatchResource struct {
	Type    string
	SubType string
	Assets  []inventory.AzureAsset `json:"assets,omitempty"`
}

func (r *AzureBatchResource) GetData() any {
	return r.Assets
}

func (r *AzureBatchResource) GetMetadata() (fetching.ResourceMetadata, error) {
	// Assuming all batch in not empty includes assets of the same subscription
	id := fmt.Sprintf("%s-%s", r.SubType, r.Assets[0].SubscriptionId)
	return fetching.ResourceMetadata{
		ID:      id,
		Type:    r.Type,
		SubType: r.SubType,
		Name:    id,
		// TODO: Make sure ActivityLogAlerts are not location scoped (benchmarks do not check location)
		Region: azurelib.GlobalRegion,
	}, nil
}

func (r *AzureBatchResource) GetElasticCommonData() (map[string]any, error) {
	return map[string]any{
		"cloud": map[string]any{
			"provider": "azure",
			"account": map[string]any{
				"id":   r.Assets[0].SubscriptionId,
				"name": r.Assets[0].SubscriptionName,
			},
			// TODO: Organization fields
		},
	}, nil
}
