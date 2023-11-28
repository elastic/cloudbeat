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

	"github.com/elastic/elastic-agent-libs/logp"
	"github.com/samber/lo"

	"github.com/elastic/cloudbeat/resources/fetching"
	"github.com/elastic/cloudbeat/resources/fetching/cycle"
	"github.com/elastic/cloudbeat/resources/providers/azurelib/inventory"
	"github.com/elastic/cloudbeat/resources/utils/strings"
)

type ManagementGroup struct {
	ID          string
	DisplayName string
}

type Subscription struct {
	ID          string
	DisplayName string
	MG          ManagementGroup
}

func (s Subscription) GetCloudAccountMetadata() fetching.CloudAccountMetadata {
	return fetching.CloudAccountMetadata{
		AccountId:        s.ID,
		AccountName:      s.DisplayName,
		OrganisationId:   s.MG.ID,
		OrganizationName: s.MG.DisplayName,
	}
}

type ProviderAPI interface {
	GetSubscriptions(ctx context.Context, cycleMetadata cycle.Metadata) (map[string]Subscription, error)
}

type provider struct {
	cache  cycle.Cache[map[string]Subscription]
	client inventory.ProviderAPI
}

func NewProvider(log *logp.Logger, client inventory.ProviderAPI) ProviderAPI {
	p := provider{
		client: client,
	}
	p.cache = cycle.NewCache(log.Named("governance"), p.scan)
	return &p
}

func (p *provider) GetSubscriptions(ctx context.Context, cycleMetadata cycle.Metadata) (map[string]Subscription, error) {
	return p.cache.GetValue(ctx, cycleMetadata)
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
			ID:          asset.Id,
			DisplayName: asset.DisplayName,
		}
	}

	subscriptions := make(map[string]Subscription)
	for _, asset := range lo.Filter(assets, typeFilter(subscriptionType)) {
		chain, ok := asset.Properties[ancestorChainPropertyName].([]any)
		if !ok || len(chain) == 0 {
			continue
		}
		parent, _ := chain[0].(map[string]any)

		mg := managementGroups[strings.FromMap(parent, "name")]
		subscriptions[asset.SubscriptionId] = Subscription{
			ID:          asset.Id,
			DisplayName: asset.Name,
			MG:          mg,
		}
	}

	return subscriptions, nil
}

func typeFilter(tpe string) func(item inventory.AzureAsset, index int) bool {
	return func(item inventory.AzureAsset, index int) bool {
		return item.Type == tpe
	}
}
