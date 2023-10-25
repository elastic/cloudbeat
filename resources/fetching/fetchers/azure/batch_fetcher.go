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

	for assetType, pair := range AzureBatchAssets {
		assets, err := f.provider.ListAllAssetTypesByName([]string{assetType})
		if err != nil {
			return err
		}

		if len(assets) == 0 {
			// Hack backported into 8.11:
			// Get the first subscription (should only be one), get the sub id and name and populate a mock resource
			// with an empty array.
			subId, subName := func() (string, string) {
				for subId, subName := range f.provider.GetSubscriptions() {
					return subId, subName
				}
				return "", ""
			}()

			success := f.write(ctx, fetching.ResourceInfo{
				Resource: &EmptyBatchResource{
					Type:    pair.Type,
					SubType: pair.SubType,
					SubId:   subId,
					SubName: subName,
				},
				CycleMetadata: cMetadata,
			})
			if !success {
				return nil
			}
			continue
		}

		success := f.write(ctx, fetching.ResourceInfo{
			Resource: &AzureBatchResource{
				// Every asset in the list has the same type and subtype
				Type:    pair.Type,
				SubType: pair.SubType,
				Assets:  assets,
			},
			CycleMetadata: cMetadata,
		})
		if !success {
			return nil
		}
	}

	return nil
}

func (f *AzureBatchAssetFetcher) write(ctx context.Context, resourceInfo fetching.ResourceInfo) bool {
	select {
	case <-ctx.Done():
		f.log.Infof("AzureBatchAssetFetcher.Fetch context err: %s", ctx.Err().Error())
		return false
	case f.resourceCh <- resourceInfo:
	}
	return true
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

type EmptyBatchResource struct {
	Type    string
	SubType string
	SubId   string
	SubName string
}

func (e *EmptyBatchResource) GetMetadata() (fetching.ResourceMetadata, error) {
	id := fmt.Sprintf("%s-%s", e.SubType, e.SubId)
	return fetching.ResourceMetadata{
		ID:      id,
		Type:    e.Type,
		SubType: e.SubType,
		Name:    id,
		Region:  azurelib.GlobalRegion,
	}, nil
}

func (e *EmptyBatchResource) GetData() any {
	return []any{}
}

func (e *EmptyBatchResource) GetElasticCommonData() (map[string]any, error) {
	return map[string]any{
		"cloud": map[string]any{
			"provider": "azure",
			"account": map[string]any{
				"id":   e.SubId,
				"name": e.SubName,
			},
		},
	}, nil
}
