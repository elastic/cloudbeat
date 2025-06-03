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

type GcpMonitoringFetcherTestSuite struct {
	suite.Suite

	resourceCh chan fetching.ResourceInfo
}

func TestGcpMonitoringFetcherTestSuite(t *testing.T) {
	s := new(GcpMonitoringFetcherTestSuite)

	suite.Run(t, s)
}

func (s *GcpMonitoringFetcherTestSuite) SetupTest() {
	s.resourceCh = make(chan fetching.ResourceInfo, 50)
}

func (s *GcpMonitoringFetcherTestSuite) TearDownTest() {
	close(s.resourceCh)
}

func (s *GcpMonitoringFetcherTestSuite) TestMonitoringFetcher_Fetch_Success() {
	t := s.T()
	ctx := t.Context()
	wg := sync.WaitGroup{}
	mockInventoryService := &inventory.MockServiceAPI{}
	fetcher := NewGcpMonitoringFetcher(ctx, testhelper.NewLogger(s.T()), s.resourceCh, mockInventoryService)
	expectedAsset := &inventory.MonitoringAsset{
		Alerts: []*inventory.ExtendedGcpAsset{
			{Asset: &assetpb.Asset{Name: "a1", AssetType: inventory.MonitoringAlertPolicyAssetType}},
		},
		LogMetrics: []*inventory.ExtendedGcpAsset{
			{Asset: &assetpb.Asset{Name: "a1", AssetType: inventory.MonitoringLogMetricAssetType}},
		},
		CloudAccount: &fetching.CloudAccountMetadata{
			AccountId: "1",
		},
	}

	mockInventoryService.EXPECT().Clear()
	mockInventoryService.On("ListMonitoringAssets", mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			ch := args.Get(1).(chan<- *inventory.MonitoringAsset)
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
	asset, ok := res.Resource.(*GcpMonitoringAsset)
	s.True(ok)
	s.Len(asset.Asset.Alerts, 1)
	s.Len(asset.Asset.LogMetrics, 1)
	s.Equal(expectedAsset.Alerts, asset.Asset.Alerts)
	s.Equal(expectedAsset.LogMetrics, asset.Asset.LogMetrics)
	s.Equal(expectedAsset.CloudAccount.AccountId, asset.Asset.CloudAccount.AccountId)
	s.Equal(fetching.MonitoringIdentity, asset.Type)
	s.Equal(fetching.GcpMonitoringType, asset.subType)
	wg.Wait()
	mockInventoryService.AssertExpectations(s.T())
}

func TestMonitoringResource_GetMetadata(t *testing.T) {
	tests := []struct {
		name     string
		resource GcpMonitoringAsset
		want     fetching.ResourceMetadata
		wantErr  bool
	}{
		{
			name: "happy path",
			resource: GcpMonitoringAsset{
				Type:    fetching.MonitoringIdentity,
				subType: fetching.GcpMonitoringType,
				Asset: &inventory.MonitoringAsset{
					CloudAccount: &fetching.CloudAccountMetadata{
						AccountId:        projectId,
						AccountName:      "a",
						OrganisationId:   "a",
						OrganizationName: "a",
					},
					LogMetrics: []*inventory.ExtendedGcpAsset{},
					Alerts:     []*inventory.ExtendedGcpAsset{},
				},
			},
			want: fetching.ResourceMetadata{
				ID:      fmt.Sprintf("%s-%s", fetching.GcpMonitoringType, projectId),
				Name:    fmt.Sprintf("%s-%s", fetching.GcpMonitoringType, projectId),
				Type:    fetching.MonitoringIdentity,
				SubType: fetching.GcpMonitoringType,
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
