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
	"testing"

	"cloud.google.com/go/asset/apiv1/assetpb"
	"github.com/elastic/cloudbeat/internal/ecs"
	"github.com/elastic/cloudbeat/internal/infra/clog"
	"github.com/elastic/cloudbeat/internal/inventory"
	"github.com/elastic/cloudbeat/internal/inventory/testutil"
	"github.com/elastic/cloudbeat/internal/resources/fetching"
	gcpinventory "github.com/elastic/cloudbeat/internal/resources/providers/gcplib/inventory"
	"github.com/samber/lo"
	"github.com/stretchr/testify/mock"
)

func TestAccountFetcher_Fetch_Assets(t *testing.T) {
	logger := clog.NewLogger("gcpfetcher_test")
	assets := []*gcpinventory.ExtendedGcpAsset{
		{
			Asset: &assetpb.Asset{
				Name: "/projects/<project UUID>/some_resource", // name is the ID
			},
			CloudAccount: &fetching.CloudAccountMetadata{
				AccountId:        "<project UUID>",
				AccountName:      "<project name>",
				OrganisationId:   "<org UUID>",
				OrganizationName: "<org name>",
			},
		},
	}

	expected := lo.Map(ResourcesToFetch, func(r ResourcesClassification, _ int) inventory.AssetEvent {
		return inventory.NewAssetEvent(
			r.classification,
			"/projects/<project UUID>/some_resource",
			"/projects/<project UUID>/some_resource",
			inventory.WithRawAsset(assets[0]),
			inventory.WithRelatedAssetIds([]string{}),
			inventory.WithCloud(ecs.Cloud{
				Provider:    inventory.GcpCloudProvider,
				AccountID:   "<project UUID>",
				AccountName: "<project name>",
				ProjectID:   "<org UUID>",
				ProjectName: "<org name>",
				ServiceName: r.assetType,
			}),
		)
	})

	provider := newMockInventoryProvider(t)
	provider.EXPECT().ListAllAssetTypesByName(mock.Anything, mock.AnythingOfType("[]string")).Return(assets, nil)
	fetcher := newAssetsInventoryFetcher(logger, provider)
	testutil.CollectResourcesAndMatch(t, fetcher, expected)
}
