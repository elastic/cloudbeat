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

	"github.com/elastic/cloudbeat/internal/resources/fetching/cycle"
	"github.com/elastic/cloudbeat/internal/resources/providers/azurelib"
	"github.com/elastic/cloudbeat/internal/resources/providers/azurelib/inventory"
)

type postgresqlEnricher struct {
	provider azurelib.ProviderAPI
}

func (p postgresqlEnricher) Enrich(ctx context.Context, _ cycle.Metadata, assets []inventory.AzureAsset) error {
	var errs error

	enrichFn := []func(context.Context, *inventory.AzureAsset) error{
		p.enrichConfigurations,
		p.enrichFirewallRules,
	}

	for i, a := range assets {
		if !isPsql(a.Type) {
			continue
		}

		for _, fn := range enrichFn {
			if err := fn(ctx, &a); err != nil {
				errs = errors.Join(errs, err)
			}
		}

		assets[i] = a
	}

	return errs
}

func (p postgresqlEnricher) enrichConfigurations(ctx context.Context, a *inventory.AzureAsset) error {
	configs, err := p.listConfigurations(ctx, a)
	if err != nil {
		return err
	}

	if len(configs) == 0 {
		return nil
	}

	a.AddExtension(inventory.ExtensionPostgresqlConfigurations, configs)
	return nil
}

func (p postgresqlEnricher) listConfigurations(ctx context.Context, a *inventory.AzureAsset) ([]inventory.AzureAsset, error) {
	if isPsqlSingleServer(a.Type) {
		return p.provider.ListSinglePostgresConfigurations(ctx, a.SubscriptionId, a.ResourceGroup, a.Name)
	}

	return p.provider.ListFlexiblePostgresConfigurations(ctx, a.SubscriptionId, a.ResourceGroup, a.Name)
}

func (p postgresqlEnricher) enrichFirewallRules(ctx context.Context, a *inventory.AzureAsset) error {
	configs, err := p.listFirewallRules(ctx, a)
	if err != nil {
		return err
	}

	if len(configs) == 0 {
		return nil
	}

	a.AddExtension(inventory.ExtensionPostgresqlFirewallRules, configs)
	return nil
}

func (p postgresqlEnricher) listFirewallRules(ctx context.Context, a *inventory.AzureAsset) ([]inventory.AzureAsset, error) {
	if isPsqlSingleServer(a.Type) {
		return p.provider.ListSinglePostgresFirewallRules(ctx, a.SubscriptionId, a.ResourceGroup, a.Name)
	}

	return p.provider.ListFlexiblePostgresFirewallRules(ctx, a.SubscriptionId, a.ResourceGroup, a.Name)
}

func isPsqlSingleServer(t string) bool {
	return t == inventory.PostgreSQLDBAssetType
}

func isPsql(t string) bool {
	return isPsqlSingleServer(t) || t == inventory.FlexiblePostgreSQLDBAssetType
}
