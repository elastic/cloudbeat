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
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/samber/lo"

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
		assetChan <- getAssetEvent(classification, item)
	}
}

func getAssetEvent(classification inventory.AssetClassification, item *gcpinventory.ExtendedGcpAsset) inventory.AssetEvent {
	// Common enrichers
	enrichers := []inventory.AssetEnricher{
		inventory.WithRawAsset(item),
		inventory.WithLabels(getAssetLabels(item)),
		inventory.WithTags(getAssetTags(item)),
		inventory.WithRelatedAssetIds(
			findRelatedAssetIds(item),
		),
		// Any asset type enrichers also setting Cloud fields will need to re-add these fields below
		inventory.WithCloud(inventory.Cloud{
			Provider:    inventory.GcpCloudProvider,
			AccountID:   item.CloudAccount.AccountId,
			AccountName: item.CloudAccount.AccountName,
			ProjectID:   item.CloudAccount.OrganisationId,
			ProjectName: item.CloudAccount.OrganizationName,
			ServiceName: item.AssetType,
		}),
	}

	// Asset type specific enrichers
	if hasResourceData(item) {
		if enricher, ok := assetEnrichers[item.AssetType]; ok {
			enrichers = append(enrichers, enricher(item, item.GetResource().GetData().AsMap())...)
		}
	}

	return inventory.NewAssetEvent(
		classification,
		item.Name,
		item.Name,
		enrichers...,
	)
}

func findRelatedAssetIds(item *gcpinventory.ExtendedGcpAsset) []string {
	ids := []string{}
	ids = append(ids, item.Ancestors...)
	if item.Resource != nil {
		ids = append(ids, item.Resource.Parent)
	}
	ids = append(ids, findRelatedAssetIdsForType(item)...)
	ids = lo.Compact(ids)
	ids = lo.Uniq(ids)
	return ids
}

func findRelatedAssetIdsForType(item *gcpinventory.ExtendedGcpAsset) []string {
	ids := []string{}
	var pb map[string]any
	if hasResourceData(item) {
		pb = item.GetResource().GetData().AsMap()
	}

	switch item.AssetType {
	case gcpinventory.ComputeInstanceAssetType:
		ids = append(ids, values([]string{"networkInterfaces", "network"}, pb)...)
		ids = append(ids, values([]string{"networkInterfaces", "subnetwork"}, pb)...)
		ids = append(ids, values([]string{"serviceAccounts", "email"}, pb)...)
		ids = append(ids, values([]string{"disks", "source"}, pb)...)
		ids = append(ids, values([]string{"machineType"}, pb)...)
		ids = append(ids, values([]string{"zone"}, pb)...)

	case gcpinventory.ComputeFirewallAssetType, gcpinventory.ComputeSubnetworkAssetType:
		ids = append(ids, values([]string{"network"}, pb)...)
	case gcpinventory.CrmProjectAssetType, gcpinventory.StorageBucketAssetType:
		if item.IamPolicy == nil {
			break
		}
		for _, binding := range item.IamPolicy.Bindings {
			ids = append(ids, binding.Role)
			ids = append(ids, binding.Members...)
		}
	default:
		return ids
	}

	return ids
}

func hasResourceData(item *gcpinventory.ExtendedGcpAsset) bool {
	return item.Resource != nil && item.Resource.Data != nil
}

func getAssetTags(item *gcpinventory.ExtendedGcpAsset) []string {
	if !hasResourceData(item) {
		return nil
	}

	return values([]string{"tags", "items"}, item.GetResource().GetData().AsMap())
}

func getAssetLabels(item *gcpinventory.ExtendedGcpAsset) map[string]string {
	if !hasResourceData(item) {
		return nil
	}

	labels, ok := item.GetResource().GetData().GetFields()["labels"]
	if !ok {
		return nil
	}

	labelsMap := make(map[string]string)
	if err := mapstructure.Decode(labels.GetStructValue().AsMap(), &labelsMap); err != nil {
		return nil
	}

	return labelsMap
}

var assetEnrichers = map[string]func(item *gcpinventory.ExtendedGcpAsset, pb map[string]any) []inventory.AssetEnricher{
	gcpinventory.IamRoleAssetType:               noopEnricher,
	gcpinventory.CrmFolderAssetType:             noopEnricher,
	gcpinventory.CrmProjectAssetType:            noopEnricher,
	gcpinventory.StorageBucketAssetType:         noopEnricher,
	gcpinventory.IamServiceAccountKeyAssetType:  noopEnricher,
	gcpinventory.CloudRunService:                enrichCloudRunService,
	gcpinventory.CrmOrgAssetType:                enrichOrganization,
	gcpinventory.ComputeInstanceAssetType:       enrichComputeInstance,
	gcpinventory.ComputeFirewallAssetType:       enrichFirewall,
	gcpinventory.ComputeSubnetworkAssetType:     enrichSubnetwork,
	gcpinventory.IamServiceAccountAssetType:     enrichServiceAccount,
	gcpinventory.GkeClusterAssetType:            enrichGkeCluster,
	gcpinventory.ComputeForwardingRuleAssetType: enrichForwardingRule,
	gcpinventory.CloudFunctionAssetType:         enrichCloudFunction,
}

func enrichOrganization(_ *gcpinventory.ExtendedGcpAsset, pb map[string]any) []inventory.AssetEnricher {
	return []inventory.AssetEnricher{
		inventory.WithOrganization(inventory.Organization{
			Name: lo.FirstOrEmpty(values([]string{"displayName"}, pb)),
		}),
	}
}

func enrichComputeInstance(item *gcpinventory.ExtendedGcpAsset, pb map[string]any) []inventory.AssetEnricher {
	return []inventory.AssetEnricher{
		inventory.WithCloud(inventory.Cloud{
			// This will override the default Cloud pb, so we re-add the common ones
			Provider:         inventory.GcpCloudProvider,
			AccountID:        item.CloudAccount.AccountId,
			AccountName:      item.CloudAccount.AccountName,
			ProjectID:        item.CloudAccount.OrganisationId,
			ProjectName:      item.CloudAccount.OrganizationName,
			ServiceName:      item.AssetType,
			InstanceID:       lo.FirstOrEmpty(values([]string{"id"}, pb)),
			InstanceName:     lo.FirstOrEmpty(values([]string{"name"}, pb)),
			MachineType:      lo.FirstOrEmpty(values([]string{"machineType"}, pb)),
			AvailabilityZone: lo.FirstOrEmpty(values([]string{"zone"}, pb)),
		}),
		inventory.WithHost(inventory.Host{
			ID: lo.FirstOrEmpty(values([]string{"id"}, pb)),
		}),
		inventory.WithNetwork(inventory.Network{
			Name: values([]string{"networkInterfaces", "name"}, pb),
		}),
	}
}

func enrichFirewall(_ *gcpinventory.ExtendedGcpAsset, pb map[string]any) []inventory.AssetEnricher {
	return []inventory.AssetEnricher{
		inventory.WithNetwork(inventory.Network{
			Name:      values([]string{"name"}, pb),
			Direction: lo.FirstOrEmpty(values([]string{"direction"}, pb)),
		}),
	}
}

func enrichSubnetwork(_ *gcpinventory.ExtendedGcpAsset, pb map[string]any) []inventory.AssetEnricher {
	return []inventory.AssetEnricher{
		inventory.WithNetwork(inventory.Network{
			Name: values([]string{"name"}, pb),
			Type: strings.ToLower(lo.FirstOrEmpty(values([]string{"stackType"}, pb))),
		}),
	}
}

func enrichServiceAccount(_ *gcpinventory.ExtendedGcpAsset, pb map[string]any) []inventory.AssetEnricher {
	return []inventory.AssetEnricher{
		inventory.WithUser(inventory.User{
			Email: lo.FirstOrEmpty(values([]string{"email"}, pb)),
			Name:  lo.FirstOrEmpty(values([]string{"displayName"}, pb)),
		}),
	}
}

func enrichGkeCluster(_ *gcpinventory.ExtendedGcpAsset, pb map[string]any) []inventory.AssetEnricher {
	return []inventory.AssetEnricher{
		inventory.WithOrchestrator(inventory.Orchestrator{
			Type:        "kubernetes",
			ClusterName: lo.FirstOrEmpty(values([]string{"name"}, pb)),
			ClusterID:   lo.FirstOrEmpty(values([]string{"id"}, pb)),
		}),
	}
}

func enrichForwardingRule(item *gcpinventory.ExtendedGcpAsset, pb map[string]any) []inventory.AssetEnricher {
	return []inventory.AssetEnricher{
		inventory.WithCloud(inventory.Cloud{
			// This will override the default Cloud pb, so we re-add the common ones
			Provider:    inventory.GcpCloudProvider,
			AccountID:   item.CloudAccount.AccountId,
			AccountName: item.CloudAccount.AccountName,
			ProjectID:   item.CloudAccount.OrganisationId,
			ProjectName: item.CloudAccount.OrganizationName,
			ServiceName: item.AssetType,
			Region:      lo.FirstOrEmpty(values([]string{"region"}, pb)),
		}),
	}
}

func enrichCloudFunction(_ *gcpinventory.ExtendedGcpAsset, pb map[string]any) []inventory.AssetEnricher {
	return []inventory.AssetEnricher{
		inventory.WithURL(inventory.URL{
			Full: lo.FirstOrEmpty(values([]string{"url"}, pb)),
		}),
		inventory.WithFass(inventory.Fass{
			Name:    lo.FirstOrEmpty(values([]string{"name"}, pb)),
			Version: lo.FirstOrEmpty(values([]string{"serviceConfig", "revision"}, pb)),
		}),
	}
}

func enrichCloudRunService(_ *gcpinventory.ExtendedGcpAsset, pb map[string]any) []inventory.AssetEnricher {
	return []inventory.AssetEnricher{
		inventory.WithContainer(inventory.Container{
			Name:      values([]string{"spec", "template", "spec", "containers", "name"}, pb),
			ImageName: values([]string{"spec", "template", "spec", "containers", "image"}, pb),
		}),
	}
}

func noopEnricher(_ *gcpinventory.ExtendedGcpAsset, _ map[string]any) []inventory.AssetEnricher {
	return []inventory.AssetEnricher{}
}

// returns string values of keys in a map/array
func values(keys []string, current any) []string {
	if len(keys) == 0 {
		switch v := current.(type) {
		case string:
			return []string{v}
		case []any:
			var results []string
			for _, item := range v {
				results = append(results, values(keys, item)...)
			}
			return results
		}
		return []string{}
	}

	switch v := current.(type) {
	case map[string]any:
		return values(keys[1:], v[keys[0]])
	case []any:
		var results []string
		for _, item := range v {
			results = append(results, values(keys, item)...)
		}
		return results
	}

	return []string{}
}
