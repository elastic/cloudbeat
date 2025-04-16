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

package fetchers

import (
	"context"
	"fmt"
	"strings"

	"github.com/huandu/xstrings"
	"github.com/samber/lo"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/elastic/cloudbeat/internal/infra/clog"
	"github.com/elastic/cloudbeat/internal/resources/fetching"
	"github.com/elastic/cloudbeat/internal/resources/fetching/cycle"
	"github.com/elastic/cloudbeat/internal/resources/providers/gcplib/inventory"
)

type GcpAssetsFetcher struct {
	log        *clog.Logger
	resourceCh chan fetching.ResourceInfo
	provider   inventory.ServiceAPI
}

type GcpAsset struct {
	Type    string
	SubType string

	ExtendedAsset *inventory.ExtendedGcpAsset `json:"asset,omitempty"`
}

// GcpAssetTypes https://cloud.google.com/asset-inventory/docs/supported-asset-types
// map of types to asset types.
// sub-type is derived from asset type by using the first and last segments of the asset type name
// example: gcp-cloudkms-crypto-key

var reversedGcpAssetTypes = map[string]string{
	inventory.CrmProjectAssetType:            fetching.ProjectManagement,
	inventory.ApiKeysKeyAssetType:            fetching.KeyManagement,
	inventory.CloudKmsCryptoKeyAssetType:     fetching.KeyManagement,
	inventory.IamServiceAccountAssetType:     fetching.CloudIdentity,
	inventory.IamServiceAccountKeyAssetType:  fetching.CloudIdentity,
	inventory.BigqueryDatasetAssetType:       fetching.CloudDatabase,
	inventory.BigqueryTableAssetType:         fetching.CloudDatabase,
	inventory.SqlDatabaseInstanceAssetType:   fetching.CloudDatabase,
	inventory.StorageBucketAssetType:         fetching.CloudStorage,
	inventory.LogBucketAssetType:             fetching.CloudStorage,
	inventory.ComputeInstanceAssetType:       fetching.CloudCompute,
	inventory.ComputeFirewallAssetType:       fetching.CloudCompute,
	inventory.ComputeDiskAssetType:           fetching.CloudCompute,
	inventory.ComputeBackendServiceAssetType: fetching.CloudCompute,
	inventory.ComputeSubnetworkAssetType:     fetching.CloudCompute,
	inventory.DnsManagedZoneAssetType:        fetching.CloudDns,
	inventory.DataprocClusterAssetType:       fetching.DataProcessing,
}

func NewGcpAssetsFetcher(_ context.Context, log *clog.Logger, ch chan fetching.ResourceInfo, provider inventory.ServiceAPI) *GcpAssetsFetcher {
	return &GcpAssetsFetcher{
		log:        log,
		resourceCh: ch,
		provider:   provider,
	}
}

func (f *GcpAssetsFetcher) Fetch(ctx context.Context, cycleMetadata cycle.Metadata) error {
	f.log.Info("GcpAssetsFetcher.Fetch start")
	defer f.log.Info("GcpAssetsFetcher.Fetch done")
	defer f.provider.Clear()

	resultsCh := make(chan *inventory.ExtendedGcpAsset)
	go f.provider.ListAssetTypes(ctx, lo.Keys(reversedGcpAssetTypes), resultsCh)

	for asset := range resultsCh {
		select {
		case <-ctx.Done():
			f.log.Debugf("GcpAssetsFetcher.Fetch context done: %v", ctx.Err())
			return nil

		case f.resourceCh <- fetching.ResourceInfo{
			CycleMetadata: cycleMetadata,
			Resource: &GcpAsset{
				Type:          reversedGcpAssetTypes[asset.AssetType],
				SubType:       getGcpSubType(asset.AssetType),
				ExtendedAsset: asset,
			},
		}:
		}
	}
	return nil
}

func (f *GcpAssetsFetcher) Stop() {
	f.provider.Close()
}

func (r *GcpAsset) GetData() any {
	return r.ExtendedAsset.Asset
}

func (r *GcpAsset) GetIds() []string {
	return []string{r.ExtendedAsset.Name}
}

func (r *GcpAsset) GetMetadata() (fetching.ResourceMetadata, error) {
	var region string

	if r.ExtendedAsset.Resource != nil {
		region = r.ExtendedAsset.Resource.Location
	}

	return fetching.ResourceMetadata{
		ID:                   r.ExtendedAsset.Name,
		Type:                 r.Type,
		SubType:              r.SubType,
		Name:                 getAssetResourceName(r.ExtendedAsset),
		Region:               region,
		CloudAccountMetadata: *r.ExtendedAsset.CloudAccount,
	}, nil
}

func (r *GcpAsset) GetElasticCommonData() (map[string]any, error) {
	m := map[string]any{}

	if r.Type == fetching.CloudIdentity {
		m["user.effective.id"] = r.ExtendedAsset.Name
		m["user.effective.name"] = getAssetResourceName(r.ExtendedAsset)
	}

	if r.Type == fetching.CloudCompute && r.ExtendedAsset.AssetType == inventory.ComputeInstanceAssetType {
		fields := getAssetDataFields(r.ExtendedAsset)
		if fields == nil {
			return m, nil
		}
		nameField, ok := fields["name"]
		if ok {
			if name := nameField.GetStringValue(); name != "" {
				m["host.name"] = name
			}
		}
		hostnameField, ok := fields["hostname"]
		if ok {
			if hostname := hostnameField.GetStringValue(); hostname != "" {
				m["host.hostname"] = hostname
			}
		}
	}

	return m, nil
}

// try to retrieve the resource name from the asset data fields (name or displayName), in case it is not set
// get the last part of the asset name (https://cloud.google.com/apis/design/resource_names#resource_id)
func getAssetResourceName(asset *inventory.ExtendedGcpAsset) string {
	fields := getAssetDataFields(asset)
	if fields != nil {
		if name, exist := fields["displayName"]; exist && name.GetStringValue() != "" {
			return name.GetStringValue()
		}

		if name, exist := fields["name"]; exist && name.GetStringValue() != "" {
			return name.GetStringValue()
		}
	}
	parts := strings.Split(asset.Name, "/")
	return parts[len(parts)-1]
}

func getGcpSubType(assetType string) string {
	dotIndex := strings.Index(assetType, ".")
	slashIndex := strings.Index(assetType, "/")

	prefix := assetType[:dotIndex]
	suffix := assetType[slashIndex+1:]

	return strings.ToLower(fmt.Sprintf("gcp-%s-%s", prefix, xstrings.ToKebabCase(suffix)))
}

// getAssetDataFields tries to retrieve asset.resource.data fields if possible.
// Returns nil otherwise.
func getAssetDataFields(asset *inventory.ExtendedGcpAsset) map[string]*structpb.Value {
	resource := asset.GetResource()
	if resource == nil {
		return nil
	}
	data := resource.GetData()
	if data == nil {
		return nil
	}
	return data.GetFields()
}
