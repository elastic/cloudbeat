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

	"cloud.google.com/go/asset/apiv1/assetpb"
	"github.com/elastic/elastic-agent-libs/logp"

	"github.com/elastic/cloudbeat/resources/fetching"
	"github.com/elastic/cloudbeat/resources/providers/gcplib/inventory"
)

type GcpAssetsFetcher struct {
	log        *logp.Logger
	resourceCh chan fetching.ResourceInfo
	provider   inventory.InventoryService
}

type GcpAsset struct {
	Type    string
	SubType string

	Asset *assetpb.Asset `json:"asset,omitempty"`
}

var GcpAssetTypes = map[string]map[string][]string{
	fetching.KeyManagement: {
		"gcp-kms": {"cloudkms.googleapis.com/CryptoKey"},
	},
	fetching.CloudIdentity: {
		"gcp-iam": {"iam.googleapis.com/ServiceAccount"},
	},
	fetching.CloudDatabase: {
		"gcp-bq-dataset": {"bigquery.googleapis.com/Dataset"},
		"gcp-bq-table":   {"bigquery.googleapis.com/Table"},
	},
	fetching.CloudStorage: {
		"gcp-gcs": {"storage.googleapis.com/Bucket"},
	},
}

func NewGcpAssetsFetcher(_ context.Context, log *logp.Logger, ch chan fetching.ResourceInfo, provider inventory.InventoryService) *GcpAssetsFetcher {
	return &GcpAssetsFetcher{
		log:        log,
		resourceCh: ch,
		provider:   provider,
	}
}

func (f *GcpAssetsFetcher) Fetch(ctx context.Context, cMetadata fetching.CycleMetadata) error {
	f.log.Info("Starting GcpAssetsFetcher.Fetch")

	for typeName, subtypes := range GcpAssetTypes {
		for subTypeName, assetTypes := range subtypes {
			assets, err := f.provider.ListAllAssetTypesByName(assetTypes)
			if err != nil {
				f.log.Errorf("Failed to list assets for type %s: %s", typeName, err)
				continue
			}

			for _, asset := range assets {
				f.resourceCh <- fetching.ResourceInfo{
					CycleMetadata: cMetadata,
					Resource: &GcpAsset{
						Type:    typeName,
						SubType: subTypeName,
						Asset:   asset,
					},
				}
			}
		}
	}

	return nil
}

func (f *GcpAssetsFetcher) Stop() {
	f.provider.Close()
}

func (r *GcpAsset) GetData() interface{} {
	return r
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
