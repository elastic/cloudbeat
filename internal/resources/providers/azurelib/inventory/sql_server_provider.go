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
	"errors"
	"fmt"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/sql/armsql"
	"github.com/samber/lo"

	"github.com/elastic/cloudbeat/internal/infra/clog"
	"github.com/elastic/cloudbeat/internal/resources/utils/pointers"
)

type sqlAzureClientWrapper struct {
	AssetEncryptionProtector                    func(ctx context.Context, subID, resourceGroup, serverName string, clientOptions *arm.ClientOptions, options *armsql.EncryptionProtectorsClientListByServerOptions) ([]armsql.EncryptionProtectorsClientListByServerResponse, error)
	AssetBlobAuditingPolicies                   func(ctx context.Context, subID, resourceGroup, serverName string, clientOptions *arm.ClientOptions, options *armsql.ServerBlobAuditingPoliciesClientGetOptions) (armsql.ServerBlobAuditingPoliciesClientGetResponse, error)
	AssetTransparentDataEncryptions             func(ctx context.Context, subID, resourceGroup, serverName, dbName string, clientOptions *arm.ClientOptions, options *armsql.TransparentDataEncryptionsClientListByDatabaseOptions) ([]armsql.TransparentDataEncryptionsClientListByDatabaseResponse, error)
	AssetDatabases                              func(ctx context.Context, subID, resourceGroup, serverName string, clientOptions *arm.ClientOptions, options *armsql.DatabasesClientListByServerOptions) ([]armsql.DatabasesClientListByServerResponse, error)
	AssetServerAdvancedThreatProtectionSettings func(ctx context.Context, subID, resourceGroup, serverName string, clientOptions *arm.ClientOptions, options *armsql.ServerAdvancedThreatProtectionSettingsClientListByServerOptions) ([]armsql.ServerAdvancedThreatProtectionSettingsClientListByServerResponse, error)
	AssetServerFirewallRules                    func(ctx context.Context, subID, resourceGroup, serverName string, clientOptions *arm.ClientOptions, listOptions *armsql.FirewallRulesClientListByServerOptions) ([]armsql.FirewallRulesClientListByServerResponse, error)
}

type SQLProviderAPI interface {
	ListSQLEncryptionProtector(ctx context.Context, subID, resourceGroup, serverName string) ([]AzureAsset, error)
	ListSQLTransparentDataEncryptions(ctx context.Context, subID, resourceGroup, serverName string) ([]AzureAsset, error)
	GetSQLBlobAuditingPolicies(ctx context.Context, subID, resourceGroup, serverName string) ([]AzureAsset, error)
	ListSQLAdvancedThreatProtectionSettings(ctx context.Context, subID, resourceGroup, serverName string) ([]AzureAsset, error)
	ListSQLFirewallRules(ctx context.Context, subID, resourceGroup, serverName string) ([]AzureAsset, error)
}

type sqlProvider struct {
	client        *sqlAzureClientWrapper
	log           *clog.Logger //nolint:unused
	clientOptions *arm.ClientOptions
}

func NewSQLProvider(log *clog.Logger, credentials azcore.TokenCredential) SQLProviderAPI {
	// We wrap the client, so we can mock it in tests
	wrapper := &sqlAzureClientWrapper{
		AssetEncryptionProtector: func(ctx context.Context, subID, resourceGroup, serverName string, clientOptions *arm.ClientOptions, options *armsql.EncryptionProtectorsClientListByServerOptions) ([]armsql.EncryptionProtectorsClientListByServerResponse, error) {
			cl, err := armsql.NewEncryptionProtectorsClient(subID, credentials, clientOptions)
			if err != nil {
				return nil, err
			}
			return readPager(ctx, cl.NewListByServerPager(resourceGroup, serverName, options))
		},
		AssetBlobAuditingPolicies: func(ctx context.Context, subID, resourceGroup, serverName string, clientOptions *arm.ClientOptions, options *armsql.ServerBlobAuditingPoliciesClientGetOptions) (armsql.ServerBlobAuditingPoliciesClientGetResponse, error) {
			cl, err := armsql.NewServerBlobAuditingPoliciesClient(subID, credentials, clientOptions)
			if err != nil {
				return armsql.ServerBlobAuditingPoliciesClientGetResponse{}, err
			}
			return cl.Get(ctx, resourceGroup, serverName, options)
		},
		AssetTransparentDataEncryptions: func(ctx context.Context, subID, resourceGroup, serverName, dbName string, clientOptions *arm.ClientOptions, options *armsql.TransparentDataEncryptionsClientListByDatabaseOptions) ([]armsql.TransparentDataEncryptionsClientListByDatabaseResponse, error) {
			cl, err := armsql.NewTransparentDataEncryptionsClient(subID, credentials, clientOptions)
			if err != nil {
				return nil, err
			}
			return readPager(ctx, cl.NewListByDatabasePager(resourceGroup, serverName, dbName, options))
		},
		AssetDatabases: func(ctx context.Context, subID, resourceGroup, serverName string, clientOptions *arm.ClientOptions, options *armsql.DatabasesClientListByServerOptions) ([]armsql.DatabasesClientListByServerResponse, error) {
			cl, err := armsql.NewDatabasesClient(subID, credentials, clientOptions)
			if err != nil {
				return nil, err
			}
			return readPager(ctx, cl.NewListByServerPager(resourceGroup, serverName, options))
		},
		AssetServerAdvancedThreatProtectionSettings: func(ctx context.Context, subID, resourceGroup, serverName string, clientOptions *arm.ClientOptions, options *armsql.ServerAdvancedThreatProtectionSettingsClientListByServerOptions) ([]armsql.ServerAdvancedThreatProtectionSettingsClientListByServerResponse, error) {
			cl, err := armsql.NewServerAdvancedThreatProtectionSettingsClient(subID, credentials, clientOptions)
			if err != nil {
				return nil, err
			}
			return readPager(ctx, cl.NewListByServerPager(resourceGroup, serverName, options))
		},
		AssetServerFirewallRules: func(ctx context.Context, subID, resourceGroup, serverName string, clientOptions *arm.ClientOptions, listOptions *armsql.FirewallRulesClientListByServerOptions) ([]armsql.FirewallRulesClientListByServerResponse, error) {
			cl, err := armsql.NewFirewallRulesClient(subID, credentials, clientOptions)
			if err != nil {
				return nil, err
			}
			return readPager(ctx, cl.NewListByServerPager(resourceGroup, serverName, listOptions))
		},
	}

	return &sqlProvider{
		log:    log,
		client: wrapper,
	}
}

func (p *sqlProvider) ListSQLEncryptionProtector(ctx context.Context, subID, resourceGroup, serverName string) ([]AzureAsset, error) {
	encryptProtectors, err := p.client.AssetEncryptionProtector(ctx, subID, resourceGroup, serverName, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("problem on listing sql encryption protectors (%w)", err)
	}

	capacity := lo.Reduce(encryptProtectors, func(acc int, i armsql.EncryptionProtectorsClientListByServerResponse, _ int) int {
		return acc + len(i.Value)
	}, 0)

	if capacity == 0 {
		return nil, nil
	}

	assets := make([]AzureAsset, 0, capacity)
	for _, epWrapper := range encryptProtectors {
		for _, ep := range epWrapper.Value {
			if ep == nil || ep.Properties == nil {
				continue
			}

			assets = append(assets, convertEncryptionProtector(ep, resourceGroup, subID))
		}
	}

	return assets, nil
}

func (p *sqlProvider) GetSQLBlobAuditingPolicies(ctx context.Context, subID, resourceGroup, serverName string) ([]AzureAsset, error) {
	policy, err := p.client.AssetBlobAuditingPolicies(ctx, subID, resourceGroup, serverName, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("problem on getting sql blob auditing policies (%w)", err)
	}

	if policy.Properties == nil {
		return nil, nil
	}

	return []AzureAsset{
		{
			Id:       pointers.Deref(policy.ID),
			Name:     pointers.Deref(policy.Name),
			Location: assetLocationGlobal,
			Properties: map[string]any{
				"state":                        string(pointers.Deref(policy.Properties.State)),
				"isAzureMonitorTargetEnabled":  pointers.Deref(policy.Properties.IsAzureMonitorTargetEnabled),
				"isDevopsAuditEnabled":         pointers.Deref(policy.Properties.IsDevopsAuditEnabled),
				"isManagedIdentityInUse":       pointers.Deref(policy.Properties.IsManagedIdentityInUse),
				"isStorageSecondaryKeyInUse":   pointers.Deref(policy.Properties.IsStorageSecondaryKeyInUse),
				"queueDelayMs":                 pointers.Deref(policy.Properties.QueueDelayMs),
				"retentionDays":                pointers.Deref(policy.Properties.RetentionDays),
				"storageAccountAccessKey":      pointers.Deref(policy.Properties.StorageAccountAccessKey),
				"storageAccountSubscriptionID": pointers.Deref(policy.Properties.StorageAccountSubscriptionID),
				"storageEndpoint":              pointers.Deref(policy.Properties.StorageEndpoint),

				"auditActionsAndGroups": lo.Map(policy.Properties.AuditActionsAndGroups, func(s *string, _ int) string {
					return pointers.Deref(s)
				}),
			},
			ResourceGroup:  resourceGroup,
			SubscriptionId: subID,
			TenantId:       "",
			Type:           pointers.Deref(policy.Type),
		},
	}, nil
}

func (p *sqlProvider) ListSQLTransparentDataEncryptions(ctx context.Context, subID, resourceGroup, serverName string) ([]AzureAsset, error) {
	pagedDbs, err := p.client.AssetDatabases(ctx, subID, resourceGroup, serverName, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("problem on listing sql databases for sql server %s (%w)", serverName, err)
	}

	dbs := lo.FlatMap(pagedDbs, func(db armsql.DatabasesClientListByServerResponse, _ int) []*armsql.Database {
		return db.Value
	})

	var errs error
	assets := make([]AzureAsset, 0)
	for _, db := range dbs {
		if db == nil {
			continue
		}

		assetsByDb, err := p.listTransparentDataEncryptionsByDB(ctx, subID, resourceGroup, serverName, pointers.Deref(db.Name))
		if err != nil {
			errs = errors.Join(errs, err)
		}

		assets = append(assets, assetsByDb...)
	}

	return assets, errs
}

func (p *sqlProvider) ListSQLAdvancedThreatProtectionSettings(ctx context.Context, subID, resourceGroup, serverName string) ([]AzureAsset, error) {
	pagedSettings, err := p.client.AssetServerAdvancedThreatProtectionSettings(ctx, subID, resourceGroup, serverName, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("problem on listing advanced threat protection settings for server %s (%w)", serverName, err)
	}

	settings := lo.FlatMap(pagedSettings, func(page armsql.ServerAdvancedThreatProtectionSettingsClientListByServerResponse, _ int) []*armsql.ServerAdvancedThreatProtection {
		return page.Value
	})

	assets := make([]AzureAsset, 0, len(settings))
	for _, s := range settings {
		if s == nil || s.Properties == nil {
			continue
		}

		assets = append(assets, convertAdvancedThreatProtectionSettings(s, resourceGroup, subID))
	}

	return assets, nil
}

func (p *sqlProvider) listTransparentDataEncryptionsByDB(ctx context.Context, subID, resourceGroup, serverName, dbName string) ([]AzureAsset, error) {
	pagedTdes, err := p.client.AssetTransparentDataEncryptions(ctx, subID, resourceGroup, serverName, dbName, nil, nil)
	if err != nil {
		return nil, err
	}

	capacity := lo.Reduce(pagedTdes, func(acc int, i armsql.TransparentDataEncryptionsClientListByDatabaseResponse, _ int) int {
		return acc + len(i.Value)
	}, 0)

	assets := make([]AzureAsset, 0, capacity)
	for _, tdes := range pagedTdes {
		for _, tde := range tdes.Value {
			if tde == nil || tde.Properties == nil {
				continue
			}

			assets = append(assets, convertTransparentDataEncryption(tde, dbName, subID, resourceGroup))
		}
	}

	return assets, nil
}

func convertTransparentDataEncryption(tde *armsql.LogicalDatabaseTransparentDataEncryption, dbName, subID, resourceGroup string) AzureAsset {
	return AzureAsset{
		Id:       pointers.Deref(tde.ID),
		Name:     pointers.Deref(tde.Name),
		Location: assetLocationGlobal,
		Properties: map[string]any{
			"databaseName": dbName,
			"state":        string(pointers.Deref(tde.Properties.State)),
		},
		ResourceGroup:  resourceGroup,
		SubscriptionId: subID,
		TenantId:       "",
		Type:           pointers.Deref(tde.Type),
	}
}

func convertEncryptionProtector(ep *armsql.EncryptionProtector, resourceGroup, subID string) AzureAsset {
	return AzureAsset{
		Id:       pointers.Deref(ep.ID),
		Name:     pointers.Deref(ep.Name),
		Location: pointers.Deref(ep.Location),
		Properties: map[string]any{
			"kind":                pointers.Deref(ep.Kind),
			"serverKeyType":       string(pointers.Deref(ep.Properties.ServerKeyType)),
			"autoRotationEnabled": pointers.Deref(ep.Properties.AutoRotationEnabled),
			"serverKeyName":       pointers.Deref(ep.Properties.ServerKeyName),
			"subregion":           pointers.Deref(ep.Properties.Subregion),
			"thumbprint":          pointers.Deref(ep.Properties.Thumbprint),
			"uri":                 pointers.Deref(ep.Properties.URI),
		},
		ResourceGroup:  resourceGroup,
		SubscriptionId: subID,
		TenantId:       "",
		Type:           pointers.Deref(ep.Type),
	}
}

func convertAdvancedThreatProtectionSettings(s *armsql.ServerAdvancedThreatProtection, resourceGroup, subID string) AzureAsset {
	return AzureAsset{
		Id:       pointers.Deref(s.ID),
		Name:     pointers.Deref(s.Name),
		Location: assetLocationGlobal,
		Properties: map[string]any{
			"state":        string(pointers.Deref(s.Properties.State)),
			"creationTime": pointers.Deref(s.Properties.CreationTime),
		},
		ResourceGroup:  resourceGroup,
		SubscriptionId: subID,
		TenantId:       "",
		Type:           pointers.Deref(s.Type),
	}
}

func (p *sqlProvider) ListSQLFirewallRules(ctx context.Context, subID, resourceGroup, serverName string) ([]AzureAsset, error) {
	responses, err := p.client.AssetServerFirewallRules(ctx, subID, resourceGroup, serverName, p.clientOptions, nil)
	if err != nil {
		return nil, err
	}

	return lo.FlatMap(responses, func(item armsql.FirewallRulesClientListByServerResponse, _ int) []AzureAsset {
		return lo.FilterMap(item.Value, func(item *armsql.FirewallRule, _ int) (AzureAsset, bool) {
			if item == nil {
				return AzureAsset{}, false
			}

			return p.convertFirewallRule(item, subID, resourceGroup), true
		})
	}), nil
}

func (p *sqlProvider) convertFirewallRule(item *armsql.FirewallRule, subID, resourceGroup string) AzureAsset {
	a := AzureAsset{
		Id:             pointers.Deref(item.ID),
		Name:           pointers.Deref(item.Name),
		Type:           strings.ToLower(pointers.Deref(item.Type)),
		SubscriptionId: subID,
		ResourceGroup:  resourceGroup,
	}

	if item.Properties != nil {
		a.Properties = map[string]any{
			"startIpAddress": pointers.Deref(item.Properties.StartIPAddress),
			"endIpAddress":   pointers.Deref(item.Properties.EndIPAddress),
		}
	}

	return a
}
