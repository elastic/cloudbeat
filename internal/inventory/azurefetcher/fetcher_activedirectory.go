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

	"github.com/microsoftgraph/msgraph-sdk-go/models"

	"github.com/elastic/cloudbeat/internal/infra/clog"
	"github.com/elastic/cloudbeat/internal/inventory"
	"github.com/elastic/cloudbeat/internal/resources/utils/pointers"
)

type activedirectoryFetcher struct {
	logger   *clog.Logger
	provider activedirectoryProvider
}

type (
	activedirectoryProvider interface {
		ListServicePrincipals(ctx context.Context) ([]*models.ServicePrincipal, error)
	}
)

func newActiveDirectoryFetcher(logger *clog.Logger, provider activedirectoryProvider) inventory.AssetFetcher {
	return &activedirectoryFetcher{
		logger:   logger,
		provider: provider,
	}
}

func (f *activedirectoryFetcher) Fetch(ctx context.Context, assetChan chan<- inventory.AssetEvent) {
	f.fetchServicePrincipals(ctx, assetChan)
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
			[]string{pointers.Deref(item.GetId())},
			pointers.Deref(item.GetDisplayName()),
			inventory.WithRawAsset(
				item.GetBackingStore().Enumerate(),
			),
<<<<<<< HEAD
			inventory.WithCloud(inventory.AssetCloud{
				Provider: inventory.AzureCloudProvider,
				Account: inventory.AssetCloudAccount{
					Id: tenantId,
				},
				Service: &inventory.AssetCloudService{
					Name: "Azure",
				},
=======
			inventory.WithCloud(inventory.Cloud{
				Provider:    inventory.AzureCloudProvider,
				AccountID:   tenantId,
				ServiceName: "Azure Entra",
			}),
			inventory.WithTags(item.GetTags()),
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
			pickName(pointers.Deref(item.GetDisplayName()), pointers.Deref(item.GetId())),
			inventory.WithRawAsset(
				item.GetBackingStore().Enumerate(),
			),
			inventory.WithCloud(inventory.Cloud{
				Provider:    inventory.AzureCloudProvider,
				AccountID:   f.tenantID,
				ServiceName: "Azure Entra",
			}),
			inventory.WithUser(inventory.User{
				ID:   pointers.Deref(item.GetId()),
				Name: pointers.Deref(item.GetDisplayName()),
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
			pickName(pointers.Deref(item.GetDisplayName()), pointers.Deref(item.GetId())),
			inventory.WithRawAsset(
				item.GetBackingStore().Enumerate(),
			),
			inventory.WithCloud(inventory.Cloud{
				Provider:    inventory.AzureCloudProvider,
				AccountID:   f.tenantID,
				ServiceName: "Azure Entra",
			}),
			inventory.WithGroup(inventory.Group{
				ID:   pointers.Deref(item.GetId()),
				Name: pointers.Deref(item.GetDisplayName()),
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
			pickName(pointers.Deref(item.GetDisplayName()), pointers.Deref(item.GetId())),
			inventory.WithRawAsset(
				item.GetBackingStore().Enumerate(),
			),
			inventory.WithCloud(inventory.Cloud{
				Provider:    inventory.AzureCloudProvider,
				AccountID:   f.tenantID,
				ServiceName: "Azure Entra",
			}),
			inventory.WithUser(inventory.User{
				ID:   pointers.Deref(item.GetId()),
				Name: pointers.Deref(item.GetDisplayName()),
>>>>>>> 7e3234f1 ([Asset Inventory][Azure] Fix Azure service names (cloud.service.name) (#3466))
			}),
		)
	}
}
