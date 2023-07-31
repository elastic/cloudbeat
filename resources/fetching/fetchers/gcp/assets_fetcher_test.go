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
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/elastic/cloudbeat/resources/fetching"
	"github.com/elastic/cloudbeat/resources/providers/gcplib/inventory"
	"github.com/elastic/cloudbeat/resources/utils/testhelper"
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
	mockInventoryService := &inventory.MockInventoryService{}
	fetcher := GcpAssetsFetcher{
		log:        testhelper.NewLogger(s.T()),
		resourceCh: s.resourceCh,
		provider:   mockInventoryService,
	}

	mockInventoryService.On("ListAllAssetTypesByName", mock.MatchedBy(func(assets []string) bool {
		return true
	})).Return(
		[]*assetpb.Asset{
			{Name: "a", AssetType: "iam.googleapis.com/ServiceAccount"},
		}, nil,
	)

	err := fetcher.Fetch(ctx, fetching.CycleMetadata{})
	s.NoError(err)
	results := testhelper.CollectResources(s.resourceCh)

	// ListAllAssetTypesByName mocked to return a single asset
	// Will be called N times, where N is the number of types in GcpAssetTypes
	s.Equal(len(GcpAssetTypes), len(results))
}
