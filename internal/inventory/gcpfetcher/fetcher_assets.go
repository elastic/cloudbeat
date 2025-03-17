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
	"google.golang.org/protobuf/types/known/structpb"

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
	{gcpinventory.ComputeNetworkAssetType, inventory.AssetClassificationGcpNetwork},
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
			findRelatedAssetIds(classification.Type, item),
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
		fields := item.GetResource().GetData().GetFields()
		if enricher, ok := assetEnrichers[item.AssetType]; ok {
			enrichers = append(enrichers, enricher(item, fields)...)
		}
	}

	return inventory.NewAssetEvent(
		classification,
		item.Name,
		item.Name,
		enrichers...,
	)
}

func findRelatedAssetIds(t inventory.AssetType, item *gcpinventory.ExtendedGcpAsset) []string {
	ids := []string{}
	ids = append(ids, item.Ancestors...)
	if item.Resource != nil {
		ids = append(ids, item.Resource.Parent)
	}

	ids = append(ids, findRelatedAssetIdsForType(t, item)...)

	ids = lo.Compact(ids)
	ids = lo.Uniq(ids)
	return ids
}

func findRelatedAssetIdsForType(t inventory.AssetType, item *gcpinventory.ExtendedGcpAsset) []string {
	ids := []string{}

	var fields map[string]*structpb.Value
	if item.Resource != nil && item.Resource.Data != nil {
		fields = item.GetResource().GetData().GetFields()
	}

	switch t {
	case inventory.AssetClassificationGcpInstance.Type:
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
	case inventory.AssetClassificationGcpFirewall.Type, inventory.AssetClassificationGcpSubnet.Type:
		ids = appendIfExists(ids, fields, "network")
	case inventory.AssetClassificationGcpProject.Type, inventory.AssetClassificationGcpBucket.Type:
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

func appendIfExists(slice []string, fields map[string]*structpb.Value, key string) []string {
	value, ok := fields[key]
	if !ok {
		return slice
	}
	return append(slice, value.GetStringValue())
}

func hasResourceData(item *gcpinventory.ExtendedGcpAsset) bool {
	return item.Resource != nil && item.Resource.Data != nil
}

func getAssetTags(item *gcpinventory.ExtendedGcpAsset) []string {
	if !hasResourceData(item) {
		return nil
	}

	tagsObj, ok := item.GetResource().GetData().GetFields()["tags"]
	if !ok {
		return nil
	}

	structValue := tagsObj.GetStructValue()
	if structValue == nil {
		return nil
	}

	items, ok := structValue.GetFields()["items"]
	if !ok {
		return nil
	}

	tagValues := items.GetListValue().GetValues()
	tags := make([]string, len(tagValues))
	for i, tag := range tagValues {
		tags[i] = tag.GetStringValue()
	}

	return tags
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

var assetEnrichers = map[string]func(item *gcpinventory.ExtendedGcpAsset, fields map[string]*structpb.Value) []inventory.AssetEnricher{
	gcpinventory.IamRoleAssetType:               noopEnricher,
	gcpinventory.CrmFolderAssetType:             noopEnricher,
	gcpinventory.CrmProjectAssetType:            noopEnricher,
	gcpinventory.StorageBucketAssetType:         noopEnricher,
	gcpinventory.IamServiceAccountKeyAssetType:  noopEnricher,
	gcpinventory.CloudRunService:                noopEnricher,
	gcpinventory.CrmOrgAssetType:                enrichOrganization,
	gcpinventory.ComputeInstanceAssetType:       enrichComputeInstance,
	gcpinventory.ComputeFirewallAssetType:       enrichFirewall,
	gcpinventory.ComputeSubnetworkAssetType:     enrichSubnetwork,
	gcpinventory.IamServiceAccountAssetType:     enrichServiceAccount,
	gcpinventory.GkeClusterAssetType:            enrichGkeCluster,
	gcpinventory.ComputeForwardingRuleAssetType: enrichForwardingRule,
	gcpinventory.CloudFunctionAssetType:         enrichCloudFunction,
	gcpinventory.ComputeNetworkAssetType:        enrichNetwork,
}

func enrichOrganization(_ *gcpinventory.ExtendedGcpAsset, fields map[string]*structpb.Value) []inventory.AssetEnricher {
	return []inventory.AssetEnricher{
		inventory.WithOrganization(inventory.Organization{
			Name: getStringValue("displayName", fields),
		}),
	}
}

func enrichComputeInstance(item *gcpinventory.ExtendedGcpAsset, fields map[string]*structpb.Value) []inventory.AssetEnricher {
	return []inventory.AssetEnricher{
		inventory.WithCloud(inventory.Cloud{
			// This will override the default Cloud fields, so we re-add the common ones
			Provider:         inventory.GcpCloudProvider,
			AccountID:        item.CloudAccount.AccountId,
			AccountName:      item.CloudAccount.AccountName,
			ProjectID:        item.CloudAccount.OrganisationId,
			ProjectName:      item.CloudAccount.OrganizationName,
			ServiceName:      item.AssetType,
			InstanceID:       getStringValue("id", fields),
			InstanceName:     getStringValue("name", fields),
			MachineType:      getStringValue("machineType", fields),
			AvailabilityZone: getStringValue("zone", fields),
		}),
		inventory.WithHost(inventory.Host{
			ID: getStringValue("id", fields),
		}),
	}
}

func enrichFirewall(_ *gcpinventory.ExtendedGcpAsset, fields map[string]*structpb.Value) []inventory.AssetEnricher {
	return []inventory.AssetEnricher{
		inventory.WithNetwork(inventory.Network{
			Name:      getStringValue("name", fields),
			Direction: getStringValue("direction", fields),
		}),
	}
}

func enrichSubnetwork(_ *gcpinventory.ExtendedGcpAsset, fields map[string]*structpb.Value) []inventory.AssetEnricher {
	return []inventory.AssetEnricher{
		inventory.WithNetwork(inventory.Network{
			Name: getStringValue("name", fields),
			Type: strings.ToLower(getStringValue("stackType", fields)),
		}),
	}
}

func enrichServiceAccount(_ *gcpinventory.ExtendedGcpAsset, fields map[string]*structpb.Value) []inventory.AssetEnricher {
	return []inventory.AssetEnricher{
		inventory.WithUser(inventory.User{
			Email: getStringValue("email", fields),
			Name:  getStringValue("displayName", fields),
		}),
	}
}

func enrichGkeCluster(_ *gcpinventory.ExtendedGcpAsset, fields map[string]*structpb.Value) []inventory.AssetEnricher {
	return []inventory.AssetEnricher{
		inventory.WithOrchestrator(inventory.Orchestrator{
			Type:        "kubernetes",
			ClusterName: getStringValue("name", fields),
			ClusterID:   getStringValue("id", fields),
		}),
	}
}

func enrichForwardingRule(_ *gcpinventory.ExtendedGcpAsset, fields map[string]*structpb.Value) []inventory.AssetEnricher {
	return []inventory.AssetEnricher{
		inventory.WithCloud(inventory.Cloud{
			Region: getStringValue("region", fields),
		}),
	}
}

func enrichCloudFunction(_ *gcpinventory.ExtendedGcpAsset, fields map[string]*structpb.Value) []inventory.AssetEnricher {
	var revision string
	if serviceConfig, ok := fields["serviceConfig"]; ok {
		serviceConfigFields := serviceConfig.GetStructValue().GetFields()
		revision = getStringValue("revision", serviceConfigFields)
	}
	return []inventory.AssetEnricher{
		inventory.WithURL(inventory.URL{
			Full: getStringValue("url", fields),
		}),
		inventory.WithFass(inventory.Fass{
			Name:    getStringValue("name", fields),
			Version: revision,
		}),
	}
}

func enrichNetwork(_ *gcpinventory.ExtendedGcpAsset, fields map[string]*structpb.Value) []inventory.AssetEnricher {
	return []inventory.AssetEnricher{
		inventory.WithNetwork(inventory.Network{
			Name: getStringValue("name", fields),
		}),
	}
}

func noopEnricher(_ *gcpinventory.ExtendedGcpAsset, _ map[string]*structpb.Value) []inventory.AssetEnricher {
	return []inventory.AssetEnricher{}
}

func getStringValue(key string, f map[string]*structpb.Value) string {
	if value, ok := f[key]; ok {
		return value.GetStringValue()
	}
	return ""
}
