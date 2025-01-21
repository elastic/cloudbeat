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
	"errors"
	"fmt"

	"github.com/elastic/cloudbeat/internal/resources/fetching"
	"github.com/elastic/cloudbeat/internal/resources/fetching/cycle"
	"github.com/elastic/cloudbeat/internal/resources/providers/azurelib"
	"github.com/elastic/cloudbeat/internal/resources/providers/azurelib/inventory"
	"github.com/elastic/cloudbeat/internal/resources/utils/clog"
)

type AzureSecurityAssetFetcher struct {
	log        *clog.Logger
	resourceCh chan fetching.ResourceInfo
	provider   azurelib.ProviderAPI
}

func NewAzureSecurityAssetFetcher(log *clog.Logger, ch chan fetching.ResourceInfo, provider azurelib.ProviderAPI) *AzureSecurityAssetFetcher {
	return &AzureSecurityAssetFetcher{
		log:        log,
		resourceCh: ch,
		provider:   provider,
	}
}

var AzureSecurityAssetTypeToTypePair = map[string]typePair{
	inventory.SecurityContactsAssetType:            newPair(fetching.AzureSecurityContactsType, fetching.MonitoringIdentity),
	inventory.SecurityAutoProvisioningSettingsType: newPair(fetching.AzureAutoProvisioningSettingsType, fetching.MonitoringIdentity),
}

func (f *AzureSecurityAssetFetcher) Fetch(ctx context.Context, cycleMetadata cycle.Metadata) error {
	f.log.Info("Starting AzureSecurityAssetFetcher.Fetch")

	subscriptions, err := f.provider.GetSubscriptions(ctx, cycleMetadata)
	if err != nil {
		return fmt.Errorf("error fetching subscription information: %w", err)
	}

	fetches := map[string]func(context.Context, string) ([]inventory.AzureAsset, error){
		inventory.SecurityAutoProvisioningSettingsType: f.provider.ListAutoProvisioningSettings,
		inventory.SecurityContactsAssetType:            f.provider.ListSecurityContacts,
	}

	var errs []error
	for _, sub := range subscriptions {
		for assetType, fn := range fetches {
			securityContacts, err := fn(ctx, sub.ShortID)
			if err != nil {
				f.log.Errorf("AzureSecurityAssetFetcher.Fetch failed to fetch %s for subscription %s: %s", assetType, sub.ShortID, err.Error())
				errs = append(errs, err)
				continue
			}
			if securityContacts == nil {
				securityContacts = []inventory.AzureAsset{}
			}

			select {
			case <-ctx.Done():
				err := ctx.Err()
				f.log.Infof("AzureSecurityAssetFetcher.Fetch context err: %s", err.Error())
				errs = append(errs, err)
				return errors.Join(errs...)
			case f.resourceCh <- fetching.ResourceInfo{
				CycleMetadata: cycleMetadata,
				Resource: &AzureBatchResource{
					typePair:     AzureSecurityAssetTypeToTypePair[assetType],
					Assets:       securityContacts,
					Subscription: sub,
				},
			}:
			}
		}
	}

	return errors.Join(errs...)
}

func (f *AzureSecurityAssetFetcher) Stop() {}
