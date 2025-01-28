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

package gcpfetcher

import (
	"context"

<<<<<<< HEAD
	"github.com/elastic/elastic-agent-libs/logp"
=======
	"github.com/samber/lo"
	"google.golang.org/protobuf/types/known/structpb"
>>>>>>> 8963da3d (Use custom logger to downgrade canceled context errors to warnings (#2936))

	"github.com/elastic/cloudbeat/internal/infra/clog"
	"github.com/elastic/cloudbeat/internal/inventory"
	gcpinventory "github.com/elastic/cloudbeat/internal/resources/providers/gcplib/inventory"
)

type (
	assetsInventory struct {
		logger   *clog.Logger
		provider inventoryProvider
	}
	inventoryProvider interface {
		ListAllAssetTypesByName(ctx context.Context, assets []string) ([]*gcpinventory.ExtendedGcpAsset, error)
	}
	ResourcesClassification struct {
		assetType      string
		classification inventory.AssetClassification
	}
)

var ResourcesToFetch = []ResourcesClassification{
	{gcpinventory.CrmOrgAssetType, inventory.AssetClassificationGcpOrganization},
	{gcpinventory.CrmFolderAssetType, inventory.AssetClassificationGcpFolder},
	{gcpinventory.CrmProjectAssetType, inventory.AssetClassificationGcpProject},
	{gcpinventory.ComputeInstanceAssetType, inventory.AssetClassificationGcpInstance},
	{gcpinventory.ComputeFirewallAssetType, inventory.AssetClassificationGcpFirewall},
	{gcpinventory.StorageBucketAssetType, inventory.AssetClassificationGcpBucket},
	{gcpinventory.ComputeSubnetworkAssetType, inventory.AssetClassificationGcpSubnet},
	{gcpinventory.IamServiceAccountAssetType, inventory.AssetClassificationGcpServiceAccount},
	{gcpinventory.IamServiceAccountKeyAssetType, inventory.AssetClassificationGcpServiceAccountKey},
	{gcpinventory.GkeClusterAssetType, inventory.AssetClassificationGcpGkeCluster},
	{gcpinventory.ComputeForwardingRuleAssetType, inventory.AssetClassificationGcpForwardingRule},
	{gcpinventory.CloudFunctionAssetType, inventory.AssetClassificationGcpCloudFunction},
	{gcpinventory.CloudRunService, inventory.AssetClassificationGcpCloudRunService},
	{gcpinventory.IamRoleAssetType, inventory.AssetClassificationGcpIamRole},
}

func newAssetsInventoryFetcher(logger *clog.Logger, provider inventoryProvider) inventory.AssetFetcher {
	return &assetsInventory{
		logger:   logger,
		provider: provider,
	}
}

func (f *assetsInventory) Fetch(ctx context.Context, assetChan chan<- inventory.AssetEvent) {
	for _, r := range ResourcesToFetch {
		f.fetch(ctx, assetChan, r.assetType, r.classification)
	}
}

func (f *assetsInventory) fetch(ctx context.Context, assetChan chan<- inventory.AssetEvent, assetType string, classification inventory.AssetClassification) {
	f.logger.Infof("Fetching %s", assetType)
	defer f.logger.Infof("Fetching %s - Finished", assetType)

	gcpAssets, err := f.provider.ListAllAssetTypesByName(ctx, []string{assetType})
	if err != nil {
		f.logger.Errorf("Could not fetch %s: %v", assetType, err)
		return
	}

	for _, item := range gcpAssets {
		assetChan <- inventory.NewAssetEvent(
			classification,
			[]string{item.Name},
			item.Name,
			inventory.WithRawAsset(item),
			inventory.WithCloud(inventory.AssetCloud{
				Provider: inventory.GcpCloudProvider,
				Account: inventory.AssetCloudAccount{
					Id:   item.CloudAccount.AccountId,
					Name: item.CloudAccount.AccountName,
				},
				Organization: inventory.AssetCloudOrganization{
					Id:   item.CloudAccount.OrganisationId,
					Name: item.CloudAccount.OrganizationName,
				},
				Service: &inventory.AssetCloudService{
					Name: assetType,
				},
			}),
		)
	}
}
