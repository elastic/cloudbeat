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
	"fmt"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resourcegraph/armresourcegraph"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/elastic/cloudbeat/resources/providers/azurelib/auth"
	"github.com/elastic/cloudbeat/resources/utils/testhelper"
)

type ProviderTestSuite struct {
	suite.Suite
	mockedClient *AzureClientWrapper
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
			"sku":            "3",
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
			"sku":            "1",
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
			"sku":            "2",
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
	s.mockedClient = &AzureClientWrapper{
		AssetQuery: func(ctx context.Context, query armresourcegraph.QueryRequest, options *armresourcegraph.ClientResourcesOptions) (armresourcegraph.ClientResourcesResponse, error) {
			if query.Options.SkipToken != nil && *query.Options.SkipToken != "" {
				return armresourcegraph.ClientResourcesResponse{
					QueryResponse: nonTruncatedResponse,
				}, nil
			} else {
				return armresourcegraph.ClientResourcesResponse{
					QueryResponse: truncatedResponse,
				}, nil
			}
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
		s.Equal(tt.want, getString(tt.data, tt.key), "getString(%v, %s) = %s", tt.data, tt.key, tt.want)
	}
}

func (s *ProviderTestSuite) TestListAllAssetTypesByName() {
	provider := &Provider{
		log:    testhelper.NewLogger(s.T()),
		client: s.mockedClient,
		Config: auth.AzureFactoryConfig{
			Credentials: &azidentity.DefaultAzureCredential{},
		},
	}

	values, err := provider.ListAllAssetTypesByName(context.Background(), []string{"test"})
	s.Require().NoError(err)
	s.Len(values, int(*nonTruncatedResponse.Count+*truncatedResponse.Count))
	lo.ForEach(values, func(r AzureAsset, index int) {
		strIndex := fmt.Sprintf("%d", index+1)
		s.Equal(AzureAsset{
			Id:               strIndex,
			Name:             strIndex,
			Location:         strIndex,
			Properties:       map[string]any{"test": "test"},
			ResourceGroup:    strIndex,
			SubscriptionId:   strIndex,
			SubscriptionName: "",
			TenantId:         strIndex,
			Type:             strIndex,
			Sku:              strIndex,
		}, r)
	})
}

func Test_generateQuery(t *testing.T) {
	tests := []struct {
		assets []string
		want   string
	}{
		{
			want: "Resources",
		},
		{
			assets: []string{"one"},
			want:   "Resources | where type == 'one'",
		},
		{
			assets: []string{"one", "two", "three four five"},
			want:   "Resources | where type == 'one' or type == 'two' or type == 'three four five'",
		},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			assert.Equal(t, tt.want, generateQuery(tt.assets))
		})
	}
}
