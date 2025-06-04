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
	"sync"
	"testing"

	"cloud.google.com/go/asset/apiv1/assetpb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/elastic/cloudbeat/internal/resources/fetching"
	"github.com/elastic/cloudbeat/internal/resources/fetching/cycle"
	"github.com/elastic/cloudbeat/internal/resources/providers/gcplib"
	"github.com/elastic/cloudbeat/internal/resources/providers/gcplib/inventory"
	"github.com/elastic/cloudbeat/internal/resources/utils/testhelper"
)

type GcpNetworksFetcherTestSuite struct {
	suite.Suite

	resourceCh chan fetching.ResourceInfo
}

func TestGcpNetworksFetcherTestSuite(t *testing.T) {
	s := new(GcpNetworksFetcherTestSuite)

	suite.Run(t, s)
}

func (s *GcpNetworksFetcherTestSuite) SetupTest() {
	s.resourceCh = make(chan fetching.ResourceInfo, 50)
}

func (s *GcpNetworksFetcherTestSuite) TearDownTest() {
	close(s.resourceCh)
}

func (s *GcpNetworksFetcherTestSuite) TestNetworksFetcher_Fetch_Success() {
	ctx := context.Background()
	wg := sync.WaitGroup{}
	mockInventoryService := &inventory.MockServiceAPI{}
	fetcher := NewGcpNetworksFetcher(ctx, testhelper.NewLogger(s.T()), s.resourceCh, mockInventoryService)
	expectedAsset := &inventory.ExtendedGcpAsset{
		Asset: &assetpb.Asset{Name: "a1", AssetType: inventory.MonitoringAlertPolicyAssetType},
		CloudAccount: &fetching.CloudAccountMetadata{
			AccountId: "1",
		},
	}
	mockInventoryService.EXPECT().Clear()
	mockInventoryService.On("ListNetworkAssets", mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			ch := args.Get(1).(chan<- *inventory.ExtendedGcpAsset)
			ch <- expectedAsset
			close(ch)
		}).Once()

	wg.Add(1)
	go func() {
		defer wg.Done()
		err := fetcher.Fetch(ctx, cycle.Metadata{})
		s.NoError(err)
	}()

	res := <-s.resourceCh
	s.NotNil(res.Resource)
	asset, ok := res.Resource.(*GcpNetworksAsset)
	s.True(ok)
	s.Equal(expectedAsset, asset.NetworkAsset)
	s.Equal(fetching.CloudCompute, asset.Type)
	s.Equal("gcp-compute-network", asset.subType)

	wg.Wait()
	mockInventoryService.AssertExpectations(s.T())
}

func TestNetworksResource_GetMetadata(t *testing.T) {
	tests := []struct {
		name     string
		resource GcpNetworksAsset
		want     fetching.ResourceMetadata
		wantErr  bool
	}{
		{
			name: "happy path",
			resource: GcpNetworksAsset{
				Type:    fetching.CloudCompute,
				subType: "gcp-compute-network",
				NetworkAsset: &inventory.ExtendedGcpAsset{
					Asset: &assetpb.Asset{
						Name: fmt.Sprintf("%s/net1", projectId),
					},
					CloudAccount: &fetching.CloudAccountMetadata{
						AccountId:        projectId,
						AccountName:      "a",
						OrganisationId:   "a",
						OrganizationName: "a",
					},
				},
			},
			want: fetching.ResourceMetadata{
				ID:      fmt.Sprintf("%s/net1", projectId),
				Name:    "net1",
				Type:    fetching.CloudCompute,
				SubType: "gcp-compute-network",
				Region:  gcplib.GlobalRegion,
				CloudAccountMetadata: fetching.CloudAccountMetadata{
					AccountId:        projectId,
					AccountName:      "a",
					OrganisationId:   "a",
					OrganizationName: "a",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.resource.GetMetadata()

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
