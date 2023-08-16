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

	"cloud.google.com/go/asset/apiv1/assetpb"
	"github.com/elastic/elastic-agent-libs/logp"
	"github.com/huandu/xstrings"

	"github.com/elastic/cloudbeat/resources/fetching"
	"github.com/elastic/cloudbeat/resources/providers/gcplib/inventory"
)

type GcpAssetsFetcher struct {
	log        *logp.Logger
	resourceCh chan fetching.ResourceInfo
	provider   inventory.ServiceAPI
}

type GcpAsset struct {
	Type    string
	SubType string

	Asset *assetpb.Asset `json:"asset,omitempty"`
}

// GcpAssetTypes https://cloud.google.com/asset-inventory/docs/supported-asset-types
// map of types to asset types.
// sub-type is derived from asset type by using the first and last segments of the asset type name
// example: gcp-cloudkms-crypto-key
var GcpAssetTypes = map[string][]string{
	fetching.ProjectManagement: {
		"cloudresourcemanager.googleapis.com/Project",
	},
	fetching.KeyManagement: {
		"apikeys.googleapis.com/Key",
		"cloudkms.googleapis.com/CryptoKey",
	},
	fetching.CloudIdentity: {
		"iam.googleapis.com/ServiceAccount",
		"iam.googleapis.com/ServiceAccountKey",
	},
	fetching.CloudDatabase: {
		"bigquery.googleapis.com/Dataset",
		"bigquery.googleapis.com/Table",
		"sqladmin.googleapis.com/Instance",
	},
	fetching.CloudStorage: {
		"storage.googleapis.com/Bucket",
	},
	fetching.CloudCompute: {
		"compute.googleapis.com/Instance",
		"compute.googleapis.com/Firewall",
		"compute.googleapis.com/Disk",
		"compute.googleapis.com/Network",
		"compute.googleapis.com/RegionBackendService",
		"compute.googleapis.com/Subnetwork",
	},
	fetching.CloudDns: {
		"dns.googleapis.com/ManagedZone",
	},
	fetching.DataProcessing: {
		"dataproc.googleapis.com/Cluster",
	},
}

func NewGcpAssetsFetcher(_ context.Context, log *logp.Logger, ch chan fetching.ResourceInfo, provider inventory.ServiceAPI) *GcpAssetsFetcher {
	return &GcpAssetsFetcher{
		log:        log,
		resourceCh: ch,
		provider:   provider,
	}
}

func (f *GcpAssetsFetcher) Fetch(ctx context.Context, cMetadata fetching.CycleMetadata) error {
	f.log.Info("Starting GcpAssetsFetcher.Fetch")

	for typeName, assetTypes := range GcpAssetTypes {
		assets, err := f.provider.ListAllAssetTypesByName(assetTypes)
		if err != nil {
			f.log.Errorf("Failed to list assets for type %s: %s", typeName, err.Error())
			continue
		}

		for _, asset := range assets {
			select {
			case <-ctx.Done():
				f.log.Infof("GcpAssetsFetcher.Fetch context err: %s", ctx.Err().Error())
				return nil
			case f.resourceCh <- fetching.ResourceInfo{
				CycleMetadata: cMetadata,
				Resource: &GcpAsset{
					Type:    typeName,
					SubType: getGcpSubType(asset.AssetType),
					Asset:   asset,
				},
			}:
			}
		}
	}

	return nil
}

func (f *GcpAssetsFetcher) Stop() {
	f.provider.Close()
}

func (r *GcpAsset) GetData() interface{} {
	return r.Asset
}

func (r *GcpAsset) GetMetadata() (fetching.ResourceMetadata, error) {
	var region string

	if r.Asset.Resource != nil {
		region = r.Asset.Resource.Location
	}

	return fetching.ResourceMetadata{
		ID:      r.Asset.Name,
		Type:    r.Type,
		SubType: r.SubType,
		Name:    r.Asset.Name,
		Region:  region,
	}, nil
}

func (r *GcpAsset) GetElasticCommonData() any { return nil }

func getGcpSubType(assetType string) string {
	dotIndex := strings.Index(assetType, ".")
	slashIndex := strings.Index(assetType, "/")

	prefix := assetType[:dotIndex]
	suffix := assetType[slashIndex+1:]

	return strings.ToLower(fmt.Sprintf("gcp-%s-%s", prefix, xstrings.ToKebabCase(suffix)))
}
