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
		assetChan <- getAssetEvent(*f.logger, classification, item)
	}
}

func getAssetEvent(log clog.Logger, classification inventory.AssetClassification, item *gcpinventory.ExtendedGcpAsset) inventory.AssetEvent {
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
			enrichers = append(enrichers, enricher(log, item, item.GetResource().GetData().AsMap())...)
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
		gcpComputeInstance := gcpComputeInstance{}
		if err := mapstructure.Decode(pb, &gcpComputeInstance); err == nil {
			for _, ni := range gcpComputeInstance.NetworkInterfaces {
				ids = append(ids, ni.Network)
				ids = append(ids, ni.Subnetwork)
			}
			for _, sa := range gcpComputeInstance.ServiceAccounts {
				ids = append(ids, sa.Email)
			}
			for _, disk := range gcpComputeInstance.Disks {
				ids = append(ids, disk.Source)
			}
			ids = append(ids, gcpComputeInstance.MachineType)
			ids = append(ids, gcpComputeInstance.Zone)

		}
	case gcpinventory.ComputeFirewallAssetType:
		gcpFirewall := gcpFirewall{}
		if err := mapstructure.Decode(pb, &gcpFirewall); err == nil {
			ids = append(ids, gcpFirewall.Network)
		}
	case gcpinventory.ComputeSubnetworkAssetType:
		gcpSubnetwork := gcpSubnetwork{}
		if err := mapstructure.Decode(pb, &gcpSubnetwork); err == nil {
			ids = append(ids, gcpSubnetwork.Network)
		}

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

type gcpResource struct {
	Tags struct {
		Items []string `mapstructure:"items"`
	}
}

func getAssetTags(item *gcpinventory.ExtendedGcpAsset) []string {
	if !hasResourceData(item) {
		return nil
	}
	var resource gcpResource
	if err := mapstructure.Decode(item.GetResource().GetData().AsMap(), &resource); err != nil {
		return nil
	}
	return resource.Tags.Items
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

var assetEnrichers = map[string]func(log clog.Logger, item *gcpinventory.ExtendedGcpAsset, pb map[string]any) []inventory.AssetEnricher{
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

type gcpOrganization struct {
	Name string `mapstructure:"displayName"`
}

func enrichOrganization(log clog.Logger, _ *gcpinventory.ExtendedGcpAsset, pb map[string]any) []inventory.AssetEnricher {
	gcpOrganization := gcpOrganization{}
	if err := mapstructure.Decode(pb, &gcpOrganization); err != nil {
		log.Errorf("Failed to decode GCP organization asset: %+v, err: %v", pb, err)
		return []inventory.AssetEnricher{}
	}
	return []inventory.AssetEnricher{
		inventory.WithOrganization(inventory.Organization{
			Name: gcpOrganization.Name,
		}),
	}
}

type gcpComputeInstance struct {
	Name              string `mapstructure:"name"`
	ID                string `mapstructure:"id"`
	MachineType       string `mapstructure:"machineType"`
	Zone              string `mapstructure:"zone"`
	NetworkInterfaces []struct {
		Network    string `mapstructure:"network"`
		Subnetwork string `mapstructure:"subnetwork"`
	} `mapstructure:"networkInterfaces"`
	ServiceAccounts []struct {
		Email string `mapstructure:"email"`
	} `mapstructure:"serviceAccounts"`
	Disks []struct {
		Source string `mapstructure:"source"`
	} `mapstructure:"disks"`
}

func enrichComputeInstance(log clog.Logger, item *gcpinventory.ExtendedGcpAsset, pb map[string]any) []inventory.AssetEnricher {
	gcpComputeInstance := gcpComputeInstance{}
	if err := mapstructure.Decode(pb, &gcpComputeInstance); err != nil {
		log.Errorf("Failed to decode GCP compute instance asset: %+v, err: %v", pb, err)
		return []inventory.AssetEnricher{}
	}

	return []inventory.AssetEnricher{
		inventory.WithCloud(inventory.Cloud{
			// This will override the default Cloud pb, so we re-add the common ones
			Provider:         inventory.GcpCloudProvider,
			AccountID:        item.CloudAccount.AccountId,
			AccountName:      item.CloudAccount.AccountName,
			ProjectID:        item.CloudAccount.OrganisationId,
			ProjectName:      item.CloudAccount.OrganizationName,
			ServiceName:      item.AssetType,
			InstanceID:       gcpComputeInstance.ID,
			InstanceName:     gcpComputeInstance.Name,
			MachineType:      gcpComputeInstance.MachineType,
			AvailabilityZone: gcpComputeInstance.Zone,
		}),
		inventory.WithHost(inventory.Host{
			ID: gcpComputeInstance.ID,
		}),
	}
}

type gcpFirewall struct {
	Name      string `mapstructure:"name"`
	Direction string `mapstructure:"direction"`
	Network   string `mapstructure:"network"`
}

func enrichFirewall(log clog.Logger, _ *gcpinventory.ExtendedGcpAsset, pb map[string]any) []inventory.AssetEnricher {
	gcpFirewall := gcpFirewall{}
	if err := mapstructure.Decode(pb, &gcpFirewall); err != nil {
		log.Errorf("Failed to decode GCP firewall asset: %+v, err: %v", pb, err)
		return []inventory.AssetEnricher{}
	}

	return []inventory.AssetEnricher{
		inventory.WithNetwork(inventory.Network{
			Name:      gcpFirewall.Name,
			Direction: gcpFirewall.Direction,
		}),
	}
}

type gcpSubnetwork struct {
	Name      string `mapstructure:"name"`
	StackType string `mapstructure:"stackType"`
	Network   string `mapstructure:"network"`
}

func enrichSubnetwork(log clog.Logger, _ *gcpinventory.ExtendedGcpAsset, pb map[string]any) []inventory.AssetEnricher {
	gcpSubnetwork := gcpSubnetwork{}
	if err := mapstructure.Decode(pb, &gcpSubnetwork); err != nil {
		log.Errorf("Failed to decode GCP subnetwork asset: %+v, err: %v", pb, err)
		return []inventory.AssetEnricher{}
	}

	return []inventory.AssetEnricher{
		inventory.WithNetwork(inventory.Network{
			Name: gcpSubnetwork.Name,
			Type: strings.ToLower(gcpSubnetwork.StackType),
		}),
	}
}

type gcpServiceAccount struct {
	Email       string `mapstructure:"email"`
	DisplayName string `mapstructure:"displayName"`
}

func enrichServiceAccount(log clog.Logger, _ *gcpinventory.ExtendedGcpAsset, pb map[string]any) []inventory.AssetEnricher {
	gcpServiceAccount := gcpServiceAccount{}
	if err := mapstructure.Decode(pb, &gcpServiceAccount); err != nil {
		log.Errorf("Failed to decode GCP service account asset: %+v, err: %v", pb, err)
		return []inventory.AssetEnricher{}
	}

	return []inventory.AssetEnricher{
		inventory.WithUser(inventory.User{
			Email: gcpServiceAccount.Email,
			Name:  gcpServiceAccount.DisplayName,
		}),
	}
}

type gkeCluster struct {
	Name string `mapstructure:"name"`
	ID   string `mapstructure:"id"`
}

func enrichGkeCluster(log clog.Logger, _ *gcpinventory.ExtendedGcpAsset, pb map[string]any) []inventory.AssetEnricher {
	gkeCluster := gkeCluster{}
	if err := mapstructure.Decode(pb, &gkeCluster); err != nil {
		log.Errorf("Failed to decode GKE cluster asset: %+v, err: %v", pb, err)
		return []inventory.AssetEnricher{}
	}
	return []inventory.AssetEnricher{
		inventory.WithOrchestrator(inventory.Orchestrator{
			Type:        "kubernetes",
			ClusterName: gkeCluster.Name,
			ClusterID:   gkeCluster.ID,
		}),
	}
}

type gcpForwardingRule struct {
	Region string `mapstructure:"region"`
}

func enrichForwardingRule(log clog.Logger, item *gcpinventory.ExtendedGcpAsset, pb map[string]any) []inventory.AssetEnricher {
	cloud := inventory.Cloud{
		Provider:    inventory.GcpCloudProvider,
		AccountID:   item.CloudAccount.AccountId,
		AccountName: item.CloudAccount.AccountName,
		ProjectID:   item.CloudAccount.OrganisationId,
		ProjectName: item.CloudAccount.OrganizationName,
		ServiceName: item.AssetType,
	}

	gcpForwardingRule := gcpForwardingRule{}
	if err := mapstructure.Decode(pb, &gcpForwardingRule); err != nil {
		log.Errorf("Failed to decode GCP forwarding rule asset: %+v, err: %v", pb, err)
		return []inventory.AssetEnricher{
			inventory.WithCloud(cloud),
		}
	}
	cloud.Region = gcpForwardingRule.Region
	return []inventory.AssetEnricher{
		inventory.WithCloud(cloud),
	}
}

type gcpCloudFunction struct {
	Name          string `mapstructure:"name"`
	URL           string `mapstructure:"url"`
	ServiceConfig struct {
		Revision string `mapstructure:"revision"`
	} `mapstructure:"serviceConfig"`
}

func enrichCloudFunction(log clog.Logger, _ *gcpinventory.ExtendedGcpAsset, pb map[string]any) []inventory.AssetEnricher {
	gcpCloudFunction := gcpCloudFunction{}
	if err := mapstructure.Decode(pb, &gcpCloudFunction); err != nil {
		log.Errorf("Failed to decode GCP cloud function asset: %+v, err: %v", pb, err)
		return []inventory.AssetEnricher{}
	}
	return []inventory.AssetEnricher{
		inventory.WithURL(inventory.URL{
			Full: gcpCloudFunction.URL,
		}),
		inventory.WithFass(inventory.Fass{
			Name:    gcpCloudFunction.Name,
			Version: gcpCloudFunction.ServiceConfig.Revision,
		}),
	}
}

type gcpNetwork struct {
	Name string `mapstructure:"name"`
}

func enrichNetwork(log clog.Logger, _ *gcpinventory.ExtendedGcpAsset, pb map[string]any) []inventory.AssetEnricher {
	gcpNetwork := gcpNetwork{}
	if err := mapstructure.Decode(pb, &gcpNetwork); err != nil {
		log.Errorf("Failed to decode GCP network asset: %+v, err: %v", pb, err)
		return []inventory.AssetEnricher{}
	}
	if gcpNetwork.Name == "" {
		return []inventory.AssetEnricher{}
	}

	return []inventory.AssetEnricher{
		inventory.WithNetwork(inventory.Network{
			Name: gcpNetwork.Name,
		}),
	}
}

func noopEnricher(_ clog.Logger, _ *gcpinventory.ExtendedGcpAsset, _ map[string]any) []inventory.AssetEnricher {
	return []inventory.AssetEnricher{}
}
