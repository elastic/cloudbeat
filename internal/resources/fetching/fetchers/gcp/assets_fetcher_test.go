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
	"google.golang.org/protobuf/types/known/structpb"

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
				CloudAccount: &fetching.CloudAccountMetadata{
					AccountId:        "prjId",
					AccountName:      "prjName",
					OrganisationId:   "orgId",
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
		metadata, err := r.Resource.GetMetadata()
		s.Require().NoError(err)
		cloudAccountMetadata := metadata.CloudAccountMetadata

		s.Equal("prjName", cloudAccountMetadata.AccountName)
		s.Equal("prjId", cloudAccountMetadata.AccountId)
		s.Equal("orgId", cloudAccountMetadata.OrganisationId)
		s.Equal("orgName", cloudAccountMetadata.OrganizationName)
		if metadata.Type == fetching.CloudIdentity {
			m, err := r.GetElasticCommonData()
			s.Require().NoError(err, "error getting Elastic Common Data")
			s.Len(m, 2)
		}
	})
}

func (s *GcpAssetsFetcherTestSuite) TestFetcher_ElasticCommonData() {
	cases := []struct {
		resourceData map[string]any
		expectedECS  map[string]any
	}{
		{
			resourceData: map[string]any{},
			expectedECS:  map[string]any{},
		},
		{
			resourceData: map[string]any{"name": ""},
			expectedECS:  map[string]any{},
		},
		{
			resourceData: map[string]any{"name": "henrys-vm"},
			expectedECS:  map[string]any{"host.name": "henrys-vm"},
		},
		{
			resourceData: map[string]any{"hostname": ""},
			expectedECS:  map[string]any{},
		},
		{
			resourceData: map[string]any{"hostname": "henrys-vm"},
			expectedECS:  map[string]any{"host.hostname": "henrys-vm"},
		},
		{
			resourceData: map[string]any{"name": "x", "hostname": "y"},
			expectedECS:  map[string]any{"host.name": "x", "host.hostname": "y"},
		},
	}

	for _, tc := range cases {
		dataStruct, err := structpb.NewStruct(tc.resourceData)
		s.Require().NoError(err)

		asset := &GcpAsset{
			Type:    fetching.CloudCompute,
			SubType: "gcp-compute-instance",
			ExtendedAsset: &inventory.ExtendedGcpAsset{
				Asset: &assetpb.Asset{
					AssetType: inventory.ComputeInstanceAssetType,
					Resource: &assetpb.Resource{
						Data: dataStruct,
					},
				},
			},
		}

		ecs, err := asset.GetElasticCommonData()
		s.Require().NoError(err)
		s.Equal(tc.expectedECS, ecs)
	}
}
