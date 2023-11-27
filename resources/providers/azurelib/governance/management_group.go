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
	"sync"

	"github.com/elastic/elastic-agent-libs/logp"
	"github.com/samber/lo"

	"github.com/elastic/cloudbeat/resources/fetching"
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
	GetSubscriptions(ctx context.Context, cycle fetching.CycleMetadata) (map[string]Subscription, error)
}

type provider struct {
	log          *logp.Logger
	lastSequence int64
	mu           sync.Mutex
	client       inventory.ProviderAPI

	cachedSubscriptions map[string]Subscription
}

func NewProvider(log *logp.Logger, client inventory.ProviderAPI) ProviderAPI {
	return &provider{
		log:          log.Named("governance"),
		client:       client,
		lastSequence: -1,
	}
}

func (p *provider) GetSubscriptions(ctx context.Context, cycle fetching.CycleMetadata) (map[string]Subscription, error) {
	err := p.maybeScan(ctx, cycle)
	return p.cachedSubscriptions, err
}

func (p *provider) maybeScan(ctx context.Context, cycle fetching.CycleMetadata) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.needsUpdate(cycle) {
		return nil
	}

	if err := p.scan(ctx); err != nil {
		if p.cachedSubscriptions == nil {
			return fmt.Errorf("failed to scan subscriptions: %w", err)
		}

		p.lastSequence = cycle.Sequence
		p.log.Errorf("Failed to scan subscriptions, re-using cached values: %v", err)
		return nil
	}

	p.lastSequence = cycle.Sequence
	return nil
}

func (p *provider) needsUpdate(cycle fetching.CycleMetadata) bool {
	return p.lastSequence < cycle.Sequence
}

func (p *provider) scan(ctx context.Context) error {
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
		return fmt.Errorf("failed to scan resources: %w", err)
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

	p.cachedSubscriptions = subscriptions
	return nil
}

func typeFilter(tpe string) func(item inventory.AzureAsset, index int) bool {
	return func(item inventory.AzureAsset, index int) bool {
		return item.Type == tpe
	}
}
