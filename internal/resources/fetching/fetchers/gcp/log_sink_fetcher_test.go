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
	"errors"
	"fmt"
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

type GcpLogSinkFetcherTestSuite struct {
	suite.Suite

	resourceCh chan fetching.ResourceInfo
}

func TestGcpLogSinkFetcherTestSuite(t *testing.T) {
	s := new(GcpLogSinkFetcherTestSuite)

	suite.Run(t, s)
}

func (s *GcpLogSinkFetcherTestSuite) SetupTest() {
	s.resourceCh = make(chan fetching.ResourceInfo, 50)
}

func (s *GcpLogSinkFetcherTestSuite) TearDownTest() {
	close(s.resourceCh)
}

func (s *GcpLogSinkFetcherTestSuite) TestFetcher_Fetch_Success() {
	ctx := context.Background()
	mockInventoryService := &inventory.MockServiceAPI{}
	fetcher := GcpLogSinkFetcher{
		log:        testhelper.NewLogger(s.T()),
		resourceCh: s.resourceCh,
		provider:   mockInventoryService,
	}

	mockInventoryService.On("ListLoggingAssets", mock.Anything).Return(
		[]*inventory.LoggingAsset{
			{
				Ecs: &fetching.EcsGcp{
					Provider:         "gcp",
					ProjectId:        "a",
					ProjectName:      "a",
					OrganizationId:   "a",
					OrganizationName: "a",
				},
				LogSinks: []*inventory.ExtendedGcpAsset{
					{Asset: &assetpb.Asset{Name: "a", AssetType: inventory.LogSinkAssetType}},
				},
			},
		}, nil,
	)

	err := fetcher.Fetch(ctx, cycle.Metadata{})
	s.Require().NoError(err)
	results := testhelper.CollectResources(s.resourceCh)

	// ListMonitoringAssets mocked to return a single asset
	s.Len(results, 1)
}

func (s *GcpLogSinkFetcherTestSuite) TestFetcher_Fetch_Error() {
	ctx := context.Background()
	mockInventoryService := &inventory.MockServiceAPI{}
	fetcher := GcpLogSinkFetcher{
		log:        testhelper.NewLogger(s.T()),
		resourceCh: s.resourceCh,
		provider:   mockInventoryService,
	}

	mockInventoryService.On("ListLoggingAssets", mock.Anything).Return(nil, errors.New("api call error"))

	err := fetcher.Fetch(ctx, cycle.Metadata{})
	s.Require().Error(err)
}

func TestLoggingAsset_GetMetadata(t *testing.T) {
	tests := []struct {
		name     string
		resource GcpLoggingAsset
		want     fetching.ResourceMetadata
		wantErr  bool
	}{
		{
			name: "retrieve successfully service usage assets",
			resource: GcpLoggingAsset{
				Type:    fetching.LoggingIdentity,
				subType: fetching.GcpLoggingType,
				Asset: &inventory.LoggingAsset{
					Ecs: &fetching.EcsGcp{
						ProjectId:        projectId,
						ProjectName:      "a",
						OrganizationId:   "a",
						OrganizationName: "a",
					},
					LogSinks: []*inventory.ExtendedGcpAsset{},
				},
			},
			want: fetching.ResourceMetadata{
				ID:      fmt.Sprintf("%s-%s", fetching.GcpLoggingType, projectId),
				Name:    fmt.Sprintf("%s-%s", fetching.GcpLoggingType, projectId),
				Type:    fetching.LoggingIdentity,
				SubType: fetching.GcpLoggingType,
				Region:  gcplib.GlobalRegion,
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

func TestGcpLoggingAsset_GetElasticCommonData(t *testing.T) {
	type fields struct {
		Type    string
		subType string
		Asset   *inventory.LoggingAsset
	}
	tests := []struct {
		name   string
		fields fields
		want   map[string]any
	}{
		{
			name: "verify elastic common data",
			fields: fields{
				Type:    fetching.LoggingIdentity,
				subType: fetching.GcpLoggingType,
				Asset: &inventory.LoggingAsset{
					Ecs: &fetching.EcsGcp{
						ProjectId:        projectId,
						ProjectName:      "a",
						OrganizationId:   "a",
						OrganizationName: "a",
					},
					LogSinks: []*inventory.ExtendedGcpAsset{
						{Asset: &assetpb.Asset{Name: "a", AssetType: inventory.LogSinkAssetType}},
					},
				},
			},
			want: map[string]any{
				"cloud": map[string]any{
					"provider": "gcp",
					"account": map[string]any{
						"id":   projectId,
						"name": "a",
					},
					"Organization": map[string]any{
						"id":   "a",
						"name": "a",
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &GcpLoggingAsset{
				Type:    tt.fields.Type,
				subType: tt.fields.subType,
				Asset:   tt.fields.Asset,
			}

			got, err := g.GetElasticCommonData()

			require.NoError(t, err)
			assert.Equalf(t, tt.want, got, "GetElasticCommonData()")
		})
	}
}
