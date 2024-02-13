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
	"testing"

	"cloud.google.com/go/asset/apiv1/assetpb"
	"github.com/samber/lo"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/elastic/cloudbeat/internal/resources/fetching"
	"github.com/elastic/cloudbeat/internal/resources/fetching/cycle"
	"github.com/elastic/cloudbeat/internal/resources/providers/gcplib/inventory"
	"github.com/elastic/cloudbeat/internal/resources/utils/testhelper"
)

type GcpAssetsFetcherTestSuite struct {
	suite.Suite

	resourceCh chan fetching.ResourceInfo
}

func TestGcpAssetsFetcherTestSuite(t *testing.T) {
	s := new(GcpAssetsFetcherTestSuite)

	suite.Run(t, s)
}

func (s *GcpAssetsFetcherTestSuite) SetupTest() {
	s.resourceCh = make(chan fetching.ResourceInfo, 50)
}

func (s *GcpAssetsFetcherTestSuite) TearDownTest() {
	close(s.resourceCh)
}

func (s *GcpAssetsFetcherTestSuite) TestFetcher_Fetch() {
	ctx := context.Background()
	mockInventoryService := &inventory.MockServiceAPI{}
	fetcher := GcpAssetsFetcher{
		log:        testhelper.NewLogger(s.T()),
		resourceCh: s.resourceCh,
		provider:   mockInventoryService,
	}

	mockInventoryService.EXPECT().ListAllAssetTypesByName(mock.Anything, mock.MatchedBy(func(_ []string) bool {
		return true
	})).Return(
		[]*inventory.ExtendedGcpAsset{
			{
				Ecs: &fetching.EcsGcp{
					Provider:         "gcp",
					ProjectId:        "prjId",
					ProjectName:      "prjName",
					OrganizationId:   "orgId",
					OrganizationName: "orgName",
				},
				Asset: &assetpb.Asset{
					Name: "a", AssetType: "iam.googleapis.com/ServiceAccount",
				},
			},
		}, nil,
	)

	err := fetcher.Fetch(ctx, cycle.Metadata{})
	s.Require().NoError(err)
	results := testhelper.CollectResources(s.resourceCh)

	s.Equal(len(GcpAssetTypes), len(results))

	lo.ForEach(results, func(r fetching.ResourceInfo, _ int) {
		ecs, err := r.Resource.GetElasticCommonData()
		s.Require().NoError(err)
		cloud := ecs["cloud"].(map[string]any)
		account := cloud["account"].(map[string]any)
		org := cloud["Organization"].(map[string]any)

		s.Equal("prjName", account["name"])
		s.Equal("prjId", account["id"])
		s.Equal("orgId", org["id"])
		s.Equal("orgName", org["name"])
		s.Equal("gcp", cloud["provider"])
	})
}
