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

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/sql/armsql"
	"github.com/elastic/elastic-agent-libs/logp"
	"github.com/samber/lo"

	"github.com/elastic/cloudbeat/resources/utils/ptrs"
)

type sqlAzureClientWrapper struct {
	AssetSQLEncryptionProtector        func(ctx context.Context, subID, resourceGroup, sqlServerName string, clientOptions *arm.ClientOptions, options *armsql.EncryptionProtectorsClientListByServerOptions) ([]armsql.EncryptionProtectorsClientListByServerResponse, error)
	AssetSQLBlobAuditingPolicies       func(ctx context.Context, subID, resourceGroup, sqlServerName string, clientOptions *arm.ClientOptions, options *armsql.ServerBlobAuditingPoliciesClientGetOptions) (armsql.ServerBlobAuditingPoliciesClientGetResponse, error)
	AssetSQLTransparentDataEncryptions func(ctx context.Context, subID, resourceGroup, sqlServerName, dbName string, clientOptions *arm.ClientOptions, options *armsql.TransparentDataEncryptionsClientListByDatabaseOptions) ([]armsql.TransparentDataEncryptionsClientListByDatabaseResponse, error)
	AssetSQLDatabases                  func(ctx context.Context, subID, resourceGroup, sqlServerName string, clientOptions *arm.ClientOptions, options *armsql.DatabasesClientListByServerOptions) ([]armsql.DatabasesClientListByServerResponse, error)
}

type SQLProviderAPI interface {
	ListSQLEncryptionProtector(ctx context.Context, subID, resourceGroup, sqlServerName string) ([]AzureAsset, error)
	ListSQLTransparentDataEncryptions(ctx context.Context, subID, resourceGroup, sqlServerName string) ([]AzureAsset, error)
	GetSQLBlobAuditingPolicies(ctx context.Context, subID, resourceGroup, sqlServerName string) ([]AzureAsset, error)
}

type sqlProvider struct {
	client *sqlAzureClientWrapper
	log    *logp.Logger //nolint:unused
}

func NewSQLProvider(log *logp.Logger, credentials azcore.TokenCredential) SQLProviderAPI {
	// We wrap the client, so we can mock it in tests
	wrapper := &sqlAzureClientWrapper{
		AssetSQLEncryptionProtector: func(ctx context.Context, subID, resourceGroup, sqlServerName string, clientOptions *arm.ClientOptions, options *armsql.EncryptionProtectorsClientListByServerOptions) ([]armsql.EncryptionProtectorsClientListByServerResponse, error) {
			cl, err := armsql.NewEncryptionProtectorsClient(subID, credentials, clientOptions)
			if err != nil {
				return nil, err
			}
			return readPager(ctx, cl.NewListByServerPager(resourceGroup, sqlServerName, options))
		},
		AssetSQLBlobAuditingPolicies: func(ctx context.Context, subID, resourceGroup, sqlServerName string, clientOptions *arm.ClientOptions, options *armsql.ServerBlobAuditingPoliciesClientGetOptions) (armsql.ServerBlobAuditingPoliciesClientGetResponse, error) {
			cl, err := armsql.NewServerBlobAuditingPoliciesClient(subID, credentials, clientOptions)
			if err != nil {
				return armsql.ServerBlobAuditingPoliciesClientGetResponse{}, err
			}
			return cl.Get(ctx, resourceGroup, sqlServerName, options)
		},
		AssetSQLTransparentDataEncryptions: func(ctx context.Context, subID, resourceGroup, sqlServerName, dbName string, clientOptions *arm.ClientOptions, options *armsql.TransparentDataEncryptionsClientListByDatabaseOptions) ([]armsql.TransparentDataEncryptionsClientListByDatabaseResponse, error) {
			cl, err := armsql.NewTransparentDataEncryptionsClient(subID, credentials, clientOptions)
			if err != nil {
				return nil, err
			}
			return readPager(ctx, cl.NewListByDatabasePager(resourceGroup, sqlServerName, dbName, options))
		},
		AssetSQLDatabases: func(ctx context.Context, subID, resourceGroup, sqlServerName string, clientOptions *arm.ClientOptions, options *armsql.DatabasesClientListByServerOptions) ([]armsql.DatabasesClientListByServerResponse, error) {
			cl, err := armsql.NewDatabasesClient(subID, credentials, clientOptions)
			if err != nil {
				return nil, err
			}
			return readPager(ctx, cl.NewListByServerPager(resourceGroup, sqlServerName, options))
		},
	}

	return &sqlProvider{
		log:    log,
		client: wrapper,
	}
}

func (p *sqlProvider) ListSQLEncryptionProtector(ctx context.Context, subID, resourceGroup, sqlServerName string) ([]AzureAsset, error) {
	encryptProtectors, err := p.client.AssetSQLEncryptionProtector(ctx, subID, resourceGroup, sqlServerName, nil, nil)
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

func (p *sqlProvider) GetSQLBlobAuditingPolicies(ctx context.Context, subID, resourceGroup, sqlServerName string) ([]AzureAsset, error) {
	policy, err := p.client.AssetSQLBlobAuditingPolicies(ctx, subID, resourceGroup, sqlServerName, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("problem on getting sql blob auditing policies (%w)", err)
	}

	if policy.Properties == nil {
		return nil, nil
	}

	return []AzureAsset{
		{

			Id:       ptrs.Deref(policy.ID),
			Name:     ptrs.Deref(policy.Name),
			Location: assetLocationGlobal,
			Properties: map[string]any{
				"state":                        string(ptrs.Deref(policy.Properties.State)),
				"isAzureMonitorTargetEnabled":  ptrs.Deref(policy.Properties.IsAzureMonitorTargetEnabled),
				"isDevopsAuditEnabled":         ptrs.Deref(policy.Properties.IsDevopsAuditEnabled),
				"isManagedIdentityInUse":       ptrs.Deref(policy.Properties.IsManagedIdentityInUse),
				"isStorageSecondaryKeyInUse":   ptrs.Deref(policy.Properties.IsStorageSecondaryKeyInUse),
				"queueDelayMs":                 ptrs.Deref(policy.Properties.QueueDelayMs),
				"retentionDays":                ptrs.Deref(policy.Properties.RetentionDays),
				"storageAccountAccessKey":      ptrs.Deref(policy.Properties.StorageAccountAccessKey),
				"storageAccountSubscriptionID": ptrs.Deref(policy.Properties.StorageAccountSubscriptionID),
				"storageEndpoint":              ptrs.Deref(policy.Properties.StorageEndpoint),

				"auditActionsAndGroups": lo.Map(policy.Properties.AuditActionsAndGroups, func(s *string, _ int) string {
					return ptrs.Deref(s)
				}),
			},
			ResourceGroup:  resourceGroup,
			SubscriptionId: subID,
			TenantId:       "",
			Type:           ptrs.Deref(policy.Type),
		},
	}, nil
}

func (p *sqlProvider) ListSQLTransparentDataEncryptions(ctx context.Context, subID, resourceGroup, sqlServerName string) ([]AzureAsset, error) {
	pagedDbs, err := p.client.AssetSQLDatabases(ctx, subID, resourceGroup, sqlServerName, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("problem on listing sql databases for sql server %s (%w)", sqlServerName, err)
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

		assetsByDb, err := p.listTransparentDataEncryptionsByDB(ctx, subID, resourceGroup, sqlServerName, ptrs.Deref(db.Name))
		if err != nil {
			errs = errors.Join(errs, err)
		}

		assets = append(assets, assetsByDb...)
	}

	return assets, errs
}

func (p *sqlProvider) listTransparentDataEncryptionsByDB(ctx context.Context, subID, resourceGroup, sqlServerName, dbName string) ([]AzureAsset, error) {
	pagedTdes, err := p.client.AssetSQLTransparentDataEncryptions(ctx, subID, resourceGroup, sqlServerName, dbName, nil, nil)
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
		Id:       ptrs.Deref(tde.ID),
		Name:     ptrs.Deref(tde.Name),
		Location: assetLocationGlobal,
		Properties: map[string]any{
			"databaseName": dbName,
			"state":        string(ptrs.Deref(tde.Properties.State)),
		},
		ResourceGroup:  resourceGroup,
		SubscriptionId: subID,
		TenantId:       "",
		Type:           ptrs.Deref(tde.Type),
	}
}

func convertEncryptionProtector(ep *armsql.EncryptionProtector, resourceGroup string, subID string) AzureAsset {
	return AzureAsset{
		Id:       ptrs.Deref(ep.ID),
		Name:     ptrs.Deref(ep.Name),
		Location: ptrs.Deref(ep.Location),
		Properties: map[string]any{
			"kind":                ptrs.Deref(ep.Kind),
			"serverKeyType":       string(ptrs.Deref(ep.Properties.ServerKeyType)),
			"autoRotationEnabled": ptrs.Deref(ep.Properties.AutoRotationEnabled),
			"serverKeyName":       ptrs.Deref(ep.Properties.ServerKeyName),
			"subregion":           ptrs.Deref(ep.Properties.Subregion),
			"thumbprint":          ptrs.Deref(ep.Properties.Thumbprint),
			"uri":                 ptrs.Deref(ep.Properties.URI),
		},
		ResourceGroup:  resourceGroup,
		SubscriptionId: subID,
		TenantId:       "",
		Type:           ptrs.Deref(ep.Type),
	}
}
