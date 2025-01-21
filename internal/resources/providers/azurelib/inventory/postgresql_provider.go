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

package inventory

import (
	"context"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/postgresql/armpostgresql"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/postgresql/armpostgresqlflexibleservers"
	"github.com/samber/lo"

	"github.com/elastic/cloudbeat/internal/resources/utils/clog"
	"github.com/elastic/cloudbeat/internal/resources/utils/pointers"
)

type psqlAzureClientWrapper struct {
	AssetSingleServerConfigurations   func(ctx context.Context, subID, resourceGroup, serverName string, clientOptions *arm.ClientOptions, options *armpostgresql.ConfigurationsClientListByServerOptions) ([]armpostgresql.ConfigurationsClientListByServerResponse, error)
	AssetFlexibleServerConfigurations func(ctx context.Context, subID, resourceGroup, serverName string, clientOptions *arm.ClientOptions, options *armpostgresqlflexibleservers.ConfigurationsClientListByServerOptions) ([]armpostgresqlflexibleservers.ConfigurationsClientListByServerResponse, error)
	AssetSingleServerFirewallRules    func(ctx context.Context, subID, resourceGroup, serverName string, clientOptions *arm.ClientOptions, options *armpostgresql.FirewallRulesClientListByServerOptions) ([]armpostgresql.FirewallRulesClientListByServerResponse, error)
	AssetFlexibleServerFirewallRules  func(ctx context.Context, subID, resourceGroup, serverName string, clientOptions *arm.ClientOptions, options *armpostgresqlflexibleservers.FirewallRulesClientListByServerOptions) ([]armpostgresqlflexibleservers.FirewallRulesClientListByServerResponse, error)
}

type PostgresqlProviderAPI interface {
	ListSinglePostgresConfigurations(ctx context.Context, subID, resourceGroup, serverName string) ([]AzureAsset, error)
	ListFlexiblePostgresConfigurations(ctx context.Context, subID, resourceGroup, serverName string) ([]AzureAsset, error)
	ListSinglePostgresFirewallRules(ctx context.Context, subID, resourceGroup, serverName string) ([]AzureAsset, error)
	ListFlexiblePostgresFirewallRules(ctx context.Context, subID, resourceGroup, serverName string) ([]AzureAsset, error)
}

type psqlProvider struct {
	client *psqlAzureClientWrapper
	log    *clog.Logger //nolint:unused
}

func NewPostgresqlProvider(log *clog.Logger, credentials azcore.TokenCredential) PostgresqlProviderAPI {
	// We wrap the client, so we can mock it in tests
	wrapper := &psqlAzureClientWrapper{
		AssetSingleServerConfigurations: func(ctx context.Context, subID, resourceGroup, serverName string, clientOptions *arm.ClientOptions, options *armpostgresql.ConfigurationsClientListByServerOptions) ([]armpostgresql.ConfigurationsClientListByServerResponse, error) {
			cl, err := armpostgresql.NewConfigurationsClient(subID, credentials, clientOptions)
			if err != nil {
				return nil, err
			}
			return readPager(ctx, cl.NewListByServerPager(resourceGroup, serverName, options))
		},
		AssetFlexibleServerConfigurations: func(ctx context.Context, subID, resourceGroup, serverName string, clientOptions *arm.ClientOptions, options *armpostgresqlflexibleservers.ConfigurationsClientListByServerOptions) ([]armpostgresqlflexibleservers.ConfigurationsClientListByServerResponse, error) {
			cl, err := armpostgresqlflexibleservers.NewConfigurationsClient(subID, credentials, clientOptions)
			if err != nil {
				return nil, err
			}
			return readPager(ctx, cl.NewListByServerPager(resourceGroup, serverName, options))
		},
		AssetSingleServerFirewallRules: func(ctx context.Context, subID, resourceGroup, serverName string, clientOptions *arm.ClientOptions, options *armpostgresql.FirewallRulesClientListByServerOptions) ([]armpostgresql.FirewallRulesClientListByServerResponse, error) {
			cl, err := armpostgresql.NewFirewallRulesClient(subID, credentials, clientOptions)
			if err != nil {
				return nil, err
			}

			return readPager(ctx, cl.NewListByServerPager(resourceGroup, serverName, options))
		},
		AssetFlexibleServerFirewallRules: func(ctx context.Context, subID, resourceGroup, serverName string, clientOptions *arm.ClientOptions, options *armpostgresqlflexibleservers.FirewallRulesClientListByServerOptions) ([]armpostgresqlflexibleservers.FirewallRulesClientListByServerResponse, error) {
			cl, err := armpostgresqlflexibleservers.NewFirewallRulesClient(subID, credentials, clientOptions)
			if err != nil {
				return nil, err
			}

			return readPager(ctx, cl.NewListByServerPager(resourceGroup, serverName, options))
		},
	}

	return &psqlProvider{
		client: wrapper,
		log:    log,
	}
}

func (p *psqlProvider) ListSinglePostgresConfigurations(ctx context.Context, subID, resourceGroup, serverName string) ([]AzureAsset, error) {
	paged, err := p.client.AssetSingleServerConfigurations(ctx, subID, resourceGroup, serverName, nil, nil)
	if err != nil {
		return nil, err
	}

	configs := lo.FlatMap(paged, func(p armpostgresql.ConfigurationsClientListByServerResponse, _ int) []*armpostgresql.Configuration {
		return p.Value
	})

	assets := make([]AzureAsset, 0, len(configs))
	for _, c := range configs {
		if c == nil || c.Properties == nil {
			continue
		}

		assets = append(assets, AzureAsset{
			Id:       pointers.Deref(c.ID),
			Name:     pointers.Deref(c.Name),
			Location: assetLocationGlobal,
			Properties: map[string]any{
				"source":       pointers.Deref(c.Properties.Source),
				"value":        strings.ToLower(pointers.Deref(c.Properties.Value)),
				"dataType":     pointers.Deref(c.Properties.DataType),
				"defaultValue": pointers.Deref(c.Properties.DefaultValue),
			},
			ResourceGroup:  resourceGroup,
			SubscriptionId: subID,
			TenantId:       "",
			Type:           pointers.Deref(c.Type),
		})
	}

	return assets, nil
}

func (p *psqlProvider) ListFlexiblePostgresConfigurations(ctx context.Context, subID, resourceGroup, serverName string) ([]AzureAsset, error) {
	paged, err := p.client.AssetFlexibleServerConfigurations(ctx, subID, resourceGroup, serverName, nil, nil)
	if err != nil {
		return nil, err
	}

	configs := lo.FlatMap(paged, func(p armpostgresqlflexibleservers.ConfigurationsClientListByServerResponse, _ int) []*armpostgresqlflexibleservers.Configuration {
		return p.Value
	})

	assets := make([]AzureAsset, 0, len(configs))
	for _, c := range configs {
		if c == nil || c.Properties == nil {
			continue
		}

		assets = append(assets, AzureAsset{
			Id:       pointers.Deref(c.ID),
			Name:     pointers.Deref(c.Name),
			Location: assetLocationGlobal,
			Properties: map[string]any{
				"source":       pointers.Deref(c.Properties.Source),
				"value":        strings.ToLower(pointers.Deref(c.Properties.Value)),
				"dataType":     string(pointers.Deref(c.Properties.DataType)),
				"defaultValue": pointers.Deref(c.Properties.DefaultValue),
			},
			ResourceGroup:  resourceGroup,
			SubscriptionId: subID,
			TenantId:       "",
			Type:           pointers.Deref(c.Type),
		})
	}

	return assets, nil
}

func (p *psqlProvider) ListSinglePostgresFirewallRules(ctx context.Context, subID, resourceGroup, serverName string) ([]AzureAsset, error) {
	paged, err := p.client.AssetSingleServerFirewallRules(ctx, subID, resourceGroup, serverName, nil, nil)
	if err != nil {
		return nil, err
	}

	configs := lo.FlatMap(paged, func(p armpostgresql.FirewallRulesClientListByServerResponse, _ int) []*armpostgresql.FirewallRule {
		return p.Value
	})

	assets := make([]AzureAsset, 0, len(configs))
	for _, fr := range configs {
		if fr == nil || fr.Properties == nil {
			continue
		}

		assets = append(assets, AzureAsset{
			Id:       pointers.Deref(fr.ID),
			Name:     pointers.Deref(fr.Name),
			Location: assetLocationGlobal,
			Properties: map[string]any{
				"startIPAddress": pointers.Deref(fr.Properties.StartIPAddress),
				"endIPAddress":   pointers.Deref(fr.Properties.EndIPAddress),
			},
			ResourceGroup:  resourceGroup,
			SubscriptionId: subID,
			TenantId:       "",
			Type:           pointers.Deref(fr.Type),
		})
	}

	return assets, nil
}

func (p *psqlProvider) ListFlexiblePostgresFirewallRules(ctx context.Context, subID, resourceGroup, serverName string) ([]AzureAsset, error) {
	paged, err := p.client.AssetFlexibleServerFirewallRules(ctx, subID, resourceGroup, serverName, nil, nil)
	if err != nil {
		return nil, err
	}

	configs := lo.FlatMap(paged, func(p armpostgresqlflexibleservers.FirewallRulesClientListByServerResponse, _ int) []*armpostgresqlflexibleservers.FirewallRule {
		return p.Value
	})

	assets := make([]AzureAsset, 0, len(configs))
	for _, fr := range configs {
		if fr == nil || fr.Properties == nil {
			continue
		}

		assets = append(assets, AzureAsset{
			Id:       pointers.Deref(fr.ID),
			Name:     pointers.Deref(fr.Name),
			Location: assetLocationGlobal,
			Properties: map[string]any{
				"startIPAddress": pointers.Deref(fr.Properties.StartIPAddress),
				"endIPAddress":   pointers.Deref(fr.Properties.EndIPAddress),
			},
			ResourceGroup:  resourceGroup,
			SubscriptionId: subID,
			TenantId:       "",
			Type:           pointers.Deref(fr.Type),
		})
	}

	return assets, nil
}
