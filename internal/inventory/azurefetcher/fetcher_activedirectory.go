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

package azurefetcher

import (
	"context"

	"github.com/elastic/elastic-agent-libs/logp"
	"github.com/microsoftgraph/msgraph-sdk-go/models"

	"github.com/elastic/cloudbeat/internal/inventory"
	"github.com/elastic/cloudbeat/internal/resources/utils/pointers"
)

type activedirectoryFetcher struct {
	logger   *logp.Logger
	provider activedirectoryProvider
}

type (
	activedirectoryProvider interface {
		ListServicePrincipals(ctx context.Context) ([]*models.ServicePrincipal, error)
		ListDirectoryRoles(context.Context) ([]*models.DirectoryRole, error)
		ListGroups(context.Context) ([]*models.Group, error)
		ListUsers(context.Context) ([]*models.User, error)
	}
)

func newActiveDirectoryFetcher(logger *logp.Logger, provider activedirectoryProvider) inventory.AssetFetcher {
	return &activedirectoryFetcher{
		logger:   logger,
		provider: provider,
	}
}

func (f *activedirectoryFetcher) Fetch(ctx context.Context, assetChan chan<- inventory.AssetEvent) {
	f.fetchServicePrincipals(ctx, assetChan)
	f.fetchDirectoryRoles(ctx, assetChan)
	f.fetchGroups(ctx, assetChan)
	f.fetchUsers(ctx, assetChan)
}

func (f *activedirectoryFetcher) fetchServicePrincipals(ctx context.Context, assetChan chan<- inventory.AssetEvent) {
	f.logger.Info("Fetching Service Principals")
	defer f.logger.Info("Fetching Service Principals - Finished")

	items, err := f.provider.ListServicePrincipals(ctx)
	if err != nil {
		f.logger.Errorf("Could not fetch Service Principals: %v", err)
	}

	for _, item := range items {
		var tenantId string
		if uuid := item.GetAppOwnerOrganizationId(); uuid != nil {
			tenantId = uuid.String()
		}
		assetChan <- inventory.NewAssetEvent(
			inventory.AssetClassificationAzureServicePrincipal,
			pointers.Deref(item.GetId()),
			pointers.Deref(item.GetDisplayName()),
			inventory.WithRawAsset(
				item.GetBackingStore().Enumerate(),
			),
			inventory.WithCloud(inventory.Cloud{
				Provider:    inventory.AzureCloudProvider,
				AccountID:   tenantId,
				ServiceName: "Azure",
			}),
		)
	}
}

func (f *activedirectoryFetcher) fetchDirectoryRoles(ctx context.Context, assetChan chan<- inventory.AssetEvent) {
	f.logger.Info("Fetching Directory Roles")
	defer f.logger.Info("Fetching Directory Roles - Finished")

	items, err := f.provider.ListDirectoryRoles(ctx)
	if err != nil {
		f.logger.Errorf("Could not fetch Directory Roles: %v", err)
	}

	for _, item := range items {
		assetChan <- inventory.NewAssetEvent(
			inventory.AssetClassificationAzureRoleDefinition,
			pointers.Deref(item.GetId()),
			pointers.Deref(item.GetDisplayName()),
			inventory.WithRawAsset(
				item.GetBackingStore().Enumerate(),
			),
			inventory.WithCloud(inventory.Cloud{
				Provider:    inventory.AzureCloudProvider,
				AccountID:   "",
				ServiceName: "Azure",
			}),
		)
	}
}

func (f *activedirectoryFetcher) fetchGroups(ctx context.Context, assetChan chan<- inventory.AssetEvent) {
	f.logger.Info("Fetching Groups")
	defer f.logger.Info("Fetching Groups - Finished")

	items, err := f.provider.ListGroups(ctx)
	if err != nil {
		f.logger.Errorf("Could not fetch Groups: %v", err)
	}

	for _, item := range items {
		assetChan <- inventory.NewAssetEvent(
			inventory.AssetClassificationAzureEntraGroup,
			pointers.Deref(item.GetId()),
			pointers.Deref(item.GetDisplayName()),
			inventory.WithRawAsset(
				item.GetBackingStore().Enumerate(),
			),
			inventory.WithCloud(inventory.Cloud{
				Provider:    inventory.AzureCloudProvider,
				AccountID:   "",
				ServiceName: "Azure",
			}),
		)
	}
}

func (f *activedirectoryFetcher) fetchUsers(ctx context.Context, assetChan chan<- inventory.AssetEvent) {
	f.logger.Info("Fetching Users")
	defer f.logger.Info("Fetching Users - Finished")

	items, err := f.provider.ListUsers(ctx)
	if err != nil {
		f.logger.Errorf("Could not fetch Users: %v", err)
	}

	for _, item := range items {
		assetChan <- inventory.NewAssetEvent(
			inventory.AssetClassificationAzureEntraUser,
			pointers.Deref(item.GetId()),
			pointers.Deref(item.GetDisplayName()),
			inventory.WithRawAsset(
				item.GetBackingStore().Enumerate(),
			),
			inventory.WithCloud(inventory.Cloud{
				Provider:    inventory.AzureCloudProvider,
				AccountID:   "",
				ServiceName: "Azure",
			}),
		)
	}
}
