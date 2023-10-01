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
	"golang.org/x/exp/maps"

	"github.com/elastic/cloudbeat/resources/fetching"
	"github.com/elastic/cloudbeat/resources/providers/azurelib/inventory"
)

type AzureActivityLogAlertsFetcher struct {
	log        *logp.Logger
	resourceCh chan fetching.ResourceInfo
	provider   inventory.ServiceAPI
}

type AzureActivityLogAlertsAsset struct {
	ActivityLogAlerts []inventory.AzureAsset `json:"activity_log_alerts,omitempty"`
}

type AzureActivityLogAlertsResource struct {
	Type    string
	SubType string
	Asset   AzureActivityLogAlertsAsset `json:"asset,omitempty"`
}

var AzureActivityLogAlertsResourceTypes = map[string]string{
	inventory.ActivityLogAlertAssetType: fetching.AzureActivityLogAlertType,
}

func NewAzureActivityLogAlertsFetcher(log *logp.Logger, ch chan fetching.ResourceInfo, provider inventory.ServiceAPI) *AzureActivityLogAlertsFetcher {
	return &AzureActivityLogAlertsFetcher{
		log:        log,
		resourceCh: ch,
		provider:   provider,
	}
}

func (f *AzureActivityLogAlertsFetcher) Fetch(ctx context.Context, cMetadata fetching.CycleMetadata) error {
	f.log.Info("Starting AzureActivityLogAlertsFetcher.Fetch")
	assets, err := f.provider.ListAllAssetTypesByName(maps.Keys(AzureActivityLogAlertsResourceTypes))
	if err != nil {
		return err
	}

	if len(assets) == 0 {
		return nil
	}

	select {
	case <-ctx.Done():
		f.log.Infof("AzureActivityLogAlertsFetcher.Fetch context err: %s", ctx.Err().Error())
		return nil
	// TODO: Groups by subscription id to create multiple batches of assets
	case f.resourceCh <- fetching.ResourceInfo{
		CycleMetadata: cMetadata,
		Resource: &AzureActivityLogAlertsResource{
			// Every asset in the list has the same type and subtype
			Type:    AzureActivityLogAlertsResourceTypes[assets[0].Type],
			SubType: getAzureActivityLogAlertsSubType(assets[0].Type),
			Asset: AzureActivityLogAlertsAsset{
				ActivityLogAlerts: assets,
			},
		},
	}:
	}

	return nil
}

func getAzureActivityLogAlertsSubType(assetType string) string {
	return ""
}

func (f *AzureActivityLogAlertsFetcher) Stop() {}

func (r *AzureActivityLogAlertsResource) GetData() any {
	return r.Asset
}

func (r *AzureActivityLogAlertsResource) GetMetadata() (fetching.ResourceMetadata, error) {
	// Assuming all batch in not empty includes assets of the same subscription
	id := fmt.Sprintf("%s-%s", r.Type, r.Asset.ActivityLogAlerts[0].SubscriptionId)
	return fetching.ResourceMetadata{
		ID:      id,
		Type:    r.Type,
		SubType: r.SubType,
		Name:    id,
		// TODO: Make sure ActivityLogAlerts are not location scoped (benchmarks do not check location)
		Region: "",
	}, nil
}

func (r *AzureActivityLogAlertsResource) GetElasticCommonData() (map[string]any, error) {
	return nil, nil
}
