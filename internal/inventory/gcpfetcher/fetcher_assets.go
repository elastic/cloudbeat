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

	"github.com/elastic/elastic-agent-libs/logp"
	"github.com/samber/lo"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/elastic/cloudbeat/internal/inventory"
	gcpinventory "github.com/elastic/cloudbeat/internal/resources/providers/gcplib/inventory"
)

type (
	assetsInventory struct {
		logger   *logp.Logger
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

func newAssetsInventoryFetcher(logger *logp.Logger, provider inventoryProvider) inventory.AssetFetcher {
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
			inventory.WithRelatedAssetIds(
				f.findRelatedAssetIds(classification.SubType, item),
			),
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

func (f *assetsInventory) findRelatedAssetIds(subType inventory.AssetSubType, item *gcpinventory.ExtendedGcpAsset) []string {
	ids := []string{}
	ids = append(ids, item.Ancestors...)
	if item.Resource != nil {
		ids = append(ids, item.Resource.Parent)
	}

	ids = append(ids, f.findRelatedAssetIdsForSubType(subType, item)...)

	ids = lo.Compact(ids)
	ids = lo.Uniq(ids)
	return ids
}

func (f *assetsInventory) findRelatedAssetIdsForSubType(subType inventory.AssetSubType, item *gcpinventory.ExtendedGcpAsset) []string {
	ids := []string{}

	var fields map[string]*structpb.Value
	if item.Resource != nil && item.Resource.Data != nil {
		fields = item.GetResource().GetData().GetFields()
	}

	switch subType {
	case inventory.SubTypeGcpInstance:
		if v, ok := fields["networkInterfaces"]; ok {
			for _, networkInterface := range v.GetListValue().GetValues() {
				networkInterfaceFields := networkInterface.GetStructValue().GetFields()
				ids = appendIfExists(ids, networkInterfaceFields, "network")
				ids = appendIfExists(ids, networkInterfaceFields, "subnetwork")
			}
		}
		if v, ok := fields["serviceAccounts"]; ok {
			for _, serviceAccount := range v.GetListValue().GetValues() {
				serviceAccountFields := serviceAccount.GetStructValue().GetFields()
				ids = appendIfExists(ids, serviceAccountFields, "email")
			}
		}
		if v, ok := fields["disks"]; ok {
			for _, disk := range v.GetListValue().GetValues() {
				diskFields := disk.GetStructValue().GetFields()
				ids = appendIfExists(ids, diskFields, "source")
			}
		}
		ids = appendIfExists(ids, fields, "machineType")
		ids = appendIfExists(ids, fields, "zone")
	case inventory.SubTypeGcpFirewall, inventory.SubTypeGcpSubnet:
		ids = appendIfExists(ids, fields, "network")
	case inventory.SubTypeGcpProject, inventory.SubTypeGcpBucket:
		if item.IamPolicy == nil {
			break
		}
		for _, binding := range item.IamPolicy.Bindings {
			ids = append(ids, binding.Role)
			ids = append(ids, binding.Members...)
		}
	default:
		f.logger.Warnf("cannot find related asset IDs for unsupported sub-type %q", subType)
		return ids
	}

	return ids
}

func appendIfExists(slice []string, fields map[string]*structpb.Value, key string) []string {
	value, ok := fields[key]
	if !ok {
		return slice
	}
	return append(slice, value.GetStringValue())
}
