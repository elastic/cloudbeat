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

package inventory

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resourcegraph/armresourcegraph"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/elastic/cloudbeat/internal/resources/utils/pointers"
	"github.com/elastic/cloudbeat/internal/resources/utils/strings"
	"github.com/elastic/cloudbeat/internal/resources/utils/testhelper"
)

type ProviderTestSuite struct {
	suite.Suite
	mockedClient *ResourceGraphAzureClientWrapper
}

var nonTruncatedResponse = armresourcegraph.QueryResponse{
	Count: to.Ptr(int64(1)),
	Data: []any{
		map[string]any{
			"id":             "3",
			"name":           "3",
			"location":       "3",
			"properties":     map[string]any{"test": "test"},
			"resourceGroup":  "3",
			"subscriptionId": "3",
			"tenantId":       "3",
			"type":           "3",
			"sku":            map[string]any{"test": "test"},
			"identity":       map[string]any{"test": "test"},
		},
	},
	ResultTruncated: to.Ptr(armresourcegraph.ResultTruncatedFalse),
}

var truncatedResponse = armresourcegraph.QueryResponse{
	Count: to.Ptr(int64(2)),
	Data: []any{
		map[string]any{
			"id":             "1",
			"name":           "1",
			"location":       "1",
			"properties":     map[string]any{"test": "test"},
			"resourceGroup":  "1",
			"subscriptionId": "1",
			"tenantId":       "1",
			"type":           "1",
			"sku":            map[string]any{"test": "test"},
			"identity":       map[string]any{"test": "test"},
		},
		map[string]any{
			"id":             "2",
			"name":           "2",
			"location":       "2",
			"properties":     map[string]any{"test": "test"},
			"resourceGroup":  "2",
			"subscriptionId": "2",
			"tenantId":       "2",
			"type":           "2",
			"sku":            map[string]any{"test": "test"},
			"identity":       map[string]any{"test": "test"},
		},
	},
	ResultTruncated: to.Ptr(armresourcegraph.ResultTruncatedTrue),
	SkipToken:       to.Ptr("token"),
}

func TestInventoryProviderTestSuite(t *testing.T) {
	s := new(ProviderTestSuite)

	suite.Run(t, s)
}

func (s *ProviderTestSuite) SetupTest() {
	s.mockedClient = &ResourceGraphAzureClientWrapper{
		AssetQuery: func(_ context.Context, query armresourcegraph.QueryRequest, _ *armresourcegraph.ClientResourcesOptions) (armresourcegraph.ClientResourcesResponse, error) {
			if query.Options.SkipToken != nil && *query.Options.SkipToken != "" {
				return armresourcegraph.ClientResourcesResponse{
					QueryResponse: nonTruncatedResponse,
				}, nil
			}
			return armresourcegraph.ClientResourcesResponse{
				QueryResponse: truncatedResponse,
			}, nil
		},
	}
}

func (s *ProviderTestSuite) TestGetString() {
	tests := []struct {
		name string
		data map[string]any
		key  string
		want string
	}{
		{
			name: "nil map",
			data: nil,
			key:  "key",
			want: "",
		},
		{
			name: "key does not exist",
			data: map[string]any{"key": "value"},
			key:  "other-key",
			want: "",
		},
		{
			name: "wrong type",
			data: map[string]any{"key": 1},
			key:  "key",
			want: "",
		},
		{
			name: "correct value",
			data: map[string]any{"key": "value", "other-key": 1},
			key:  "key",
			want: "value",
		},
	}
	for _, tt := range tests {
		s.Equal(tt.want, strings.FromMap(tt.data, tt.key), "getString(%v, %s) = %s", tt.data, tt.key, tt.want)
	}
}

func (s *ProviderTestSuite) TestListAllAssetTypesByName() {
	provider := &ResourceGraphProvider{
		log:    testhelper.NewLogger(s.T()),
		client: s.mockedClient,
	}

	t := s.T()
	values, err := provider.ListAllAssetTypesByName(t.Context(), "test", []string{"test"})
	s.Require().NoError(err)
	s.Len(values, int(*nonTruncatedResponse.Count+*truncatedResponse.Count))
	lo.ForEach(values, func(r AzureAsset, index int) {
		strIndex := fmt.Sprintf("%d", index+1)
		s.Equal(AzureAsset{
			Id:             strIndex,
			Name:           strIndex,
			Location:       strIndex,
			Properties:     map[string]any{"test": "test"},
			ResourceGroup:  strIndex,
			SubscriptionId: strIndex,
			TenantId:       strIndex,
			Type:           strIndex,
			Sku:            map[string]any{"test": "test"},
			Identity:       map[string]any{"test": "test"},
		}, r)
	})
}

// TestListAllAssetTypesByName_ContinuesAcrossAllPagesRegardlessOfResultTruncated is a
// regression test for the pagination bug where most Azure subscriptions
// returned only subscription-level findings with zero resource-level findings.
//
// Real Azure Resource Graph responses set ResultTruncated=false on every ordinary
// paginated page - ResultTruncated only communicates that paging itself is unavailable
// (e.g. because the query uses a `limit`/`take` operator), not whether more pages exist.
// The authoritative "more data available" signal, per Microsoft's own reference
// implementations (which loop "while SkipToken is not nil"), is solely the presence of
// SkipToken. Before the fix, ListAllAssetTypesByName stopped fetching as soon as it saw
// ResultTruncated == false, which meant it always stopped after the very first page and
// silently dropped every subsequent page of resources - even though SkipToken indicated
// more pages were available. This test fails against the pre-fix implementation (it
// returns only the first page's single asset) and passes once pagination is driven
// exclusively by SkipToken.
func (s *ProviderTestSuite) TestListAllAssetTypesByName_ContinuesAcrossAllPagesRegardlessOfResultTruncated() {
	pages := []armresourcegraph.QueryResponse{
		{
			Count:           to.Ptr(int64(1)),
			Data:            []any{rawAsset("1")},
			ResultTruncated: to.Ptr(armresourcegraph.ResultTruncatedFalse),
			SkipToken:       to.Ptr("page-2-token"),
		},
		{
			Count:           to.Ptr(int64(1)),
			Data:            []any{rawAsset("2")},
			ResultTruncated: to.Ptr(armresourcegraph.ResultTruncatedFalse),
			SkipToken:       to.Ptr("page-3-token"),
		},
		{
			Count:           to.Ptr(int64(1)),
			Data:            []any{rawAsset("3")},
			ResultTruncated: to.Ptr(armresourcegraph.ResultTruncatedFalse),
			SkipToken:       nil,
		},
	}
	expectedIncomingSkipTokens := []string{"", "page-2-token", "page-3-token"}

	callCount := 0
	client := &ResourceGraphAzureClientWrapper{
		AssetQuery: func(_ context.Context, query armresourcegraph.QueryRequest, _ *armresourcegraph.ClientResourcesOptions) (armresourcegraph.ClientResourcesResponse, error) {
			s.Require().Lessf(callCount, len(pages), "provider requested a %dth page, but the mock server only has %d pages", callCount+1, len(pages))
			s.Equal(expectedIncomingSkipTokens[callCount], pointers.Deref(query.Options.SkipToken), "unexpected SkipToken sent for call #%d", callCount)

			response := armresourcegraph.ClientResourcesResponse{QueryResponse: pages[callCount]}
			callCount++
			return response, nil
		},
	}

	provider := &ResourceGraphProvider{
		log:    testhelper.NewLogger(s.T()),
		client: client,
	}

	values, err := provider.ListAllAssetTypesByName(s.T().Context(), "test", []string{"test"})
	s.Require().NoError(err)
	s.Equal(len(pages), callCount, "expected the provider to fetch every page")
	s.Len(values, len(pages), "expected assets from all pages to be aggregated, not just the first page")

	ids := lo.Map(values, func(a AzureAsset, _ int) string { return a.Id })
	s.ElementsMatch([]string{"1", "2", "3"}, ids)
}

func rawAsset(id string) map[string]any {
	return map[string]any{
		"id":             id,
		"name":           id,
		"location":       id,
		"properties":     map[string]any{"test": "test"},
		"resourceGroup":  id,
		"subscriptionId": id,
		"tenantId":       id,
		"type":           id,
		"sku":            map[string]any{"test": "test"},
		"identity":       map[string]any{"test": "test"},
	}
}

func Test_generateQuery(t *testing.T) {
	tests := []struct {
		assetsGroup string
		assets      []string
		want        string
	}{
		{
			assetsGroup: "empty assets",
			want:        "empty assets",
		},
		{
			assetsGroup: "resources",
			assets:      []string{"one"},
			want:        "resources | where type == 'one'",
		},
		{
			assetsGroup: "resources",
			assets:      []string{"one", "two", "three four five"},
			want:        "resources | where type == 'one' or type == 'two' or type == 'three four five'",
		},
		{
			assetsGroup: "authorizationresources",
			assets:      []string{"one"},
			want:        "authorizationresources | where type == 'one'",
		},
		{
			assetsGroup: "authorizationresources",
			assets:      []string{"one", "two", "three four five"},
			want:        "authorizationresources | where type == 'one' or type == 'two' or type == 'three four five'",
		},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			assert.Equal(t, tt.want, generateQuery(tt.assetsGroup, tt.assets))
		})
	}
}

func TestReadPager(t *testing.T) {
	tests := map[string]struct {
		moreFn    func(i int) bool
		fetchFn   func(context.Context, *int) (int, error)
		expectErr bool
		expected  []int
	}{
		"happy path, four pages": {
			moreFn: func(i int) bool {
				return i < 4
			},
			fetchFn: func(_ context.Context, i *int) (int, error) {
				if i == nil {
					i = to.Ptr(0)
				}
				*i++
				return *i, nil
			},
			expectErr: false,
			expected:  []int{1, 2, 3, 4},
		},

		"error at third of four pages": {
			moreFn: func(i int) bool {
				return i < 4
			},
			fetchFn: func(_ context.Context, i *int) (int, error) {
				if i == nil {
					i = to.Ptr(0)
				}
				*i++
				if *i == 3 {
					return *i, errors.New("mock error")
				}
				return *i, nil
			},
			expectErr: true,
			expected:  nil,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			pagerHandlerMock := runtime.PagingHandler[int]{
				More:    tc.moreFn,
				Fetcher: tc.fetchFn,
			}
			pager := runtime.NewPager[int](pagerHandlerMock)
			intSlice, err := readPager[int](t.Context(), pager)

			if tc.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			require.Equal(t, tc.expected, intSlice)
		})
	}
}
