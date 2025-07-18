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

package governance

import (
	"context"
	"fmt"

	"github.com/samber/lo"

	"github.com/elastic/cloudbeat/internal/infra/clog"
	"github.com/elastic/cloudbeat/internal/resources/fetching"
	"github.com/elastic/cloudbeat/internal/resources/fetching/cycle"
	"github.com/elastic/cloudbeat/internal/resources/providers/azurelib/inventory"
	"github.com/elastic/cloudbeat/internal/resources/utils/strings"
)

type ManagementGroup struct {
	// FullyQualifiedID is the fully qualified ID for the management group.
	// For example, /providers/Microsoft.Management/managementGroups/0000000-0000-0000-0000-000000000000
	FullyQualifiedID string
	DisplayName      string
}

type Subscription struct {
	// FullyQualifiedID is the fully qualified ID for the subscription.
	// For example, /subscriptions/8d65815f-a5b6-402f-9298-045155da7d74
	FullyQualifiedID string
	// ShortID is the ID of the subscription which is the SubscriptionId field of the asset.
	// For example, 8d65815f-a5b6-402f-9298-045155da7d74.
	ShortID         string
	DisplayName     string
	ManagementGroup ManagementGroup
}

func (s Subscription) GetCloudAccountMetadata() fetching.CloudAccountMetadata {
	return fetching.CloudAccountMetadata{
		AccountId:        s.FullyQualifiedID,
		AccountName:      s.DisplayName,
		OrganisationId:   s.ManagementGroup.FullyQualifiedID,
		OrganizationName: s.ManagementGroup.DisplayName,
	}
}

type ProviderAPI interface {
	GetSubscriptions(ctx context.Context, cycleMetadata cycle.Metadata) (map[string]Subscription, error)
}

type provider struct {
	cache  *cycle.Cache[map[string]Subscription]
	client inventory.ResourceGraphProviderAPI
}

func NewProvider(log *clog.Logger, client inventory.ResourceGraphProviderAPI) ProviderAPI {
	p := provider{
		client: client,
	}
	p.cache = cycle.NewCache[map[string]Subscription](log.Named("governance"))
	return &p
}

func (p *provider) GetSubscriptions(ctx context.Context, cycleMetadata cycle.Metadata) (map[string]Subscription, error) {
	return p.cache.GetValue(ctx, cycleMetadata, p.scan)
}

func (p *provider) scan(ctx context.Context) (map[string]Subscription, error) {
	const (
		managementGroupType       = "microsoft.management/managementgroups"
		subscriptionType          = "microsoft.resources/subscriptions"
		ancestorChainPropertyName = "managementGroupAncestorsChain"
	)

	assets, err := p.client.ListAllAssetTypesByName(
		ctx,
		"resourcecontainers",
		[]string{managementGroupType, subscriptionType},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to scan resources: %w", err)
	}

	managementGroups := make(map[string]ManagementGroup)
	for _, asset := range lo.Filter(assets, typeFilter(managementGroupType)) {
		managementGroups[asset.Name] = ManagementGroup{
			FullyQualifiedID: asset.Id,
			DisplayName:      strings.FirstNonEmpty(asset.DisplayName, asset.Name),
		}
	}

	subscriptions := make(map[string]Subscription)
	for _, asset := range lo.Filter(assets, typeFilter(subscriptionType)) {
		chain, ok := asset.Properties[ancestorChainPropertyName].([]any)
		if !ok || len(chain) == 0 {
			continue
		}
		parent, ok := chain[0].(map[string]any)
		if !ok {
			continue
		}

		mg := managementGroups[strings.FromMap(parent, "name")]
		subscriptions[asset.SubscriptionId] = Subscription{
			FullyQualifiedID: asset.Id,
			ShortID:          asset.SubscriptionId,
			DisplayName:      strings.FirstNonEmpty(asset.DisplayName, asset.Name),
			ManagementGroup:  mg,
		}
	}

	return subscriptions, nil
}

func typeFilter(tpe string) func(item inventory.AzureAsset, index int) bool {
	return func(item inventory.AzureAsset, _ int) bool {
		return item.Type == tpe
	}
}
