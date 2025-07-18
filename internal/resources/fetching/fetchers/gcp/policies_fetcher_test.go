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

type GcpPoliciesFetcherTestSuite struct {
	suite.Suite

	resourceCh chan fetching.ResourceInfo
}

func TestGcpPoliciesFetcherTestSuite(t *testing.T) {
	s := new(GcpPoliciesFetcherTestSuite)

	suite.Run(t, s)
}

func (s *GcpPoliciesFetcherTestSuite) SetupTest() {
	s.resourceCh = make(chan fetching.ResourceInfo, 50)
}

func (s *GcpPoliciesFetcherTestSuite) TearDownTest() {
	close(s.resourceCh)
}

func (s *GcpPoliciesFetcherTestSuite) TestPoliciesFetcher_Fetch_Success() {
	t := s.T()
	ctx := t.Context()
	mockInventoryService := &inventory.MockServiceAPI{}
	fetcher := NewGcpPoliciesFetcher(ctx, testhelper.NewLogger(s.T()), s.resourceCh, mockInventoryService)
	wg := sync.WaitGroup{}
	expectedAsset := &inventory.ProjectPoliciesAsset{
		Policies: []*inventory.ExtendedGcpAsset{
			{Asset: &assetpb.Asset{Name: "a1", AssetType: inventory.CrmProjectAssetType}},
		},
		CloudAccount: &fetching.CloudAccountMetadata{
			AccountId: "1",
		},
	}

	mockInventoryService.EXPECT().Clear()
	mockInventoryService.On("ListProjectsAncestorsPolicies", mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			ch, ok := args.Get(1).(chan<- *inventory.ProjectPoliciesAsset)
			require.True(t, ok, "expected chan<- *inventory.ProjectPoliciesAsset")
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
	asset, ok := res.Resource.(*GcpPoliciesAsset)
	s.True(ok)
	s.Len(asset.Asset.Policies, 1)
	s.Equal(expectedAsset.Policies[0], asset.Asset.Policies[0])
	s.Equal(expectedAsset.CloudAccount.AccountId, asset.Asset.CloudAccount.AccountId)
	s.Equal(fetching.ProjectManagement, asset.Type)
	s.Equal(fetching.GcpPolicies, asset.subType)
	wg.Wait()
	mockInventoryService.AssertExpectations(s.T())
}

func TestGcpPoliciesAsset_GetMetadata(t *testing.T) {
	const projectId = "1"
	tests := []struct {
		name     string
		resource GcpPoliciesAsset
		want     fetching.ResourceMetadata
		wantErr  bool
	}{
		{
			name: "retrieve successfully project  policies asset",
			resource: GcpPoliciesAsset{
				Type:    fetching.ProjectManagement,
				subType: fetching.GcpPolicies,
				Asset: &inventory.ProjectPoliciesAsset{
					CloudAccount: &fetching.CloudAccountMetadata{
						AccountId:        projectId,
						AccountName:      "a",
						OrganisationId:   "a",
						OrganizationName: "a",
					},
					Policies: []*inventory.ExtendedGcpAsset{},
				},
			},
			want: fetching.ResourceMetadata{
				ID:      fmt.Sprintf("%s-%s", fetching.GcpPolicies, projectId),
				Name:    fmt.Sprintf("%s-%s", fetching.GcpPolicies, projectId),
				Type:    fetching.ProjectManagement,
				SubType: fetching.GcpPolicies,
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
